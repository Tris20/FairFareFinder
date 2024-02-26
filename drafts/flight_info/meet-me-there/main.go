package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Flight struct {
	ID               int    `json:"id"`
	FlightNumber     string `json:"flightNumber"`
	DepartureAirport string `json:"departureAirport"`
	ArrivalAirport   string `json:"arrivalAirport"`
	DepartureTime    string `json:"departureTime"`
	ArrivalTime      string `json:"arrivalTime"`
	Direction        string `json:"direction"`
}

func main() {
	var departureDatesString, arrivalDatesString, iataCodesString string
	flag.StringVar(&departureDatesString, "departureDates", "", "List of departure dates (comma-separated)")
	flag.StringVar(&arrivalDatesString, "arrivalDates", "", "List of arrival dates (comma-separated)")
	flag.StringVar(&iataCodesString, "iataCodes", "", "List of IATA codes (comma-separated)")
	flag.Parse()

	departureDates := strings.Split(departureDatesString, ",")
	arrivalDates := strings.Split(arrivalDatesString, ",")
	iataCodes := strings.Split(iataCodesString, ",")

	db, err := sql.Open("sqlite3", "./input/flights.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	allFlights := make([]Flight, 0)

	// Handle departures
	for _, departureDate := range departureDates {
		for _, code := range iataCodes {
			query := `SELECT * FROM flights WHERE DATE(departureTime) = ? AND departureAirport = ?`
			processFlights(db, &allFlights, query, departureDate, code, true, iataCodes)
		}
	}

	// Handle arrivals
	for _, arrivalDate := range arrivalDates {
		for _, code := range iataCodes {
			query := `SELECT * FROM flights WHERE DATE(arrivalTime) = ? AND arrivalAirport = ?`
			processFlights(db, &allFlights, query, arrivalDate, code, false, iataCodes)
		}
	}

	// New code to collect and deduplicate airports
	airportMap := make(map[string]struct{}) // Using a map to ensure uniqueness
	for _, flight := range allFlights {
		airportMap[flight.DepartureAirport] = struct{}{}
		airportMap[flight.ArrivalAirport] = struct{}{}
	}

	// Convert the map keys to a slice
	uniqueAirports := make([]string, 0, len(airportMap))
	for airport := range airportMap {
		uniqueAirports = append(uniqueAirports, airport)
	}

	// Creating a combined output structure
	output := struct {
		Flights      []Flight `json:"flights"`
		Destinations []string `json:"destinations"`
	}{
		Flights:      allFlights,
		Destinations: uniqueAirports,
	}

	// Serialize the combined output to JSON
	outputJSON, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	// Save to file
	if err := os.WriteFile("flights.json", outputJSON, 0644); err != nil {
		panic(err)
	}
	fmt.Println("All flights data saved to flights.json")
}

func processFlights(db *sql.DB, allFlights *[]Flight, query string, date string, code string, isDeparture bool, iataCodes []string) {
	rows, err := db.Query(query, date, code)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	tempFlights := make([]Flight, 0)
	for rows.Next() {
		var flight Flight
		if err := rows.Scan(&flight.ID, &flight.FlightNumber, &flight.DepartureAirport, &flight.ArrivalAirport, &flight.DepartureTime, &flight.ArrivalTime, &flight.Direction); err != nil {
			panic(err)
		}
		tempFlights = append(tempFlights, flight)
	}

	for _, flight := range tempFlights {
		valid := true
		for _, checkCode := range iataCodes {
			var checkQuery string
			if isDeparture {
				checkQuery = `SELECT COUNT(*) FROM flights WHERE arrivalAirport = ? AND departureAirport = ?`
			} else {
				checkQuery = `SELECT COUNT(*) FROM flights WHERE departureAirport = ? AND arrivalAirport = ?`
			}
			var count int
			err := db.QueryRow(checkQuery, flight.ArrivalAirport, checkCode).Scan(&count)
			if err != nil || count == 0 {
				valid = false
				break
			}
		}
		if valid {
			*allFlights = append(*allFlights, flight)
		}
	}
}
