package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
  "github.com/Tris20/FairFareFinder/src/go_files/discourse"
	"github.com/Tris20/FairFareFinder/src/go_files/json_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/server"
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

  
	switch os.Args[1] {
  case "db_test":
  flightdb.GetCitiesAndIATACodes()

  case "web":

		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {
        //Berlin
				generateflightsdotjson.GenerateLinks("input/berlin-destinations.json", "input/berlin-flights.json", "ber")
				GenerateAndPostCityRankings("input/berlin-flights.json", "src/html/berlin-flight-destinations.html")
        // Glasgow
        generateflightsdotjson.GenerateLinks("input/glasgow-edi-destinations.json", "input/glasgow-edi-flights.json", "gla")
				GenerateAndPostCityRankings("input/glasgow-edi-flights.json", "src/html/glasgow-flight-destinations.html")

        fmt.Println("hello")
			}
		}()
		fffwebserver.SetupFFFWebServer()

	case "init-db":
		user_db.Init_database(dbPath)
		user_db.Insert_test_user(dbPath)
	default:
		// Check if the argument is a json file
		if strings.HasSuffix(os.Args[1], ".json") {
      out := fmt.Sprintf("input/%s-flights.json",os.Args[1:] )
			GenerateAndPostCityRankings(os.Args[1], out)
		} else {
			// Assuming it's a city name
			location := strings.Join(os.Args[1:], " ")
			weather_pleasantry.ProcessLocation(location)
		}
	}
}

func GenerateAndPostCityRankings(jsonFile string, target_url string) {
	var flights []struct {
		CityName      string `json:"City_name"`
		SkyScannerURL string `json:"SkyScannerURL"`
		AirbnbURL     string `json:"airbnbURL"`
		BookingURL    string `json:"bookingURL"`
	}

	fileContents, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		log.Fatalf("Error reading %s file: %v", jsonFile, err)
	}

	err = json.Unmarshal(fileContents, &flights)
	if err != nil {
		log.Fatalf("Error parsing JSON file: %v", err)
	}

	var cityWPIs []CityAverageWPI
	for _, flight := range flights {
		wpi := weather_pleasantry.ProcessLocation(flight.CityName)
		if !math.IsNaN(wpi) {
			cityWPIs = append(cityWPIs, CityAverageWPI{
				Name:          flight.CityName,
				WPI:           wpi,
				SkyScannerURL: replaceSpaceWithURLEncoding(flight.SkyScannerURL),
				AirbnbURL:     replaceSpaceWithURLEncoding(flight.AirbnbURL),
				BookingURL:    replaceSpaceWithURLEncoding(flight.BookingURL),
			})
		}
	}

	// Sort the cities by WPI in descending order
	sort.Slice(cityWPIs, func(i, j int) bool {
		return cityWPIs[i].WPI > cityWPIs[j].WPI
	})

	content := buildContentString(cityWPIs)
	fmt.Println(content)

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	discourse.PostToDiscourse(content)

	// Call the function to convert markdown to HTML and save it
	err = mdtabletohtml.ConvertMarkdownToHTML(content, target_url)
	if err != nil {
		log.Fatalf("Failed to convert markdown to HTML: %v", err)
	}

}

// replaceSpaceWithURLEncoding replaces space characters with %20 in the URL
func replaceSpaceWithURLEncoding(urlString string) string {
	return strings.ReplaceAll(urlString, " ", "%20")
}

func buildContentString(cityWPIs []CityAverageWPI) string {
	var contentBuilder strings.Builder
	// Add image to topic
	contentBuilder.WriteString("![image|690x394](upload://jGDO8BaFIvS1MVO53MDmqlS27vQ.jpeg)\n")
	// Header for the content
	contentBuilder.WriteString("|City Name | WPI | Flights | Accommodation | Things to Do|\n")
	contentBuilder.WriteString("|--|--|--|--|--|\n") // Additional line after headers

	// Loop through cityWPIs and append each to the contentBuilder
	for _, cityWPI := range cityWPIs {
		line := fmt.Sprintf("| [%s](https://www.google.com/maps/place/%s) | [%.2f](https://www.google.com/search?q=weather+%s) | [SkyScanner](%s) | [Airbnb](%s) [Booking.com](%s) | [Google Results](https://www.google.com/search?q=things+to+do+this+weekend+%s)| \n", cityWPI.Name, replaceSpaceWithURLEncoding(cityWPI.Name), cityWPI.WPI, replaceSpaceWithURLEncoding(cityWPI.Name), cityWPI.SkyScannerURL, cityWPI.AirbnbURL, cityWPI.BookingURL, cityWPI.Name)
		contentBuilder.WriteString(line)
	}

	// Convert the StringBuilder content to a string and return it
	return contentBuilder.String()
}
