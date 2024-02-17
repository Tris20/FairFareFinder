
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

// Define a struct for the JSON data
type Destination struct {
	IATACode string `json:"IATA_code"`
	CityName string `json:"city_name"`
}

// Define a struct for the output with the URL added
type DestinationWithURL struct {
	IATA     string `json:"IATA"`
	CityName string `json:"City_name"`
	URL      string `json:"URL"`
}

func main() {
	// Load the JSON data from file
	var destinations []Destination
	data, err := ioutil.ReadFile("destinations.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	// Unmarshal the JSON data into the slice of Destination structs
	err = json.Unmarshal(data, &destinations)
	if err != nil {
		fmt.Println("Error unmarshaling JSON data:", err)
		return
	}

	// Prepare the base URL with a placeholder for the IATA code
	baseURL := "https://www.skyscanner.de/transport/fluge/ber/$$$/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1"

	// Create a new slice for the modified destinations
	var destinationsWithUrls []DestinationWithURL

	// Replace the placeholder in the URL with the actual IATA code for each destination
	for _, dest := range destinations {
		url := replacePlaceholder(baseURL, dest.IATACode)
		destinationsWithUrls = append(destinationsWithUrls, DestinationWithURL{
			IATA:     dest.IATACode,
			CityName: dest.CityName,
			URL:      url,
		})
	}

	// Marshal the modified slice into JSON
	modifiedData, err := json.MarshalIndent(destinationsWithUrls, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling modified data to JSON:", err)
		return
	}

	// Save the modified JSON data to "flights.json"
	err = ioutil.WriteFile("flights.json", modifiedData, 0644)
	if err != nil {
		fmt.Println("Error writing modified data to file:", err)
		return
	}

	fmt.Println("Modified data has been saved to flights.json")
}

// Function to replace the placeholder in the URL with the actual IATA code
func replacePlaceholder(url, iataCode string) string {
	return strings.Replace(url, "$$$", iataCode, 1)
}
