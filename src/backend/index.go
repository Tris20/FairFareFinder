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
	// Ensure cityCountryPairs is loaded
	LoadCityCountryPairs(db) // sync.Once ensures it only runs once

	// Pass city-country pairs and backend constants to the template
	err := tmpl.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"CityCountryPairs":  GetCityCountryPairs(), // Use a getter for consistency
		"MinFlightPrice":    MinFlightPrice,
		"MidFlightPrice":    MidFlightPrice,
		"MaxFlightPrice":    MaxFlightPrice,
		"MinAccomPrice":     MinAccomPrice,
		"MidAccomPrice":     MidAccomPrice,
		"MaxAccomPrice":     MaxAccomPrice,
		"DefaultAccomPrice": DefaultAccomPrice,
		"DefaultSortOption": DefaultSortOption,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
