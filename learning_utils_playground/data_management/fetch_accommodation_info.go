package data_management

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

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

type PriceBreakdown struct {
	GrossPrice struct {
		Value    float64 `json:"value"`
		Currency string  `json:"currency"`
	} `json:"grossPrice"`
	StrikethroughPrice struct {
		Value    float64 `json:"value"`
		Currency string  `json:"currency"`
	} `json:"strikethroughPrice"`
	BenefitBadges []interface{} `json:"benefitBadges"`
}

type Hotel struct {
	HotelID  int `json:"hotel_id"`
	Property struct {
		CountryCode    string         `json:"countryCode"`
		Longitude      float64        `json:"longitude"`
		Latitude       float64        `json:"latitude"`
		Name           string         `json:"name"`
		PriceBreakdown PriceBreakdown `json:"priceBreakdown"`
		ReviewScore    float64        `json:"reviewScore"`
		ReviewCount    int            `json:"reviewCount"`
		CheckinDate    string         `json:"checkinDate"`
		CheckoutDate   string         `json:"checkoutDate"`
		PhotoUrls      []string       `json:"photoUrls"`
		IsPreferred    bool           `json:"isPreferred"`
		Currency       string         `json:"currency"`
	} `json:"property"`
}

type Meta struct {
	Title string `json:"title"`
}

func getDestinationID_BookingCom(db *sql.DB) {
	// expects booking.db???
	// creates the city table if it doesn't exist
	cityFetch := db_manager.CityFetch{}
	_, err := db.Exec(cityFetch.CreateTableQuery())
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// expects locations.db
	// Step 1: Look up city names from the 'locations' database
	cities, err := getCityNames(db)
	if err != nil {
		log.Fatalf("Error retrieving city names: %v", err)
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

// what is the structure of the locations.db?
// getCityNames fetches cities from the locations.db where include_tf == 1
func getCityNames(db *sql.DB) ([]db_manager.CityFetch, error) {
	rows, err := db.Query(`SELECT city, iso2 FROM city WHERE include_tf = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []db_manager.CityFetch
	for rows.Next() {
		var city db_manager.CityFetch
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

	body, err := io.ReadAll(resp.Body)
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
func insertDestinationID(db *sql.DB, city db_manager.CityFetch, destinationID string) error {
	_, err := db.Exec(`
		INSERT INTO city (city, country, destination_id) 
		VALUES (?, ?, ?)`,
		city.CityName, city.CountryCode, destinationID)
	return err
}

// Input is startDestID, which is the destination ID to start fetching properties from
func GetProperties_BookingCom(startDestID, bookingDBPath, secretsPath string) error {
	db, err := sql.Open("sqlite3", bookingDBPath)
	if err != nil {
		log.Printf("Error creating accommodation.db: %v", err)
		return err
	}
	defer db.Close()

	// creates the property table if it doesn't exist
	propertyFetch := db_manager.PropertyFetch{}
	_, err = db.Exec(propertyFetch.CreateTableQuery())
	if err != nil {
		log.Printf("Error creating tables: %v", err)
		return err
	}

	// Step 1: Look up city names from the 'locations' database
	cities, err := getCityNamesAndDestinationIDs(db)
	if err != nil {
		log.Printf("Error retrieving city names: %v", err)
		return err
	}

	// Read the API key once at the beginning
	apiKey, err := readAPIKey(secretsPath)
	if err != nil {
		log.Printf("Error reading API key: %v", err)
		return err
	}

	start := false
	if startDestID == "" {
		start = true // Start from the beginning if no startDestID provided
	}

	// Step 3: Fetch data from the API and insert it into the database with a progress bar
	bar := progressbar.Default(int64(len(cities)), "Fetching data for cities")
	for _, city := range cities {

		if strings.TrimSpace(city.DestinationID) == strings.TrimSpace(startDestID) {
			start = true // Start processing from this city
		}
		if !start {
			bar.Add(1)
			continue // Skip until we reach the starting city
		}

		bar.Add(1)

		fmt.Printf("Fetching property data for city: %s (%s)\n", city.CityName, city.CountryCode)

		err = processCityProperties(city.DestinationID, db, apiKey, city)
		if err != nil {
			log.Printf("Error processing properties for city %s: %v", city.CityName, err)
		}
	}
	return nil
}

// fetchTotalProperties updated to use dynamic date range
func fetchTotalProperties(destinationID, apiKey string) (int, error) {
	arrivalDate, departureDate := getWednesdayRange()

	apiURL := fmt.Sprintf("https://booking-com15.p.rapidapi.com/api/v1/hotels/searchHotels?dest_id=%s&search_type=CITY&arrival_date=%s&departure_date=%s&adults=1&children_age=0,17&room_qty=1&page_number=1&units=metric&temperature_unit=c&languagecode=en-us&currency_code=EUR", destinationID, arrivalDate, departureDate)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("x-rapidapi-host", "booking-com15.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var apiResponse APIResponse_Accomodation
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return 0, err
	}

	var totalProperties int
	if len(apiResponse.Data.Meta) > 0 {
		totalPropertiesStr := apiResponse.Data.Meta[0].Title
		totalProperties, err = strconv.Atoi(strings.Fields(totalPropertiesStr)[0])
		if err != nil {
			return 0, fmt.Errorf("failed to parse total properties: %v", err)
		}
	} else {
		// Handle the case where no meta data is available
		return 0, nil
	}

	return totalProperties, nil
}

// fetchPropertiesByPage updated to use dynamic date range
func fetchPropertiesByPage(destinationID, apiKey string, pageNumber int) ([]db_manager.PropertyFetch, error) {
	arrivalDate, departureDate := getWednesdayRange()

	apiURL := fmt.Sprintf("https://booking-com15.p.rapidapi.com/api/v1/hotels/searchHotels?dest_id=%s&search_type=CITY&arrival_date=%s&departure_date=%s&adults=1&children_age=0,17&room_qty=1&page_number=%d&units=metric&temperature_unit=c&languagecode=en-us&currency_code=EUR", destinationID, arrivalDate, departureDate, pageNumber)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-rapidapi-host", "booking-com15.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse_Accomodation
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	var properties []db_manager.PropertyFetch
	for _, hotel := range apiResponse.Data.Hotels {
		property := db_manager.PropertyFetch{
			HotelID:            hotel.HotelID,
			AccessibilityLabel: hotel.Property.Name,
			CountryCode:        hotel.Property.CountryCode,
			PhotoUrls:          hotel.Property.PhotoUrls,
			IsPreferred:        hotel.Property.IsPreferred,
			Longitude:          hotel.Property.Longitude,
			Latitude:           hotel.Property.Latitude,
			Name:               hotel.Property.Name,
			GrossPrice:         hotel.Property.PriceBreakdown.GrossPrice.Value,
			Currency:           hotel.Property.PriceBreakdown.GrossPrice.Currency,
			ReviewScore:        hotel.Property.ReviewScore,
			ReviewCount:        hotel.Property.ReviewCount,
			CheckinDate:        hotel.Property.CheckinDate,
			CheckoutDate:       hotel.Property.CheckoutDate,
		}
		properties = append(properties, property)
	}

	return properties, nil
}

func processCityProperties(destinationID string, db *sql.DB, apiKey string, city db_manager.CityFetch) error {
	// Fetch the total number of properties for the destination
	totalProperties, err := fetchTotalProperties(destinationID, apiKey)
	if err != nil {
		log.Printf("Error fetching total properties for %s: %v", city.CityName, err)
		return err
	} // Calculate the number of pages (each page has 20 properties)

	totalPages := totalProperties / 20
	if totalProperties%20 != 0 {
		totalPages++ // Add an extra page if there's a remainder
	}

	if totalPages >= 5 {
		totalPages = 5
	}

	fmt.Printf("Total Properties: %d, Total Pages: %d\n", totalProperties, totalPages)

	// Loop through all the pages and fetch properties for each page
	for page := 1; page <= totalPages; page++ {
		fmt.Printf("Fetching page %d for city %s\n", page, city.CityName)

		properties, err := fetchPropertiesByPage(destinationID, apiKey, page)
		if err != nil {
			fmt.Printf("Error fetching properties for page %d: %v\n", page, err)
			continue
		}

		// Insert properties into the database
		err = db_manager.InsertProperties(db, properties, city)
		if err != nil {
			fmt.Printf("Error inserting properties for page %d: %v\n", page, err)
			continue
		}
	}

	return nil
}

// getWednesdayRange calculates "this Wednesday" and "next Wednesday" in YYYY-MM-DD format
func getWednesdayRange() (string, string) {
	// Get current date
	now := time.Now()

	// Calculate days until this week's Wednesday
	thisWednesdayOffset := (3 - int(now.Weekday()) + 7) % 7 // 3 corresponds to Wednesday (0 = Sunday)
	// Calculate this Wednesday date
	thisWednesday := now.AddDate(0, 0, thisWednesdayOffset)
	// Calculate next Wednesday (7 days after this Wednesday)
	nextWednesday := thisWednesday.AddDate(0, 0, 7)

	// Format dates to "YYYY-MM-DD"
	thisWednesdayStr := thisWednesday.Format("2006-01-02")
	nextWednesdayStr := nextWednesday.Format("2006-01-02")

	return thisWednesdayStr, nextWednesdayStr
}

// airbnb could also be included here

////// UNUSED CODE

// fetchPropertyData now takes the dynamic Wednesday-to-Wednesday date range into account
func fetchPropertyData(destinationID, apiKey string) ([]db_manager.PropertyFetch, error) {
	// Get the dynamic Wednesday-to-Wednesday date range
	arrivalDate, departureDate := getWednesdayRange()

	// Use the date range in the API URL
	apiURL := fmt.Sprintf("https://booking-com15.p.rapidapi.com/api/v1/hotels/searchHotels?dest_id=%s&search_type=CITY&arrival_date=%s&departure_date=%s&adults=1&children_age=0,17&room_qty=1&page_number=1&units=metric&temperature_unit=c&languagecode=en-us&currency_code=EUR", destinationID, arrivalDate, departureDate)

	fmt.Println("API URL with dynamic date range:", apiURL)

	urlBase := "booking-com15.p.rapidapi.com"

	resp, err := fetchWithRetry(apiURL, urlBase, apiKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse_Accomodation
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	var properties []db_manager.PropertyFetch
	for _, hotel := range apiResponse.Data.Hotels {
		property := db_manager.PropertyFetch{
			HotelID:            hotel.HotelID,
			AccessibilityLabel: hotel.Property.Name,
			CountryCode:        hotel.Property.CountryCode,
			PhotoUrls:          hotel.Property.PhotoUrls,
			IsPreferred:        hotel.Property.IsPreferred,
			Longitude:          hotel.Property.Longitude,
			Latitude:           hotel.Property.Latitude,
			Name:               hotel.Property.Name,
			GrossPrice:         hotel.Property.PriceBreakdown.GrossPrice.Value,
			Currency:           hotel.Property.PriceBreakdown.GrossPrice.Currency,
			ReviewScore:        hotel.Property.ReviewScore,
			ReviewCount:        hotel.Property.ReviewCount,
			CheckinDate:        hotel.Property.CheckinDate,
			CheckoutDate:       hotel.Property.CheckoutDate,
		}
		properties = append(properties, property)
	}

	return properties, nil
}
