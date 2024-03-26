
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

type Flight struct {
	ID               int
	FlightNumber     string
	DepartureAirport string
	ArrivalAirport   string
	DepartureTime    string
	ArrivalTime      string
}

var (
	db            *sql.DB
	templatesPath = "templates"
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", "../../../../data/flights.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/", serveTemplate("index.html"))
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/airports", airportsHandler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveTemplate(fileName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"DepartureAirports": fetchAirports(`SELECT DISTINCT departureAirport FROM flights ORDER BY departureAirport`),
		}
		executeTemplate(w, fileName, data)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    departureAirport := r.URL.Query().Get("departureAirport")
    arrivalAirport := r.URL.Query().Get("arrivalAirport")
    departureDate := r.URL.Query().Get("departureDate")
    arrivalDate := r.URL.Query().Get("arrivalDate")

    flights := fetchFlights(departureAirport, arrivalAirport, departureDate, arrivalDate)
    executeTemplate(w, "results.html", flights)
}

func airportsHandler(w http.ResponseWriter, r *http.Request) {
	departureAirport := r.URL.Query().Get("departureAirport")
	options := fetchAirports(`SELECT DISTINCT arrivalAirport FROM flights WHERE departureAirport = ? ORDER BY arrivalAirport`, departureAirport)
	w.Header().Set("Content-Type", "text/html")
	for _, airport := range options {
		fmt.Fprintf(w, `<option value="%s">%s</option>`, airport, airport)
	}
}

func fetchAirports(query string, args ...interface{}) []string {
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil
	}
	defer rows.Close()

	var airports []string
	for rows.Next() {
		var airport string
		if err := rows.Scan(&airport); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		airports = append(airports, airport)
	}
	return airports
}

func fetchFlights(departureAirport, arrivalAirport, departureDate, arrivalDate string) []Flight {
    query := "SELECT ID, FlightNumber, DepartureAirport, ArrivalAirport, DepartureTime, ArrivalTime FROM flights WHERE 1=1"
    args := []interface{}{}

    if departureAirport != "" {
        query += " AND DepartureAirport = ?"
        args = append(args, departureAirport)
    }
    if arrivalAirport != "" {
        query += " AND ArrivalAirport = ?"
        args = append(args, arrivalAirport)
    }
    if departureDate != "" && arrivalDate != "" {
        query += " AND DepartureTime BETWEEN ? AND ?"
        args = append(args, departureDate, arrivalDate)
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        log.Printf("Failed to execute query: %v", err)
        return nil
    }
    defer rows.Close()

    var flights []Flight
    for rows.Next() {
        var f Flight
        if err := rows.Scan(&f.ID, &f.FlightNumber, &f.DepartureAirport, &f.ArrivalAirport, &f.DepartureTime, &f.ArrivalTime); err != nil {
            log.Printf("Failed to scan row: %v", err)
            continue
        }
        flights = append(flights, f)
    }
    return flights
}
func executeTemplate(w http.ResponseWriter, fileName string, data interface{}) {
	tmpl, err := template.ParseFiles(filepath.Join(templatesPath, fileName))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Printf("Error parsing template: %v", err)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", 500)
		log.Printf("Error executing template: %v", err)
	}
}

