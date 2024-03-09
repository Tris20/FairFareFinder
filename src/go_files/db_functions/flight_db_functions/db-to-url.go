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

// executeQueries connects to the SQLite database and executes the SQL queries.
func executeQueries(db *sql.DB, queries []string) error {
	for _, query := range queries {
		fmt.Println("Executing query:", query)
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Iterate through the result set.
		for rows.Next() {
			var flight Flight
			if err := rows.Scan(&flight.Id, &flight.FlightNumber, &flight.DepartureAirport, &flight.ArrivalAirport, &flight.DepartureTime, &flight.ArrivalTime, &flight.Direction); err != nil {
				return err
			}
			fmt.Printf("Flight: %#v\n", flight)
		}

		// Check for errors from iterating over rows.
		if err = rows.Err(); err != nil {
			return err
		}
	}

	return nil
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
	queries := []string{
		"SELECT * FROM flights WHERE departureAirport = 'BER' AND departureTime BETWEEN '2024-03-08' AND '2024-03-09'",
		"SELECT * FROM flights WHERE arrivalAirport = 'BER' AND arrivalTime BETWEEN '2024-04-10' AND '2024-04-13'",
	}

	// Connect to the SQLite database.
	db, err := sql.Open("sqlite3", "./data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Execute the queries.
	if err := executeQueries(db, queries); err != nil {
		log.Fatal(err)
	}
}
