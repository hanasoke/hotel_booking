package controllers

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
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
		"toString": func(num int) string {
			return strconv.Itoa(num)
		},
		"formatDate": func(date time.Time) string {
			return date.Format("Monday, 2 January 2006")
		},
		"formatDateShort": func(date time.Time) string {
			return date.Format("02 Jan 2006")
		},
		"formatPhone": func(phone string) string {
			if strings.HasPrefix(phone, "+62") {
				return "0" + phone[3:]
			}
			return phone
		},
		"getDurationDays": func(checkIn, checkOut time.Time) string {
			if checkIn.IsZero() || checkOut.IsZero() {
				return "0 hari"
			}

			// Hitung selisih dalam hari
			duration := checkOut.Sub(checkIn)
			hours := duration.Hours()
			days := int(hours / 24)

			// Jika ada sisa jam, tambah 1 hari
			if hours > float64(days*24) {
				days++
			}

			// Minimal 1 hari
			if days < 1 {
				days = 1
			}

			return fmt.Sprintf("%d hari", days)
		},
	}

	return template.New("").Funcs(funcMap).ParseFiles(files...)
}
