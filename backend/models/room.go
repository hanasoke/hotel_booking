package models

import (
	"time"
)

type Room struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RoomNumber  string    `json:"room_number" gorm:"unique;not null"`
	Type        string    `json:"type" gorm:"not null"` // single, double, deluxe, suite
	Price       float64   `json:"price" gorm:"not null"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity" gorm:"not null"`
	IsAvailable bool      `json:"is_available" gorm:"default:true"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Booking struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	RoomID        uint      `json:"room_id" gorm:"not null"`
	CustomerName  string    `json:"customer_name" gorm:"not null"`
	CustomerEmail string    `json:"customer_email" gorm:"not null"`
	CustomerPhone string    `json:"customer_phone" gorm:"not null"`
	CheckInDate   time.Time `json:"check_in_date" gorm:"not null"`
	CheckOutDate  time.Time `json:"check_out_date" gorm:"not null"`
	TotalAmount   float64   `json:"total_amount" gorm:"not null"`
	Status        string    `json:"status" gorm:"default:'pending'"` // pending, confirmed, cancelled
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Room          Room      `json:"room" gorm:"foreignKey:RoomID"`
}
