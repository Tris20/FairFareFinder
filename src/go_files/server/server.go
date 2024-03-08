package fffwebserver


import (
   "github.com/Tris20/FairFareFinder/src/go_files/web_pages"
   "net/http"
   "fmt"
   "log"
)


func SetupFFFWebServer(){
// Handle starting the web server
		http.HandleFunc("/", fffwebpages.HomeHandler)
//		http.HandleFunc("/forecast", forecastHandler)
//		http.HandleFunc("/getforecast", getForecastHandler)
		http.HandleFunc("/berlin-flight-destinations", fffwebpages.PresentBerlinFlightDestinations)
		// Serve static files from the `images` directory
		fs := http.FileServer(http.Dir("src/images"))
		http.Handle("/images/", http.StripPrefix("/images/", fs))

		// Start the web server
		fmt.Println("Starting server on :6969")
		if err := http.ListenAndServe(":6969", nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
  }






