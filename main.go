package main

import (
	"hotel_booking/config"
	"log"
	"net/http"
)

func main() {
	config.ConnectDB()

	log.Println("Server running on port 2026")
	http.ListenAndServe(":2026", nil)
}
