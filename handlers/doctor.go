package handlers

import (
	"github.com/gin-gonic/gin"
)

func GetDoctorSchedule(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Your schedule"})
}
