package entities

import (
	"mime/multipart"
	"time"
)

type Booking struct {
	ID               int
	RoomID           int
	Room             Room
	UserName         string
	UserPhone        string
	CheckInDate      time.Time
	CheckOutDate     time.Time
	TotalPrice       int
	TransactionProof string
	CreatedAt        time.Time
}

type BookingForm struct {
	RoomID           int
	UserName         string
	UserPhone        string
	CheckInDate      string
	CheckOutDate     string
	TotalPrice       int
	TransactionProof *multipart.FileHeader
}
