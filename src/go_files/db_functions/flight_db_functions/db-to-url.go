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







