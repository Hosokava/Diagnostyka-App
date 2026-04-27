package handlers

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"gin-quickstart/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetActiveSchedule(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var appointments []models.Appointment

	database.DB.Preload("Examination").Preload("Patient").
		Where("doctor_id = ? AND is_finished = ?", userID, false).
		Find(&appointments)

	response := make([]gin.H, 0)
	for _, a := range appointments {
		response = append(response, gin.H{
			"id":               a.ID,
			"date":             a.CreatedAt,
			"examination_name": a.Examination.Name,
			"patient_name":     a.Patient.FirstName + " " + a.Patient.LastName,
		})
	}
	c.JSON(http.StatusOK, response)
}

func GetDoctorHistory(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var appointments []models.Appointment

	database.DB.Preload("Examination").Preload("Patient").
		Where("doctor_id = ? AND is_finished = ?", userID, true).
		Find(&appointments)

	response := make([]gin.H, 0)
	for _, a := range appointments {
		response = append(response, gin.H{
			"id":               a.ID,
			"date":             a.CompletionDate,
			"examination_name": a.Examination.Name,
			"patient_name":     a.Patient.FirstName + " " + a.Patient.LastName,
			"result":           a.Result,
		})
	}
	c.JSON(http.StatusOK, response)
}

func UpdateDoctorProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		FirstName      *string `json:"first_name"`
		LastName       *string `json:"last_name"`
		Specialization *string `json:"specialization"`
		ExaminationIDs *[]uint `json:"examination_ids"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var doctor models.Doctor
	if err := database.DB.Preload("Examinations").First(&doctor, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	if body.FirstName != nil {
		doctor.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		doctor.LastName = *body.LastName
	}
	if body.Specialization != nil {
		doctor.Specialization = *body.Specialization
	}

	tx := database.DB.Begin()
	if err := tx.Save(&doctor).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update doctor info"})
		return
	}

	if body.ExaminationIDs != nil {
		var exams []models.Examination
		if len(*body.ExaminationIDs) > 0 {
			if err := tx.Find(&exams, *body.ExaminationIDs).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch examinations"})
				return
			}
		}

		if err := tx.Model(&doctor).Association("Examinations").Replace(exams); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update examination list"})
			return
		}
	}

	tx.Commit()

	database.DB.Preload("Examinations").First(&doctor, userID)

	var exams []gin.H
	for _, e := range doctor.Examinations {
		exams = append(exams, gin.H{
			"id":    e.ID,
			"name":  e.Name,
			"price": e.Price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
		"profile": gin.H{
			"id":             doctor.ID,
			"email":          doctor.Email,
			"first_name":     doctor.FirstName,
			"last_name":      doctor.LastName,
			"specialization": doctor.Specialization,
			"managed_exams":  exams,
		},
	})
}

func GetDoctorProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var doctor models.Doctor
	if err := database.DB.Preload("Examinations").First(&doctor, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "doctor not found"})
		return
	}

	var exams []gin.H
	for _, e := range doctor.Examinations {
		exams = append(exams, gin.H{
			"id":    e.ID,
			"name":  e.Name,
			"price": e.Price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             doctor.ID,
		"email":          doctor.Email,
		"first_name":     doctor.FirstName,
		"last_name":      doctor.LastName,
		"specialization": doctor.Specialization,
		"managed_exams":  exams,
	})
}

func CompleteAppointment(c *gin.Context) {
	doctorID := c.MustGet("userID").(uint)

	appointmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid appointment ID"})
		return
	}

	var body struct {
		Result          string `json:"result" binding:"required"`
		DiagnosticNotes string `json:"notes"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var appointment models.Appointment
	if err := database.DB.Preload("Patient").First(&appointment, appointmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		return
	}

	if appointment.DoctorID != doctorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not assigned to this appointment"})
		return
	}

	if appointment.IsFinished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "appointment already completed"})
		return
	}

	now := time.Now()
	appointment.Result = body.Result
	appointment.DiagnosticNotes = body.DiagnosticNotes
	appointment.IsFinished = true
	appointment.CompletionDate = &now

	if err := database.DB.Save(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save results"})
		return
	}

	services.SendResultsNotification(appointment.Patient.Email, appointment.QRCodeHash)

	c.JSON(http.StatusOK, gin.H{"message": "appointment completed successfully"})
}
