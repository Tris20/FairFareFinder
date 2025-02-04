package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Origin defines one origin input.
type Origin struct {
	City       string
	Country    string
	PriceLimit float64 // the slider value for that origin
}

// FlightResult holds one row of the query output.
type FlightResult struct {
	DestinationCity    string
	DestinationCountry string
	FlightPrice        float64
}

func buildQueryForActiveOrigin(active Origin, origins []Origin, globalThreshold, accomLimit float64) string {
	// Build a slice of OR conditions, one per origin.
	// For the active origin, we use the global threshold.
	// For all others, we use their individual PriceLimit.
	var conditions []string
	for _, o := range origins {
		var cond string
		if o.City == active.City && o.Country == active.Country {
			cond = fmt.Sprintf(
				"(f.origin_city_name = '%s' AND f.origin_country = '%s' AND f.price_next_week <= %.2f)",
				o.City, o.Country, globalThreshold,
			)
		} else {
			cond = fmt.Sprintf(
				"(f.origin_city_name = '%s' AND f.origin_country = '%s' AND f.price_next_week <= %.2f)",
				o.City, o.Country, o.PriceLimit,
			)
		}
		conditions = append(conditions, cond)
	}
	// Join the conditions with OR.
	whereClause := fmt.Sprintf("(%s)", strings.Join(conditions, " OR "))

	// Build the complete query with an ORDER BY clause.
	query := fmt.Sprintf(`
SELECT DISTINCT 
  f.destination_city_name,
  f.destination_country,
  f.price_next_week AS FlightPrice
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
WHERE %s
  AND a.booking_pppn IS NOT NULL
  AND a.booking_pppn <= %.2f
ORDER BY f.price_next_week ASC;
`, whereClause, accomLimit)

	return query
}

func main() {
	// Define input origins.
	origins := []Origin{
		{City: "Berlin", Country: "DE", PriceLimit: 38.00},
		{City: "Glasgow", Country: "GB", PriceLimit: 33.00},
		{City: "Frankfurt", Country: "DE", PriceLimit: 55.00},
	}
	// Accommodation price slider.
	accomPriceLimit := 31.00
	// Global flight price threshold for the active origin.
	globalFlightPriceThreshold := 2500.00

	// Open the database (adjust the DSN as needed).
	db, err := sql.Open("sqlite3", "../../data/compiled/main.db")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}
	defer db.Close()

	// For each origin, build the corresponding query and execute it.
	// We'll store the results in a map from active origin to the results array.
	resultsMap := make(map[string][]FlightResult)

	for _, active := range origins {
		query := buildQueryForActiveOrigin(active, origins, globalFlightPriceThreshold, accomPriceLimit)
		log.Printf("Query for active origin %s:\n%s", active.City, query)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatalf("DB query failed for active origin %s: %v", active.City, err)
		}

		var resArr []FlightResult
		for rows.Next() {
			var fr FlightResult
			if err := rows.Scan(&fr.DestinationCity, &fr.DestinationCountry, &fr.FlightPrice); err != nil {
				log.Fatalf("Row scan failed for active origin %s: %v", active.City, err)
			}
			resArr = append(resArr, fr)
		}
		if err := rows.Err(); err != nil {
			log.Fatalf("Rows error for active origin %s: %v", active.City, err)
		}
		rows.Close()

		resultsMap[active.City] = resArr
	}

	// Print the results for each active origin.
	for _, o := range origins {
		fmt.Printf("Results for active origin %s:\n", o.City)
		for _, r := range resultsMap[o.City] {
			fmt.Printf("  Destination: %s, %s | Flight Price: %.2f\n",
				r.DestinationCity, r.DestinationCountry, r.FlightPrice)
		}
		fmt.Println()
	}
}
