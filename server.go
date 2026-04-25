package main

import (
	"gin-quickstart/database"
	"gin-quickstart/handlers"
	"gin-quickstart/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	router := gin.Default()

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/refresh", handlers.Refresh)
		authGroup.POST("/logout", handlers.Logout)
	}

	api := router.Group("/api")
	api.Use(middleware.RequireAuth())
	{
		api.GET("/me", handlers.GetMe)

		patient := api.Group("/patient")
		patient.Use(middleware.RequireRole("patient"))
		{
			patient.GET("/appointments", handlers.GetPatientAppointments)
		}

		doctor := api.Group("/doctor")
		doctor.Use(middleware.RequireRole("doctor"))
		{
			doctor.GET("/schedule", handlers.GetDoctorSchedule)
		}
	}

	router.Run() //8080
}
