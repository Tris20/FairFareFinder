package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Flight represents a flight record in the database
type Flight struct {
	ID               int
	FlightNumber     string
	DepartureAirport string
	ArrivalAirport   string
	DepartureTime    string
	ArrivalTime      string
	Direction        string
}

var (
	db            *sql.DB
	templatesPath = "templates"
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", "../../../data/flights.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
    
  // Serve static files from the "css" directory
    fs := http.FileServer(http.Dir("css"))
    http.Handle("/css/", http.StripPrefix("/css/", fs))

    log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(filepath.Join(templatesPath, "index.html"))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Println(err)
		return
	}
	tmpl.Execute(w, nil)
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse form inputs
	departureAirport := r.URL.Query().Get("departureAirport")
	arrivalAirport := r.URL.Query().Get("arrivalAirport")
	// For simplicity, dates will be handled as strings. In a real application, consider using proper date handling.
	departureDate := r.URL.Query().Get("departureDate")
	arrivalDate := r.URL.Query().Get("arrivalDate")

	// Create the base SQL query
	query := "SELECT * FROM flights WHERE 1=1"
	args := []interface{}{}

	// Dynamically build the query based on input
	if departureAirport != "" {
		query += " AND departureAirport = ?"
		args = append(args, departureAirport)
	}
	if arrivalAirport != "" {
		query += " AND arrivalAirport = ?"
		args = append(args, arrivalAirport)
	}
	if departureDate != "" && arrivalDate != "" {
		query += " AND departureTime BETWEEN ? AND ?"
		args = append(args, departureDate, arrivalDate)
	}
log.Println("Executing query:", query, args)

	rows, err := db.Query(query, args...)
  
  if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Println("Failed to execute query:", err)
		return
	}
	defer rows.Close()
  w.Header().Set("Content-Type", "text/html")
	flights := make([]Flight, 0)
	for rows.Next() {
		var f Flight

		if err := rows.Scan(&f.ID, &f.FlightNumber, &f.DepartureAirport, &f.ArrivalAirport, &f.DepartureTime, &f.ArrivalTime, &f.Direction); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			log.Println("Failed to scan row:", err)
			return
		}

fmt.Printf("Flight: %s, Departure: %s, Arrival: %s\n", f.FlightNumber, f.DepartureAirport, f.ArrivalAirport)
		flights = append(flights, f)
	}
  //fmt.Println(flights)
	// Render the results back to the client
  path := filepath.Join(templatesPath, "results.html")
  fmt.Println(path)
  tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Println(err)
		return
	}
    // Execute the template, writing the generated HTML directly to the response writer
    err = tmpl.Execute(w, flights)
    if err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Internal Server Error", 500)
        return
    }
}

