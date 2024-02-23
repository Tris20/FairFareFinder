package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Tris20/FairFareFinder/src/go_files"
	"gopkg.in/yaml.v2"
)

type WeatherData struct {
	Dt   int64 `json:"dt"` // Unix timestamp of the forecasted data
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
}

type Favourites struct {
	Locations []string `yaml:"locations"`
}

type ForecastResponse struct {
	List []WeatherData `json:"list"`
}

// Secrets represents the structure of the secrets.yaml file.
type Secrets struct {
	APIKeys map[string]string `yaml:"api_keys"`
}

type CityAverageWPI struct {
	Name          string
	WPI           float64
	SkyScannerURL string
	AirbnbURL     string
	BookingURL    string
}

func main() {
	go_files.Setup_database()
	dbPath := "user_database.db"
	if len(os.Args) < 2 {
		log.Fatal("Error: No argument provided. Please provide a location, 'web', or a YAML file.")
	}

	switch os.Args[1] {
	case "web":

		// Update WPI data every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		go func() {
			for range ticker.C {
				generateLinks()
				handleFavourites("flights.json")
				fmt.Println("hello")
			}
		}()
		// Handle starting the web server
		http.HandleFunc("/", homeHandler)
		http.HandleFunc("/forecast", forecastHandler)
		http.HandleFunc("/getforecast", getForecastHandler)
		http.HandleFunc("/berlin-flight-destinations", presentBerlinFlightDestinations)
		// Serve static files from the `images` directory
		fs := http.FileServer(http.Dir("src/images"))
		http.Handle("/images/", http.StripPrefix("/images/", fs))

		// Start the web server
		fmt.Println("Starting server on :6969")
		if err := http.ListenAndServe(":6969", nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	case "init-db":
		init_database(dbPath)
		insert_test_user(dbPath)
	default:
		// Check if the argument is a YAML file
		if strings.HasSuffix(os.Args[1], ".json") {
			handleFavourites(os.Args[1])
		} else {
			// Assuming it's a city name
			location := strings.Join(os.Args[1:], " ")
			processLocation(location)
		}
	}
}

func init_database(dbPath string) {
	// Check if the database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// Database does not exist; create it
		go_files.CreateDatabase(dbPath)
	} else {
		// Database exists
		log.Println("Database already exists.")
	}
}

func insert_test_user(dbPath string) {
	username := "newuser"
	email := "newuser@example.com"
	preference := "dark mode"

	go_files.AddNewUserWithPreferences(dbPath, username, email, preference)
}

// handles requests to the home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Serve the HTML landing page
		pageContent, err := ioutil.ReadFile("src/html/landingPage.html")
		if err != nil {
			log.Printf("Error reading landing page file: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(pageContent)
	} else if strings.HasSuffix(r.URL.Path, ".css") {
		// Serve CSS files
		cssPath := "src/css" + r.URL.Path
		cssContent, err := ioutil.ReadFile(cssPath)
		if err != nil {
			log.Printf("Error reading CSS file: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(cssContent)
	} else {
		// 404 Not Found for other paths
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
}

func getForecastHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Handling request to /getforecast")

	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// fmt.Println("Handling POST request")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}
	city := r.FormValue("city")
	// fmt.Println("City:", city)

	// Call the processLocation function
	wpi := processLocation(city)

	response := fmt.Sprintf("The Weather Pleasantness Index (WPI) for %s is %.2f", city, wpi)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, response)
}

// handles requests to the forecast page
func forecastHandler(w http.ResponseWriter, r *http.Request) {
	// serving a static file
	pageContent, err := ioutil.ReadFile("src/html/forecast.html")
	if err != nil {
		log.Printf("Error reading forecast page file: %v", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(pageContent)
}

// handles requests to the forecast page
func presentBerlinFlightDestinations(w http.ResponseWriter, r *http.Request) {
	// serving a static file
	pageContent, err := ioutil.ReadFile("src/html/berlin-flight-destinations.html")
	if err != nil {
		log.Printf("Error reading forecast page file: %v", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(pageContent)
}

func handleFavourites(jsonFile string) {
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
		wpi := processLocation(flight.CityName)
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

	// Initialize a StringBuilder to efficiently build the content string
	var contentBuilder strings.Builder
	// add image to topic
	contentBuilder.WriteString("![image|690x394](upload://jGDO8BaFIvS1MVO53MDmqlS27vQ.jpeg)\n")
	// Header for the content
	contentBuilder.WriteString("|City Name | WPI | Flights | Accommodation | Things to Do|\n")
	contentBuilder.WriteString("|--|--|--|--|--|\n") // Additional line after headers

	// Loop through sorted results and append each to the contentBuilder
	for _, cityWPI := range cityWPIs {
		line := fmt.Sprintf("| [%s](https://www.google.com/maps/place/%s) | [%.2f](https://www.google.com/search?q=weather+%s) | [SkyScanner](%s) | [Airbnb](%s) [Booking.com](%s) | [Google Results](https://www.google.com/search?q=things+to+do+dubai+this+weekend+%s)| \n", cityWPI.Name, replaceSpaceWithURLEncoding(cityWPI.Name), cityWPI.WPI, replaceSpaceWithURLEncoding(cityWPI.Name), cityWPI.SkyScannerURL, cityWPI.AirbnbURL, cityWPI.BookingURL, cityWPI.Name)
		contentBuilder.WriteString(line)
	}

	// Convert the StringBuilder content to a string
	content := contentBuilder.String()
	fmt.Println(content)

	// Now content holds the full message to be posted, and you can pass it to the PostToDiscourse function
	PostToDiscourse(content)

	// Call the function to convert markdown to HTML and save it
	err = ConvertMarkdownToHTML(content, "src/html/berlin-flight-destinations.html")
	if err != nil {
		log.Fatalf("Failed to convert markdown to HTML: %v", err)
	}

}

func processLocation(location string) float64 {
	// Load API key from secrets.yaml
	apiKey, err := loadApiKey("ignore/secrets.yaml", "openweathermap.org")
	if err != nil {
		log.Fatal("Error loading API key:", err)
	}

	// Build the forecast API URL with the provided city
	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?q=%s&appid=%s&units=metric", location, apiKey)

	// Make the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	var forecast ForecastResponse
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Fatalf("Error reading response body: %v", err)
	} else if err := json.Unmarshal(body, &forecast); err != nil {
		log.Fatalf("Error parsing JSON response: %v", err)
	}

	// Load weather pleasantness config
	config, err := LoadWeatherPleasantnessConfig("weatherPleasantness.yaml")
	if err != nil {
		log.Fatal("Error loading weather pleasantness config:", err)
	}

	dailyDetails, overallAverage := ProcessForecastData(forecast.List, config)
	displayForecastData(location, dailyDetails)

	return overallAverage
}

func displayForecastData(location string, dailyDetails map[time.Weekday]DailyWeatherDetails) {
	orderedDays := []time.Weekday{time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday, time.Monday, time.Tuesday}

	fmt.Printf("Weather Pleasantness Index (WPI) for %s:\n", location)
	for _, day := range orderedDays {
		details, ok := dailyDetails[day]
		wind_kmh := 3.6 * details.AverageWind
		if ok {
			fmt.Printf("%s: Avg Temp: %.2f°C, Weather: %s, Wind: %.2fkm/h, WPI: %.2f\n",
				day.String(), details.AverageTemp, details.CommonWeather, wind_kmh, details.WPI)
		}
	}
}

// loadApiKey loads the API key for a given domain from a YAML file
func loadApiKey(filePath, domain string) (string, error) {
	var secrets Secrets

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(yamlFile, &secrets)
	if err != nil {
		return "", err
	}

	apiKey, ok := secrets.APIKeys[domain]
	if !ok {
		return "", fmt.Errorf("API key for %s not found", domain)
	}

	return apiKey, nil
}

// replaceSpaceWithURLEncoding replaces space characters with %20 in the URL
func replaceSpaceWithURLEncoding(urlString string) string {
	return strings.ReplaceAll(urlString, " ", "%20")
}
