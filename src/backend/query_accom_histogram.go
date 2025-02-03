package backend

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend/config"
	"log"
)

func ExecuteAccommodationPricesHistogramQuery(input *FilterInput) ([]float64, error) {
	// Build query with fixed maxAccommodationPrice = 550.0
	allPricesQuery, allPricesArgs := BuildMainQuery(input.LogicalExpression, config.MaxAccomPrice, input.Cities, input.OrderClause)

	if !config.MutePrints {
		fmt.Println("Generated SQL Query (ACCOMMODATION HISTOGRAM PRICES):")
		fmt.Println(allPricesQuery)
		fmt.Println("Arguments:", allPricesArgs)
	}

	fullAllPricesQuery, err := ReplacePlaceholdersWithArgs(allPricesQuery, allPricesArgs)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(fullAllPricesQuery)
	}

	log.Printf("Full ALL-PRICES Query:\n%s\n", fullAllPricesQuery)

	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	rows, err := db.Query(allPricesQuery, allPricesArgs...)
	if err != nil {
		log.Printf("Error querying all accommodation prices: %v", err)
		return nil, err
	}
	defer rows.Close()

	flightsAll, err := ProcessFlightRows(rows)
	if err != nil {
		log.Printf("Error processing flight rows for all accommodation prices: %v", err)
		return nil, err
	}

	var prices []float64
	for _, f := range flightsAll {
		if f.BookingPppn.Valid {
			prices = append(prices, f.BookingPppn.Float64)
		}
	}
	return prices, nil
}
