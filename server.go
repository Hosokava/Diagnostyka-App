package main

import (
	"gin-quickstart/database"
	"gin-quickstart/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/refresh", handlers.Refresh)
		authGroup.POST("/logout", handlers.Logout)
	}

	router.Run() //8080
}
