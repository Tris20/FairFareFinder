package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

type Airport struct {
	IATACode    string `json:"iata_code"`
	Municipality string `json:"municipality"`
}

type Destination struct {
	IATACode    string `json:"iata"`
	Municipality string `json:"municipality"`
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

	for _, departureDate := range departureDates {
		for _, code := range iataCodes {
			query := `SELECT * FROM flights WHERE DATE(departureTime) = ? AND departureAirport = ?`
			processFlights(db, &allFlights, query, departureDate, code, true, iataCodes)
		}
	}

	for _, arrivalDate := range arrivalDates {
		for _, code := range iataCodes {
			query := `SELECT * FROM flights WHERE DATE(arrivalTime) = ? AND arrivalAirport = ?`
			processFlights(db, &allFlights, query, arrivalDate, code, false, iataCodes)
		}
	}


	// Load and parse the airports.json file
	airportsJSON, err := ioutil.ReadFile("input/airports.json")
	if err != nil {
		panic(err)
	}
	var airports []Airport
	if err := json.Unmarshal(airportsJSON, &airports); err != nil {
		panic(err)
	}

	// Map IATA codes to municipalities
	iataToMunicipality := make(map[string]string)
	for _, airport := range airports {
		if airport.IATACode != "" { // Ensure we have an IATA code
			iataToMunicipality[airport.IATACode] = airport.Municipality
		}
	}

	// Create a map to hold unique destinations based on municipality to ensure no duplicates
	uniqueDestinations := make(map[string]Destination)
	for _, flight := range allFlights {
		municipality, exists := iataToMunicipality[flight.ArrivalAirport]
		if !exists {
			municipality = "Unknown" // Use "Unknown" for airports not found in the JSON file
		}
		// Use municipality as the key to ensure uniqueness
		uniqueDestinations[municipality] = Destination{
			IATACode:    flight.ArrivalAirport,
			Municipality: municipality,
		}
	}

	// Convert the map to a slice for the final JSON output
	destinations := make([]Destination, 0, len(uniqueDestinations))
	for _, dest := range uniqueDestinations {
		destinations = append(destinations, dest)
	}

	// Output structure now includes destinations with separate iata and municipality
	output := struct {
		Flights      []Flight      `json:"flights"`
		Destinations []Destination `json:"destinations"`
	}{
		Flights:      allFlights,
		Destinations: destinations,
	}

	// Serialize and save the output to a JSON file
	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("output/flights_with_destinations.json", outputJSON, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Output saved to output/flights_with_destinations.json")
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

