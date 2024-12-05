package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// CityInput represents the input for each city
type CityInput struct {
	Name       string
	PriceLimit float64
}

func main() {
	// Example inputs
	cities := []CityInput{
		{Name: "Berlin", PriceLimit: 150},
		{Name: "Munich", PriceLimit: 200},
	}
	logicalOperator := "AND" // Change to "OR" for OR logic

	// Build the query
	query, args := buildQuery(cities, logicalOperator)

	// Output the query for debugging
	fmt.Println("Generated SQL Query:")
	fmt.Println(query)
	fmt.Println("Arguments:")
	fmt.Println(args)

	// Connect to the SQLite database (replace "your_database.db" with your actual database file)
	db, err := sql.Open("sqlite3", "../../../data/compiled/main.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Process the results
	for rows.Next() {
		var destinationCityName, destinationCountry, originCities string
		var price float64
		// Add other columns as needed

		err = rows.Scan(&destinationCityName, &destinationCountry, &originCities, &price /*, other columns*/)
		if err != nil {
			panic(err)
		}

		// Output the results
		fmt.Printf("Destination: %s, %s\n", destinationCityName, destinationCountry)
		fmt.Printf("Origin Cities: %s\n", originCities)
		fmt.Printf("Price: %.2f\n", price)
		// Output other columns as needed
	}
}

func buildQuery(cities []CityInput, logicalOperator string) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}

	// Begin the query
	queryBuilder.WriteString(`
WITH Flights AS (
`)

	// Build individual city queries
	cityQueries := make([]string, len(cities))
	for i, city := range cities {
		cityQuery := fmt.Sprintf(`
    SELECT 
        f.origin_city_name,
        f.destination_city_name,
        f.destination_country,
        MIN(f.price_next_week) AS price,
        MIN(f.skyscanner_url_next_week) AS url
    FROM flight f
    WHERE f.origin_city_name = ?
          AND f.price_next_week < ?
    GROUP BY f.origin_city_name, f.destination_city_name, f.destination_country
`)

		// Add to args
		args = append(args, city.Name, city.PriceLimit)

		// If not the first city, prepend UNION ALL
		if i > 0 {
			cityQuery = "UNION ALL" + cityQuery
		}

		cityQueries[i] = cityQuery
	}

	// Join all city queries
	queryBuilder.WriteString(strings.Join(cityQueries, "\n"))

	// Close the Flights CTE
	queryBuilder.WriteString(`
),
DestinationCounts AS (
    SELECT 
        destination_city_name,
        destination_country,
        COUNT(DISTINCT origin_city_name) AS num_origins
    FROM Flights
    GROUP BY destination_city_name, destination_country
),
TotalOrigins AS (
    SELECT COUNT(DISTINCT origin_city_name) AS total_origins
    FROM Flights
)
SELECT 
    f.destination_city_name,
    f.destination_country,
    GROUP_CONCAT(DISTINCT f.origin_city_name) AS origin_cities,
    MIN(f.price) AS price
    -- Add other columns as needed
FROM Flights f
JOIN DestinationCounts d ON f.destination_city_name = d.destination_city_name 
                          AND f.destination_country = d.destination_country
JOIN TotalOrigins t
JOIN location l ON f.destination_city_name = l.city 
                 AND f.destination_country = l.country
JOIN weather w ON w.city = f.destination_city_name 
                AND w.country = f.destination_country
LEFT JOIN accommodation a ON a.city = f.destination_city_name 
                           AND a.country = f.destination_country
WHERE l.avg_wpi BETWEEN 1.0 AND 10.0 
  AND w.date >= date('now')
`)

	// Add logical operator condition
	if logicalOperator == "AND" {
		queryBuilder.WriteString(`
  AND d.num_origins = t.total_origins
`)
	}

	// Close the query
	queryBuilder.WriteString(`
GROUP BY f.destination_city_name, f.destination_country
ORDER BY l.avg_wpi DESC;
`)

	return queryBuilder.String(), args
}
