package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v2"
)

// City represents the city data to be inserted into the accommodation.db
type City struct {
	CityName      string
	CountryCode   string
	DestinationID string
}

// DestinationResponse represents the structure of the API response for destination search
type DestinationResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      []struct {
		DestID     string  `json:"dest_id"`
		SearchType string  `json:"search_type"`
		CityName   string  `json:"city_name"`
		Country    string  `json:"country"`
		Region     string  `json:"region"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Type       string  `json:"type"`
		Name       string  `json:"name"`
		Label      string  `json:"label"`
	} `json:"data"`
}

// Secrets struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
}

// readAPIKey reads the API key from a YAML file
func readAPIKey(filepath string) (string, error) {
	var secrets Secrets
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &secrets)
	if err != nil {
		return "", err
	}
	return secrets.APIKeys.Aerodatabox, nil
}

func main() {
	// Step 1: Look up city names from the 'locations' database
	cities, err := getCityNames()
	if err != nil {
		log.Fatalf("Error retrieving city names: %v", err)
	}

	// Step 2: Create a new SQLite database for storing destination IDs
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/accommocation/booking-com/booking.db")
	if err != nil {
		log.Fatalf("Error creating accommodation.db: %v", err)
	}
	defer db.Close()

	// Create the city table if it doesn't exist
	err = createTables(db)
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// Step 3: Fetch destination IDs from the API and insert them into the database with a progress bar
	bar := progressbar.Default(int64(len(cities)), "Fetching destination IDs for cities")
	for _, city := range cities {
		bar.Add(1)

		fmt.Printf("Fetching destination ID for city: %s (%s)\n", city.CityName, city.CountryCode)
		destinationID := getDestinationID(city.CityName)
		if destinationID == "" {
			log.Printf("No destination ID found for city: %s\n", city.CityName)
			continue
		}

		// Insert the destination ID into the database
		err = insertDestinationID(db, city, destinationID)
		if err != nil {
			log.Printf("Error inserting destination ID for %s: %v", city.CityName, err)
		}

		// Simulate a short delay between API requests to avoid overloading the server
		time.Sleep(10 * time.Millisecond)
	}
}

// getCityNames fetches cities from the locations.db where include_tf == 1
func getCityNames() ([]City, error) {
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/locations/locations.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT city, iso2 FROM city WHERE include_tf = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var city City
		err := rows.Scan(&city.CityName, &city.CountryCode)
		if err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, nil
}

// getDestinationID fetches the destination ID from the API based on the city name with retries
func getDestinationID(city string) string {
	encodedCity := url.QueryEscape(city)
	apiURL := fmt.Sprintf("https://booking-com15.p.rapidapi.com/api/v1/hotels/searchDestination?query=%s", encodedCity)

	apiKey, err := readAPIKey("../../../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
	}

	var resp *http.Response
	client := &http.Client{}
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			log.Fatalf("Error creating API request: %v", err)
		}

		req.Header.Add("x-rapidapi-host", "booking-com15.p.rapidapi.com")
		req.Header.Add("x-rapidapi-key", apiKey)

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("Error making API request (attempt %d): %v", attempt+1, err)
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("API request failed with status: %s for city: %s (attempt %d)", resp.Status, city, attempt+1)
			time.Sleep(time.Duration(1<<attempt) * time.Second)
			continue
		}

		break
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to fetch data for city: %s after %d attempts", city, maxRetries)
		return ""
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading API response: %v", err)
	}

	var destinationResponse DestinationResponse
	err = json.Unmarshal(body, &destinationResponse)
	if err != nil {
		log.Fatalf("Error parsing API response JSON for city: %s: %v", city, err)
	}

	for _, dest := range destinationResponse.Data {
		if dest.SearchType == "city" {
			fmt.Printf("Found city: %s, Destination ID: %s\n", dest.Name, dest.DestID)
			return dest.DestID
		}
	}

	return ""
}

// insertDestinationID inserts the city and its destination ID into the city table in the database
func insertDestinationID(db *sql.DB, city City, destinationID string) error {
	_, err := db.Exec(`
		INSERT INTO city (city, country, destination_id) 
		VALUES (?, ?, ?)`,
		city.CityName, city.CountryCode, destinationID)
	return err
}

// createTables creates the necessary tables in the accommodation.db
func createTables(db *sql.DB) error {
	createCityTable := `CREATE TABLE IF NOT EXISTS city (
		city TEXT,
		country TEXT,
		destination_id TEXT
	);`
	_, err := db.Exec(createCityTable)
	return err
}
