package main

import (
	"database/sql"
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/config_handlers"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/flightutils"
	"github.com/Tris20/FairFareFinder/src/go_files/server"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
	"github.com/Tris20/FairFareFinder/src/go_files/url_generators"
	"github.com/Tris20/FairFareFinder/src/go_files/weather_pleasantness"
	"github.com/Tris20/FairFareFinder/src/go_files/web_pages/html_generators"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"
)

type Favourites struct {
	Locations []string `yaml:"locations"`
}

type CityAverageWPI struct {
	Name          string
	WPI           float64
	SkyScannerURL string
	AirbnbURL     string
	BookingURL    string
}

var checkFlightPrices = false
var checkprice_init = false

var origins []model.OriginInfo

func main() {
	user_db.Setup_database()
	dbPath := "user_database.db"
	if len(os.Args) < 2 {
		log.Fatal("Error: No argument provided. Please provide a location, 'web', or a json file.")
	}
	// Load IATA, skyscanenrID etc of origins(Berlin, Glasgow, Edi)
	originsConfig, _ := config_handlers.LoadOrigins("input/origins.yaml")
	origins := config_handlers.ConvertConfigToModel(originsConfig)
	// Update dates
	origins = update_origin_dates(origins)

	switch os.Args[1] {
	case "dev":

		origins = update_origin_dates(origins)
		ProcessOriginConfigurations(origins)
		fmt.Println("\nStarting Webserver")
		fffwebserver.SetupFFFWebServer()

	case "updateSkyPrices":
		flightutils.UpdateSkyscannerPrices(origins)

	case "web":
		fmt.Printf("WEB")

		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {
				origins = update_origin_dates(origins)
				ProcessOriginConfigurations(origins)
			}
		}()
		fffwebserver.SetupFFFWebServer()
		// Start a goroutine to check and execute a task every Monday

	case "init-db":
		user_db.Init_database(dbPath)
		user_db.Insert_test_user(dbPath)

	default:
		// Check if the argument is a json file
		if strings.HasSuffix(os.Args[1], ".json") {
			out := fmt.Sprintf("input/%s-flights.json", os.Args[1:])
			fmt.Sprintf(out)
			//      GenerateAndPostCityRankings(os.Args[1], out)
		} else {
			// Assuming it's a city name

			var location model.DestinationInfo
			location.City = strings.Join(os.Args[1:], " ")
			location.Country = ("")
			weather_pleasantry.ProcessLocation(location)
		}
	}
}

// GenerateCityRankings processes destinations and updates them with various info including weather and flight prices.
func GenerateCityRankings(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo) {
	// Open SQLite database
	db, err := sql.Open("sqlite3", "./data/flights.db")
	if err != nil {
		log.Fatalf("Failed to open flights.db: %v", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	for i, destination := range destinationsWithUrls {
		// Process location for weather pleasantry index (WPI)
		wpi, dailyDetails := weather_pleasantry.ProcessLocation(destination)
		if math.IsNaN(wpi) {
			log.Printf("WPI calculation returned NaN for destination %v", destination)
			continue
		}
		
		destinationsWithUrls[i].WPI = wpi // Update the WPI in the destination info
		log.Printf("Updated WPI for destination %v: %f", destination, wpi)

		// Read price data from the database table
		price, err := flightutils.GetPriceForRoute(db, origin.SkyScannerID, destination.SkyScannerID)
		if err != nil {
			log.Printf("Failed to get price for route from %v to %v: %v", origin.SkyScannerID, destination.SkyScannerID, err)
			continue
		}
		
		destinationsWithUrls[i].SkyScannerPrice = price
		log.Printf("Retrieved SkyScanner price for %v -> %v: %.2f", origin.SkyScannerID, destination.SkyScannerID, price)

		// Update URLs with URL encoding
		destinationsWithUrls[i].SkyScannerURL = replaceSpaceWithURLEncoding(destination.SkyScannerURL)
		destinationsWithUrls[i].AirbnbURL = replaceSpaceWithURLEncoding(destination.AirbnbURL)
		destinationsWithUrls[i].BookingURL = replaceSpaceWithURLEncoding(destination.BookingURL)

		// Transfer daily weather details
		var weatherDetailsSlice []model.DailyWeatherDetails
		for _, details := range dailyDetails {
			weatherDetailsSlice = append(weatherDetailsSlice, details)
		}
		destinationsWithUrls[i].WeatherDetails = weatherDetailsSlice
	}

	// Sort the cities by WPI in descending order
	sort.Slice(destinationsWithUrls, func(i, j int) bool {
		return destinationsWithUrls[i].WPI > destinationsWithUrls[j].WPI
	})
	log.Println("Sorted destinations by WPI in descending order.")

	generate_html_table(origin, destinationsWithUrls)
	log.Println("Generated HTML table for city rankings.")
}
// replaceSpaceWithURLEncoding replaces space characters with %20 in the URL
func replaceSpaceWithURLEncoding(urlString string) string {
	return strings.ReplaceAll(urlString, " ", "%20")
}

func generate_html_table(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo) {

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	target_url := fmt.Sprintf("src/html/%s-flight-destinations.html", strings.ToLower(origin.City))

	err := htmltablegenerator.GenerateHtmlTable(target_url, destinationsWithUrls)
	if err != nil {
		log.Fatalf("Failed to convert markdown to HTML: %v", err)
	}
}

func update_origin_dates(origins []model.OriginInfo) []model.OriginInfo {

	for i := range origins {
		origins[i].DepartureStartDate, origins[i].DepartureEndDate = timeutils.UpcomingWedToSat()
		origins[i].ArrivalStartDate, origins[i].ArrivalEndDate = timeutils.UpcomingSunToWedFromSat(origins[i].DepartureEndDate)

		// Print updated origin info for verification
		fmt.Printf("Updated Origin: %+v\n", origins[i])

	}
	return origins
}

// ProcessOriginConfigurations processes each origin configuration
func ProcessOriginConfigurations(origins []model.OriginInfo) {
	for _, origin := range origins {
		// Build a list of airports from the given origin and dates
		airportDetailsList := flightdb.DetermineFlightsFromConfig(origin)

		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList)
		//Generate WPI,  sort by WPI, update webpages
		GenerateCityRankings(origin, destinationsWithUrls)
	}
}
