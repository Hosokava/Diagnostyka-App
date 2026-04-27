package handlers

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/services"
	"gin-quickstart/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetActiveAppointments(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var appointments []models.Appointment
	database.DB.Preload("Examination").Preload("Doctor").
		Where("patient_id = ? AND is_finished = ?", userID, false).
		Find(&appointments)

	response := make([]gin.H, 0)
	for _, a := range appointments {
		response = append(response, gin.H{
			"id":               a.ID,
			"date":             a.CreatedAt,
			"examination_name": a.Examination.Name,
			"doctor_name":      a.Doctor.FirstName + " " + a.Doctor.LastName,
		})
	}
	c.JSON(http.StatusOK, response)
}

func GetAppointmentHistory(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var appointments []models.Appointment
	database.DB.Preload("Examination").Preload("Doctor").
		Where("patient_id = ? AND is_finished = ?", userID, true).
		Find(&appointments)

	response := make([]gin.H, 0)
	for _, a := range appointments {
		response = append(response, gin.H{
			"id":               a.ID,
			"date":             a.CompletionDate,
			"examination_name": a.Examination.Name,
			"doctor_name":      a.Doctor.FirstName + " " + a.Doctor.LastName,
			"results":          a.Result,
			"notes":            a.DiagnosticNotes,
		})
	}
	c.JSON(http.StatusOK, response)
}

func UpdatePatientProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		PESEL     *string `json:"pesel"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var patient models.Patient
	if err := database.DB.First(&patient, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	if body.FirstName != nil {
		patient.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		patient.LastName = *body.LastName
	}
	if body.PESEL != nil && *body.PESEL != "" {
		if len(*body.PESEL) > 7 && (*body.PESEL)[:7] == "XXXXXXX" {
		} else {
			if !utils.IsValidPESEL(*body.PESEL) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid PESEL number"})
				return
			}
			encryptedPesel, err := utils.EncryptAES(*body.PESEL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt PESEL"})
				return
			}
			patient.PESEL = encryptedPesel
		}
	}

	if err := database.DB.Save(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

func GetPatientProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var patient models.Patient
	if err := database.DB.First(&patient, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	decryptedPesel, _ := utils.DecryptAES(patient.PESEL)

	c.JSON(http.StatusOK, gin.H{
		"id":         patient.ID,
		"email":      patient.Email,
		"first_name": patient.FirstName,
		"last_name":  patient.LastName,
		"pesel":      utils.MaskPESEL(decryptedPesel),
	})
}

func RevealPESEL(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var patient models.Patient
	if err := database.DB.First(&patient, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(patient.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "incorrect password"})
		return
	}

	decryptedPesel, err := utils.DecryptAES(patient.PESEL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decrypt PESEL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pesel": decryptedPesel,
	})
}

func BookAppointment(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		ExaminationID uint `json:"examination_id" binding:"required"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var existing models.Appointment
	err := database.DB.Where("patient_id = ? AND examination_id = ? AND is_finished = ?", userID, body.ExaminationID, false).First(&existing).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have an active appointment for this examination"})
		return
	}

	doctorID, err := services.FindLeastBusyDoctor(body.ExaminationID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	appointment := models.Appointment{
		PatientID:     userID,
		DoctorID:      doctorID,
		ExaminationID: body.ExaminationID,
		QRCodeHash:    utils.GenerateRandomHash(16),
		IsFinished:    false,
	}

	if err := database.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create appointment"})
		return
	}

	database.DB.Preload("Doctor").Preload("Examination").First(&appointment, appointment.ID)

	var patient models.Patient
	database.DB.First(&patient, userID)

	services.SendBookingConfirmation(
		patient.Email,
		appointment.Doctor.FirstName+" "+appointment.Doctor.LastName,
		appointment.Examination.Name,
	)

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Appointment booked successfully",
		"appointment_id": appointment.ID,
		"doctor":         appointment.Doctor.FirstName + " " + appointment.Doctor.LastName,
		"examination":    appointment.Examination.Name,
		"date":           appointment.CreatedAt,
		"qr_code_hash":   appointment.QRCodeHash,
	})
}

func CancelAppointment(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment ID"})
		return
	}

	var appointment models.Appointment
	if err := database.DB.First(&appointment, appointmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	if appointment.PatientID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only cancel your own appointments"})
		return
	}

	if appointment.IsFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel a completed appointment"})
		return
	}

	if err := database.DB.Delete(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel appointment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled successfully"})
}
