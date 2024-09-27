
package backend

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

// Declare the required variables (db and tmpl) to be accessible from this package
var db *sql.DB
var tmpl *template.Template

// Set the database and templates
func Init(dbConn *sql.DB, templates *template.Template) {
	db = dbConn
	tmpl = templates
}

// IndexHandler serves the home page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Query to fetch distinct origin city names
	rows, err := db.Query("SELECT DISTINCT origin_city_name FROM flight")
	if err != nil {
		log.Printf("Error querying cities: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			log.Printf("Error scanning city: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		cities = append(cities, city)
	}

	// Pass cities to template
	if err := tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Cities": cities,
	}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
