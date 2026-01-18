package roomcontroller

import (
	"hotel_booking/controllers"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := controllers.LoadTemplate(
		"views/room/index.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", nil)
}

func Add(w http.ResponseWriter, r *http.Request) {

	tmpl, err := controllers.LoadTemplate(
		"views/room/crud/add.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", nil)
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
