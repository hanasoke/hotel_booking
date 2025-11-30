package handlers

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRooms(c *gin.Context) {
	var rooms []models.Room

	query := database.DB.Model(&models.Room{})

	// Filter by type
	if roomType := c.Query("type"); roomType != "" {
		query = query.Where("type = ?", roomType)
	}

	// Filter by availability
	if available := c.Query("available"); available != "" {
		isAvailable := available == "true"
		query = query.Where("is_available = ?", isAvailable)
	}

	if err := query.Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rooms,
		"count": len(rooms),
	})
}

func GetRoom(c *gin.Context) {
	id := c.Param("id")

	var room models.Room
	if err := database.DB.First(&room, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func CreateBooking(c *gin.Context) {
	var booking models.Booking

	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if room exists and is available
	var room models.Room
	if err := database.DB.First(&room, booking.RoomID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room not found"})
		return
	}

	if !room.IsAvailable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room is not available"})
		return
	}

	// Calculate total amount based on days
	days := booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24
	if days < 1 {
		days = 1
	}
	booking.TotalAmount = room.Price * float64(int(days))

	// Save booking
	if err := database.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Booking created successfully",
		"data":    booking,
	})
}

func GetBookings(c *gin.Context) {
	var bookings []models.Booking

	if err := database.DB.Preload("Room").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  bookings,
		"count": len(bookings),
	})
}

func UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")

	var booking models.Booking
	if err := database.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	if err := database.DB.Model(&booking).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking status updated successfully",
		"data":    booking,
	})
}
