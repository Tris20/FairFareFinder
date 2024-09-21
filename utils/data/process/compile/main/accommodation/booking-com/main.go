
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
	query := `SELECT city, country, gross_price FROM property WHERE review_score > 7`
	rows, err := rawDb.Query(query)
	if err != nil {
		log.Fatalf("Failed to query property table: %v", err)
	}
	defer rows.Close()

	// Step 4: Collect prices for each unique location
	locationData := make(map[string]LocationPrices)

	for rows.Next() {
		var acc Accommodation
		err := rows.Scan(&acc.City, &acc.Country, &acc.GrossPrice)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
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

		// Collect the lowest 10% and highest 10% separately
		lowestPrices := loc.Prices[:dropCount]
		highestPrices := loc.Prices[numEntries-dropCount:]

		// Collect the remaining prices (middle 80%)
		remainingPrices := loc.Prices[dropCount : numEntries-dropCount]

		// Sort the lowest and highest dropped prices
		sort.Float64s(lowestPrices)  // Already sorted, but to be consistent
		sort.Float64s(highestPrices) // Already sorted

		// Print remaining prices (lowest to highest)
		fmt.Printf("Remaining prices for %s, %s: ", loc.City, loc.Country)
		for _, price := range remainingPrices {
			fmt.Printf("%.2f ", price)
		}
		fmt.Println()

		// Print dropped prices (lowest to highest: first lowest 10%, then highest 10%)
		fmt.Printf("Dropped prices for %s, %s: ", loc.City, loc.Country)
		for _, price := range lowestPrices {
			fmt.Printf("%.2f ", price)
		}
		for _, price := range highestPrices {
			fmt.Printf("%.2f ", price)
		}
		fmt.Println()
	}
}

