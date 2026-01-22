package bookingmodel

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
	"regexp"
	"time"
)

// Helper function untuk normalisasi nomor telepon
func normalizePhoneNumber(phone string) string {
	// Hapus semua karakter non-digit
	re := regexp.MustCompile(`\D`)
	cleanPhone := re.ReplaceAllString(phone, "")

	// Jika diawali 0, ganti dengan +62
	if len(cleanPhone) > 0 && cleanPhone[0] == '0' {
		cleanPhone = "62" + cleanPhone[1:]
	}

	// Jika diawali 62, tambahkan +
	if len(cleanPhone) > 0 && cleanPhone[:2] == "62" {
		cleanPhone = "+" + cleanPhone
	}

	return cleanPhone
}

// Get all available rooms for booking
func GetAvailableRooms(checkIn, checkOut time.Time) ([]entities.Room, error) {
	query := `
		SELECT r.id, r.name, r.type, r.description, r.capacity, 
			   r.status, r.image, r.price_per_day, r.created_at
		FROM rooms r
		WHERE r.status = 'active'
		AND r.id NOT IN (
			SELECT b.room_id 
			FROM bookings b
			WHERE (
				(b.check_in_date <= ? AND b.check_out_date >= ?) OR
				(b.check_in_date <= ? AND b.check_out_date >= ?) OR
				(? <= b.check_out_date AND ? >= b.check_in_date)
			)
		)
		ORDER BY r.name
	`

	rows, err := config.DB.Query(query, checkIn, checkIn, checkOut, checkOut, checkIn, checkOut)
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

// Get all bookings
func GetAll() ([]entities.Booking, error) {
	query := `
		SELECT b.id, b.room_id, b.user_name, b.user_phone, 
			   b.check_in_date, b.check_out_date, b.total_price,
			   b.transaction_proof, b.created_at,
			   r.name as room_name, r.image as room_image
		FROM bookings b
		LEFT JOIN rooms r ON b.room_id = r.id
		ORDER BY b.created_at DESC
	`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []entities.Booking
	for rows.Next() {
		var booking entities.Booking
		var roomName, roomImage sql.NullString

		err := rows.Scan(
			&booking.ID,
			&booking.RoomID,
			&booking.UserName,
			&booking.UserPhone,
			&booking.CheckInDate,
			&booking.CheckOutDate,
			&booking.TotalPrice,
			&booking.TransactionProof,
			&booking.CreatedAt,
			&roomName,
			&roomImage,
		)
		if err != nil {
			return nil, err
		}

		// Set room info
		if roomName.Valid {
			booking.Room = entities.Room{
				ID:    booking.RoomID,
				Name:  roomName.String,
				Image: roomImage.String,
			}
		}

		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// Insert booking dengan validasi
func Insert(booking entities.Booking, proofFile *multipart.FileHeader) (error, string) {
	// Validasi data null/kosong
	if booking.UserName == "" {
		return errors.New("Nama lengkap wajib diisi"), ""
	}
	if booking.UserPhone == "" {
		return errors.New("Nomor telepon wajib diisi"), ""
	}
	if booking.RoomID == 0 {
		return errors.New("Ruangan wajib dipilih"), ""
	}
	if booking.TotalPrice <= 0 {
		return errors.New("Total harga tidak valid"), ""
	}
	if booking.CheckInDate.IsZero() {
		return errors.New("Tanggal check-in wajib diisi"), ""
	}
	if booking.CheckOutDate.IsZero() {
		return errors.New("Tanggal check-out wajib diisi"), ""
	}

	// Validasi format nomor telepon (Indonesia)
	phoneRegex := regexp.MustCompile(`^(\+62|62|0)8[1-9][0-9]{6,9}$`)
	if !phoneRegex.MatchString(booking.UserPhone) {
		return errors.New("Format nomor telepon tidak valid. Gunakan format Indonesia (contoh: 081234567890 / 6285819536158)"), ""
	}

	// Normalisasi nomor telepon ke format +62
	normalizedPhone := normalizePhoneNumber(booking.UserPhone)
	booking.UserPhone = normalizedPhone

	// Validasi tanggal
	today := time.Now().Truncate(24 * time.Hour)

	// Check-in tidak boleh sebelum hari ini
	if booking.CheckInDate.Before(today) {
		return errors.New("Tanggal check-in tidak boleh sebelum hari ini"), ""
	}

	// Check-out tidak boleh sebelum check-in
	if !booking.CheckOutDate.After(booking.CheckInDate) {
		return errors.New("Tanggal check-out harus setelah tanggal check-in"), ""
	}

	// Validasi room tersedia
	var roomStatus string
	var roomPrice int
	queryCheckRoom := `SELECT status, price_per_day FROM rooms WHERE id = ?`
	err := config.DB.QueryRow(queryCheckRoom, booking.RoomID).Scan(&roomStatus, &roomPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Ruangan tidak ditemukan"), ""
		}
		return fmt.Errorf("Gagal memeriksa ketersediaan ruangan: %v", err), ""
	}

	if roomStatus != "active" {
		return errors.New("Ruangan tidak tersedia untuk booking"), ""
	}

	// Validasi tanggal tidak bentrok dengan booking lain
	var count int
	queryCheckDate := `
		SELECT COUNT(*) FROM bookings 
		WHERE room_id = ? AND (
			(check_in_date <= ? AND check_out_date >= ?) OR
			(check_in_date <= ? AND check_out_date >= ?) OR
			(? <= check_out_date AND ? >= check_in_date)
		) AND id != ?
	`
	err = config.DB.QueryRow(queryCheckDate,
		booking.RoomID,
		booking.CheckInDate, booking.CheckInDate,
		booking.CheckOutDate, booking.CheckOutDate,
		booking.CheckInDate, booking.CheckOutDate,
		0).Scan(&count)

	if err != nil {
		return fmt.Errorf("Gagal memeriksa ketersediaan tanggal: %v", err), ""
	}

	if count > 0 {
		return errors.New("Ruangan sudah dibooking pada tanggal yang dipilih"), ""
	}

	// Handle upload bukti transaksi
	var proofPath string
	if proofFile != nil && proofFile.Size > 0 {
		// Validasi tipe file gambar
		allowedTypes := map[string]bool{
			"image/jpeg":      true,
			"image/jpg":       true,
			"image/png":       true,
			"image/gif":       true,
			"application/pdf": true,
		}

		mimeType := proofFile.Header.Get("Content-Type")
		if !allowedTypes[mimeType] {
			return errors.New("Format file tidak didukung. Gunakan JPG, PNG, GIF, atau PDF"), ""
		}

		// Validasi ukuran file (max 5MB)
		if proofFile.Size > 5*1024*1024 {
			return errors.New("Ukuran file terlalu besar. Maksimal 5MB"), ""
		}

		// Generate unique filename
		fileExt := filepath.Ext(proofFile.Filename)
		newFilename := fmt.Sprintf("proof_%d%s", time.Now().UnixNano(), fileExt)

		// Create upload directory if not exists
		uploadDir := "uploads/bookings"
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

		src, err := proofFile.Open()
		if err != nil {
			return fmt.Errorf("Gagal membuka file upload: %v", err), ""
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("Gagal menyimpan file: %v", err), ""
		}

		proofPath = newFilename
	} else {
		return errors.New("Bukti transaksi wajib diunggah"), ""
	}

	// Hitung total harga berdasarkan jumlah hari
	days := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
	if days < 1 {
		days = 1
	}
	calculatedPrice := days * roomPrice

	// Validasi total harga
	if booking.TotalPrice != calculatedPrice {
		return fmt.Errorf("Total harga tidak sesuai. Seharusnya Rp %d untuk %d hari",
			calculatedPrice, days), ""
	}

	// Insert ke database
	query := `
		INSERT INTO bookings 
		(room_id, user_name, user_phone, check_in_date, check_out_date, 
		 total_price, transaction_proof, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW())
	`

	_, err = config.DB.Exec(
		query,
		booking.RoomID,
		booking.UserName,
		booking.UserPhone,
		booking.CheckInDate,
		booking.CheckOutDate,
		calculatedPrice, // Gunakan harga yang dihitung
		proofPath,
	)

	if err != nil {
		// Delete uploaded file if database insert fails
		if proofPath != "" {
			os.Remove(filepath.Join("uploads/bookings", proofPath))
		}
		return fmt.Errorf("Gagal menyimpan data: %v", err), ""
	}

	return nil, "Booking berhasil ditambahkan"
}

// Get booking by ID
func GetByID(id int) (entities.Booking, error) {
	var booking entities.Booking

	query := `
		SELECT b.id, b.room_id, b.user_name, b.user_phone, 
			   b.check_in_date, b.check_out_date, b.total_price,
			   b.transaction_proof, b.created_at,
			   r.id as room_id, r.name as room_name, r.type as room_type,
			   r.description as room_description, r.capacity as room_capacity,
			   r.status as room_status, r.image as room_image,
			   r.price_per_day as room_price_per_day, r.created_at as room_created_at
		FROM bookings b
		LEFT JOIN rooms r ON b.room_id = r.id
		WHERE b.id = ?
	`

	err := config.DB.QueryRow(query, id).Scan(
		&booking.ID,
		&booking.Room.ID,
		&booking.UserName,
		&booking.UserPhone,
		&booking.CheckInDate,
		&booking.CheckOutDate,
		&booking.TotalPrice,
		&booking.TransactionProof,
		&booking.CreatedAt,
		&booking.Room.ID,
		&booking.Room.Name,
		&booking.Room.Type,
		&booking.Room.Description,
		&booking.Room.Capacity,
		&booking.Room.Status,
		&booking.Room.Image,
		&booking.Room.PricePerDay,
		&booking.Room.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return booking, errors.New("booking not found")
		}
		return booking, err
	}

	return booking, nil
}

// Update booking dengan validasi
func Update(booking entities.Booking, proofFile *multipart.FileHeader, bookingID int) (error, string) {
	// Validasi data null/kosong
	if booking.UserName == "" {
		return errors.New("Nama lengkap wajib diisi"), ""
	}
	if booking.UserPhone == "" {
		return errors.New("Nomor telepon wajib diisi"), ""
	}
	if booking.RoomID == 0 {
		return errors.New("Ruangan wajib dipilih"), ""
	}
	if booking.TotalPrice <= 0 {
		return errors.New("Total harga tidak valid"), ""
	}
	if booking.CheckInDate.IsZero() {
		return errors.New("Tanggal check-in wajib diisi"), ""
	}
	if booking.CheckOutDate.IsZero() {
		return errors.New("Tanggal check-out wajib diisi"), ""
	}

	// Validasi format nomor telepon (Indonesia)
	phoneRegex := regexp.MustCompile(`^(\+62|62|0)8[1-9][0-9]{6,9}$`)
	if !phoneRegex.MatchString(booking.UserPhone) {
		return errors.New("Format nomor telepon tidak valid. Gunakan format Indonesia (contoh: 081234567890 / 6285819536158)"), ""
	}

	// Normalisasi nomor telepon ke format +62
	normalizedPhone := normalizePhoneNumber(booking.UserPhone)
	booking.UserPhone = normalizedPhone

	// Validasi tanggal
	today := time.Now().Truncate(24 * time.Hour)

	// Check-in tidak boleh sebelum hari ini (kecuali booking lama)
	currentBooking, err := GetByID(bookingID)
	if err != nil {
		return errors.New("Booking tidak ditemukan"), ""
	}

	// Hanya validasi jika tanggal check-in diubah ke sebelum hari ini
	if !booking.CheckInDate.Equal(currentBooking.CheckInDate) && booking.CheckInDate.Before(today) {
		return errors.New("Tanggal check-in tidak boleh sebelum hari ini"), ""
	}

	// Check-out tidak boleh sebelum check-in
	if !booking.CheckOutDate.After(booking.CheckInDate) {
		return errors.New("Tanggal check-out harus setelah tanggal check-in"), ""
	}

	// Validasi room tersedia
	var roomStatus string
	var roomPrice int
	queryCheckRoom := `SELECT status, price_per_day FROM rooms WHERE id = ?`
	err = config.DB.QueryRow(queryCheckRoom, booking.RoomID).Scan(&roomStatus, &roomPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Ruangan tidak ditemukan"), ""
		}
		return fmt.Errorf("Gagal memeriksa ketersediaan ruangan: %v", err), ""
	}

	if roomStatus != "active" {
		return errors.New("Ruangan tidak tersedia untuk booking"), ""
	}

	// VALIDASI DUPLIKAT ROOM - Periksa apakah room sudah dipakai orang lain
	var count int
	queryCheckDate := `
        SELECT COUNT(*) FROM bookings 
        WHERE room_id = ? 
        AND id != ?
        AND (
            (check_in_date <= ? AND check_out_date >= ?) OR
            (check_in_date <= ? AND check_out_date >= ?) OR
            (? <= check_out_date AND ? >= check_in_date)
        )
    `
	err = config.DB.QueryRow(queryCheckDate,
		booking.RoomID,
		bookingID, // exclude current booking
		booking.CheckInDate, booking.CheckInDate,
		booking.CheckOutDate, booking.CheckOutDate,
		booking.CheckInDate, booking.CheckOutDate).Scan(&count)

	if err != nil {
		return fmt.Errorf("Gagal memeriksa ketersediaan tanggal: %v", err), ""
	}

	if count > 0 {
		return errors.New("Ruangan sudah dibooking pada tanggal yang dipilih"), ""
	}

	// Handle upload bukti transaksi jika ada perubahan
	var proofPath string
	var oldProofPath string

	// Get old proof path
	queryGetOldProof := `SELECT transaction_proof FROM bookings WHERE id = ?`
	err = config.DB.QueryRow(queryGetOldProof, bookingID).Scan(&oldProofPath)
	if err != nil {
		return fmt.Errorf("Gagal mengambil data bukti lama: %v", err), ""
	}

	if proofFile != nil && proofFile.Size > 0 {
		// Validasi tipe file gambar
		allowedTypes := map[string]bool{
			"image/jpeg":      true,
			"image/jpg":       true,
			"image/png":       true,
			"image/gif":       true,
			"application/pdf": true,
		}

		mimeType := proofFile.Header.Get("Content-Type")
		if !allowedTypes[mimeType] {
			return errors.New("Format file tidak didukung. Gunakan JPG, PNG, GIF, atau PDF"), ""
		}

		// Validasi ukuran file (max 5MB)
		if proofFile.Size > 5*1024*1024 {
			return errors.New("Ukuran file terlalu besar. Maksimal 5MB"), ""
		}

		// Generate unique filename
		fileExt := filepath.Ext(proofFile.Filename)
		newFilename := fmt.Sprintf("proof_%d%s", time.Now().UnixNano(), fileExt)

		// Create upload directory if not exists
		uploadDir := "uploads/bookings"
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

		src, err := proofFile.Open()
		if err != nil {
			return fmt.Errorf("Gagal membuka file upload: %v", err), ""
		}
		defer src.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("Gagal menyimpan file: %v", err), ""
		}

		proofPath = newFilename

		// Delete old proof file if exists
		if oldProofPath != "" {
			oldFilePath := filepath.Join(uploadDir, oldProofPath)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}
	} else {
		// Jika tidak upload file baru, gunakan yang lama
		proofPath = oldProofPath
	}

	// Hitung total harga berdasarkan jumlah hari
	days := int(booking.CheckOutDate.Sub(booking.CheckInDate).Hours() / 24)
	if days < 1 {
		days = 1
	}
	calculatedPrice := days * roomPrice

	// Validasi total harga
	if booking.TotalPrice != calculatedPrice {
		return fmt.Errorf("Total harga tidak sesuai. Seharusnya Rp %d untuk %d hari",
			calculatedPrice, days), ""
	}

	// Update ke database
	query := `
        UPDATE bookings 
        SET room_id = ?, 
            user_name = ?, 
            user_phone = ?, 
            check_in_date = ?, 
            check_out_date = ?, 
            total_price = ?, 
            transaction_proof = ?
        WHERE id = ?
    `

	_, err = config.DB.Exec(
		query,
		booking.RoomID,
		booking.UserName,
		booking.UserPhone,
		booking.CheckInDate,
		booking.CheckOutDate,
		calculatedPrice,
		proofPath,
		bookingID,
	)

	if err != nil {
		// Delete uploaded file if database update fails
		if proofPath != "" && proofPath != oldProofPath {
			os.Remove(filepath.Join("uploads/bookings", proofPath))
		}
		return fmt.Errorf("Gagal mengupdate data: %v", err), ""
	}

	return nil, "Booking berhasil diupdate"
}
