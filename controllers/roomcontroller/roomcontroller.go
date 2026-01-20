package roomcontroller

import (
	"hotel_booking/controllers"
	"hotel_booking/entities"
	"hotel_booking/models/roommodel"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

// Struct untuk passing data ke template
type PageData struct {
	Rooms    []entities.Room
	Room     entities.Room // Tambahkan ini untuk detail room
	Error    string
	Success  string
	FormData map[string]string
}

// Helper function untuk validasi integer
func isValidInt(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

// Helper function untuk mengekstrak form data
func extractFormData(r *http.Request) map[string]string {
	return map[string]string{
		"name":          strings.TrimSpace(r.FormValue("name")),
		"type":          strings.TrimSpace(r.FormValue("type")),
		"description":   strings.TrimSpace(r.FormValue("description")),
		"capacity":      strings.TrimSpace(r.FormValue("capacity")),
		"status":        strings.TrimSpace(r.FormValue("status")),
		"price_per_day": strings.TrimSpace(r.FormValue("price_per_day")),
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	// Get success message from session/query parameter
	successMsg := r.URL.Query().Get("success")

	// Get all rooms
	rooms, err := roommodel.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Rooms:   rooms,
		Success: successMsg,
	}

	tmpl, err := controllers.LoadTemplate(
		"views/room/index.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", data)
}

func Add(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := controllers.LoadTemplate(
			"views/room/crud/add.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Pass empty form data
		data := PageData{
			FormData: make(map[string]string),
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
		name := r.FormValue("name")
		roomType := r.FormValue("type")
		description := r.FormValue("description")
		capacity, _ := strconv.Atoi(r.FormValue("capacity"))
		status := r.FormValue("status")
		pricePerDay, _ := strconv.Atoi(r.FormValue("price_per_day"))

		// Get uploaded file
		file, header, err := r.FormFile("image")
		var imageFile *multipart.FileHeader
		if err == nil {
			imageFile = header
			file.Close()
		}

		// Create room entity
		room := entities.Room{
			Name:        name,
			Type:        roomType,
			Description: description,
			Capacity:    capacity,
			Status:      status,
			PricePerDay: pricePerDay,
		}

		// Insert dengan validasi
		err, successMsg := roommodel.Insert(room, imageFile)

		if err != nil {
			// Jika error, tampilkan form kembali dengan data sebelumnya
			formData := map[string]string{
				"name":          name,
				"type":          roomType,
				"description":   description,
				"capacity":      r.FormValue("capacity"),
				"status":        status,
				"price_per_day": r.FormValue("price_per_day"),
			}

			data := PageData{
				Error:    err.Error(),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/room/crud/add.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Jika sukses, redirect ke halaman utama dengan pesan success
		http.Redirect(w, r, "/rooms?success="+successMsg, http.StatusSeeOther)
	}
}

func Edit(w http.ResponseWriter, r *http.Request) {
	// GET Method: Tampilkan form edit
	if r.Method == "GET" {
		// Get room ID from query parameter
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "ID ruangan tidak ditemukan", http.StatusBadRequest)
			return
		}

		// Convert ID to integer
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "ID ruangan tidak valid", http.StatusBadRequest)
			return
		}

		// Get room data from database
		room, err := roommodel.GetById(id)
		if err != nil {
			if err.Error() == "Ruangan tidak ditemukan" {
				http.Error(w, "Ruangan tidak ditemukan", http.StatusNotFound)
				return
			}
			http.Error(w, "Gagal mengambil data ruangan: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		// Prepare template data
		data := PageData{
			Room: room,
		}

		// Load and execute template
		tmpl, err := controllers.LoadTemplate(
			"views/room/crud/edit.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base.html", data)
		return
	}

	// POST Method: Proses update
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

		if formData["name"] == "" {
			errors = append(errors, "Nama ruangan wajib diisi")
		}
		if formData["type"] == "" {
			errors = append(errors, "Tipe ruangan wajib diisi")
		}
		if formData["description"] == "" {
			errors = append(errors, "Deskripsi wajib diisi")
		}
		if !isValidInt(formData["capacity"]) {
			errors = append(errors, "Kapasitas harus berupa angka")
		} else if capacity, _ := strconv.Atoi(formData["capacity"]); capacity <= 0 {
			errors = append(errors, "Kapasitas harus lebih dari 0")
		}
		if formData["status"] == "" {
			errors = append(errors, "Status wajib dipilih")
		}
		if !isValidInt(formData["price_per_day"]) {
			errors = append(errors, "Harga per hari harus berupa angka")
		} else if price, _ := strconv.Atoi(formData["price_per_day"]); price <= 0 {
			errors = append(errors, "Harga per hari harus lebih dari 0")
		}

		// Get room ID
		idStr := r.FormValue("id")
		if idStr == "" {
			errors = append(errors, "ID ruangan tidak ditemukan")
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			errors = append(errors, "ID ruangan tidak valid")
		}

		// Jika ada error validasi
		if len(errors) > 0 {
			// Get room data for form
			room, _ := roommodel.GetById(id)

			data := PageData{
				Room:     room,
				Error:    strings.Join(errors, "<br>"),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/room/crud/edit.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Convert string to int
		capacity, _ := strconv.Atoi(formData["capacity"])
		pricePerDay, _ := strconv.Atoi(formData["price_per_day"])

		// Get uploaded file
		var imageFile *multipart.FileHeader
		file, header, err := r.FormFile("image")
		if err == nil {
			imageFile = header
			file.Close()
		}

		// Create room entity
		room := entities.Room{
			ID:          id,
			Name:        formData["name"],
			Type:        formData["type"],
			Description: formData["description"],
			Capacity:    capacity,
			Status:      formData["status"],
			PricePerDay: pricePerDay,
		}

		// Get old image path
		oldRoom, _ := roommodel.GetById(id)
		oldImage := oldRoom.Image

		// Update dengan validasi
		err, successMsg := roommodel.Update(room, imageFile, oldImage)

		if err != nil {
			// Jika error, tampilkan form kembali dengan data sebelumnya
			data := PageData{
				Room:     oldRoom,
				Error:    err.Error(),
				FormData: formData,
			}

			tmpl, tmplErr := controllers.LoadTemplate(
				"views/room/crud/edit.html",
			)
			if tmplErr != nil {
				http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.ExecuteTemplate(w, "base.html", data)
			return
		}

		// Jika sukses, redirect ke halaman detail dengan pesan success
		http.Redirect(w, r, "/detail_room?id="+idStr+"&success="+successMsg, http.StatusSeeOther)
	}
}

func Detail(w http.ResponseWriter, r *http.Request) {
	// Get room ID from query parameter
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID ruangan tidak ditemukan", http.StatusBadRequest)
		return
	}

	// Convert ID to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID ruangan tidak valid", http.StatusBadRequest)
		return
	}

	// Get room data from database
	room, err := roommodel.GetById(id)
	if err != nil {
		// Jika room tidak ditemukan
		if err.Error() == "Ruangan tidak ditemukan" {
			http.Error(w, "Ruangan tidak ditemukan", http.StatusNotFound)
			return
		}

		http.Error(w, "Gagal mengambil data ruangan: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := PageData{
		Room: room,
	}

	// Load and execute template
	tmpl, err := controllers.LoadTemplate(
		"views/room/crud/detail.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", data)
}
