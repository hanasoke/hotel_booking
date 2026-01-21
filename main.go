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

	// Static files untuk uploads
	http.Handle("/uploads/",
		http.StripPrefix("/uploads/",
			http.FileServer(http.Dir("uploads"))))

	// Homepage
	http.HandleFunc("/", homepage.Index)

	// Booking
	http.HandleFunc("/bookings", bookingcontroller.Index)
	http.HandleFunc("/add_booking", bookingcontroller.Add)
	http.HandleFunc("/edit_booking", bookingcontroller.Edit)
	http.HandleFunc("/detail_booking", bookingcontroller.Detail)

	// Room
	http.HandleFunc("/rooms", roomcontroller.Index)
	http.HandleFunc("/add_room", roomcontroller.Add)
	http.HandleFunc("/edit_room", roomcontroller.Edit)
	http.HandleFunc("/detail_room", roomcontroller.Detail)
	http.HandleFunc("/delete_room", roomcontroller.Delete)

	log.Println("Server running on port 2026")
	http.ListenAndServe(":2026", nil)
}
