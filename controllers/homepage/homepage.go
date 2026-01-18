package homepage

import (
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(
		"views/templates/base.html",
		"views/homepage/index.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base.html", nil)
}
