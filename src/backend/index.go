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

// IndexHandler serves the home page
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Check if cityCountryPairs is already loaded in memory
	if len(cityCountryPairs) == 0 {
		log.Println("City-country pairs are not loaded; loading from database.")
		LoadCityCountryPairs(db) // Ensure cityCountryPairs is populated
	}

	// Pass the city-country pairs to the template
	if err := tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"CityCountryPairs": cityCountryPairs,
	}); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
