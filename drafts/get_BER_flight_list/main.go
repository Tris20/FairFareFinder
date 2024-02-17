
package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "time"
    "os"
    "strings"
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
    DestinationCity       string `json:"arr_airport_name"`
}

// Define a struct to hold city information with labels for JSON
type Destination struct {
    IATACode string `json:"IATA_code"`
    CityName string `json:"city_name"`
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
    departureIATACodes := make(map[string]string)

    // Filter and store unique departure IATA codes for Thursday and Friday departures
    for _, flight := range departureResponseData.Data.Items {
        departureTime, err := time.Parse(time.RFC3339, flight.ScheduledDepartureTime)
        if err != nil {
            continue
        }
        if departureTime.Weekday() == time.Thursday || departureTime.Weekday() == time.Friday {
            departureIATACodes[flight.OriginIATACode] = flight.OriginCity
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
    arrivalIATACodes := make(map[string]string)

    // Filter and store unique arrival IATA codes for Sunday, Monday, and Tuesday arrivals
    for _, flight := range arrivalResponseData.Data.Items {
        arrivalTime, err := time.Parse(time.RFC3339, flight.ScheduledDepartureTime)
        if err != nil {
            continue
        }
        if arrivalTime.Weekday() == time.Sunday || arrivalTime.Weekday() == time.Monday || arrivalTime.Weekday() == time.Tuesday {
            arrivalIATACodes[flight.OriginIATACode] = flight.OriginCity
        }
    }

    // Create a map to store combinations of departures and arrivals
    combinations := make(map[string][]string)

    // Populate the combinations map
    combinations["Thursday to Tuesday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)
    combinations["Thursday to Monday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)
    combinations["Thursday to Sunday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)
    combinations["Friday to Tuesday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)
    combinations["Friday to Monday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)
    combinations["Friday to Sunday"] = getMatchingFlights(departureIATACodes, arrivalIATACodes)

    // Print the final table
    fmt.Println("Final Table:")
    for key, value := range combinations {
        fmt.Printf("%s: %v\n", key, value)
    }

    // Create a map to store unique cities and their IATA codes
    uniqueCities := make(map[string]string)

    // Populate the uniqueCities map with departures
    for code, city := range departureIATACodes {
        uniqueCities[code] = city
    }

    // Populate the uniqueCities map with arrivals
    for code, city := range arrivalIATACodes {
        uniqueCities[code] = city
    }

    // Print the final list of unique cities and their IATA codes
    fmt.Println("\nFinal List of Unique Cities:")
    for code, city := range uniqueCities {
        fmt.Printf("%s - %s\n", code, city)
    }
    saveDestinationsAsJSON(uniqueCities)
}

// getMatchingFlights returns a list of matching IATA codes for departures and arrivals
func getMatchingFlights(departureCodes, arrivalCodes map[string]string) []string {
    matchingFlights := []string{}
    for code := range departureCodes {
        if arrivalCodes[code] != "" {
            matchingFlights = append(matchingFlights, code)
        }
    }
    return matchingFlights
}



// Function to save destinations to JSON with appropriate labels
func saveDestinationsAsJSON(cities map[string]string) {
    uniqueDestinations := make(map[string]Destination)

    for iataCode, cityName := range cities {
        // Check if the city name contains the IATA code and remove it
        if strings.Contains(cityName, iataCode) {
            cityName = strings.Replace(cityName, iataCode, "", -1)
            cityName = strings.TrimSpace(cityName)
        }

        // Use cityName as the key to ensure uniqueness
        if _, exists := uniqueDestinations[cityName]; !exists {
            uniqueDestinations[cityName] = Destination{
                IATACode: iataCode,
                CityName: cityName,
            }
        }
    }

    // Convert the map to a slice for JSON marshaling
    destinationsSlice := make([]Destination, 0, len(uniqueDestinations))
    for _, dest := range uniqueDestinations {
        destinationsSlice = append(destinationsSlice, dest)
    }

    // Marshal the slice to JSON
    jsonData, err := json.MarshalIndent(destinationsSlice, "", "    ")
    if err != nil {
        fmt.Println("Error marshaling destinations to JSON:", err)
        return
    }

    // Write the JSON data to a file
    err = os.WriteFile("destinations.json", jsonData, 0644)
    if err != nil {
        fmt.Println("Error writing destinations to file:", err)
        return
    }

    fmt.Println("Saved destinations and their IATA codes to destinations.json")
}
