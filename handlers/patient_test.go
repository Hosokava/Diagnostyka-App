package handlers_test

import (
	"encoding/json"
	"fmt"
	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/utils"
	"testing"
)

func TestPatientProfile_Update_Partial(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	p := createPatient(t, "p@test.com", "pass1234")
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "PATCH", "/api/patient/profile",
		jsonBody(map[string]interface{}{"first_name": "Jan"}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Patient
	database.DB.First(&updated, p.ID)
	if updated.FirstName != "Jan" {
		t.Fatalf("expected Jan, got %s", updated.FirstName)
	}
	if updated.LastName != "" {
		t.Fatalf("expected empty last_name, got %s", updated.LastName)
	}
}

func TestPatientProfile_Update_InvalidPESEL(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	p := createPatient(t, "p@test.com", "pass1234")
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "PATCH", "/api/patient/profile",
		jsonBody(map[string]interface{}{"pesel": "12345678900"}), c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for bad PESEL checksum, got %d", w.Code)
	}
}

func TestPatientProfile_Update_MaskedPESELIgnored(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "PATCH", "/api/patient/profile",
		jsonBody(map[string]interface{}{"pesel": "XXXXXXX8369"}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var updated models.Patient
	database.DB.First(&updated, p.ID)
	if updated.PESEL != enc {
		t.Fatal("masked PESEL should not overwrite encrypted value")
	}
}

func TestPatientProfile_Get_ReturnsMasked(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "GET", "/api/patient/profile", nil, c)
	res := parseJSON(w)
	pesel, _ := res["pesel"].(string)
	if pesel != "XXXXXXX8369" {
		t.Fatalf("expected masked PESEL, got %s", pesel)
	}
}

func TestRevealPESEL_CorrectPassword(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/profile/pesel",
		jsonBody(map[string]interface{}{"password": "pass1234"}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	res := parseJSON(w)
	if res["pesel"] != "06301268369" {
		t.Fatalf("expected full PESEL, got %v", res["pesel"])
	}
}

func TestRevealPESEL_WrongPassword(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/profile/pesel",
		jsonBody(map[string]interface{}{"password": "wrongpass"}), c)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProfileIncomplete_BlocksBooking(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	p := createPatient(t, "p@test.com", "pass1234")
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": 1}), c)
	if w.Code != 403 {
		t.Fatalf("expected 403 for incomplete profile, got %d", w.Code)
	}
}

func TestBookAppointment_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": 1}), c)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBookAppointment_DuplicateBlocked(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})
	c := getAuthCookies(t, p.ID, "patient")

	doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": 1}), c)

	w := doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": 1}), c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for duplicate booking, got %d", w.Code)
	}
}

func TestBookAppointment_NoDoctor(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": 1}), c)
	if w.Code != 503 {
		t.Fatalf("expected 503 when no doctor available, got %d", w.Code)
	}
}

func TestBookAppointment_StringID(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "POST", "/api/patient/book",
		jsonBody(map[string]interface{}{"examination_id": "abc"}), c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for string exam ID, got %d", w.Code)
	}
}

func TestCancelAppointment_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})
	c := getAuthCookies(t, p.ID, "patient")

	appt := models.Appointment{PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1, QRCodeHash: "testhash"}
	database.DB.Create(&appt)

	w := doRequest(r, "DELETE", fmt.Sprintf("/api/patient/appointments/%d", appt.ID), nil, c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCancelAppointment_InvalidID(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "DELETE", "/api/patient/appointments/abc", nil, c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for non-numeric ID, got %d", w.Code)
	}
}

func TestCancelAppointment_OtherPatient(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p1 := createCompletePatient(t, "p1@test.com", "pass1234", enc)
	p2 := createCompletePatient(t, "p2@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{PatientID: p1.ID, DoctorID: d.ID, ExaminationID: 1, QRCodeHash: "hash1"}
	database.DB.Create(&appt)

	c2 := getAuthCookies(t, p2.ID, "patient")
	w := doRequest(r, "DELETE", fmt.Sprintf("/api/patient/appointments/%d", appt.ID), nil, c2)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCancelAppointment_Completed(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1, QRCodeHash: "hash2", IsFinished: true}
	database.DB.Create(&appt)

	c := getAuthCookies(t, p.ID, "patient")
	w := doRequest(r, "DELETE", fmt.Sprintf("/api/patient/appointments/%d", appt.ID), nil, c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for completed appointment, got %d", w.Code)
	}
}

func TestActiveAppointments_Empty(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	c := getAuthCookies(t, p.ID, "patient")

	w := doRequest(r, "GET", "/api/patient/appointments/active", nil, c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d items", len(result))
	}
}
