package roommodel

import (
	"database/sql"
	"errors"
	"fmt"
	"hotel_booking/config"
	"hotel_booking/entities"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// Insert function with validations
func Insert(room entities.Room, imageFile *multipart.FileHeader) (error, string) {
	// Validasi data null/kosong
	if room.Name == "" {
		return errors.New("Nama ruangan wajib diisi"), ""
	}
	if room.Type == "" {
		return errors.New("Tipe ruangan wajib diisi"), ""
	}
	if room.Description == "" {
		return errors.New("Deskripsi wajib diisi"), ""
	}
	if room.Capacity <= 0 {
		return errors.New("Kapasitas harus lebih dari 0"), ""
	}
	if room.Status == "" {
		return errors.New("Status wajib dipilih"), ""
	}
	if room.PricePerDay <= 0 {
		return errors.New("Harga per hari harus lebih dari 0"), ""
	}

	// Validasi duplikat nama
	var exists bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM rooms WHERE name = ?)`
	err := config.DB.QueryRow(queryCheck, room.Name).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("Gagal memeriksa duplikasi: %v", err), ""
	}
	if exists {
		return errors.New("Nama ruangan sudah digunakan"), ""
	}

	// Handle upload gambar
	var imagePath string
	if imageFile != nil && imageFile.Size > 0 {
		// Validasi tipe file gambar
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
		}

		mimeType := imageFile.Header.Get("Content-Type")
		if !allowedTypes[mimeType] {
			return errors.New("Format gambar tidak didukung. Gunakan JPG, PNG, atau GIF"), ""
		}

		// Validasi ukuran file (max 2MB)
		if imageFile.Size > 2*1024*1024 {
			return errors.New("Ukuran gambar terlalu besar. Maksimal 2MB"), ""
		}

		// Generate unique filename
		fileExt := filepath.Ext(imageFile.Filename)
		newFilename := fmt.Sprintf("room_%d%s", time.Now().UnixNano(), fileExt)

		// Create upload directory if not exists
		uploadDir := "uploads/rooms"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return fmt.Errorf("Gagal membuat folder upload: %v", err), ""
		}

		// Save file
		dstPath := filepath.Join(uploadDir, newFilename)
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("Gagal membuat file: %v", err), ""
		}
		defer dst.Close()

		src, err := imageFile.Open()
		if err != nil {
			return fmt.Errorf("Gagal membuka file upload: %v", err), ""
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("Gagal menyimpan file: %v", err), ""
		}

		imagePath = newFilename
	} else {
		return errors.New("Gambar ruangan wajib diunggah"), ""
	}

	// Insert ke database
	query := `
        INSERT INTO rooms 
        (name, type, description, capacity, status, image, price_per_day, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
    `

	_, err = config.DB.Exec(
		query,
		room.Name,
		room.Type,
		room.Description,
		room.Capacity,
		room.Status,
		imagePath,
		room.PricePerDay,
	)

	if err != nil {
		// Delete uploaded file if database insert fails
		if imagePath != "" {
			os.Remove(filepath.Join("uploads/rooms", imagePath))
		}
		return fmt.Errorf("Gagal menyimpan data: %v", err), ""
	}

	return nil, "Data ruangan berhasil ditambahkan"
}

// Fungsi untuk mendapatkan semua rooms (untuk alert success)
func GetAll() ([]entities.Room, error) {
	rows, err := config.DB.Query(`
        SELECT id, name, type, description, capacity, status, 
               image, price_per_day, created_at 
        FROM rooms 
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []entities.Room
	for rows.Next() {
		var room entities.Room
		err := rows.Scan(
			&room.ID,
			&room.Name,
			&room.Type,
			&room.Description,
			&room.Capacity,
			&room.Status,
			&room.Image,
			&room.PricePerDay,
			&room.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}
