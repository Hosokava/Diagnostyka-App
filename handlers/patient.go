package handlers

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPatientAppointments(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Your appointments"})
}

func UpdatePatientProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		PESEL     string `json:"pesel" binding:"required"`
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

	patient.FirstName = body.FirstName
	patient.LastName = body.LastName
	patient.PESEL = body.PESEL

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

	c.JSON(http.StatusOK, gin.H{
		"id":         patient.ID,
		"email":      patient.Email,
		"first_name": patient.FirstName,
		"last_name":  patient.LastName,
		"pesel":      patient.PESEL,
	})
}
