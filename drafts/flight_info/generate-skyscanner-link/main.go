
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// Define a struct for the JSON data
type Destination struct {
	IATACode string `json:"IATA_code"`
	CityName string `json:"city_name"`
}

// Define a struct for the output with the URLs added
type DestinationWithURL struct {
	IATA          string `json:"IATA"`
	CityName      string `json:"City_name"`
	SkyScannerURL string `json:"SkyScannerURL"`
	AirbnbURL     string `json:"airbnbURL"`
	BookingURL    string `json:"bookingURL"`
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

	// Prepare the base URLs with placeholders
	baseSkyScannerURL := "https://www.skyscanner.de/transport/fluge/ber/$$$/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1"

	// Create a new slice for the modified destinations
	var destinationsWithUrls []DestinationWithURL

	// Replace the placeholders in the URLs with actual values for each destination
	for _, dest := range destinations {
		skyScannerURL := replacePlaceholder(baseSkyScannerURL, dest.IATACode)
		airbnbURL := generateAirbnbURL(dest.CityName)
		bookingURL := generateBookingURL(dest.CityName)
		destinationsWithUrls = append(destinationsWithUrls, DestinationWithURL{
			IATA:          dest.IATACode,
			CityName:      dest.CityName,
			SkyScannerURL: skyScannerURL,
			AirbnbURL:     airbnbURL,
			BookingURL:    bookingURL,
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

// Function to generate the Airbnb URL with dynamic values
func generateAirbnbURL(cityName string) string {
	checkin := nextThursdayFridaySaturday()
	checkout := checkin.Add(3 * 24 * time.Hour) // Adding 3 days for a one-week stay
	return fmt.Sprintf("https://www.airbnb.de/s/%s/homes?adults=1&checkin=%s&checkout=%s&flexible_trip_lengths%%5B%%5D=one_week&price_filter_num_nights=3&price_max=112", cityName, checkin.Format("2006-01-02"), checkout.Format("2006-01-02"))
}

// Function to generate the Booking.com URL with dynamic values
func generateBookingURL(cityName string) string {
	checkin := nextThursdayFridaySaturday()
	checkout := checkin.Add(3 * 24 * time.Hour) // Adding 3 days for a one-week stay
	return fmt.Sprintf("https://www.booking.com/searchresults.en-gb.html?ss=%s&group_adults=1&no_rooms=1&group_children=0&nflt=price%%3DEUR-min-110-1%%3Breview_score%%3D80&flex_window=2&checkin=%s&checkout=%s", cityName, checkin.Format("2006-01-02"), checkout.Format("2006-01-02"))
}

// Function to get the date of the next Thursday, Friday, or Saturday from today's date
func nextThursdayFridaySaturday() time.Time {
	today := time.Now()
	for {
		switch today.Weekday() {
		case time.Thursday, time.Friday, time.Saturday:
			return today
		}
		today = today.Add(24 * time.Hour)
	}
}

