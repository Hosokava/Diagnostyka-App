package handlers

import (
	"gin-quickstart/database"
	"gin-quickstart/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListExaminations(c *gin.Context) {
	var examinations []models.Examination
	if err := database.DB.Find(&examinations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch examinations"})
		return
	}

	response := make([]gin.H, 0)
	for _, e := range examinations {
		response = append(response, gin.H{
			"id":          e.ID,
			"name":        e.Name,
			"description": e.Description,
			"price":       e.Price,
		})
	}

	c.JSON(http.StatusOK, response)
}
