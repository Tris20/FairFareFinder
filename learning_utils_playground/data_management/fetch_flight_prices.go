package data_management

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/config_handlers"
	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/time_utils"
	"github.com/Tris20/FairFareFinder/src/backend/model" //import types

	// "github.com/Tris20/FairFareFinder/config/handlers"
	// "github.com/Tris20/FairFareFinder/utils/data/process/generate/urls"
	// "github.com/Tris20/FairFareFinder/utils/time-and-date"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

var apiKey string

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Itineraries []Itinerary `json:"itineraries"`
	} `json:"data"`
}

type Itinerary struct {
	Price struct {
		Raw       float64 `json:"raw"`
		Formatted string  `json:"formatted"`
	} `json:"price"`
	Legs []struct {
		DurationInMinutes int `json:"durationInMinutes"`
	} `json:"legs"`
}

// var origins []model.OriginInfo

func FetchFlightPrices(originsYamlPath string, flightsDBPath, locationsDBPath string) error {
	// Load IATA, skyscanenrID etc of origins(Berlin, Glasgow, Edi)
	originsConfig, _ := config_handlers.LoadOrigins(originsYamlPath)
	origins := config_handlers.ConvertConfigToModel(originsConfig)
	origins = update_origin_dates(origins)
	//"../../../../../data/raw/flights/flights.db"
	UpdateSkyscannerPrices(origins, flightsDBPath, locationsDBPath)
	return nil
}

func GetBestPrice(origin model.OriginInfo, destination model.DestinationInfo) (float64, int, error) {
	// Load API key from secrets.yaml
	var err error
	apiKey, err = config_handlers.LoadApiKey("../../../../../ignore/secrets.yaml", "skyscanner")
	if err != nil {
		return 0, 0, fmt.Errorf("error loading API key: %v", err)
	}

	departureDates, err := time_utils.ListDatesBetween(origin.NextDepartureStartDate, origin.NextDepartureEndDate)
	if err != nil {
		return 0, 0, fmt.Errorf("error generating departure dates: %v", err)
	}

	returnDates, err := time_utils.ListDatesBetween(origin.NextArrivalStartDate, origin.NextArrivalEndDate)
	if err != nil {
		return 0, 0, fmt.Errorf("error generating return dates: %v", err)
	}

	// Get departure price and duration
	depPrice, depDuration, err := GetBestPriceForGivenDates(origin.SkyScannerID, destination.SkyScannerID, departureDates)
	if err != nil {
		return 0, 0, err
	}

	// Get return price and duration
	returnPrice, returnDuration, err := GetBestPriceForGivenDates(destination.SkyScannerID, origin.SkyScannerID, returnDates)
	if err != nil {
		return 0, 0, err
	}
	// Total price and duration

	totalPrice := depPrice + returnPrice
	totalDuration := depDuration + returnDuration // Total round-trip duration

	return totalPrice, totalDuration, nil

	// price = (get_lowest_departure_price() + get_lowest_arrival_price())

	//return SearchOneWay(origin.SkyScannerID, destination.SkyScannerID )
}

func GetBestPriceForGivenDates(departureSkyScannerID string, arrivalSkyScannerID string, dates []string) (float64, int, error) {
	var lowestDayPrice float64 = math.MaxFloat64
	var err error
	var lowestDuration int = math.MaxInt
	for _, date := range dates {
		fmt.Printf("\n\nsearching %s", date)
		price, duration, err := SearchOneWay(departureSkyScannerID, arrivalSkyScannerID, date)
		if err != nil {
			// Handle the error according to your error policy.
			// For example, you can return the error or continue to try other dates.
			fmt.Println("\nError fetching price for date:", date, "Error:", err)
			continue
		}

		if price < lowestDayPrice {
			lowestDayPrice = price
		}
		if duration < lowestDuration {
			lowestDuration = duration
		}

	}

	// Check if lowestDayPrice was updated, return an error if not
	if lowestDayPrice == math.MaxFloat64 {
		lowestDayPrice = 0.0
		//	return 0, fmt.Errorf("no valid prices found")
	}

	// Check if lowestDayPrice was updated, return an error if not
	if lowestDuration == math.MaxInt {
		lowestDuration = 0
		//	return 0, fmt.Errorf("no valid prices found")
	}

	fmt.Printf("\nLowest Price, %.2f", lowestDayPrice)
	return lowestDayPrice, lowestDuration, err
}

func SearchOneWay(Departure_SkyScannerID string, Arrival_SkyScannerID string, date string) (float64, int, error) {
	url := fmt.Sprintf("https://skyscanner80.p.rapidapi.com/api/v1/flights/search-one-way?fromId=%s&toId=%s&departDate=%s&adults=1&currency=EUR&market=US&locale=en-US", Departure_SkyScannerID, Arrival_SkyScannerID, date)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", "skyscanner80.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	// Parse the JSON response and determine the best price and duration
	return determineBestPriceFromResponse(body)
}
func determineBestPriceFromResponse(body []byte) (float64, int, error) {
	var response Response

	err := json.Unmarshal(body, &response)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(response.Data.Itineraries) == 0 {
		return 0, 0, fmt.Errorf("no itineraries found")
	}

	bestPrice := response.Data.Itineraries[0].Price.Raw
	bestDuration := response.Data.Itineraries[0].Legs[0].DurationInMinutes

	for _, itinerary := range response.Data.Itineraries {
		if itinerary.Price.Raw < bestPrice {
			bestPrice = itinerary.Price.Raw
			bestDuration = itinerary.Legs[0].DurationInMinutes
		}
	}

	// Convert duration to the nearest hour
	durationInHours := int(math.Round(float64(bestDuration) / 60))

	return bestPrice, durationInHours, nil
}

/*

func GetFlightDates(origin, destination)
//Search flights.db for origin to destination W TH FR SA

//Search flights.db for origin to destination SU Mon Tue Wed


func sumcosts()

*/

func UpdateSkyscannerPrices(origins []model.OriginInfo, flightsDBPath, locationsDBPath string) error {
	// Open SQLite database
	flightDB, err := sql.Open("sqlite3", flightsDBPath)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}
	defer flightDB.Close()

	locationsDB, err := sql.Open("sqlite3", locationsDBPath)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}

	fmt.Printf("\n\n\n COPY WEEKEND")

	// todo: there needs to be a check to see if the table exists, if not, create it
	skyScannerPrice := db_manager.SkyScannerPrice{}
	_, err = flightDB.Exec(skyScannerPrice.CreateTableQuery())
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return err
	}

	// SQL statement to copy "next_weekend" to "this_weekend"
	query := `UPDATE skyscannerprices SET this_weekend = next_weekend`

	// Execute the update query
	_, err = flightDB.Exec(query)
	if err != nil {
		log.Printf("Failed to update the table: %v", err)
		return err
	}
	log.Println("Table updated successfully.")

	/*
		// Update if entry exists
		updateStmt, err := db.Prepare("UPDATE skyscannerprices SET next_weekend = ? WHERE origin_skyscanner_id = ? AND destination_skyscanner_id = ?")
		if err != nil {
			log.Fatalf("Failed to prepare update statement: %v", err)
		}
		defer updateStmt.Close()
		// Create if entry for dest AND origin does not exist
		insertStmt, err := db.Prepare("INSERT INTO skyscannerprices (origin_city, origin_country, origin_iata, origin_skyscanner_id, destination_city, destination_country, destination_iata, destination_skyscanner_id, next_weekend) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatalf("Failed to prepare insert statement: %v", err)
		}
		defer insertStmt.Close()
	*/

	// HOTFIX setting both this weekend and nextweekend to price value because we don't use both prices in the output table yet
	updateStmt, err := flightDB.Prepare(`
    UPDATE skyscannerprices 
    SET next_weekend = ?, this_weekend = ?, duration = ? 
    WHERE origin_skyscanner_id = ? 
    AND destination_skyscanner_id = ?`)
	if err != nil {
		log.Fatalf("Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	insertStmt, err := flightDB.Prepare(`
    INSERT INTO skyscannerprices 
    (origin_city, origin_country, origin_iata, origin_skyscanner_id, destination_city, destination_country, destination_iata, destination_skyscanner_id, next_weekend, this_weekend, duration) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer insertStmt.Close()

	println("HERE4\n")
	totalDestinations := calculateTotalDestinations(origins, flightDB, locationsDB) // Function to sum up all destinations for all origins

	// Create a new progress bar
	bar := progressbar.Default(int64(totalDestinations))

	for _, origin := range origins {
		// Assume DetermineFlightsFromConfig and GenerateFlightsAndHotelsURLs are functions that return valid results
		// todo: this also happens in calculateTotalDestinations
		airportDetailsList := DetermineFlightsFromConfig(origin, flightDB, locationsDB)
		destinationsWithUrls := GenerateFlightsAndHotelsURLs(origin, airportDetailsList)

		println("HERE5\n")
		for _, destination := range destinationsWithUrls {
			price, duration, err := GetBestPrice(origin, destination)
			if err != nil {
				log.Printf("Error getting best price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				bar.Add(1) // Increment progress bar even on error
				continue   // Continue with the next destination if there's an error
			}

			// Execute the update statement for each origin-destination pair with the new price
			result, err := updateStmt.Exec(price, price, duration, origin.SkyScannerID, destination.SkyScannerID)
			if err != nil {
				log.Printf("Failed to update price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				bar.Add(1) // Increment progress bar even on error
				continue
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				log.Printf("Error checking rows affected for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				bar.Add(1) // Increment progress bar even on error
				continue
			}

			// If no rows were updated, insert a new row
			if rowsAffected == 0 {
				_, err = insertStmt.Exec(origin.City, origin.Country,
					origin.IATA, origin.SkyScannerID, destination.City, destination.Country, destination.IATA, destination.SkyScannerID, price, price, duration)
				if err != nil {
					log.Printf("Failed to insert price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				} else {
					log.Printf("Successfully inserted price for %s(%s) to %s(%s): €%.2f", origin.IATA, origin.SkyScannerID, destination.IATA, destination.SkyScannerID, price)
				}
			} else {
				log.Printf("Successfully updated price for %s(%s) to %s(%s): €%.2f", origin.IATA, origin.SkyScannerID, destination.IATA, destination.SkyScannerID, price)

			}
			bar.Add(1) // Increment progress bar even on error
		}
	}
	return nil
}

// Function to get price for a given pair of skyscanner IDs
func GetPriceForRoute(db *sql.DB, weekend string, origin string, destination string) (float64, error) {
	var price float64

	query := fmt.Sprintf("SELECT %s FROM skyscannerprices WHERE origin = ? AND destination = ?", weekend)
	err := db.QueryRow(query, origin, destination).Scan(&price)
	if err != nil {
		return 0, err // Return 0 and the error
	}

	return price, nil // Return the found price and no error
}

// this determines the length of the progress bar
func calculateTotalDestinations(origins []model.OriginInfo, flightDB, airportsDB *sql.DB) int {
	total := 0
	for _, origin := range origins {
		airportDetailsList := DetermineFlightsFromConfig(origin, flightDB, airportsDB)
		destinationsWithUrls := GenerateFlightsAndHotelsURLs(origin, airportDetailsList)
		total += len(destinationsWithUrls)
	}
	return total
}

func update_origin_dates(origins []model.OriginInfo) []model.OriginInfo {

	for i := range origins {
		origins[i].DepartureStartDate, origins[i].DepartureEndDate, origins[i].ArrivalStartDate, origins[i].ArrivalEndDate = time_utils.CalculateWeekendRange(0)

		origins[i].NextDepartureStartDate, origins[i].NextDepartureEndDate, origins[i].NextArrivalStartDate, origins[i].NextArrivalEndDate = time_utils.CalculateWeekendRange(1)

		// Print updated origin info for verification
		// Print updated origin info for verification
		fmt.Printf("Origin #%d: %s\n", i+1, origins[i].City)
		fmt.Printf("  Upcoming Departure: %s to %s\n", origins[i].DepartureStartDate, origins[i].DepartureEndDate)
		fmt.Printf("  Upcoming Arrival: %s to %s\n", origins[i].ArrivalStartDate, origins[i].ArrivalEndDate)
		fmt.Printf("  Next Departure: %s to %s\n", origins[i].NextDepartureStartDate, origins[i].NextDepartureEndDate)
		fmt.Printf("  Next Arrival: %s to %s\n\n", origins[i].NextArrivalStartDate, origins[i].NextArrivalEndDate)

		//		fmt.Printf("Updated Origin: %+v\n", origins[i])

	}
	return origins
}
