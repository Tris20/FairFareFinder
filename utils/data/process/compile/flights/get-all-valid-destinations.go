package flightdb

import (
	"database/sql"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend" //import types
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// FlightRequest represents the structure of the YAML input.
type FlightRequest struct {
	Flights []FlightCriteria `yaml:"flights"`
}

// FlightCriteria represents each flight's criteria within the YAML input.
type FlightCriteria struct {
	Direction string `yaml:"direction"`
	Airport   string `yaml:"airport"`
	StartDate string `yaml:"startDate"`
	EndDate   string `yaml:"endDate"`
}

// Flight represents a row from the flights table.
type Flight struct {
	Id               int
	FlightNumber     string
	DepartureAirport string
	ArrivalAirport   string
	DepartureTime    string
	ArrivalTime      string
	Direction        string
}

// executeQueryForAirports executes a given SQL query and returns a set of airports.
func executeQueryForAirports(db *sql.DB, query string) (map[string]bool, error) {
	airports := make(map[string]bool)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var airport string
		if err := rows.Scan(&airport); err != nil {
			return nil, err
		}
		airports[airport] = true
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

// intersectSets finds the intersection of an array of sets.
func intersectSets(sets []map[string]bool) []string {
	intersection := make([]string, 0)
	if len(sets) == 0 {
		return intersection
	}

	// Initialize intersection with the first set's elements.
	for item := range sets[0] {
		intersection = append(intersection, item)
	}

	// Intersect with remaining sets.
	for _, set := range sets[1:] {
		temp := intersection[:0] // reuse the existing slice but start filling from the beginning
		for _, item := range intersection {
			if set[item] {
				temp = append(temp, item)
			}
		}
		intersection = temp
	}

	return intersection
}

func DetermineFlightsFromConfig(origin model.OriginInfo) []model.DestinationInfo {

	// Assuming the YAML to SQL query conversion is done elsewhere and we have the queries ready.
	// Connect to the SQLite database.
	db, err := sql.Open("sqlite3", "data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Define your queries here.
	queries := []string{
		fmt.Sprintf("SELECT arrivalAirport FROM flights WHERE departureAirport = '%s' AND departureTime BETWEEN '%s' AND '%s'", origin.IATA, origin.DepartureStartDate, origin.DepartureEndDate),
		fmt.Sprintf("SELECT departureAirport FROM flights WHERE arrivalAirport = '%s' AND arrivalTime BETWEEN '%s' AND '%s'", origin.IATA, origin.ArrivalStartDate, origin.ArrivalEndDate),
		//"SELECT arrivalAirport FROM flights WHERE departureAirport = 'EDI' AND departureTime BETWEEN '2024-03-20' AND '2024-03-22'",
		//"SELECT departureAirport FROM flights WHERE arrivalAirport = 'EDI' AND arrivalTime BETWEEN '2024-03-24' AND '2024-03-26'",

		// Add your third, fourth, ... queries here.
	}

	// Execute all queries and collect their results in a slice of sets.
	var sets []map[string]bool
	for _, query := range queries {
		//fmt.Printf(query)
		airports, err := executeQueryForAirports(db, query)
		// fmt.Printf("AIRPOTS %s", airports)
		if err != nil {
			log.Fatal("Error executing query:", err)
		}
		sets = append(sets, airports)
	}

	// Find the intersection of all sets.
	intersection := intersectSets(sets)

	fmt.Println("Airports meeting all conditions:", intersection)
	airportDetailsList := buildAirportDetails(db, intersection)
	for _, airportInfo := range airportDetailsList {
		fmt.Printf("%s: %s, %s\n", airportInfo.IATA, airportInfo.City, airportInfo.Country)
	}
	return airportDetailsList
}

// printAirportDetails prints the details for each airport IATA code.
func buildAirportDetails(db *sql.DB, iataCodes []string) []model.DestinationInfo {
	var airportDetailsList []model.DestinationInfo
	for _, iata := range iataCodes {
		//skip empty or blank IATA codes
		if iata == "" {
			continue
		}
		city, country, skyscannerid, err := fetchAirportDetails(db, iata)
		if err != nil {
			log.Printf("Error fetching details for IATA %s: %v", iata, err)
			continue
		}
		// Append the fetched details to the list
		airportDetailsList = append(airportDetailsList, model.DestinationInfo{
			IATA:         iata,
			City:         city,
			Country:      country,
			SkyScannerID: skyscannerid,
		})
	}
	return airportDetailsList
}

// fetchAirportDetails executes a query to fetch city and country for a given IATA code.
func fetchAirportDetails(db *sql.DB, iataCode string) (string, string, string, error) {
	var city, country, skyscannerid string
	query := "SELECT city, country, skyscannerid FROM airport_info WHERE iata = ?"
	err := db.QueryRow(query, iataCode).Scan(&city, &country, &skyscannerid)
	if err != nil {
		return "", "", "", err
	}
	return city, country, skyscannerid, nil
}
