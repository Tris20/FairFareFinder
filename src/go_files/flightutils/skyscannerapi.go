package flightutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/config_handlers"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
	"github.com/Tris20/FairFareFinder/src/go_files/url_generators"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
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
}

func GetBestPrice(origin model.OriginInfo, destination model.DestinationInfo) (float64, error) {
	// Load API key from secrets.yaml
	var err error
	apiKey, err = config_handlers.LoadApiKey("ignore/secrets.yaml", "skyscanner")
	if err != nil {
		return 0, fmt.Errorf("error loading API key: %v", err)
	}
	fmt.Printf("\nDeparture")
	departure_dates, err := timeutils.ListDatesBetween(origin.NextDepartureStartDate, origin.NextDepartureEndDate)
	fmt.Printf("\nReturn")
	return_dates, err := timeutils.ListDatesBetween(origin.NextArrivalStartDate, origin.NextArrivalEndDate)

	fmt.Printf("\nGetting Departure Price")
	dep_price, err := GetBestPriceForGivenDates(origin.SkyScannerID, destination.SkyScannerID, departure_dates)
	fmt.Printf("\nGetting Return Price")
	return_price, err := GetBestPriceForGivenDates(destination.SkyScannerID, origin.SkyScannerID, return_dates)

	return (dep_price + return_price), err

	// price = (get_lowest_departure_price() + get_lowest_arrival_price())

	//return SearchOneWay(origin.SkyScannerID, destination.SkyScannerID )
}

func GetBestPriceForGivenDates(departureSkyScannerID string, arrivalSkyScannerID string, dates []string) (float64, error) {
	var lowestDayPrice float64 = math.MaxFloat64
	var err error

	for _, date := range dates {
		fmt.Printf("\n\nsearching %s", date)
		price, err := SearchOneWay(departureSkyScannerID, arrivalSkyScannerID, date)
		if err != nil {
			// Handle the error according to your error policy.
			// For example, you can return the error or continue to try other dates.
      fmt.Println("\nError fetching price for date:", date, "Error:", err)
			continue
		}

		if price < lowestDayPrice {
			lowestDayPrice = price
		}
	}

	// Check if lowestDayPrice was updated, return an error if not
	if lowestDayPrice == math.MaxFloat64 {
		lowestDayPrice = 0.0
		//	return 0, fmt.Errorf("no valid prices found")
	}
	fmt.Printf("\nLowest Price, %.2f", lowestDayPrice)
	return lowestDayPrice, err
}

func SearchOneWay(Departure_SkyScannerID string, Arrival_SkyScannerID string, date string) (float64, error) {
	url := fmt.Sprintf("https://skyscanner80.p.rapidapi.com/api/v1/flights/search-one-way?fromId=%s&toId=%s&departDate=%s&adults=1&currency=EUR&market=US&locale=en-US", Departure_SkyScannerID, Arrival_SkyScannerID, date)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", "skyscanner80.p.rapidapi.com")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	// Parse the JSON response and determine the best price
	return determineBestPriceFromResponse(body)
}

func determineBestPriceFromResponse(body []byte) (float64, error) {
	var response Response

	err := json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("error parsing JSON: %v", err)
	}

	if len(response.Data.Itineraries) == 0 {
		return 0, fmt.Errorf("no itineraries found")
	}

	// Assume the first itinerary has the best (lowest) price to start
	bestPrice := response.Data.Itineraries[0].Price.Raw
	for _, itinerary := range response.Data.Itineraries {
		if itinerary.Price.Raw < bestPrice {
			bestPrice = itinerary.Price.Raw
		}
	}

	return bestPrice, nil
}

/*

func GetFlightDates(origin, destination)
//Search flights.db for origin to destination W TH FR SA

//Search flights.db for origin to destination SU Mon Tue Wed


func sumcosts()

*/

func UpdateSkyscannerPrices(origins []model.OriginInfo) {
	// Open SQLite database
	db, err := sql.Open("sqlite3", "./data/flights.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	departureEndDate, err := time.Parse("2006-01-02", origins[0].DepartureEndDate)
	if time.Now().After(departureEndDate) {
		fmt.Printf("\n\n\n COPY WEEKEND")

		// SQL statement to copy "next_weekend" to "this_weekend"
		query := `UPDATE skyscannerprices SET this_weekend = next_weekend`

		// Execute the update query
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal("Failed to update the table:", err)
		}
		log.Println("Table updated successfully.")

	}

	// Update if entry exists
	updateStmt, err := db.Prepare("UPDATE skyscannerprices SET next_weekend = ? WHERE origin = ? AND destination = ?")
	if err != nil {
		log.Fatalf("Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	// Create if entry for dest AND origin does not exist
	insertStmt, err := db.Prepare("INSERT INTO skyscannerprices (origin, destination, next_weekend) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer insertStmt.Close()


	totalDestinations := calculateTotalDestinations(origins) // Function to sum up all destinations for all origins

	// Create a new progress bar
	bar := progressbar.Default(int64(totalDestinations))

	for _, origin := range origins {
		// Assume DetermineFlightsFromConfig and GenerateFlightsAndHotelsURLs are functions that return valid results
		airportDetailsList := flightdb.DetermineFlightsFromConfig(origin)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList)

		for _, destination := range destinationsWithUrls {
			price, err := GetBestPrice(origin, destination)
			if err != nil {
				log.Printf("Error getting best price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
bar.Add(1) // Increment progress bar even on error
				continue // Continue with the next destination if there's an error
			}

			// Execute the update statement for each origin-destination pair with the new price
			result, err := updateStmt.Exec(price, origin.SkyScannerID, destination.SkyScannerID)
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
				_, err = insertStmt.Exec(origin.SkyScannerID, destination.SkyScannerID, price)
				if err != nil {
					log.Printf("Failed to insert price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				} else {
					log.Printf("Successfully inserted price for %s to %s: €%.2f", origin.SkyScannerID, destination.SkyScannerID, price)
				}
			} else {
				log.Printf("Successfully updated price for %s to %s: €%.2f", origin.SkyScannerID, destination.SkyScannerID, price)
			}
bar.Add(1) // Increment progress bar even on error
		}
	}
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
func calculateTotalDestinations(origins  []model.OriginInfo) int {
	total := 0
	for _, origin := range origins {
		airportDetailsList := flightdb.DetermineFlightsFromConfig(origin)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList)
		total += len(destinationsWithUrls)
	}
	return total
}
