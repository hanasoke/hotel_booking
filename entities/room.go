package entities

import (
	"mime/multipart"
	"time"
)

type Room struct {
	ID          int
	Name        string
	Type        string
	Description string
	Capacity    int
	Status      string
	Image       string
	PricePerDay int
	CreatedAt   time.Time
}

type RoomForm struct {
	Name        string
	Type        string
	Description string
	Capacity    int
	Status      string
	PricePerDay int
	ImageFile   *multipart.FileHeader
}
