package controllers

import "html/template"

func LoadTemplate(files ...string) (*template.Template, error) {
	base := []string{"views/templates/base.html"}
	files = append(base, files...)
	return template.ParseFiles(files...)
}
