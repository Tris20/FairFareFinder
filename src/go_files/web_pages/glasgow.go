package fffwebpages

import (
	"io/ioutil"
	"log"
	"net/http"
)

// handles requests to the forecast page
func PresentGlasgowFlightDestinations(w http.ResponseWriter, r *http.Request) {
	// serving a static file
	pageContent, err := ioutil.ReadFile("src/html/glasgow-flight-destinations.html")
	if err != nil {
		log.Printf("Error reading forecast page file: %v", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(pageContent)
}
