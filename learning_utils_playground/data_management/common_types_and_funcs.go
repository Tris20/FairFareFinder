package data_management

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	"gopkg.in/yaml.v2"
)

// Secrets struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
}

// ApiResponse defines the structure to parse the JSON response
type ApiResponse_Airports struct {
	Data []struct {
		Id string `json:"id"`
	} `json:"data"`
}

type APIResponse_Accomodation struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Hotels []Hotel `json:"hotels"`
		Meta   []Meta  `json:"meta"`
	} `json:"data"`
}

// Assuming a part of the JSON response structure from OpenWeatherMap API for simplification
type ApiResponse_Weather struct {
	List []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Main string `json:"main"`
			Icon string `json:"icon"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	} `json:"list"`
}

// FlightRequest represents the structure of the YAML input.
type FlightRequest struct {
	Flights []FlightCriteria `yaml:"flights"`
}

// FlightCriteria represents each flight's criteria within the YAML input.
type FlightCriteria struct {
	Direction string `yaml:"direction"`
	Airport   string `yaml:"airport"`
	StartDate string `yaml:"startDate"`
	EndDate   string `yaml:"endDate"`
}

// readAPIKey reads the API key from a YAML file
func readAPIKey(filepath string) (string, error) {
	var secrets Secrets
	file, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &secrets)
	if err != nil {
		return "", err
	}
	return secrets.APIKeys.Aerodatabox, nil
}

// CopyFile copies a file from source to destination
func CopyFile(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err == nil {
		fmt.Printf("Copied file from %s to %s\n", source, dest)
	}

	return err
}

// isImage checks if a file is an image based on its extension
func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// executeQueryForAirports executes a given SQL query and returns a set of airports.
func executeQueryForAirports(db *sql.DB, query string) (map[string]bool, error) {
	airports := make(map[string]bool)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var airport string
		if err := rows.Scan(&airport); err != nil {
			return nil, err
		}
		airports[airport] = true
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return airports, nil
}

// intersectSets finds the intersection of an array of sets.
func intersectSets(sets []map[string]bool) []string {
	intersection := make([]string, 0)
	if len(sets) == 0 {
		return intersection
	}

	// Initialize intersection with the first set's elements.
	for item := range sets[0] {
		intersection = append(intersection, item)
	}

	// Intersect with remaining sets.
	for _, set := range sets[1:] {
		temp := intersection[:0] // reuse the existing slice but start filling from the beginning
		for _, item := range intersection {
			if set[item] {
				temp = append(temp, item)
			}
		}
		intersection = temp
	}

	return intersection
}

func DetermineFlightsFromConfig(origin model.OriginInfo, flightDB, locationsDB *sql.DB) []model.DestinationInfo {

	// Define your queries here.
	queries := []string{
		fmt.Sprintf("SELECT arrivalAirport FROM schedule WHERE departureAirport = '%s' AND departureTime BETWEEN '%s' AND '%s'", origin.IATA, origin.DepartureStartDate, origin.DepartureEndDate),
		fmt.Sprintf("SELECT departureAirport FROM schedule WHERE arrivalAirport = '%s' AND arrivalTime BETWEEN '%s' AND '%s'", origin.IATA, origin.ArrivalStartDate, origin.ArrivalEndDate),
		//"SELECT arrivalAirport FROM flights WHERE departureAirport = 'EDI' AND departureTime BETWEEN '2024-03-20' AND '2024-03-22'",
		//"SELECT departureAirport FROM flights WHERE arrivalAirport = 'EDI' AND arrivalTime BETWEEN '2024-03-24' AND '2024-03-26'",

		// Add your third, fourth, ... queries here.
	}

	// Execute all queries and collect their results in a slice of sets.
	var sets []map[string]bool
	for _, query := range queries {
		//fmt.Printf(query)
		airports, err := executeQueryForAirports(flightDB, query)
		// fmt.Printf("AIRPOTS %s", airports)
		if err != nil {
			log.Fatal("Error executing query:", err)
		}
		sets = append(sets, airports)
	}

	// Find the intersection of all sets.
	intersection := intersectSets(sets)

	fmt.Println("Airports meeting all conditions:", intersection)
	airportDetailsList := buildAirportDetails(locationsDB, intersection)
	for _, airportInfo := range airportDetailsList {
		fmt.Printf("%s: %s, %s\n", airportInfo.IATA, airportInfo.City, airportInfo.Country)
	}
	return airportDetailsList
}

// printAirportDetails prints the details for each airport IATA code.
func buildAirportDetails(db *sql.DB, iataCodes []string) []model.DestinationInfo {
	var airportDetailsList []model.DestinationInfo
	for _, iata := range iataCodes {
		//skip empty or blank IATA codes
		if iata == "" {
			continue
		}
		city, country, skyscannerid, err := fetchAirportDetails(db, iata)
		if err != nil {
			log.Printf("Error fetching details for IATA %s: %v", iata, err)
			continue
		}
		// Append the fetched details to the list
		airportDetailsList = append(airportDetailsList, model.DestinationInfo{
			IATA:         iata,
			City:         city,
			Country:      country,
			SkyScannerID: skyscannerid,
		})
	}
	return airportDetailsList
}

// fetchAirportDetails executes a query to fetch city and country for a given IATA code.
func fetchAirportDetails(db *sql.DB, iataCode string) (string, string, string, error) {
	var city, country, skyscannerid string
	query := "SELECT city, country, skyscannerid FROM airport_info WHERE iata = ?"
	err := db.QueryRow(query, iataCode).Scan(&city, &country, &skyscannerid)
	if err != nil {
		return "", "", "", err
	}
	return city, country, skyscannerid, nil
}

func fetchWithRetry(url, urlBase, apiKey string) (*http.Response, error) {
	client := &http.Client{}
	var resp *http.Response
	var err error

	for i := 0; i < 13; i++ { // Try a maximum of 3 times
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("x-rapidapi-host", urlBase)
		req.Header.Add("x-rapidapi-key", apiKey)

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil // Success
		}

		if resp != nil {
			resp.Body.Close() // Close the previous response body
		}

		log.Printf("Request failed: %v. Retrying...", err)
		time.Sleep(time.Duration(2^(i+1)) * time.Second) // Exponential backoff
	}

	return nil, err // Return the last error
}

func getCityNamesAndDestinationIDs(db *sql.DB) ([]db_manager.CityFetch, error) {
	err := db_manager.CreateTable(db, &db_manager.CityFetch{})
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(`SELECT city, country, destination_id FROM city 
	WHERE destination_id IS NOT NULL AND destination_id != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []db_manager.CityFetch
	for rows.Next() {
		var city db_manager.CityFetch
		err := rows.Scan(&city.CityName, &city.CountryCode, &city.DestinationID)
		if err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, nil
}
