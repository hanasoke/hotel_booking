package bookingcontroller

import (
	"hotel_booking/controllers"
	"hotel_booking/entities"
	"hotel_booking/models/bookingmodel"
	"hotel_booking/models/roommodel"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Struct untuk passing data ke template
type PageData struct {
	Bookings       []entities.Booking
	Booking        entities.Booking
	Rooms          []entities.Room
	AvailableRooms []entities.Room
	Error          string
	Success        string
	FormData       map[string]string
	CheckInDate    string
	CheckOutDate   string
}

// Helper function untuk mengekstrak form data
func extractFormData(r *http.Request) map[string]string {
	return map[string]string{
		"user_name":      strings.TrimSpace(r.FormValue("user_name")),
		"user_phone":     strings.TrimSpace(r.FormValue("user_phone")),
		"room_id":        strings.TrimSpace(r.FormValue("room_id")),
		"check_in_date":  strings.TrimSpace(r.FormValue("check_in_date")),
		"check_out_date": strings.TrimSpace(r.FormValue("check_out_date")),
		"total_price":    strings.TrimSpace(r.FormValue("total_price")),
	}
}

// Helper function untuk validasi phone number
func isValidPhone(phone string) bool {
	// Hapus +62 atau 0 di depan
	cleanPhone := strings.TrimPrefix(phone, "+62")
	cleanPhone = strings.TrimPrefix(cleanPhone, "62")
	cleanPhone = strings.TrimPrefix(cleanPhone, "0")

	// Cek panjang dan hanya angka
	if len(cleanPhone) < 10 || len(cleanPhone) > 13 {
		return false
	}

	for _, c := range cleanPhone {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

func Index(w http.ResponseWriter, r *http.Request) {
	// Get messages from query parameters
	successMsg := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	// Get all bookings
	bookings, err := bookingmodel.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Bookings: bookings,
		Success:  successMsg,
		Error:    errorMsg,
	}

	tmpl, err := controllers.LoadTemplate(
		"views/booking/index.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", data)
}

func Add(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Get all active rooms
		rooms, err := roommodel.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Filter only active rooms
		var activeRooms []entities.Room
		for _, room := range rooms {
			if room.Status == "active" {
				activeRooms = append(activeRooms, room)
			}
		}

		data := PageData{
			Rooms: activeRooms,
			FormData: map[string]string{
				"check_in_date":  time.Now().Format("2006-01-02"),
				"check_out_date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
			},
		}

		tmpl, err := controllers.LoadTemplate(
			"views/booking/crud/add.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	if r.Method == "POST" {
		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			http.Error(w, "Gagal parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get form values
		formData := extractFormData(r)

		// Validasi required fields
		errors := []string{}

		if formData["user_name"] == "" {
			errors = append(errors, "Nama lengkap wajib diisi")
		}
		if formData["user_phone"] == "" {
			errors = append(errors, "Nomor telepon wajib diisi")
		} else if !isValidPhone(formData["user_phone"]) {
			errors = append(errors, "Format nomor telepon tidak valid. Gunakan format Indonesia")
		}
		if formData["room_id"] == "" {
			errors = append(errors, "Ruangan wajib dipilih")
		}
		if formData["check_in_date"] == "" {
			errors = append(errors, "Tanggal check-in wajib diisi")
		}
		if formData["check_out_date"] == "" {
			errors = append(errors, "Tanggal check-out wajib diisi")
		}
		if formData["total_price"] == "" {
			errors = append(errors, "Total harga wajib diisi")
		}

		// Parse dates
		var checkInDate, checkOutDate time.Time
		if formData["check_in_date"] != "" {
			checkInDate, err = time.Parse("2006-01-02", formData["check_in_date"])
			if err != nil {
				errors = append(errors, "Format tanggal check-in tidak valid")
			}
		}
		if formData["check_out_date"] != "" {
			checkOutDate, err = time.Parse("2006-01-02", formData["check_out_date"])
			if err != nil {
				errors = append(errors, "Format tanggal check-out tidak valid")
			}
		}

		// Validasi tanggal jika parse berhasil
		if !checkInDate.IsZero() && !checkOutDate.IsZero() {
			today := time.Now().Truncate(24 * time.Hour)

			// Check-in tidak boleh sebelum hari ini
			if checkInDate.Before(today) {
				errors = append(errors, "Tanggal check-in tidak boleh sebelum hari ini")
			}

			// Check-out tidak boleh sebelum atau sama dengan check-in
			if !checkOutDate.After(checkInDate) {
				errors = append(errors, "Tanggal check-out harus setelah tanggal check-in")
			}
		}

		// Get uploaded file
		var proofFile *multipart.FileHeader
		file, header, err := r.FormFile("transaction_proof")
		if err == nil {
			proofFile = header
			file.Close()
		} else {
			errors = append(errors, "Bukti transaksi wajib diunggah")
		}

		// Get room price untuk validasi total harga
		roomID, _ := strconv.Atoi(formData["room_id"])
		var roomPrice int
		if roomID > 0 {
			room, err := roommodel.GetById(roomID)
			if err != nil {
				errors = append(errors, "Ruangan tidak ditemukan")
			} else {
				roomPrice = room.PricePerDay
			}
		}

		// Hitung total harga yang seharusnya
		totalPrice, _ := strconv.Atoi(formData["total_price"])
		if !checkInDate.IsZero() && !checkOutDate.IsZero() && roomPrice > 0 {
			days := int(checkOutDate.Sub(checkInDate).Hours() / 24)
			if days < 1 {
				days = 1
			}
			calculatedPrice := days * roomPrice

			if totalPrice != calculatedPrice {
				errors = append(errors,
					"Total harga tidak sesuai. Seharusnya Rp "+strconv.Itoa(calculatedPrice))
			}
		}

		// Jika ada error validasi
		if len(errors) > 0 {
			// Get all active rooms untuk form
			rooms, _ := roommodel.GetAll()
			var activeRooms []entities.Room
			for _, room := range rooms {
				if room.Status == "active" {
					activeRooms = append(activeRooms, room)
				}
			}

			data := PageData{
				Rooms:    activeRooms,
				Error:    strings.Join(errors, "<br>"),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/booking/crud/add.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Create booking entity
		booking := entities.Booking{
			RoomID:       roomID,
			UserName:     formData["user_name"],
			UserPhone:    formData["user_phone"],
			CheckInDate:  checkInDate,
			CheckOutDate: checkOutDate,
			TotalPrice:   totalPrice,
		}

		// Insert dengan validasi
		err, successMsg := bookingmodel.Insert(booking, proofFile)

		if err != nil {
			// Jika error, tampilkan form kembali dengan data sebelumnya
			rooms, _ := roommodel.GetAll()
			var activeRooms []entities.Room
			for _, room := range rooms {
				if room.Status == "active" {
					activeRooms = append(activeRooms, room)
				}
			}

			data := PageData{
				Rooms:    activeRooms,
				Error:    err.Error(),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/booking/crud/add.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Jika sukses, redirect ke halaman utama dengan pesan success
		http.Redirect(w, r, "/bookings?success="+successMsg, http.StatusSeeOther)
	}
}

func Detail(w http.ResponseWriter, r *http.Request) {
	// Get ID from query parameter
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		return
	}

	// Convert ID to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		return
	}

	// Get booking data from model
	booking, err := bookingmodel.GetByID(id)
	if err != nil {
		// Jika booking tidak ditemukan, redirect ke list
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		return
	}

	// Format phone number for display
	displayPhone := booking.UserPhone
	if strings.HasPrefix(displayPhone, "+62") {
		displayPhone = "0" + displayPhone[3:]
	}

	// Prepare data for template
	data := PageData{
		Booking: booking,
	}

	// Load and execute template
	tmpl, err := controllers.LoadTemplate(
		"views/booking/crud/detail.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute template
	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Edit(w http.ResponseWriter, r *http.Request) {
	// Get ID from query parameter
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		return
	}

	// Convert ID to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		// Get booking data
		booking, err := bookingmodel.GetByID(id)
		if err != nil {
			// Jika booking tidak ditemukan, redirect ke list
			http.Redirect(w, r, "/bookings?error=Booking+tidak+ditemukan", http.StatusSeeOther)
			return
		}

		// Get all active rooms
		rooms, err := roommodel.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Filter only active rooms
		var activeRooms []entities.Room
		for _, room := range rooms {
			if room.Status == "active" || room.ID == booking.Room.ID {
				activeRooms = append(activeRooms, room)
			}
		}

		// Format dates for input field
		checkInStr := booking.CheckInDate.Format("2006-01-02")
		checkOutStr := booking.CheckOutDate.Format("2006-01-02")

		data := PageData{
			Booking: booking,
			Rooms:   activeRooms,
			FormData: map[string]string{
				"user_name":      booking.UserName,
				"user_phone":     booking.UserPhone,
				"room_id":        strconv.Itoa(booking.Room.ID),
				"check_in_date":  checkInStr,
				"check_out_date": checkOutStr,
				"total_price":    strconv.Itoa(booking.TotalPrice),
			},
		}

		tmpl, err := controllers.LoadTemplate(
			"views/booking/crud/edit.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	if r.Method == "POST" {
		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			http.Error(w, "Gagal parsing form: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get form values
		formData := extractFormData(r)

		// Validasi required fields
		errors := []string{}

		if formData["user_name"] == "" {
			errors = append(errors, "Nama lengkap wajib diisi")
		}
		if formData["user_phone"] == "" {
			errors = append(errors, "Nomor telepon wajib diisi")
		} else if !isValidPhone(formData["user_phone"]) {
			errors = append(errors, "Format nomor telepon tidak valid. Gunakan format Indonesia")
		}
		if formData["room_id"] == "" {
			errors = append(errors, "Ruangan wajib dipilih")
		}
		if formData["check_in_date"] == "" {
			errors = append(errors, "Tanggal check-in wajib diisi")
		}
		if formData["check_out_date"] == "" {
			errors = append(errors, "Tanggal check-out wajib diisi")
		}
		if formData["total_price"] == "" {
			errors = append(errors, "Total harga wajib diisi")
		}

		// Parse dates
		var checkInDate, checkOutDate time.Time
		if formData["check_in_date"] != "" {
			checkInDate, err = time.Parse("2006-01-02", formData["check_in_date"])
			if err != nil {
				errors = append(errors, "Format tanggal check-in tidak valid")
			}
		}
		if formData["check_out_date"] != "" {
			checkOutDate, err = time.Parse("2006-01-02", formData["check_out_date"])
			if err != nil {
				errors = append(errors, "Format tanggal check-out tidak valid")
			}
		}

		// Validasi tanggal jika parse berhasil
		if !checkInDate.IsZero() && !checkOutDate.IsZero() {
			// Check-out tidak boleh sebelum atau sama dengan check-in
			if !checkOutDate.After(checkInDate) {
				errors = append(errors, "Tanggal check-out harus setelah tanggal check-in")
			}
		}

		// Get uploaded file (optional untuk edit)
		var proofFile *multipart.FileHeader
		file, header, err := r.FormFile("transaction_proof")
		if err == nil {
			proofFile = header
			file.Close()
		}

		// Get room price untuk validasi total harga
		roomID, _ := strconv.Atoi(formData["room_id"])
		var roomPrice int
		if roomID > 0 {
			room, err := roommodel.GetById(roomID)
			if err != nil {
				errors = append(errors, "Ruangan tidak ditemukan")
			} else {
				roomPrice = room.PricePerDay
			}
		}

		// Hitung total harga yang seharusnya
		totalPrice, _ := strconv.Atoi(formData["total_price"])
		if !checkInDate.IsZero() && !checkOutDate.IsZero() && roomPrice > 0 {
			days := int(checkOutDate.Sub(checkInDate).Hours() / 24)
			if days < 1 {
				days = 1
			}
			calculatedPrice := days * roomPrice

			if totalPrice != calculatedPrice {
				errors = append(errors,
					"Total harga tidak sesuai. Seharusnya Rp "+strconv.Itoa(calculatedPrice))
			}
		}

		// Jika ada error validasi
		if len(errors) > 0 {
			// Get booking data dan rooms
			booking, _ := bookingmodel.GetByID(id)
			rooms, _ := roommodel.GetAll()
			var activeRooms []entities.Room
			for _, room := range rooms {
				if room.Status == "active" || room.ID == booking.Room.ID {
					activeRooms = append(activeRooms, room)
				}
			}

			data := PageData{
				Booking:  booking,
				Rooms:    activeRooms,
				Error:    strings.Join(errors, "<br>"),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/booking/crud/edit.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Create booking entity
		booking := entities.Booking{
			RoomID:       roomID,
			UserName:     formData["user_name"],
			UserPhone:    formData["user_phone"],
			CheckInDate:  checkInDate,
			CheckOutDate: checkOutDate,
			TotalPrice:   totalPrice,
		}

		// Update dengan validasi
		err, successMsg := bookingmodel.Update(booking, proofFile, id)

		if err != nil {
			// Jika error, tampilkan form kembali dengan data sebelumnya
			rooms, _ := roommodel.GetAll()
			var activeRooms []entities.Room
			for _, room := range rooms {
				if room.Status == "active" || room.ID == booking.RoomID {
					activeRooms = append(activeRooms, room)
				}
			}

			data := PageData{
				Rooms:    activeRooms,
				Error:    err.Error(),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/booking/crud/edit.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Jika sukses, redirect ke halaman utama dengan pesan success
		http.Redirect(w, r, "/bookings?success="+successMsg, http.StatusSeeOther)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// Hanya menerima method POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get ID from form
	idStr := r.FormValue("id")
	if idStr == "" {
		http.Redirect(w, r, "/bookings?error=ID+booking+tidak+valid", http.StatusSeeOther)
		return
	}

	// Convert ID to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/bookings?error=ID+booking+tidak+valid", http.StatusSeeOther)
		return
	}

	// Call delete function
	err, successMsg := bookingmodel.Delete(id)

	if err != nil {
		// Redirect dengan pesan error
		errorMsg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/bookings?error="+errorMsg, http.StatusSeeOther)
		return
	}

	// Redirect dengan pesan success
	successMsgEncoded := url.QueryEscape(successMsg)
	http.Redirect(w, r, "/bookings?success="+successMsgEncoded, http.StatusSeeOther)
}
