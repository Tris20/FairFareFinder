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

// LogicalOperator represents a logical operator (AND, OR)
type LogicalOperator string

const (
	AndOperator LogicalOperator = "AND"
	OrOperator  LogicalOperator = "OR"
)

// Expression represents a logical expression
type Expression interface{}

// CityCondition represents a condition for a single city
type CityCondition struct {
	City CityInput
}

// LogicalExpression represents a logical combination of expressions
type LogicalExpression struct {
	Operator LogicalOperator
	Left     Expression
	Right    Expression
}

func main() {
	// Example logical expression: (Berlin AND Munich) OR Edinburgh
	expr := &LogicalExpression{
		Operator: AndOperator,
		Left: &LogicalExpression{
			Operator: AndOperator,
			Left:     &CityCondition{City: CityInput{Name: "Berlin", PriceLimit: 150}},
			Right:    &CityCondition{City: CityInput{Name: "Munich", PriceLimit: 200}},
		},
		Right: &CityCondition{City: CityInput{Name: "Edinburgh", PriceLimit: 180}},
	}

	// Define the maximum accommodation price per person per night
	maxAccommodationPrice := 70.0 // Or get this value from user input

	// Build the query
	query, args := buildQuery(expr, maxAccommodationPrice)

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

	var flights []Flight
	for rows.Next() {
		var flight Flight
		var weather Weather
		var imageUrl sql.NullString
		var bookingUrl sql.NullString
		var priceFnaf sql.NullFloat64

		err = rows.Scan(
			&flight.DestinationCityName,
			&flight.PriceCity1,
			&flight.UrlCity1,
			&weather.Date,
			&weather.AvgDaytimeTemp,
			&weather.WeatherIcon,
			&weather.GoogleUrl,
			&flight.AvgWpi,
			&imageUrl,
			&bookingUrl,
			&flight.BookingPppn,
			&priceFnaf,
		)
		if err != nil {
			panic(err)
		}

		// Assign additional fields
		flight.RandomImageURL = imageUrl.String
		flight.BookingUrl = bookingUrl
		flight.FiveNightsFlights = priceFnaf
		flight.WeatherForecast = []Weather{weather} // Assuming one weather record per destination

		flights = append(flights, flight)
	}

	// Output the results
	for _, flight := range flights {
		fmt.Printf("Destination: %s\n", flight.DestinationCityName)
		fmt.Printf("Price City1: %.2f\n", flight.PriceCity1.Float64)
		fmt.Printf("URL City1: %s\n", flight.UrlCity1)
		fmt.Printf("Weather Date: %s\n", flight.WeatherForecast[0].Date)
		fmt.Printf("Avg Daytime Temp: %.2f\n", flight.WeatherForecast[0].AvgDaytimeTemp.Float64)
		fmt.Printf("Weather Icon: %s\n", flight.WeatherForecast[0].WeatherIcon)
		fmt.Printf("Google URL: %s\n", flight.WeatherForecast[0].GoogleUrl)
		fmt.Printf("Avg WPI: %.2f\n", flight.AvgWpi.Float64)
		fmt.Printf("Random Image URL: %s\n", flight.RandomImageURL)
		fmt.Printf("Booking URL: %s\n", flight.BookingUrl.String)
		fmt.Printf("Booking PPPN: %.2f\n", flight.BookingPppn.Float64)
		fmt.Printf("Five Nights Flights: %.2f\n", flight.FiveNightsFlights.Float64)
		fmt.Println("----------------------------")
	}
}

func buildQuery(expr Expression, maxAccommodationPrice float64) (string, []interface{}) {
	var queryBuilder strings.Builder
	var args []interface{}

	// Begin the query with the DestinationSet CTE
	queryBuilder.WriteString("WITH DestinationSet AS (\n")

	// Build the subquery based on the logical expression
	subquery, subqueryArgs := buildSubquery(expr)
	queryBuilder.WriteString(subquery)
	queryBuilder.WriteString("\n)")
	args = append(args, subqueryArgs...)

	// Build the rest of the query
	queryBuilder.WriteString(`
    SELECT 
        ds.destination_city_name,
        MIN(f.price_next_week) AS price_city1,
        MIN(f.skyscanner_url_next_week) AS url_city1,
        w.date,
        w.avg_daytime_temp,
        w.weather_icon,
        w.google_url,
        l.avg_wpi,
        l.image_1,
        a.booking_url,
        a.booking_pppn,
        fnf.price_fnaf
    FROM DestinationSet ds
    JOIN flight f ON ds.destination_city_name = f.destination_city_name 
                   AND ds.destination_country = f.destination_country
    JOIN location l ON ds.destination_city_name = l.city 
                     AND ds.destination_country = l.country
    JOIN weather w ON w.city = ds.destination_city_name 
                    AND w.country = ds.destination_country
    LEFT JOIN accommodation a ON a.city = ds.destination_city_name 
                               AND a.country = ds.destination_country
    LEFT JOIN (
        SELECT 
            fnf.origin_city,
            fnf.origin_country,
            fnf.destination_city,
            fnf.destination_country,
            MIN(fnf.price_fnaf) AS price_fnaf
        FROM five_nights_and_flights fnf
        GROUP BY fnf.origin_city, fnf.origin_country, fnf.destination_city, fnf.destination_country
    ) fnf ON fnf.destination_city = ds.destination_city_name
           AND fnf.destination_country = ds.destination_country
           AND fnf.origin_city = f.origin_city_name
           AND fnf.origin_country = f.origin_country
    WHERE l.avg_wpi BETWEEN 1.0 AND 10.0 
      AND w.date >= date('now')
      AND f.price_next_week < ?
      AND f.origin_city_name IN `)

	// Build the IN clause dynamically based on the number of origin cities
	originCities := []string{"Berlin", "Munich", "Edinburgh"}
	placeholders := make([]string, len(originCities))
	for i := range originCities {
		placeholders[i] = "?"
	}
	inClause := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	queryBuilder.WriteString(inClause)
	queryBuilder.WriteString(`
      AND a.booking_pppn IS NOT NULL
      AND a.booking_pppn <= ?
    GROUP BY ds.destination_city_name, ds.destination_country
    ORDER BY l.avg_wpi DESC;
    `)

	// Add price limits and origin city names to args
	maxPrice := 2000.0 // Set a max price or calculate based on inputs
	args = append(args, maxPrice)

	// Correctly append originCities to args
	for _, city := range originCities {
		args = append(args, city)
	}

	args = append(args, maxAccommodationPrice)

	return queryBuilder.String(), args
}

func buildSubquery(expr Expression) (string, []interface{}) {
	switch e := expr.(type) {
	case *CityCondition:
		// Return the subquery for a city condition
		subquery := `
            SELECT 
                f.destination_city_name,
                f.destination_country
            FROM flight f
            WHERE f.origin_city_name = ? AND f.price_next_week < ?
            GROUP BY f.destination_city_name, f.destination_country
        `
		args := []interface{}{e.City.Name, e.City.PriceLimit}
		return subquery, args
	case *LogicalExpression:
		// Build the left and right subqueries
		leftSubquery, leftArgs := buildSubquery(e.Left)
		rightSubquery, rightArgs := buildSubquery(e.Right)
		var operator string
		if e.Operator == AndOperator {
			operator = "INTERSECT"
		} else if e.Operator == OrOperator {
			operator = "UNION"
		} else {
			panic("Unknown operator")
		}
		// Combine subqueries without unnecessary parentheses
		combinedSubquery := fmt.Sprintf("%s\n%s\n%s", leftSubquery, operator, rightSubquery)
		args := append(leftArgs, rightArgs...)
		return combinedSubquery, args
	default:
		panic("Unknown expression type")
	}
}

// Define your structs here
type Weather struct {
	Date           string
	AvgDaytimeTemp sql.NullFloat64
	WeatherIcon    string
	GoogleUrl      string
	AvgDaytimeWpi  sql.NullFloat64
}

type Flight struct {
	DestinationCityName string
	RandomImageURL      string
	PriceCity1          sql.NullFloat64
	UrlCity1            string
	WeatherForecast     []Weather
	AvgWpi              sql.NullFloat64
	BookingUrl          sql.NullString
	BookingPppn         sql.NullFloat64
	FiveNightsFlights   sql.NullFloat64
}
