package main

import (
	"backend/database"
	"backend/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database.InitDB()

	// Initialize router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	api := router.Group("/api")
	{
		// Room routes
		rooms := api.Group("/rooms")
		{
			rooms.GET("", handlers.GetRooms)
			rooms.GET("/:id", handlers.GetRoom)
		}

		// Booking routes
		bookings := api.Group("/bookings")
		{
			bookings.GET("", handlers.GetBookings)
			bookings.POST("", handlers.CreateBooking)
			bookings.PUT("/:id/status", handlers.UpdateBookingStatus)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "Hotel Booking API is running",
		})
	})

	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
