package backend

import (
	"database/sql"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"
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

type Match struct {
	CityCountry CityCountry
	Score       int
}

// Fuzzy matching logic that prioritizes substring matches

func FuzzyMatch(query string, city, country string) int {
	queryLower := strings.ToLower(query)
	cityLower := strings.ToLower(city)
	countryLower := strings.ToLower(country)

	// Exact substring match scores highest
	if strings.HasPrefix(cityLower, queryLower) || strings.HasPrefix(countryLower, queryLower) {
		return 0
	}
	if strings.Contains(cityLower, queryLower) || strings.Contains(countryLower, queryLower) {
		return 1
	}

	// Use Levenshtein distance as a fallback
	return Levenshtein(queryLower, cityLower) + Levenshtein(queryLower, countryLower)
}

func AutocompleteCitiesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	var matches []CityCountry
	for _, pair := range cityCountryPairs {
		if strings.Contains(strings.ToLower(pair.City), strings.ToLower(query)) {
			matches = append(matches, pair)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
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

		var pairs []CityCountry
		for rows.Next() {
			var city, country string
			if err := rows.Scan(&city, &country); err == nil {
				pairs = append(pairs, CityCountry{City: city, Country: country})
			}
		}

		cityCountryPairs = pairs
		log.Printf("Loaded %d city-country pairs into memory.", len(cityCountryPairs))
	})
}

func Levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Convert strings to lowercase
	a, b = strings.ToLower(a), strings.ToLower(b)

	// Initialize distance matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Fill in base cases
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Compute distances
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = int(math.Min(
				math.Min(
					float64(matrix[i-1][j]+1), // Deletion
					float64(matrix[i][j-1]+1), // Insertion
				),
				float64(matrix[i-1][j-1]+cost), // Substitution
			))
		}
	}
	return matrix[len(a)][len(b)]
}
