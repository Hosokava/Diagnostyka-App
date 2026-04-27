package handlers_test

import (
	"fmt"
	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/utils"
	"testing"
)

func TestDoctorProfile_PartialUpdate(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	d := createDoctor(t, "d@test.com", "pass1234")
	c := getAuthCookies(t, d.ID, "doctor")

	w := doRequest(r, "PATCH", "/api/doctor/profile",
		jsonBody(map[string]interface{}{"first_name": "Gregory"}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	res := parseJSON(w)
	profile := res["profile"].(map[string]interface{})
	if profile["first_name"] != "Gregory" {
		t.Fatalf("expected Gregory, got %v", profile["first_name"])
	}
}

func TestDoctorProfile_ReturnsUpdatedData(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	d := createDoctor(t, "d@test.com", "pass1234")
	c := getAuthCookies(t, d.ID, "doctor")

	w := doRequest(r, "PATCH", "/api/doctor/profile",
		jsonBody(map[string]interface{}{
			"first_name": "House", "last_name": "MD",
			"specialization": "Diagnostics", "examination_ids": []int{1, 2},
		}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	res := parseJSON(w)
	if res["profile"] == nil {
		t.Fatal("expected profile in response")
	}
}

func TestCompleteAppointment_Success(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1, QRCodeHash: "hash1"}
	database.DB.Create(&appt)

	c := getAuthCookies(t, d.ID, "doctor")
	w := doRequest(r, "POST", fmt.Sprintf("/api/doctor/appointments/%d/complete", appt.ID),
		jsonBody(map[string]interface{}{"result": "All clear", "notes": "Follow up in 6mo"}), c)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Appointment
	database.DB.First(&updated, appt.ID)
	if !updated.IsFinished {
		t.Fatal("expected appointment to be finished")
	}
	if updated.CompletionDate == nil {
		t.Fatal("expected completion date to be set")
	}
}

func TestCompleteAppointment_AlreadyDone(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1, QRCodeHash: "hash2", IsFinished: true}
	database.DB.Create(&appt)

	c := getAuthCookies(t, d.ID, "doctor")
	w := doRequest(r, "POST", fmt.Sprintf("/api/doctor/appointments/%d/complete", appt.ID),
		jsonBody(map[string]interface{}{"result": "test"}), c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for already completed, got %d", w.Code)
	}
}

func TestCompleteAppointment_WrongDoctor(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d1 := createCompleteDoctor(t, "d1@test.com", "pass1234", "General", []uint{1})
	d2 := createCompleteDoctor(t, "d2@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{PatientID: p.ID, DoctorID: d1.ID, ExaminationID: 1, QRCodeHash: "hash3"}
	database.DB.Create(&appt)

	c := getAuthCookies(t, d2.ID, "doctor")
	w := doRequest(r, "POST", fmt.Sprintf("/api/doctor/appointments/%d/complete", appt.ID),
		jsonBody(map[string]interface{}{"result": "test"}), c)
	if w.Code != 403 {
		t.Fatalf("expected 403 for wrong doctor, got %d", w.Code)
	}
}

func TestCompleteAppointment_InvalidID(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})
	c := getAuthCookies(t, d.ID, "doctor")

	w := doRequest(r, "POST", "/api/doctor/appointments/abc/complete",
		jsonBody(map[string]interface{}{"result": "test"}), c)
	if w.Code != 400 {
		t.Fatalf("expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestPublicResults_ValidHash(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{
		PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1,
		QRCodeHash: "publichash123", IsFinished: true, Result: "Normal",
	}
	database.DB.Create(&appt)

	w := doRequest(r, "GET", "/api/results/publichash123", nil, nil)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPublicResults_InvalidHash(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := doRequest(r, "GET", "/api/results/nonexistent", nil, nil)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPublicResults_UnfinishedAppointment(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()
	seedExaminations(t)
	enc, _ := utils.EncryptAES("06301268369")
	p := createCompletePatient(t, "p@test.com", "pass1234", enc)
	d := createCompleteDoctor(t, "d@test.com", "pass1234", "General", []uint{1})

	appt := models.Appointment{
		PatientID: p.ID, DoctorID: d.ID, ExaminationID: 1,
		QRCodeHash: "unfinished", IsFinished: false,
	}
	database.DB.Create(&appt)

	w := doRequest(r, "GET", "/api/results/unfinished", nil, nil)
	if w.Code != 403 {
		t.Fatalf("expected 403 for unfinished, got %d", w.Code)
	}
}
