package handlers_test

import (
	"bytes"
	"encoding/json"
	"gin-quickstart/auth"
	"gin-quickstart/database"
	"gin-quickstart/handlers"
	"gin-quickstart/middleware"
	"gin-quickstart/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("AES_KEY", "12345678901234567890123456789012")
	os.Setenv("EMAILS_ENABLED", "false")
}

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(
		&models.Patient{},
		&models.Doctor{},
		&models.Examination{},
		&models.Appointment{},
		&models.RefreshToken{},
	)
	database.DB = db
}

func setupRouter() *gin.Engine {
	r := gin.New()

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/refresh", handlers.Refresh)
		authGroup.POST("/logout", handlers.Logout)
	}

	r.GET("/api/results/:hash", handlers.GetPublicResults)
	r.GET("/api/results/:hash/qr", handlers.GetQRCode)

	api := r.Group("/api")
	api.Use(middleware.RequireAuth())
	{
		api.GET("/me", handlers.GetMe)
		api.GET("/examinations", handlers.ListExaminations)

		patient := api.Group("/patient")
		patient.Use(middleware.RequireRole("patient"))
		{
			patient.PATCH("/profile", handlers.UpdatePatientProfile)
			patient.GET("/profile", handlers.GetPatientProfile)

			restricted := patient.Group("/")
			restricted.Use(middleware.RequireProfileComplete())
			{
				restricted.GET("/appointments/active", handlers.GetActiveAppointments)
				restricted.GET("/appointments/history", handlers.GetAppointmentHistory)
				restricted.POST("/book", handlers.BookAppointment)
				restricted.DELETE("/appointments/:id", handlers.CancelAppointment)
				restricted.POST("/profile/pesel", handlers.RevealPESEL)
			}
		}

		doctor := api.Group("/doctor")
		doctor.Use(middleware.RequireRole("doctor"))
		{
			doctor.PATCH("/profile", handlers.UpdateDoctorProfile)
			doctor.GET("/profile", handlers.GetDoctorProfile)

			restricted := doctor.Group("/")
			restricted.Use(middleware.RequireProfileComplete())
			{
				restricted.GET("/schedule/active", handlers.GetActiveSchedule)
				restricted.GET("/schedule/history", handlers.GetDoctorHistory)
				restricted.POST("/appointments/:id/complete", handlers.CompleteAppointment)
			}
		}
	}

	return r
}

func jsonBody(data map[string]interface{}) *bytes.Reader {
	b, _ := json.Marshal(data)
	return bytes.NewReader(b)
}

func doRequest(r *gin.Engine, method, path string, body *bytes.Reader, cookies []*http.Cookie) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func parseJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

func createPatient(t *testing.T, email, password string) models.Patient {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	patient := models.Patient{Email: email, PasswordHash: string(hash)}
	database.DB.Create(&patient)
	return patient
}

func createCompletePatient(t *testing.T, email, password, pesel string) models.Patient {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	patient := models.Patient{
		Email: email, PasswordHash: string(hash),
		FirstName: "Test", LastName: "Patient", PESEL: pesel,
	}
	database.DB.Create(&patient)
	return patient
}

func createDoctor(t *testing.T, email, password string) models.Doctor {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	doctor := models.Doctor{Email: email, PasswordHash: string(hash)}
	database.DB.Create(&doctor)
	return doctor
}

func createCompleteDoctor(t *testing.T, email, password, spec string, examIDs []uint) models.Doctor {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	var exams []models.Examination
	database.DB.Find(&exams, examIDs)
	doctor := models.Doctor{
		Email: email, PasswordHash: string(hash),
		FirstName: "Dr", LastName: "Test",
		Specialization: spec, Examinations: exams,
	}
	database.DB.Create(&doctor)
	return doctor
}

func seedExaminations(t *testing.T) {
	t.Helper()
	exams := []models.Examination{
		{Name: "Blood Test", Description: "CBC", Price: 50.0},
		{Name: "X-Ray", Description: "Chest", Price: 100.0},
		{Name: "MRI", Description: "Brain", Price: 500.0},
	}
	for _, e := range exams {
		database.DB.Create(&e)
	}
}

func getAuthCookies(t *testing.T, userID uint, role string) []*http.Cookie {
	t.Helper()
	token, err := auth.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return []*http.Cookie{{Name: "access_token", Value: token}}
}
