package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Tris20/FairFareFinder/config/handlers"
	"github.com/Tris20/FairFareFinder/src/backend/model"
	"github.com/Tris20/FairFareFinder/utils/data/process/generate/urls"
	"github.com/Tris20/FairFareFinder/utils/time-and-date"
	"github.com/schollz/progressbar/v3"
	"io/ioutil"
	"log"
	"math"
	"net/http"
  "sync"
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

var origins []model.OriginInfo

func main() {
	// Load IATA, skyscanenrID etc of origins(Berlin, Glasgow, Edi)
	originsConfig, _ := config_handlers.LoadOrigins("../../../../../config/origins.yaml")
	origins := config_handlers.ConvertConfigToModel(originsConfig)
	origins = update_origin_dates(origins)
	UpdateSkyscannerPrices(origins)
}

func GetBestPrice(origin model.OriginInfo, destination model.DestinationInfo) (float64, error) {
	// Load API key from secrets.yaml
	var err error
	apiKey, err = config_handlers.LoadApiKey("../../../../../ignore/secrets.yaml", "skyscanner")
	if err != nil {
		return 0, fmt.Errorf("error loading API key: %v", err)
	}

	println("HERE6\n")
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
	var (
		lowestDayPrice = math.MaxFloat64
		mu             sync.Mutex // Protects lowestDayPrice
		wg             sync.WaitGroup
	)

	for _, date := range dates {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			fmt.Printf("\n\nsearching %s", d)
			price, err := SearchOneWay(departureSkyScannerID, arrivalSkyScannerID, d)
			if err != nil {
				fmt.Println("\nError fetching price for date:", d, "Error:", err)
				return
			}
			// Safely update lowestDayPrice
			mu.Lock()
			if price < lowestDayPrice {
				lowestDayPrice = price
			}
			mu.Unlock()
		}(date)
	}

	wg.Wait() // Wait for all Goroutines to finish

	// Check if lowestDayPrice was updated
	if lowestDayPrice == math.MaxFloat64 {
		lowestDayPrice = 0.0
	}
	fmt.Printf("\nLowest Price, %.2f", lowestDayPrice)
	return lowestDayPrice, nil
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

if res.StatusCode == 429 {
	log.Println("Rate limit hit: Too many requests. Waiting to retry...")
	return 0, fmt.Errorf("rate limit exceeded: status %d", res.StatusCode)
} else if res.StatusCode >= 500 {
	log.Printf("Server error: %d. Retrying might help.", res.StatusCode)
	return 0, fmt.Errorf("server error: status %d", res.StatusCode)
}
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
	db, err := sql.Open("sqlite3", "../../../../../data/raw/flights/flights.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	fmt.Printf("\n\n\n COPY WEEKEND")

	// SQL statement to copy "next_weekend" to "this_weekend"
	query := `UPDATE skyscannerprices SET this_weekend = next_weekend`

	// Execute the update query
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Failed to update the table:", err)
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
	updateStmt, err := db.Prepare(`
    UPDATE skyscannerprices 
    SET next_weekend = ?, this_weekend = ? 
    WHERE origin_skyscanner_id = ? 
    AND destination_skyscanner_id = ?`)
	if err != nil {
		log.Fatalf("Failed to prepare update statement: %v", err)
	}
	defer updateStmt.Close()

	insertStmt, err := db.Prepare(`
    INSERT INTO skyscannerprices 
    (origin_city, origin_country, origin_iata, origin_skyscanner_id, destination_city, destination_country, destination_iata, destination_skyscanner_id, next_weekend, this_weekend) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatalf("Failed to prepare insert statement: %v", err)
	}
	defer insertStmt.Close()

	println("HERE4\n")
	totalDestinations := calculateTotalDestinations(origins) // Function to sum up all destinations for all origins

	// Create a new progress bar
	bar := progressbar.Default(int64(totalDestinations))

	for _, origin := range origins {
		// Assume DetermineFlightsFromConfig and GenerateFlightsAndHotelsURLs are functions that return valid results
		airportDetailsList := DetermineFlightsFromConfig(origin)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList)

		println("HERE5\n")
		for _, destination := range destinationsWithUrls {
			price, err := GetBestPrice(origin, destination)
			if err != nil {
				log.Printf("Error getting best price for %s to %s: %v", origin.SkyScannerID, destination.SkyScannerID, err)
				bar.Add(1) // Increment progress bar even on error
				continue   // Continue with the next destination if there's an error
			}

			// Execute the update statement for each origin-destination pair with the new price
			result, err := updateStmt.Exec(price, price, origin.SkyScannerID, destination.SkyScannerID)
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
					origin.IATA, origin.SkyScannerID, destination.City, destination.Country, destination.IATA, destination.SkyScannerID, price, price)
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
func calculateTotalDestinations(origins []model.OriginInfo) int {
	total := 0
	for _, origin := range origins {
		airportDetailsList := DetermineFlightsFromConfig(origin)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList)
		total += len(destinationsWithUrls)
	}
	return total
}

func update_origin_dates(origins []model.OriginInfo) []model.OriginInfo {

	for i := range origins {
		origins[i].DepartureStartDate, origins[i].DepartureEndDate, origins[i].ArrivalStartDate, origins[i].ArrivalEndDate = timeutils.CalculateWeekendRange(0)

		origins[i].NextDepartureStartDate, origins[i].NextDepartureEndDate, origins[i].NextArrivalStartDate, origins[i].NextArrivalEndDate = timeutils.CalculateWeekendRange(1)

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
