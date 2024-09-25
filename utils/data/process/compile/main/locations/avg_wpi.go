package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

func UpdateAvgWPI(db *sql.DB) {
	// Query unique city-country pairs from the 'location' table
	rows, err := db.Query("SELECT DISTINCT city, country FROM location")
	if err != nil {
		fmt.Println("Error fetching city-country pairs:", err)
		return
	}
	defer rows.Close()

	// Collect all city-country pairs for progress bar initialization
	cityCountryPairs := make([]struct{ city, country string }, 0)
	for rows.Next() {
		var city, country string
		if err := rows.Scan(&city, &country); err != nil {
			fmt.Println("Error scanning city-country pair:", err)
			continue
		}
		cityCountryPairs = append(cityCountryPairs, struct{ city, country string }{city, country})
	}

	// Initialize the progress bar
	bar := progressbar.NewOptions(len(cityCountryPairs),
		progressbar.OptionSetDescription("Updating average WPI"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("pairs"),
	)

	// Loop over each city-country pair to calculate and update avg_wpi
	for _, pair := range cityCountryPairs {
		var avgResult sql.NullFloat64 // Use sql.NullFloat64 to handle NULL values
		err = db.QueryRow("SELECT AVG(avg_daytime_wpi) FROM weather WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", pair.city, pair.country).Scan(&avgResult)
		if err != nil {
			fmt.Printf("Failed to calculate average WPI for %s, %s: %v\n", pair.city, pair.country, err)
			bar.Add(1) // Increment the progress bar even in case of an error
			continue
		}

		if avgResult.Valid { // Check if the result is valid (not NULL)
			_, err = db.Exec("UPDATE location SET avg_wpi = ? WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", avgResult.Float64, pair.city, pair.country)
			if err != nil {
				fmt.Printf("Failed to update avg_wpi for %s, %s: %v\n", pair.city, pair.country, err)
			}
		} else {
			fmt.Printf("No valid average WPI found for %s, %s, setting avg_wpi to NULL\n", pair.city, pair.country)
			_, err = db.Exec("UPDATE location SET avg_wpi = NULL WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", pair.city, pair.country)
			if err != nil {
				fmt.Printf("Failed to set avg_wpi to NULL for %s, %s: %v\n", pair.city, pair.country, err)
			}
		}

		bar.Add(1) // Increment the progress bar after processing each pair
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error during rows iteration:", err)
	}

	fmt.Println("Updated avg_wpi based on weather data successfully.")
}
