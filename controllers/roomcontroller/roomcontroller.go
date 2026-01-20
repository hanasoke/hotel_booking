package roomcontroller

import (
	"hotel_booking/controllers"
	"hotel_booking/entities"
	"hotel_booking/models/roommodel"
	"mime/multipart"
	"net/http"
	"strconv"
)

// Struct untuk passing data ke template
type PageData struct {
	Rooms    []entities.Room
	Room     entities.Room // Tambahkan ini untuk detail room
	Error    string
	Success  string
	FormData map[string]string
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
	tmpl, err := controllers.LoadTemplate(
		"views/room/crud/edit.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", nil)
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
