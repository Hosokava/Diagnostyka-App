package middleware

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireProfileComplete() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.MustGet("userID").(uint)
		role := c.MustGet("role").(string)

		if role == "patient" {
			var patient models.Patient
			if err := database.DB.First(&patient, userID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "patient not found"})
				return
			}
			if patient.FirstName == "" || patient.PESEL == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "profile_incomplete",
					"message": "Please complete your patient profile first",
				})
				return
			}
		} else if role == "doctor" {
			var doctor models.Doctor
			if err := database.DB.Preload("Examinations").First(&doctor, userID).Error; err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "doctor not found"})
				return
			}
			if doctor.FirstName == "" || doctor.Specialization == "" || len(doctor.Examinations) == 0 {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "profile_incomplete",
					"message": "Please complete your doctor profile (including at least one examination) first",
				})
				return
			}
		}

		c.Next()
	}
}
