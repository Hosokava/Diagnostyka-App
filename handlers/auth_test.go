package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// --- REGISTER ---

func TestRegister_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "test@example.com", "password": "password123", "role": "patient"}), nil)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	body := jsonBody(map[string]interface{}{"email": "dup@example.com", "password": "password123", "role": "patient"})
	doRequest(r, "POST", "/auth/register", body, nil)

	body2 := jsonBody(map[string]interface{}{"email": "dup@example.com", "password": "password456", "role": "patient"})
	w := doRequest(r, "POST", "/auth/register", body2, nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for duplicate email, got %d", w.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "notanemail", "password": "password123", "role": "patient"}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for invalid email, got %d", w.Code)
	}
	res := parseJSON(w)
	if res["error"] != "invalid email format" {
		t.Fatalf("expected 'invalid email format', got %v", res["error"])
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "test@example.com", "password": "short", "role": "patient"}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for short password, got %d", w.Code)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "test@example.com", "password": "password123", "role": "admin"}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for invalid role, got %d", w.Code)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for missing fields, got %d", w.Code)
	}
}

func TestRegister_EmailCaseInsensitive(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "Test@Example.COM", "password": "password123", "role": "patient"}), nil)

	w := doRequest(r, "POST", "/auth/register",
		jsonBody(map[string]interface{}{"email": "test@example.com", "password": "password456", "role": "patient"}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400 for case-duplicate email, got %d", w.Code)
	}
}

// --- LOGIN ---

func TestLogin_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	createPatient(t, "login@example.com", "password123")

	w := doRequest(r, "POST", "/auth/login",
		jsonBody(map[string]interface{}{"email": "login@example.com", "password": "password123", "role": "patient"}), nil)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	cookies := w.Result().Cookies()
	hasAccess, hasRefresh := false, false
	for _, c := range cookies {
		if c.Name == "access_token" {
			hasAccess = true
		}
		if c.Name == "refresh_token" {
			hasRefresh = true
		}
	}
	if !hasAccess || !hasRefresh {
		t.Fatal("expected both access_token and refresh_token cookies")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	createPatient(t, "login@example.com", "password123")

	w := doRequest(r, "POST", "/auth/login",
		jsonBody(map[string]interface{}{"email": "login@example.com", "password": "wrongpassword", "role": "patient"}), nil)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLogin_WrongRole(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	createPatient(t, "login@example.com", "password123")

	w := doRequest(r, "POST", "/auth/login",
		jsonBody(map[string]interface{}{"email": "login@example.com", "password": "password123", "role": "doctor"}), nil)

	if w.Code != 401 {
		t.Fatalf("expected 401 for wrong role, got %d", w.Code)
	}
}

func TestLogin_InvalidRoleValue(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/login",
		jsonBody(map[string]interface{}{"email": "x@x.com", "password": "pass", "role": "admin"}), nil)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	res := parseJSON(w)
	if res["error"] != "role must be 'patient' or 'doctor'" {
		t.Fatalf("unexpected error msg: %v", res["error"])
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/login",
		jsonBody(map[string]interface{}{"email": "ghost@example.com", "password": "password123", "role": "patient"}), nil)

	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// --- UNAUTHENTICATED ACCESS ---

func TestUnauthenticatedAccess(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/me"},
		{"GET", "/api/examinations"},
		{"GET", "/api/patient/profile"},
		{"PATCH", "/api/patient/profile"},
		{"GET", "/api/doctor/profile"},
	}

	for _, ep := range endpoints {
		w := doRequest(r, ep.method, ep.path, nil, nil)
		if w.Code != 401 {
			t.Errorf("%s %s: expected 401, got %d", ep.method, ep.path, w.Code)
		}
	}
}

// --- ROLE ENFORCEMENT ---

func TestPatientCannotAccessDoctorEndpoints(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	patient := createPatient(t, "p@test.com", "password123")
	cookies := getAuthCookies(t, patient.ID, "patient")

	w := doRequest(r, "GET", "/api/doctor/profile", nil, cookies)
	if w.Code != 403 {
		t.Fatalf("expected 403 for patient on doctor endpoint, got %d", w.Code)
	}
}

func TestDoctorCannotAccessPatientEndpoints(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	doctor := createDoctor(t, "d@test.com", "password123")
	cookies := getAuthCookies(t, doctor.ID, "doctor")

	w := doRequest(r, "GET", "/api/patient/profile", nil, cookies)
	if w.Code != 403 {
		t.Fatalf("expected 403 for doctor on patient endpoint, got %d", w.Code)
	}
}

// --- LOGOUT ---

func TestLogout_ClearsCookies(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "POST", "/auth/logout", nil, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	for _, c := range w.Result().Cookies() {
		if (c.Name == "access_token" || c.Name == "refresh_token") && c.MaxAge >= 0 {
			t.Fatalf("cookie %s should have been cleared (MaxAge < 0)", c.Name)
		}
	}
}

// --- GET ME ---

func TestGetMe_IncompleteProfile(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	patient := createPatient(t, "me@test.com", "password123")
	cookies := getAuthCookies(t, patient.ID, "patient")

	w := doRequest(r, "GET", "/api/me", nil, cookies)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	res := parseJSON(w)
	if res["profile_complete"] != false {
		t.Fatalf("expected profile_complete=false, got %v", res["profile_complete"])
	}
}

func TestGetMe_CompleteProfile(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	patient := createCompletePatient(t, "complete@test.com", "password123", "encrypted_pesel")
	cookies := getAuthCookies(t, patient.ID, "patient")

	w := doRequest(r, "GET", "/api/me", nil, cookies)
	res := parseJSON(w)
	if res["profile_complete"] != true {
		t.Fatalf("expected profile_complete=true, got %v", res["profile_complete"])
	}
}

// --- EXAMINATIONS ---

func TestListExaminations_Empty(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	patient := createPatient(t, "exam@test.com", "password123")
	cookies := getAuthCookies(t, patient.ID, "patient")

	w := doRequest(r, "GET", "/api/examinations", nil, cookies)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "[]" {
		t.Fatalf("expected empty array, got %s", w.Body.String())
	}
}

func TestListExaminations_WithData(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	patient := createPatient(t, "exam@test.com", "password123")
	cookies := getAuthCookies(t, patient.ID, "patient")

	w := doRequest(r, "GET", "/api/examinations", nil, cookies)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 3 {
		t.Fatalf("expected 3 examinations, got %d", len(result))
	}
}
