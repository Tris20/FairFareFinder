package main

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
//"github.com/Tris20/FairFareFinder/src/go_files/flightutils"
	"github.com/Tris20/FairFareFinder/src/go_files/server"
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

func main() {
	user_db.Setup_database()
	dbPath := "user_database.db"
	if len(os.Args) < 2 {
		log.Fatal("Error: No argument provided. Please provide a location, 'web', or a json file.")
	}

	berlin_config := model.OriginInfo{
		IATA:               "BER",
		City:               "Berlin",
		Country:            "Germany",
		DepartureStartDate: "2024-03-20",
		DepartureEndDate:   "2024-03-22",
		ArrivalStartDate:   "2024-03-24",
		ArrivalEndDate:     "2024-03-26",
	}

	glasgow_config := model.OriginInfo{
		IATA:               "GLA",
		City:               "Glasgow",
		Country:            "Scotland",
		DepartureStartDate: "2024-03-20",
		DepartureEndDate:   "2024-03-22",
		ArrivalStartDate:   "2024-03-24",
		ArrivalEndDate:     "2024-03-26",
	}

	switch os.Args[1] {
	case "dev":
		airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
		GenerateCityRankings(berlin_config, destinationsWithUrls)

		fmt.Println("\nStarting Webserver")

		fffwebserver.SetupFFFWebServer()

	case "web":
		//Update Berlin and Glasgow immediately
		airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
		GenerateCityRankings(berlin_config, destinationsWithUrls)

		airportDetailsList = flightdb.DetermineFlightsFromConfig(glasgow_config)
		destinationsWithUrls = urlgenerators.GenerateFlightsAndHotelsURLs(glasgow_config, airportDetailsList)
		GenerateCityRankings(glasgow_config, destinationsWithUrls)
		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {

				airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
				destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
				GenerateCityRankings(berlin_config, destinationsWithUrls)

				airportDetailsList = flightdb.DetermineFlightsFromConfig(glasgow_config)
				destinationsWithUrls = urlgenerators.GenerateFlightsAndHotelsURLs(glasgow_config, airportDetailsList)
				GenerateCityRankings(glasgow_config, destinationsWithUrls)

			}
		}()
		fffwebserver.SetupFFFWebServer()

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

func GenerateCityRankings(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo) {

	for i := range destinationsWithUrls {
		wpi, dailyDetails := weather_pleasantry.ProcessLocation(destinationsWithUrls[i])
		if !math.IsNaN(wpi) {
			destinationsWithUrls[i].WPI = wpi // Directly write the WPI to the struct
			// Update URLs or any other info as needed
			destinationsWithUrls[i].SkyScannerURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].SkyScannerURL)
			destinationsWithUrls[i].AirbnbURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].AirbnbURL)
			destinationsWithUrls[i].BookingURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].BookingURL)

			var weatherDetailsSlice []model.DailyWeatherDetails

			for _, details := range dailyDetails {
				weatherDetailsSlice = append(weatherDetailsSlice, details)
			}

			destinationsWithUrls[i].WeatherDetails = weatherDetailsSlice
		}
	}

	// Sort the cities by WPI in descending order
	sort.Slice(destinationsWithUrls, func(i, j int) bool {
		return destinationsWithUrls[i].WPI > destinationsWithUrls[j].WPI
	})

/*
	for i := range destinationsWithUrls {
    if destinationsWithUrls[i].WPI > 7 
    {
      flightutils.search()
    }

*/
  generate_html_table(origin, destinationsWithUrls)

}

// replaceSpaceWithURLEncoding replaces space characters with %20 in the URL
func replaceSpaceWithURLEncoding(urlString string) string {
	return strings.ReplaceAll(urlString, " ", "%20")
}


func generate_html_table(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo ){

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	target_url := fmt.Sprintf("src/html/%s-flight-destinations.html", strings.ToLower(origin.City))

	err := htmltablegenerator.GenerateHtmlTable(target_url, destinationsWithUrls )
	if err != nil {
		log.Fatalf("Failed to convert markdown to HTML: %v", err)
	}
}
