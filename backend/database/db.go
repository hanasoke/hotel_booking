package database

import (
	"backend/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Ganti dengan konfigurasi database XAMPP Anda
	dsn := "root:@tcp(127.0.0.1:3306)/hotel_booking?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate tables - disable foreign key constraints sementara
	err = DB.Exec("SET FOREIGN_KEY_CHECKS=0").Error
	if err != nil {
		log.Fatal("Failed to disable foreign key checks:", err)
	}

	// Migrate tables
	err = DB.AutoMigrate(&models.Room{}, &models.Booking{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Enable foreign key constraints
	err = DB.Exec("SET FOREIGN_KEY_CHECKS=1").Error
	if err != nil {
		log.Fatal("Failed to enable foreign key checks:", err)
	}

	fmt.Println("Database connected and migrated successfully")

	// Seed sample data
	seedSampleData()
}

func seedSampleData() {
	// Check if rooms already exist
	var count int64
	DB.Model(&models.Room{}).Count(&count)

	if count == 0 {
		rooms := []models.Room{
			{
				RoomNumber:  "101",
				Type:        "Single",
				Price:       250000,
				Description: "Kamar single dengan tempat tidur nyaman",
				Capacity:    1,
				IsAvailable: true,
				ImageURL:    "/images/room1.jpg",
			},
			{
				RoomNumber:  "102",
				Type:        "Double",
				Price:       400000,
				Description: "Kamar double untuk pasangan",
				Capacity:    2,
				IsAvailable: true,
				ImageURL:    "/images/room2.jpg",
			},
			{
				RoomNumber:  "201",
				Type:        "Deluxe",
				Price:       600000,
				Description: "Kamar deluxe dengan view bagus",
				Capacity:    2,
				IsAvailable: true,
				ImageURL:    "/images/room3.jpg",
			},
			{
				RoomNumber:  "202",
				Type:        "Suite",
				Price:       1000000,
				Description: "Kamar suite mewah dengan fasilitas lengkap",
				Capacity:    4,
				IsAvailable: true,
				ImageURL:    "/images/room4.jpg",
			},
		}

		DB.Create(&rooms)
		fmt.Println("Sample data seeded successfully")
	}
}
