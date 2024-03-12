package main

import (
//	"encoding/json"
	"fmt"
//	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
  "github.com/Tris20/FairFareFinder/src/go_files/discourse"
	//"github.com/Tris20/FairFareFinder/src/go_files/json_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/server"
	"github.com/Tris20/FairFareFinder/src/go_files/weather_pleasantness"
	"github.com/Tris20/FairFareFinder/src/go_files/web_pages/html_generators"
	"github.com/Tris20/FairFareFinder/src/go_files/url_generators"
	"github.com/Tris20/FairFareFinder/src/go_files"

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
  case "dev":

    origin:= model.OriginInfo{
			IATA:          "BER",
			City:          "Berlin",
			Country:       "Germany",
      DepartureStartDate: "2024-03-20", 
      DepartureEndDate: "2024-03-22", 
      ArrivalStartDate: "2024-03-24",
      ArrivalEndDate: "2024-03-26",
		}


    airportDetailsList := flightdb.DetermineFlightsFromConfig(origin)
    destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList )
// Iterate through the slice and print details of each DestinationInfo
	for _, dest := range destinationsWithUrls {
		fmt.Printf("IATA: %s, City: %s, Country: %s\n", dest.IATA, dest.City, dest.Country)
		fmt.Printf("SkyScannerURL: %s\n", dest.SkyScannerURL)
		fmt.Printf("AirbnbURL: %s\n", dest.AirbnbURL)
		fmt.Printf("BookingURL: %s\n\n", dest.BookingURL)
  }


	 GenerateCityRankings(origin, destinationsWithUrls)

origin = model.OriginInfo{
			IATA:          "GLA",
			City:          "Glasgow",
			Country:       "Scotland",
      DepartureStartDate: "2024-03-20", 
      DepartureEndDate: "2024-03-22", 
      ArrivalStartDate: "2024-03-24",
      ArrivalEndDate: "2024-03-26",
		}
    airportDetailsList = flightdb.DetermineFlightsFromConfig(origin)
    destinationsWithUrls = urlgenerators.GenerateFlightsAndHotelsURLs(origin, airportDetailsList )
	 GenerateCityRankings(origin, destinationsWithUrls)


  fmt.Println("\nStarting Webserver")
		
	fffwebserver.SetupFFFWebServer()

  case "web":

		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {
       /*
        //Berlin
				generateflightsdotjson.GenerateLinks("input/berlin-destinations.json", "input/berlin-flights.json", "ber")
				GenerateAndPostCityRankings("input/berlin-flights.json", "src/html/berlin-flight-destinations.html")
        // Glasgow
        generateflightsdotjson.GenerateLinks("input/glasgow-edi-destinations.json", "input/glasgow-edi-flights.json", "gla")
				GenerateAndPostCityRankings("input/glasgow-edi-flights.json", "src/html/glasgow-flight-destinations.html")
*/
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
		wpi := weather_pleasantry.ProcessLocation(destinationsWithUrls[i].City)
		if !math.IsNaN(wpi) {
			destinationsWithUrls[i].WPI = wpi // Directly write the WPI to the struct
			// Update URLs or any other info as needed
			destinationsWithUrls[i].SkyScannerURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].SkyScannerURL)
			destinationsWithUrls[i].AirbnbURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].AirbnbURL)
			destinationsWithUrls[i].BookingURL = replaceSpaceWithURLEncoding(destinationsWithUrls[i].BookingURL)
		}
	}

	// Sort the cities by WPI in descending order
	sort.Slice(destinationsWithUrls, func(i, j int) bool {
		return destinationsWithUrls[i].WPI > destinationsWithUrls[j].WPI
	})

	content := buildContentString(destinationsWithUrls)
	fmt.Println(content)

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	discourse.PostToDiscourse(content)
  target_url := fmt.Sprintf("src/html/%s-flight-destinations.html", strings.ToLower(origin.City)) 
	// Call the function to convert markdown to HTML and save it
  err := mdtabletohtml.ConvertMarkdownToHTML(content, target_url)
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
	for _, destination := range destinations {
    line := fmt.Sprintf("| [%s](https://www.google.com/maps/place/%s) | [%.2f](https://www.google.com/search?q=weather+%s) | [SkyScanner](%s) | [Airbnb](%s) [Booking.com](%s) | [Google Results](https://www.google.com/search?q=things+to+do+this+weekend+%s)| \n", destination.City, replaceSpaceWithURLEncoding(destination.City), destination.WPI, replaceSpaceWithURLEncoding(destination.City), destination.SkyScannerURL, destination.AirbnbURL, destination.BookingURL, destination.City)
		contentBuilder.WriteString(line)
	}

	// Convert the StringBuilder content to a string and return it
	return contentBuilder.String()
}
