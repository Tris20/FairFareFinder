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
    iataCodes := strings.Split(iataCodesString, ",")

    db, err := sql.Open("sqlite3", "./input/flights.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    refinedFlights := make([]Flight, 0)

    // Step 1: Fetch initial list of flights based on departure dates and IATA codes
    for _, departureDate := range departureDates {
        for _, code := range iataCodes {
            query := `SELECT * FROM flights WHERE DATE(departureTime) = ? AND departureAirport = ?`
            rows, err := db.Query(query, departureDate, code)
            if err != nil {
                panic(err)
            }
            defer rows.Close()

            flights := make([]Flight, 0)
            for rows.Next() {
                var flight Flight
                if err := rows.Scan(&flight.ID, &flight.FlightNumber, &flight.DepartureAirport, &flight.ArrivalAirport, &flight.DepartureTime, &flight.ArrivalTime, &flight.Direction); err != nil {
                    panic(err)
                }
                flights = append(flights, flight)
            }

            // Step 2: Refine the list for each flight
            for _, flight := range flights {
                valid := true
                for _, checkCode := range iataCodes {
                    checkQuery := `SELECT COUNT(*) FROM flights WHERE departureAirport = ? AND arrivalAirport = ?`
                    var count int
                    err := db.QueryRow(checkQuery, flight.ArrivalAirport, checkCode).Scan(&count)
                    if err != nil || count == 0 {
                        valid = false
                        break
                    }
                }
                if valid {
                    refinedFlights = append(refinedFlights, flight)
                }
            }
        }
    }

    // Serialize the refined flights to JSON
    flightsJSON, err := json.MarshalIndent(refinedFlights, "", "    ")
    if err != nil {
        panic(err)
    }

    // Save to file
    if err := os.WriteFile("flights.json", flightsJSON, 0644); err != nil {
        panic(err)
    }

    fmt.Println("Refined flights data saved to flights.json")
}

