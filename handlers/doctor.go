package handlers

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetDoctorSchedule(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Your schedule"})
}

func UpdateDoctorProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var body struct {
		FirstName      string `json:"first_name" binding:"required"`
		LastName       string `json:"last_name" binding:"required"`
		Specialization string `json:"specialization" binding:"required"`
		ExaminationIDs []uint `json:"examination_ids" binding:"required"`
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

	var exams []models.Examination
	if len(body.ExaminationIDs) > 0 {
		if err := database.DB.Find(&exams, body.ExaminationIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch examinations"})
			return
		}
	}

	if len(exams) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one valid examination ID must be selected"})
		return
	}

	doctor.FirstName = body.FirstName
	doctor.LastName = body.LastName
	doctor.Specialization = body.Specialization

	tx := database.DB.Begin()
	if err := tx.Save(&doctor).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update doctor info"})
		return
	}

	if err := tx.Model(&doctor).Association("Examinations").Replace(exams); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update examination list"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
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
