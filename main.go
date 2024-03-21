package main

import (
	"fmt"
	"github.com/Tris20/FairFareFinder/src/go_files"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/flight_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/db_functions/user_db_functions"
	"github.com/Tris20/FairFareFinder/src/go_files/flightutils"
	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
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
 "encoding/json"
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

type PriceKey struct {
    OriginID      string
    DestinationID string
}

// PriceData now just holds the price, as the IDs are embedded in the key of the map.
type PriceData struct {
    Price float64
}

type PersistentPrices struct {
    Data map[PriceKey]PriceData
}

var checkFlightPrices = false
var checkprice_init = false
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
		SkyScannerID:       "eyJzIjoiQkVSIiwiZSI6Ijk1NjczMzgzIiwiaCI6IjI3NTQ3MDUzIn0=",
	}

	glasgow_config := model.OriginInfo{
		IATA:               "GLA",
		City:               "Glasgow",
		Country:            "Scotland",
		DepartureStartDate: "2024-03-20",
		DepartureEndDate:   "2024-03-22",
		ArrivalStartDate:   "2024-03-24",
		ArrivalEndDate:     "2024-03-26",
		SkyScannerID:       "eyJzIjoiR0xBUyIsImUiOiIyNzU0MTg1MiIsImgiOiIyNzU0MTg1MiJ9",
	}

    berlin_config.DepartureStartDate, berlin_config.DepartureEndDate = timeutils.UpcomingWedToSat()
    berlin_config.ArrivalStartDate, berlin_config.ArrivalEndDate = timeutils.UpcomingSunToWedFromSat(berlin_config.DepartureEndDate)
    glasgow_config.DepartureStartDate, glasgow_config.DepartureEndDate = timeutils.UpcomingWedToSat()
    glasgow_config.ArrivalStartDate, glasgow_config.ArrivalEndDate = timeutils.UpcomingSunToWedFromSat(glasgow_config.DepartureEndDate)

	switch os.Args[1] {
	case "dev":
		airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
		GenerateCityRankings(berlin_config, destinationsWithUrls)

		fmt.Println("\nStarting Webserver")

		fffwebserver.SetupFFFWebServer()

	case "web":
    fmt.Printf("INIT")
    /*
    checkprice_init = true //get prices whenever we reset server 
		checkFlightPrices = true

		
    //Update Berlin and Glasgow immediately
		airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
		destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
		GenerateCityRankings(berlin_config, destinationsWithUrls)
		checkFlightPrices = true
		airportDetailsList = flightdb.DetermineFlightsFromConfig(glasgow_config)
		destinationsWithUrls = urlgenerators.GenerateFlightsAndHotelsURLs(glasgow_config, airportDetailsList)
		GenerateCityRankings(glasgow_config, destinationsWithUrls)
     */
     checkprice_init = false

		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {
				// If today is tuesday, preapre flag so we check prices on wednesday
				if time.Now().Weekday() == time.Tuesday {
					checkFlightPrices = true
          //PERF can optimise. This runs 4 times on tuesday, but only needs to be ran once
    berlin_config.DepartureStartDate, berlin_config.DepartureEndDate = timeutils.UpcomingWedToSat()
    berlin_config.ArrivalStartDate, berlin_config.ArrivalEndDate = timeutils.UpcomingSunToWedFromSat(berlin_config.DepartureEndDate)

    glasgow_config.DepartureStartDate, glasgow_config.DepartureEndDate = timeutils.UpcomingWedToSat()
    glasgow_config.ArrivalStartDate, glasgow_config.ArrivalEndDate = timeutils.UpcomingSunToWedFromSat(glasgow_config.DepartureEndDate)
				}

				airportDetailsList := flightdb.DetermineFlightsFromConfig(berlin_config)
				destinationsWithUrls := urlgenerators.GenerateFlightsAndHotelsURLs(berlin_config, airportDetailsList)
				GenerateCityRankings(berlin_config, destinationsWithUrls)

				airportDetailsList = flightdb.DetermineFlightsFromConfig(glasgow_config)
				destinationsWithUrls = urlgenerators.GenerateFlightsAndHotelsURLs(glasgow_config, airportDetailsList)
				GenerateCityRankings(glasgow_config, destinationsWithUrls)

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

func GenerateCityRankings(origin model.OriginInfo, destinationsWithUrls []model.DestinationInfo) {
	// Load existing prices
	persistentPrices := &PersistentPrices{}
	err := persistentPrices.Load("prices.json")
	if err != nil {
		log.Fatal("Error loading prices:", err)
	}


	  for i := range destinationsWithUrls {
        wpi, dailyDetails := weather_pleasantry.ProcessLocation(destinationsWithUrls[i])
        if !math.IsNaN(wpi) {
            destinationsWithUrls[i].WPI = wpi // Directly write the WPI to the struct

            priceKey := PriceKey{
                OriginID:      origin.SkyScannerID,
                DestinationID: destinationsWithUrls[i].SkyScannerID,
            }

            if checkFlightPrices {
                if (time.Now().Weekday() == time.Wednesday) || checkprice_init {
  //                  if destinationsWithUrls[i].WPI > 6.5 {
                        fmt.Printf("\n\nSkyscannerID: %s", destinationsWithUrls[i].SkyScannerID)
                        price, err := flightutils.GetBestPrice(origin, destinationsWithUrls[i])
                        if err != nil {
                            log.Fatal("Error getting best price:", err)
                        }
                        fmt.Printf("\n\n Best Price: €%.2f", price)

                        // Update the destination with the new price and update the persistent data
                        destinationsWithUrls[i].SkyScannerPrice = price
                        persistentPrices.Data[priceKey] = PriceData{Price: price}
        //            }
                }
            } else {
                // Use the persisted price if available
                if priceData, ok := persistentPrices.Data[priceKey]; ok {
                    destinationsWithUrls[i].SkyScannerPrice = priceData.Price
                }
            }
        
    

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
  //Reset the price check flag 
	checkFlightPrices = false

	// Sort the cities by WPI in descending order
	sort.Slice(destinationsWithUrls, func(i, j int) bool {
		return destinationsWithUrls[i].WPI > destinationsWithUrls[j].WPI
	})
    // Save updated prices at the end
    err = persistentPrices.Save("prices.json")
    if err != nil {
        log.Fatal("Error saving prices:", err)
    }
	generate_html_table(origin, destinationsWithUrls)

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




// Load reads the price data from a file and unmarshals it into the PersistentPrices struct.
func (p *PersistentPrices) Load(fileName string) error {
    file, err := os.ReadFile(fileName)
    if err != nil {
        if os.IsNotExist(err) {
            p.Data = make(map[PriceKey]PriceData) // Corrected to use the PriceKey type
            return nil // No file means first run; proceed with an empty map.
        }
        return err
    }
    // Use json.Unmarshal to leverage the custom UnmarshalJSON method of PersistentPrices
    return json.Unmarshal(file, p)
}

// Save marshals the PersistentPrices struct and writes it to a file.
func (p *PersistentPrices) Save(fileName string) error {
    // Use json.Marshal to leverage the custom MarshalJSON method of PersistentPrices
    data, err := json.Marshal(p)
    if err != nil {
        return err
    }
    // Use os.WriteFile (which replaces ioutil.WriteFile)
    return os.WriteFile(fileName, data, 0644)
}



func (pp *PersistentPrices) MarshalJSON() ([]byte, error) {
    helper := make(map[string]PriceData)
    for key, value := range pp.Data {
        keyStr := fmt.Sprintf("%s_%s", key.OriginID, key.DestinationID)
        helper[keyStr] = value
    }
    return json.Marshal(helper)
}



func (pp *PersistentPrices) UnmarshalJSON(data []byte) error {
    helper := make(map[string]PriceData)
    if err := json.Unmarshal(data, &helper); err != nil {
        return err
    }

    pp.Data = make(map[PriceKey]PriceData)
    for keyStr, value := range helper {
        ids := strings.Split(keyStr, "_")
        if len(ids) != 2 {
            return fmt.Errorf("unexpected key format: %s", keyStr)
        }
        pp.Data[PriceKey{OriginID: ids[0], DestinationID: ids[1]}] = value
    }
    return nil
}
