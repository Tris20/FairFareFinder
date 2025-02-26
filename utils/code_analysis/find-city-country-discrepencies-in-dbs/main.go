package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

// Config holds the YAML configuration with a list of airport IATA codes.
type Config struct {
	Airports []string `yaml:"airports"`
}

// loadConfig reads and decodes the YAML configuration file.
func loadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// fileExists checks if the file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// City represents a canonical city record from locations.db.
type City struct {
	Name    string // from city_ascii
	Country string // from iso2 (e.g., "DE")
}

func main() {
	// Load configuration from config.yaml.
	cfg, err := loadConfig("../../../config/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	fmt.Printf("Loaded %d airports from config.\n\n", len(cfg.Airports))

	// Define file paths.
	locPath := "../../../data/raw/locations/locations.db"
	weatherPath := "../../../data/raw/weather/weather.db"
	bookingPath := "../../../data/raw/accommocation/booking-com/booking.db"

	// Check that each database file exists.
	paths := map[string]string{
		"locations.db": locPath,
		"weather.db":   weatherPath,
		"booking.db":   bookingPath,
	}
	for name, path := range paths {
		if !fileExists(path) {
			log.Fatalf("%s not found at path: %s", name, path)
		}
	}

	// Open locations.db (contains both the canonical city table and the airport table).
	locDB, err := sql.Open("sqlite3", locPath)
	if err != nil {
		log.Fatalf("failed to open locations.db: %v", err)
	}
	defer locDB.Close()

	// Retrieve canonical cities from locations.db table "city" (using city_ascii and iso2).
	cityRows, err := locDB.Query("SELECT city_ascii, iso2 FROM city")
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	defer cityRows.Close()

	var cities []City
	for cityRows.Next() {
		var cityName, iso2 sql.NullString
		if err := cityRows.Scan(&cityName, &iso2); err != nil {
			log.Printf("failed to scan row: %v", err)
			continue
		}
		// Skip rows with NULL values in either column.
		if !cityName.Valid || !iso2.Valid {
			log.Printf("skipping row with NULL value(s): city_ascii=%v, iso2=%v", cityName, iso2)
			continue
		}
		cities = append(cities, City{Name: cityName.String, Country: iso2.String})
	}

	// Open weather.db (table "weather": columns city_name and country_code).
	weatherDB, err := sql.Open("sqlite3", weatherPath)
	if err != nil {
		log.Fatalf("failed to open weather.db: %v", err)
	}
	defer weatherDB.Close()

	// Open booking.db (table "city": columns city and country).
	bookingDB, err := sql.Open("sqlite3", bookingPath)
	if err != nil {
		log.Fatalf("failed to open booking.db: %v", err)
	}
	defer bookingDB.Close()

	// Prepare lookup statement for the weather database.
	weatherStmt, err := weatherDB.Prepare("SELECT 1 FROM weather WHERE city_name = ? AND country_code = ? LIMIT 1")
	if err != nil {
		log.Fatalf("failed to prepare weather statement: %v", err)
	}
	defer weatherStmt.Close()

	// Prepare lookup statement for the booking database.
	bookingStmt, err := bookingDB.Prepare("SELECT 1 FROM city WHERE city = ? AND country = ? LIMIT 1")
	if err != nil {
		log.Fatalf("failed to prepare booking statement: %v", err)
	}
	defer bookingStmt.Close()

	// Build dynamic query for the airport lookup using the IATA codes from config.
	// We assume the "airport" table has an "iata" column.
	placeholders := make([]string, len(cfg.Airports))
	for i := range cfg.Airports {
		placeholders[i] = "?"
	}
	airportIataQuery := fmt.Sprintf(
		"SELECT GROUP_CONCAT(iata, ',') FROM airport WHERE city = ? AND country = ? AND iata IN (%s)",
		strings.Join(placeholders, ","),
	)
	airportIataStmt, err := locDB.Prepare(airportIataQuery)
	if err != nil {
		log.Fatalf("failed to prepare airport iata statement: %v", err)
	}
	defer airportIataStmt.Close()

	// Output header.
	fmt.Println("City, Country         | IATA             | In Weather DB | In Booking DB")
	fmt.Println("--------------------------------------------------------------------------")

	// For each canonical city, check for a matching airport record first.
	for _, c := range cities {
		// Prepare parameters for the airport lookup:
		// first two parameters are city and country, followed by all IATA codes from config.
		args := []interface{}{c.Name, c.Country}
		for _, code := range cfg.Airports {
			args = append(args, code)
		}

		// Query the airport table to get matching IATA codes.
		var matchingIATAs sql.NullString
		err = airportIataStmt.QueryRow(args...).Scan(&matchingIATAs)
		if err != nil || !matchingIATAs.Valid || matchingIATAs.String == "" {
			// No matching airport with an IATA code from config found, skip this city.
			continue
		}

		// Now check the weather and booking databases.
		var weatherFound, bookingFound bool
		var tmp int

		// Check weather database.
		err = weatherStmt.QueryRow(c.Name, c.Country).Scan(&tmp)
		if err == nil {
			weatherFound = true
		}

		// Check booking database.
		err = bookingStmt.QueryRow(c.Name, c.Country).Scan(&tmp)
		if err == nil {
			bookingFound = true
		}

		// Print result for the city with the matching IATA codes.
		fmt.Printf("%-22s %-7s | %-16s | %-13t | %-13t\n", c.Name, c.Country, matchingIATAs.String, weatherFound, bookingFound)
	}
}
