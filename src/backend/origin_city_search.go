package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
)

var (
	cityCountryPairs []CityCountry
	dataLoadOnce     sync.Once // Ensure the data is loaded only once
)

type CityCountry struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// LoadCityCountryPairs loads unique city-country pairs into memory
func LoadCityCountryPairs(db *sql.DB) {
	dataLoadOnce.Do(func() { // Only load once
		rows, err := db.Query(`
			SELECT DISTINCT origin_city_name, origin_country
			FROM flight
		`)
		if err != nil {
			log.Fatalf("Failed to load city-country pairs: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var city, country string
			if err := rows.Scan(&city, &country); err == nil {
				cityCountryPairs = append(cityCountryPairs, CityCountry{City: city, Country: country})
			}
		}

		// Sort the city-country pairs alphabetically by city name
		sort.Slice(cityCountryPairs, func(i, j int) bool {
			return cityCountryPairs[i].City < cityCountryPairs[j].City
		})

		log.Printf("Loaded and sorted %d city-country pairs into memory.", len(cityCountryPairs))
	})
}

// GetCityCountryPairs returns the loaded city-country pairs
func GetCityCountryPairs() []CityCountry {
	return cityCountryPairs
}

func CityCountryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cityCountryPairs); err != nil {
		http.Error(w, "Failed to encode city-country pairs", http.StatusInternalServerError)
	}
}
