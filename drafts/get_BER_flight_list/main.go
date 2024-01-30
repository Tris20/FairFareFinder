
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
    DestinationCity       string `json:"arr_airport_name"`
    DestinationIATACode   string `json:"arr_airport_iata"`
    ScheduledDepartureTime string `json:"scheduled_time"`
}

func main() {
    // Get today's date
    today := time.Now()
    // Get the date 7 days from today
    sevenDaysLater := today.Add(7 * 24 * time.Hour)

    // Replace the dates in the URL
    url := "https://ber.berlin-airport.de//api.flights.json?arrivalDeparture=D" +
        "&dateFrom=" + today.Format("2006-01-02") +
        "&dateUntil=" + sevenDaysLater.Format("2006-01-02") +
        "&search=&lang=en&page=1&terminal=&itemsPerPage=1600"

    // Make an HTTP GET request to the URL
    response, err := http.Get(url)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer response.Body.Close()

    // Decode the JSON response into a struct
    var flightResponse FlightResponse
    err = json.NewDecoder(response.Body).Decode(&flightResponse)
    if err != nil {
        fmt.Println("Error decoding JSON:", err)
        return
    }

    // Create a map to store unique cities
    uniqueCities := make(map[string]string)

    // Filter cities that appear on Thursday or Friday
    for _, flight := range flightResponse.Data.Items {
        departureTime, err := time.Parse("2006-01-02T15:04:05-07:00", flight.ScheduledDepartureTime)
        if err != nil {
            fmt.Println("Error parsing time:", err)
            continue
        }

        if departureTime.Weekday() == time.Thursday || departureTime.Weekday() == time.Friday {
            uniqueCities[flight.DestinationIATACode] = flight.DestinationCity
        }
    }

    // Print the unique cities and their IATI codes
    for code, city := range uniqueCities {
        fmt.Printf("City: %s, IATI Code: %s\n", city, code)
    }
}

