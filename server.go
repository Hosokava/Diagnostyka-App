package main

import (
	"gin-quickstart/database"
	"gin-quickstart/handlers"
	"gin-quickstart/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/refresh", handlers.Refresh)
		authGroup.POST("/logout", handlers.Logout)
	}

	router.GET("/api/results/:hash", handlers.GetPublicResults)
	router.GET("/api/results/:hash/qr", handlers.GetQRCode)

	api := router.Group("/api")
	api.Use(middleware.RequireAuth())
	{
		api.GET("/me", handlers.GetMe)
		api.GET("/examinations", handlers.ListExaminations)

		patient := api.Group("/patient")
		patient.Use(middleware.RequireRole("patient"))
		{
			patient.PATCH("/profile", handlers.UpdatePatientProfile)
			patient.GET("/profile", handlers.GetPatientProfile)

			restricted := patient.Group("/")
			restricted.Use(middleware.RequireProfileComplete())
			{
				restricted.GET("/appointments/active", handlers.GetActiveAppointments)
				restricted.GET("/appointments/history", handlers.GetAppointmentHistory)
				restricted.POST("/book", handlers.BookAppointment)
				restricted.DELETE("/appointments/:id", handlers.CancelAppointment)
				restricted.POST("/profile/pesel", handlers.RevealPESEL)
			}
		}

		doctor := api.Group("/doctor")
		doctor.Use(middleware.RequireRole("doctor"))
		{
			doctor.PATCH("/profile", handlers.UpdateDoctorProfile)
			doctor.GET("/profile", handlers.GetDoctorProfile)

			restricted := doctor.Group("/")
			restricted.Use(middleware.RequireProfileComplete())
			{
				restricted.GET("/schedule/active", handlers.GetActiveSchedule)
				restricted.GET("/schedule/history", handlers.GetDoctorHistory)
				restricted.POST("/appointments/:id/complete", handlers.CompleteAppointment)
			}
		}
	}

	router.Run() //8080
}
