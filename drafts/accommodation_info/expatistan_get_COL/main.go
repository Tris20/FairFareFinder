package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Define a struct to match the JSON data structure you expect to receive
type CostOfLivingData struct {
	// Update these fields according to the structure of the Expatistan data
	CityName     string `json:"city_name"`
	OverallIndex int    `json:"overall_index"`
	// Add other fields as necessary
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go [city]")
		os.Exit(1)
	}
	city := os.Args[1]

	// URL construction (hypothetical, as Expatistan does not have a public API)
	url := fmt.Sprintf("https://api.expatistan.com/cost-of-living/%s", strings.ToLower(city))

	// HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making HTTP request: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check if the HTTP request was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Non-OK HTTP status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %s\n", err)
		os.Exit(1)
	}

	// Unmarshal JSON data
	var data CostOfLivingData
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("Error unmarshalling JSON: %s\n", err)
		os.Exit(1)
	}

	// Output the data
	fmt.Printf("Cost of Living Data for %s: %+v\n", city, data)
}
