package main

import (
	"hotel_booking/config"
	"hotel_booking/controllers/bookingcontroller"
	"hotel_booking/controllers/homepage"
	"hotel_booking/controllers/roomcontroller"
	"log"
	"net/http"
)

func main() {
	config.ConnectDB()

	// Static files
	http.Handle("/assets/",
		http.StripPrefix("/assets/",
			http.FileServer(http.Dir("assets"))))

	// Homepage
	http.HandleFunc("/", homepage.Index)

	// Booking
	http.HandleFunc("/bookings", bookingcontroller.Index)

	// Room
	http.HandleFunc("/rooms", roomcontroller.Index)

	log.Println("Server running on port 2026")
	http.ListenAndServe(":2026", nil)
}
