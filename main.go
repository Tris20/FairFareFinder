package main

import (
	//	"encoding/json"
	"fmt"
	//	"io/ioutil"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
	_ "github.com/Tris20/FairFareFinder/src/go_files/discourse"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	//"github.com/Tris20/FairFareFinder/src/go_files/json_functions"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/server"
	"github.com/Tris20/FairFareFinder/src/go_files/url_generators"
	"github.com/Tris20/FairFareFinder/src/go_files/weather_pleasantness"
	"github.com/Tris20/FairFareFinder/src/go_files/web_pages/html_generators"
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
			location := strings.Join(os.Args[1:], " ")
			weather_pleasantry.ProcessLocation(location)
		}
	}
}

func GenerateCityRankings(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo) {

	for i := range destinationsWithUrls {
		wpi, dailyDetails := weather_pleasantry.ProcessLocation(destinationsWithUrls[i].City)
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

	content := buildContentString(destinationsWithUrls)
	fmt.Println(content)

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	//discourse.PostToDiscourse(content)
	target_url := fmt.Sprintf("src/html/%s-flight-destinations.html", strings.ToLower(origin.City))
	// Call the function to convert markdown to HTML and save it
	//	err := htmltablegenerator.ConvertMarkdownToHTML(content, target_url)

	err := htmltablegenerator.GenerateHtmlTable(target_url, destinationsWithUrls)
	if err != nil {
		log.Fatalf("Failed to convert markdown to HTML: %v", err)
	}

}

// replaceSpaceWithURLEncoding replaces space characters with %20 in the URL
func replaceSpaceWithURLEncoding(urlString string) string {
	return strings.ReplaceAll(urlString, " ", "%20")
}

func buildContentString(destinations []model.DestinationInfo) string {
	var contentBuilder strings.Builder
	// Add image to topic
	contentBuilder.WriteString("![image|690x394](upload://jGDO8BaFIvS1MVO53MDmqlS27vQ.jpeg)\n")
	// Header for the content
	contentBuilder.WriteString("|City Name | WPI | Flights | Accommodation | Things to Do|\n")
	contentBuilder.WriteString("|--|--|--|--|--|\n") // Additional line after headers

	// Loop through cityWPIs and append each to the contentBuilder
	//	iconCode := item.Weather[0].Icon                      // Original icon code, e.g., "10n"
	//	iconCodeDay := strings.Replace(iconCode, "n", "d", 1) // Replace "n" with "d"
	iconCodeDay := "01d"
	iconURL := fmt.Sprintf("http://openweathermap.org/img/wn/%s.png", iconCodeDay)

	for _, destination := range destinations {

		weather_icons := fmt.Sprintf("[(%s)](https://www.google.com/search?q=weather+%s) [(%s)](https://www.google.com/search?q=weather+%s) [(%s)](https://www.google.com/search?q=weather+%s) [(%s)](https://www.google.com/search?q=weather+%s) [(%s)](https://www.google.com/search?q=weather+%s)", randomizeIconURL(iconURL), replaceSpaceWithURLEncoding(destination.City), randomizeIconURL(iconURL), replaceSpaceWithURLEncoding(destination.City), randomizeIconURL(iconURL), replaceSpaceWithURLEncoding(destination.City), randomizeIconURL(iconURL), replaceSpaceWithURLEncoding(destination.City), randomizeIconURL(iconURL), replaceSpaceWithURLEncoding(destination.City))

		/*
		   		weather_icons := fmt.Sprintf(`
		   <td style="white-space: nowrap;">
		     <span style="display: inline-block; text-align: center; width: 100px;">
		       Mon<br>
		       <a href="https://www.google.com/search?q=weather+Dubai">
		         <img src="http://openweathermap.org/img/wn/02d.png" alt="Image" style="max-width:100px;">
		       </a>
		     </span>
		     <span style="display: inline-block; text-align: center; width: 100px;">
		       Tue<br>
		       <a href="https://www.google.com/search?q=weather+Dubai">
		         <img src="http://openweathermap.org/img/wn/03d.png" alt="Image" style="max-width:100px;">
		       </a>
		     </span>
		     <span style="display: inline-block; text-align: center; width: 100px;">
		       Wed<br>
		       <a href="https://www.google.com/search?q=weather+Dubai">
		         <img src="http://openweathermap.org/img/wn/02d.png" alt="Image" style="max-width:100px;">
		       </a>
		     </span>
		     <span style="display: inline-block; text-align: center; width: 100px;">
		       Thu<br>
		       <a href="https://www.google.com/search?q=weather+Dubai">
		         <img src="http://openweathermap.org/img/wn/04d.png" alt="Image" style="max-width:100px;">
		       </a>
		     </span>
		     <span style="display: inline-block; text-align: center; width: 100px;">
		       Fri<br>
		       <a href="https://www.google.com/search?q=weather+Dubai">
		         <img src="http://openweathermap.org/img/wn/01d.png" alt="Image" style="max-width:100px;">
		       </a>
		     </span>
		   </td>
		   `)

		*/

		line := fmt.Sprintf("| [%s](https://www.google.com/maps/place/%s) | %s | [SkyScanner](%s) | [Airbnb](%s) [Booking.com](%s) | [Google Results](https://www.google.com/search?q=things+to+do+this+weekend+%s)| \n", destination.City, replaceSpaceWithURLEncoding(destination.City), weather_icons, destination.SkyScannerURL, destination.AirbnbURL, destination.BookingURL, destination.City)
		contentBuilder.WriteString(line)
	}

	// Convert the StringBuilder content to a string and return it
	return contentBuilder.String()
}

// randomizeIconURL replaces the digit before "d.png" or "n.png" with a random number between 1 and 4.
func randomizeIconURL(iconURL string) string {
	// Seed the random number generator (consider doing this once at the start of your program instead)
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 1 and 4
	randomNumber := rand.Intn(4) + 1 // rand.Intn(n) generates a number in [0, n), so +1 to shift to [1, 4]

	// Find and replace the digit before "d.png" or "n.png"
	re := regexp.MustCompile(`(\d)(d\.png|n\.png)`)
	newIconURL := re.ReplaceAllStringFunc(iconURL, func(m string) string {
		// m is the full match, so replace only the digit part
		return strings.Replace(m, string(m[0]), fmt.Sprintf("%d", randomNumber), 1)
	})

	return newIconURL
}
