package flightutils

import (
	"encoding/json"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/config_handlers"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
	"io/ioutil"
	"math"
	"net/http"
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
	departure_dates, err := timeutils.ListDatesBetween(origin.DepartureStartDate, origin.DepartureEndDate)
	fmt.Printf("\nReturn")
	return_dates, err := timeutils.ListDatesBetween(origin.ArrivalStartDate, origin.ArrivalEndDate)

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
