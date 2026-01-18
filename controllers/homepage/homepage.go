package homepage

import (
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("views/homepage/index.html")
	if err != nil {
		panic(err)
	}

	temp.Execute(w, nil)
}
