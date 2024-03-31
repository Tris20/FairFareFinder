package fffwebserver

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files/web_pages"
	"log"
	"net/http"
)

func SetupFFFWebServer() {
	// Handle starting the web server
	http.HandleFunc("/", fffwebpages.HomeHandler)
	http.HandleFunc("/forecast", fffwebpages.ForecastHandler)
	http.HandleFunc("/getforecast", fffwebpages.GetForecastHandler)
	http.HandleFunc("/berlin-flight-destinations", fffwebpages.PresentBerlinFlightDestinations)
	http.HandleFunc("/edinburgh-flight-destinations", fffwebpages.PresentEdinburghFlightDestinations)
	http.HandleFunc("/glasgow-flight-destinations", fffwebpages.PresentGlasgowFlightDestinations)
http.HandleFunc("/range", fffwebpages.HtmxPriceRange2)

  // Demo and Debug pages
  	http.HandleFunc("/htmx-price-range", fffwebpages.HtmxPriceRange)
	// Serve static files from the `images` directory
	fs := http.FileServer(http.Dir("src/images"))
	http.Handle("/images/", http.StripPrefix("/images/", fs))

	// Start the web server
	fmt.Println("Starting server on :6969")
	if err := http.ListenAndServe(":6969", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
