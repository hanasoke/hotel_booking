package entities

import "time"

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
