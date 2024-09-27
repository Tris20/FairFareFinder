package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

type Property struct {
	HotelID            int      `json:"hotel_id"`
	AccessibilityLabel string   `json:"accessibilityLabel"`
	CountryCode        string   `json:"countryCode"`
	PhotoUrls          []string `json:"photoUrls"`
	IsPreferred        bool     `json:"isPreferred"`
	Longitude          float64  `json:"longitude"`
	Latitude           float64  `json:"latitude"`
	Name               string   `json:"name"`
	GrossPrice         float64  `json:"gross_price"`
	Currency           string   `json:"currency"`
	ReviewScore        float64  `json:"review_score"`
	ReviewCount        int      `json:"review_count"`
	CheckinDate        string   `json:"checkin_date"`
	CheckoutDate       string   `json:"checkout_date"`
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

type APIResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Hotels []Hotel `json:"hotels"`
		Meta   []Meta  `json:"meta"`
	} `json:"data"`
}

type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
}

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

	// Accept a destinationID as a command-line argument
	var startDestID string
	flag.StringVar(&startDestID, "ID", "", "Start fetching properties from this destination ID")
	flag.Parse()

	// Step 1: Look up city names from the 'locations' database
	cities, err := getCityNamesAndDestinationIDs()
	if err != nil {
		log.Fatalf("Error retrieving city names: %v", err)
	}

	// Step 2: Create a new SQLite database for accommodation and property information
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/accommocation/booking-com/booking.db")
	if err != nil {
		log.Fatalf("Error creating accommodation.db: %v", err)
	}
	defer db.Close()

	// Create city and property tables
	err = createPropertyTable(db)
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// Read the API key once at the beginning
	apiKey, err := readAPIKey("../../../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
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
}

func getCityNamesAndDestinationIDs() ([]City, error) {
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/accommocation/booking-com/booking.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT city, country, destination_id FROM city WHERE destination_id IS NOT NULL AND destination_id != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []City
	for rows.Next() {
		var city City
		err := rows.Scan(&city.CityName, &city.CountryCode, &city.DestinationID)
		if err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, nil
}

func createPropertyTable(db *sql.DB) error {
	createPropertyTable := `CREATE TABLE IF NOT EXISTS property (
		hotel_id INTEGER,
		accessibility_label TEXT,
		country_code TEXT,
		photo_urls TEXT,
		is_preferred BOOLEAN,
		longitude REAL,
		latitude REAL,
		name TEXT,
		gross_price REAL,
		currency TEXT,
		review_score REAL,
		review_count INTEGER,
		checkin_date TEXT,
		checkout_date TEXT,
		city TEXT,
		country TEXT
	);`
	_, err := db.Exec(createPropertyTable)
	return err
}

func fetchWithRetry(url string, apiKey string) (*http.Response, error) {
	client := &http.Client{}
	var resp *http.Response
	var err error

	for i := 0; i < 13; i++ { // Try a maximum of 3 times
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("x-rapidapi-host", "booking-com15.p.rapidapi.com")
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

// fetchPropertyData now takes the dynamic Wednesday-to-Wednesday date range into account
func fetchPropertyData(destinationID, apiKey string) ([]Property, error) {
	// Get the dynamic Wednesday-to-Wednesday date range
	arrivalDate, departureDate := getWednesdayRange()

	// Use the date range in the API URL
	apiURL := fmt.Sprintf("https://booking-com15.p.rapidapi.com/api/v1/hotels/searchHotels?dest_id=%s&search_type=CITY&arrival_date=%s&departure_date=%s&adults=1&children_age=0,17&room_qty=1&page_number=1&units=metric&temperature_unit=c&languagecode=en-us&currency_code=EUR", destinationID, arrivalDate, departureDate)

	fmt.Println("API URL with dynamic date range:", apiURL)

	resp, err := fetchWithRetry(apiURL, apiKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	var properties []Property
	for _, hotel := range apiResponse.Data.Hotels {
		property := Property{
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

// insertProperties remains unchanged as it inserts data into the database
func insertProperties(db *sql.DB, properties []Property, city City) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO property (hotel_id, accessibility_label, country_code, photo_urls, is_preferred, longitude, latitude, name, gross_price, currency, review_score, review_count, checkin_date, checkout_date, city, country)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, property := range properties {
		photoUrls := strings.Join(property.PhotoUrls, ",")

		fmt.Printf("Inserting Property: %s, City: %s, Price: %.2f %s\n",
			property.Name, city.CityName, property.GrossPrice, property.Currency)

		_, err := stmt.Exec(
			property.HotelID, property.AccessibilityLabel, property.CountryCode, photoUrls, property.IsPreferred,
			property.Longitude, property.Latitude, property.Name, property.GrossPrice, property.Currency, property.ReviewScore,
			property.ReviewCount, property.CheckinDate, property.CheckoutDate, city.CityName, city.CountryCode,
		)
		if err != nil {
			log.Printf("Error inserting property %s: %v", property.Name, err)
			continue
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
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

	var apiResponse APIResponse
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
func fetchPropertiesByPage(destinationID, apiKey string, pageNumber int) ([]Property, error) {
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

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	var properties []Property
	for _, hotel := range apiResponse.Data.Hotels {
		property := Property{
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

func processCityProperties(destinationID string, db *sql.DB, apiKey string, city City) error {
	// Fetch the total number of properties for the destination
    totalProperties, err := fetchTotalProperties(destinationID, apiKey)
    if err != nil {
        log.Printf("Error fetching total properties for %s: %v", city.CityName, err)
        return err
    }	// Calculate the number of pages (each page has 20 properties)

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
		err = insertProperties(db, properties, city)
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
