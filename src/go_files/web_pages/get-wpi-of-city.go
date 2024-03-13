package fffwebpages

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files/weather_pleasantness"
	"io/ioutil"
	"log"
	"net/http"
)

func GetForecastHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Handling request to /getforecast")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// fmt.Println("Handling POST request")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}
	city := r.FormValue("city")
	// fmt.Println("City:", city)

	// Call the processLocation function
	wpi, _ := weather_pleasantry.ProcessLocation(city)

	response := fmt.Sprintf("The Weather Pleasantness Index (WPI) for %s is %.2f", city, wpi)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, response)
}

// handles requests to the forecast page
func ForecastHandler(w http.ResponseWriter, r *http.Request) {
	// serving a static file
	pageContent, err := ioutil.ReadFile("src/html/forecast.html")
	if err != nil {
		log.Printf("Error reading forecast page file: %v", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(pageContent)
}
