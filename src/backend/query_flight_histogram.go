package backend

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/backend/config"
	"log"
)

func ExecuteFlightPricesHistogramQuery(input *FilterInput) ([]float64, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	//Note we check for weather and location entries to ensure the histogram
	// only shows bars where we have all the data needed to render the destination card
	query := `
    SELECT f.price_this_week 
    FROM flight f
    JOIN accommodation a 
        ON f.destination_city_name = a.city 
        AND f.destination_country = a.country
    JOIN weather w
        ON f.destination_city_name = w.city 
        AND f.destination_country = w.country
    JOIN location l
        ON f.destination_city_name = l.city 
        AND f.destination_country = l.country
    WHERE f.origin_city_name = ?
    AND f.price_this_week <= 2500
    AND a.booking_pppn > 9.99
    AND a.booking_pppn <= ?;
`

	args := []interface{}{input.Cities[0], input.MaxAccommodationPrice}

	if !config.MutePrints {
		fmt.Println("Generated SQL Query (FLIGHT HISTOGRAM PRICES):", query)
		fmt.Println("Arguments:", args)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying flight histogram: %v", err)
		return nil, err
	}
	defer rows.Close()

	var flightPrices []float64
	for rows.Next() {
		var price float64
		if err := rows.Scan(&price); err != nil {
			log.Printf("Error scanning flight price row: %v", err)
			continue
		}
		flightPrices = append(flightPrices, price)
	}

	//	fmt.Printf("Flight Prices:", flightPrices)
	return flightPrices, nil
}
