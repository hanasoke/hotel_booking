package homepage

import (
	"hotel_booking/controllers"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := controllers.LoadTemplate(
		"views/homepage/index.html",
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", nil)
}
