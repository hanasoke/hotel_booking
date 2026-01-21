package entities

import "time"

type Booking struct {
	ID                int
	RoomID            Room
	UserName          string
	UserPhone         int
	check_in_date     time.Time
	check_out_date    time.Time
	total_price       int
	transaction_proof string
	CreatedAt         time.Time
}
