package handlers

import (
	"fmt"
	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetPublicResults(c *gin.Context) {
	hash := c.Param("hash")

	var appointment models.Appointment
	if err := database.DB.Preload("Patient").Preload("Doctor").Preload("Examination").
		Where("qr_code_hash = ?", hash).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Results not found or invalid link"})
		return
	}

	if !appointment.IsFinished {
		c.JSON(http.StatusForbidden, gin.H{"error": "Examination is still in progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patient_name":      appointment.Patient.FirstName + " " + appointment.Patient.LastName,
		"doctor_name":       appointment.Doctor.FirstName + " " + appointment.Doctor.LastName,
		"examination":       appointment.Examination.Name,
		"date":              appointment.CompletionDate,
		"diagnostic_result": appointment.Result,
		"notes":             appointment.DiagnosticNotes,
	})
}

func GetQRCode(c *gin.Context) {
	hash := c.Param("hash")

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	data := fmt.Sprintf("%s/api/results/%s", appURL, hash)

	png, err := utils.GenerateQRCodePNG(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
