package backend

import (
	"encoding/json"
	"html/template"
	"time"
)

// getDayOfWeek returns the three-letter abbreviation of the day for today plus the given offset.
// For example, an offset of 0 returns today's day (e.g., "Mon"), 1 returns tomorrow's, etc.
func getDayOfWeek(offset int) string {
	now := time.Now()
	forecastDate := now.AddDate(0, 0, offset)
	return forecastDate.Format("Mon")
}

// InitializeTemplates initializes the templates and returns a pointer to the template.
func InitializeTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"mod": Mod,
		"add": Add,
		"toJson": func(v interface{}) (string, error) {
			a, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(a), nil
		},
		"getDayOfWeek": getDayOfWeek,
	}

	// Parse all required templates
	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		"./src/frontend/html/index.html",
		"./src/frontend/html/table.html",
		"./src/frontend/html/seo.html",
		"./src/frontend/html/dev_and_debug/cities.html",
		"./src/frontend/html/dev_and_debug/all-cities.html",
	)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
