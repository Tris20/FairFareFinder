package flightdb

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

type Airport struct {
	IATACode string
	CityName string
}


func GetCitiesAndIATACodes() {

  DetermineFlightsFromConfig() 

  // Example list of IATA codes
	iataCodes := []string{"CDG", "KOI", "AYT", "BRS", "LSI", "DXB", "KEF", "STN", "BEB", "FCO", "SYY", "PRG", "FRA", "AMS", "ALC", "FAO", "ILY", "BER", "DUB", "LGW", "BCN", "BRR", "BHX", "TRE", "LPA", "LHR", "LTN", "TFS", "ACE", "BFS", "SOU", "BHD", "LCY", "LIS", "ATH", "CRL", "VCE", "FUE", "GVA", "NQY", "CGN", "MAD", "IST", "DOH", "LYS", "BGY", "EXT", "MUC", "ARN", "ORK", "CWL", "OSL", "EWR", "KRK", "MXP", "RTM", "HEL"}
  fmt.Println("Opening Flights")
	// Open the SQLite database
	db, err := sql.Open("sqlite3", "data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Prepare the query statement for performance
	stmt, err := db.Prepare("SELECT city FROM airport_info WHERE iata = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	var airports []Airport

	for _, code := range iataCodes {
		var cityName string
		err := stmt.QueryRow(code).Scan(&cityName)
		if err != nil {
			log.Printf("Failed to get city for IATA code %s: %v", code, err)
			continue // Skip to the next code
		}
		airports = append(airports, Airport{IATACode: code, CityName: cityName})
	}

	// Print the result
	for _, airport := range airports {
		fmt.Printf("IATA Code: %s, City Name: %s\n", airport.IATACode, airport.CityName)
	}
}



/* GET IATIA */





// FlightRequest represents the structure of the YAML input.
type FlightRequest struct {
	Flights []FlightCriteria `yaml:"flights"`
}

// FlightCriteria represents each flight's criteria within the YAML input.
type FlightCriteria struct {
	Direction  string `yaml:"direction"`
	Airport    string `yaml:"airport"`
	StartDate  string `yaml:"startDate"`
	EndDate    string `yaml:"endDate"`
}

// Flight represents a row from the flights table.
type Flight struct {
	Id              int
	FlightNumber    string
	DepartureAirport string
	ArrivalAirport  string
	DepartureTime   string
	ArrivalTime     string
	Direction       string
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



func DetermineFlightsFromConfig() {
	// Example YAML input.

/*
  yamlInput := []byte(`
flights:
  - direction: "Departure"
    airport: "BER"
    startDate: "2024-03-08"
    endDate: "2024-03-09"
  - direction: "Arrival"
    airport: "BER"
    startDate: "2024-04-10"
    endDate: "2024-04-13"
`)

*/
	// Assuming the YAML to SQL query conversion is done elsewhere and we have the queries ready.
// Connect to the SQLite database.
	db, err := sql.Open("sqlite3", "data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

// Define your queries here.
	queries := []string{
		"SELECT arrivalAirport FROM flights WHERE departureAirport = 'BER' AND departureTime BETWEEN '2024-03-20' AND '2024-03-22'",
		"SELECT departureAirport FROM flights WHERE arrivalAirport = 'BER' AND arrivalTime BETWEEN '2024-03-24' AND '2024-03-26'",
		"SELECT arrivalAirport FROM flights WHERE departureAirport = 'GLA' AND departureTime BETWEEN '2024-03-20' AND '2024-03-22'",
    "SELECT departureAirport FROM flights WHERE arrivalAirport = 'GLA' AND arrivalTime BETWEEN '2024-03-24' AND '2024-03-26'",

    // Add your third, fourth, ... queries here.
	}

	// Execute all queries and collect their results in a slice of sets.
	var sets []map[string]bool
	for _, query := range queries {
		airports, err := executeQueryForAirports(db, query)
		if err != nil {
			log.Fatal("Error executing query:", err)
		}
		sets = append(sets, airports)
	}

	// Find the intersection of all sets.
	intersection := intersectSets(sets)

	fmt.Println("Airports meeting all conditions:", intersection)
}
