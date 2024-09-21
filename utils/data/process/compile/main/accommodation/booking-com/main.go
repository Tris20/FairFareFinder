
package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

// Accommodation represents the filtered data we will extract
type Accommodation struct {
	City      string
	Country   string
	GrossPrice float64
	Checkin    string
	Checkout   string
}

// LocationPrices holds prices for a specific location (city + country)
type LocationPrices struct {
	City    string
	Country string
	Prices  []float64
}

func main() {
	// Step 1: Open (or create) "new_main.db"
	newDb, err := sql.Open("sqlite3", "../../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatalf("Failed to open new_main.db: %v", err)
	}
	defer newDb.Close()

	// Step 2: Open the "raw/booking.db"
	rawDb, err := sql.Open("sqlite3", "../../../../../../../data/raw/accommocation/booking-com/booking.db")
	if err != nil {
		log.Fatalf("Failed to open booking.db: %v", err)
	}
	defer rawDb.Close()

	// Step 3: Query the 'property' table for records where review_score > 7
	query := `SELECT city, country, gross_price, checkin_date, checkout_date FROM property WHERE review_score > 7`
	rows, err := rawDb.Query(query)
	if err != nil {
		log.Fatalf("Failed to query property table: %v", err)
	}
	defer rows.Close()

	// Variables to collect checkin and checkout dates once
	var checkinDate, checkoutDate string
	// Step 4: Collect prices for each unique location
	locationData := make(map[string]LocationPrices)

	for rows.Next() {
		var acc Accommodation
		err := rows.Scan(&acc.City, &acc.Country, &acc.GrossPrice, &acc.Checkin, &acc.Checkout)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Collect checkin_date and checkout_date only once
		if checkinDate == "" && checkoutDate == "" {
			checkinDate = acc.Checkin
			checkoutDate = acc.Checkout
		}

		locationKey := fmt.Sprintf("%s,%s", acc.City, acc.Country)
		if _, exists := locationData[locationKey]; !exists {
			locationData[locationKey] = LocationPrices{
				City:    acc.City,
				Country: acc.Country,
				Prices:  []float64{},
			}
		}

		location := locationData[locationKey]
		location.Prices = append(location.Prices, acc.GrossPrice)
		locationData[locationKey] = location
	}

	// Step 5: Process each location's prices
	for _, loc := range locationData {
		fmt.Printf("\nProcessing location: %s, %s\n", loc.City, loc.Country)

		// Sort the prices (lowest to highest)
		sort.Float64s(loc.Prices)

		// Calculate the 10% drop count
		numEntries := len(loc.Prices)
		if numEntries < 10 {
			fmt.Printf("Not enough entries for %s, %s to drop 10%% outliers\n", loc.City, loc.Country)
			continue
		}

		dropCount := int(math.Floor(float64(numEntries) * 0.10))

		// Collect the remaining prices (middle 80%)
		remainingPrices := loc.Prices[dropCount : numEntries-dropCount]

		// Sort remaining prices (lowest to highest)
		sort.Float64s(remainingPrices)

		// Step 6: Calculate median of remaining prices
		medianPrice := calculateMedian(remainingPrices)

		// Step 7: Calculate avg_pppn by dividing the median by 14 and rounding to 2 decimal places
		avgPppn := roundToTwoDecimalPlaces(medianPrice / 14)

		// Step 8: Create the booking URL for this location
		bookingURL := fmt.Sprintf("https://www.booking.com/searchresults.en-gb.html?ss=%s&group_adults=1&no_rooms=1&group_children=0&nflt=price%%3DEUR-min-110-1%%3Breview_score%%3D80&flex_window=2&checkin=%s&checkout=%s", loc.City, checkinDate, checkoutDate)

		// Print results
		fmt.Printf("Remaining prices for %s, %s: ", loc.City, loc.Country)
		for _, price := range remainingPrices {
			fmt.Printf("%.2f ", price)
		}
		fmt.Println()

		fmt.Printf("Median price: %.2f, avg_pppn: %.2f\n", medianPrice, avgPppn)
		fmt.Printf("Booking URL: %s\n", bookingURL)
	}
}

// Function to calculate the median of a sorted list of prices
func calculateMedian(prices []float64) float64 {
	n := len(prices)
	if n%2 == 0 {
		return (prices[n/2-1] + prices[n/2]) / 2
	}
	return prices[n/2]
}

// Function to round to two decimal places
func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}

