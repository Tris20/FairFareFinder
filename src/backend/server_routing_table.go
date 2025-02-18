package backend

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/Tris20/FairFareFinder/src/backend/dev_tools"
	"github.com/gorilla/sessions"
)

// SetupRoutes sets up all the HTTP routes for the application
func SetupRoutes(store *sessions.CookieStore, db *sql.DB, tmpl *template.Template) {
	// Application routes
	http.HandleFunc("/", IndexHandler)

	http.HandleFunc("/update-slider-price", UpdateSliderPriceHandler)

	// Static file routes
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./src/frontend/css/"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./src/frontend/images"))))
	http.Handle("/location-images/", http.StripPrefix("/location-images/", http.FileServer(http.Dir("./ignore/location-images"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./src/frontend/js/")))) // New JS route
	//Android
	http.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./manifest.json")
	})

	http.Handle("/.well-known/", http.StripPrefix("/.well-known/", http.FileServer(http.Dir("./src/frontend/.well-known"))))

	// API routes
	http.HandleFunc("/city-country-pairs", CityCountryHandler)

	// Privacy policy route
	http.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./src/frontend/html/privacy-policy.html") // Ensure the path is correct
	})

	// Dev tools routes
	http.HandleFunc("/all-cities", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		clientID := session.ID
		dev_tools.AllCitiesHandler(db, tmpl, clientID)(w, r)
	})
	http.HandleFunc("/load-more-cities", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		clientID := session.ID
		dev_tools.LoadMoreCities(tmpl, clientID)(w, r)
	})
}
