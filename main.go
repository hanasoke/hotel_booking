package main

import (
	"hotel_booking/config"
	"hotel_booking/controllers/homepage"
	"log"
	"net/http"
)

func main() {
	config.ConnectDB()

	// Homepage
	http.HandleFunc("/", homepage.Index)

	log.Println("Server running on port 2026")
	http.ListenAndServe(":2026", nil)
}
