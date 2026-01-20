package controllers

import (
	"html/template"
	"strconv"
)

func LoadTemplate(files ...string) (*template.Template, error) {
	base := []string{
		"views/templates/base.html",
	}

	files = append(base, files...)

	// Add template functions
	funcMap := template.FuncMap{
		"formatNumber": func(num int) string {
			s := strconv.Itoa(num)
			var result string
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += "."
				}
				result += string(c)
			}
			return result
		},
		"add": func(a, b int) int {
			return a + b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	return template.New("").Funcs(funcMap).ParseFiles(files...)
}
