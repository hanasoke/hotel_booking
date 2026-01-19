package roommodel

import (
	"hotel_booking/config"
	"hotel_booking/entities"
)

func Insert(room entities.Room) error {

	query := `
		INSERT INTO rooms
		(name, type, description, capacity, status, image, price_per_day, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`

	_, err := config.DB.Exec(
		query,
		room.Name,
		room.Type,
		room.Description,
		room.Capacity,
		room.Status,
		room.Image,
		room.PricePerDay,
	)

	return err
}
