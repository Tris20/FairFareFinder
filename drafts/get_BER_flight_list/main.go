
package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "time"
)

type FlightResponse struct {
    Data struct {
        Items []Flight `json:"items"`
    } `json:"data"`
}

type Flight struct {
    ArrivalDeparture      bool   `json:"is_departure"`
    OriginCity            string `json:"dep_airport_name"`
    OriginIATACode        string `json:"dep_airport_iata"`
    DestinationIATACode   string `json:"arr_airport_iata"`
    ScheduledDepartureTime string `json:"scheduled_time"`
}

func main() {
    // Get today's date
    today := time.Now()
    // Get the date 7 days from today
    sevenDaysLater := today.Add(7 * 24 * time.Hour)

    // Replace the dates in the URL for departures (Thursday and Friday)
    departureURL := "https://ber.berlin-airport.de//api.flights.json?arrivalDeparture=D" +
        "&dateFrom=" + today.Format("2006-01-02") +
        "&dateUntil=" + sevenDaysLater.Format("2006-01-02") +
        "&search=&lang=en&page=1&terminal=&itemsPerPage=1600"

    // Make an HTTP GET request for departures
    departureResponse, err := http.Get(departureURL)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer departureResponse.Body.Close()

    // Decode the JSON response for departures into a struct
    var departureResponseData FlightResponse
    err = json.NewDecoder(departureResponse.Body).Decode(&departureResponseData)
    if err != nil {
        fmt.Println("Error decoding departure JSON:", err)
        return
    }

    // Create a map to store unique departure IATA codes for Thursday and Friday departures
    departureIATACodes := make(map[string]bool)

    // Filter and store unique departure IATA codes for Thursday and Friday departures
    for _, flight := range departureResponseData.Data.Items {
        departureTime, err := time.Parse(time.RFC3339, flight.ScheduledDepartureTime)
        if err != nil {
            continue
        }
        if departureTime.Weekday() == time.Thursday || departureTime.Weekday() == time.Friday {
            departureIATACodes[flight.DestinationIATACode] = true
        }
    }

    // Replace the dates in the URL for arrivals (Sunday, Monday, and Tuesday)
    arrivalURL := "https://ber.berlin-airport.de//api.flights.json?arrivalDeparture=A" +
        "&dateFrom=" + today.Format("2006-01-02") +
        "&dateUntil=" + sevenDaysLater.Format("2006-01-02") +
        "&search=&lang=en&page=1&terminal=&itemsPerPage=1600"

    // Make an HTTP GET request for arrivals
    arrivalResponse, err := http.Get(arrivalURL)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer arrivalResponse.Body.Close()

    // Decode the JSON response for arrivals into a struct
    var arrivalResponseData FlightResponse
    err = json.NewDecoder(arrivalResponse.Body).Decode(&arrivalResponseData)
    if err != nil {
        fmt.Println("Error decoding arrival JSON:", err)
        return
    }

    // Create a map to store unique arrival IATA codes for Sunday, Monday, and Tuesday arrivals
    arrivalIATACodes := make(map[string]bool)

    // Filter and store unique arrival IATA codes for Sunday, Monday, and Tuesday arrivals
    for _, flight := range arrivalResponseData.Data.Items {
        arrivalTime, err := time.Parse(time.RFC3339, flight.ScheduledDepartureTime)
        if err != nil {
            continue
        }
        if arrivalTime.Weekday() == time.Sunday || arrivalTime.Weekday() == time.Monday || arrivalTime.Weekday() == time.Tuesday {
            arrivalIATACodes[flight.OriginIATACode] = true
        }
    }

    // Create a map to store combinations of departures and arrivals
    combinations := make(map[string][]string)

    // Populate the combinations map
    combinations["Thursday to Tuesday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Thursday}, []time.Weekday{time.Tuesday})
    combinations["Thursday to Monday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Thursday}, []time.Weekday{time.Monday})
    combinations["Thursday to Sunday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Thursday}, []time.Weekday{time.Sunday})
    combinations["Friday to Tuesday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Friday}, []time.Weekday{time.Tuesday})
    combinations["Friday to Monday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Friday}, []time.Weekday{time.Monday})
    combinations["Friday to Sunday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes, []time.Weekday{time.Friday}, []time.Weekday{time.Sunday})

    // Print the final table
    fmt.Println("Final Table:")
    for key, value := range combinations {
        fmt.Printf("%s: %v\n", key, value)
    }
}

// getMatchingFlights returns a list of matching IATA codes for departures and arrivals
func getMatchingFlights(departureCodes, arrivalCodes map[string]bool, departureDays, arrivalDays []time.Weekday) []string {
    matchingFlights := []string{}
    for code := range departureCodes {
        if arrivalCodes[code] {
            matchingFlights = append(matchingFlights, code)
        }
    }
    return matchingFlights
}

