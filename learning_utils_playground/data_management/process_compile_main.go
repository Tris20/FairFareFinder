package data_management

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

func ProcessCompile() {
	fmt.Println("ProcessCompile")
}

// main.go

/*

IF 3AM Monday Morning
  Create a brand new DB and send it to the website
ELSE IF 6 hours since last data upadte
  Fetch latest weather, compile weather data and calculate new WPI scores and send it to the website

Scripts need to be run in specific orders. The order is typically is:
Fetch
Calculate/Generate
Compile


Some Scripts have dependencies on others. For example, the 5 day WPI of a location(Compile Location) requires the weather the be fetched, but also the WPI of each day to be calculated (Compile Weather)

NOTE: Fetch and Compile Properties, gest the prices of the nearest wednesday to wednesday, so should weally be run on a monday
*/

// Helper function to run a command in a specific directory
func runExecutableInDir(dir string, executable string) {
	// Change to the specified directory
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("Failed to change directory to %s: %v", dir, err)
	}
	log.Printf("Changed to directory: %s\n", dir)

	// Execute the executable
	cmd := exec.Command("./" + executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Running executable: %s\n", executable)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run executable %s in directory %s: %v", executable, dir, err)
	}

	log.Printf("Successfully executed: %s\n", executable)
}

// Helper function to get the current weekday and hour
func GetCurrentTime() (time.Weekday, int) {
	now := time.Now()
	return now.Weekday(), now.Hour()
}

/*logging*/
// Global log file and date variables
var currentLogFile *os.File
var currentLogDate string

// Generate the log file path based on the current date (year/month/output.log)
func getDailyLogFilePath() string {
	currentTime := time.Now()
	year := currentTime.Format("2006")       // Year in YYYY format
	month := currentTime.Format("01")        // Month in MM format
	fileName := currentTime.Format("02.log") // Log file name as DD.log
	return filepath.Join("logs", year, month, fileName)
}

// Ensure log directories exist for year and month
func ensureLogDirExists(logFilePath string) error {
	dir := filepath.Dir(logFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %s: %v", dir, err)
		}
	}
	return nil
}

// Function to update the log file if a new day has started
func UpdateLogFile() error {
	newLogDate := time.Now().Format("2006-01-02") // YYYY-MM-DD
	if newLogDate != currentLogDate {
		if currentLogFile != nil {
			currentLogFile.Close()
		}

		logFilePath := getDailyLogFilePath()
		if err := ensureLogDirExists(logFilePath); err != nil {
			return err
		}

		var err error
		currentLogFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %v", logFilePath, err)
		}

		multiWriter := io.MultiWriter(os.Stdout, currentLogFile)
		log.SetOutput(multiWriter)
		currentLogDate = newLogDate
	}
	return nil
}

func CloseLogFile() {
	if currentLogFile != nil {
		currentLogFile.Close()
	}
}

// database-setup.go

func InitializeDatabase(dbPath string) error {
	// Open the database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Failed to open database: %v", err)
		return err
	}
	defer db.Close()

	// Create accommodation table
	err = db_manager.CreateTable(db, &db_manager.MainDBAccommodation{})
	if err != nil {
		return err
	}

	// Create five_nights_and_flights table
	err = db_manager.CreateTable(db, &db_manager.MainDBFiveNightsAndFlights{})
	if err != nil {
		return err
	}

	// Create flight_prices table
	err = db_manager.CreateTable(db, &db_manager.MainDBFlight{})
	if err != nil {
		return err
	}

	// Create Locations table
	err = db_manager.CreateTable(db, &db_manager.MainDBLocation{})
	if err != nil {
		return err
	}

	// Create Weather table
	err = db_manager.CreateTable(db, &db_manager.MainDBWeather{})
	if err != nil {
		return err
	}

	return nil
}

// Helper function to delete new_main.db if it exists
func DeleteDatabase(dbPath string) error {
	// Check if new_main.db exists
	if _, err := os.Stat(dbPath); err == nil {
		// Delete new_main.db
		err := os.Remove(dbPath)
		if err != nil {
			return fmt.Errorf("failed to delete db: %v", err)
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred, but not a "file doesn't exist" error
		return fmt.Errorf("failed to check db: %v", err)
	}

	return nil
}

// Helper function to copy main.db to new_main.db
func copyMainDB(srcPath, destPath string) error {
	// Open source main.db
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open main.db: %v", err)
	}
	defer srcFile.Close()

	// Create destination new_main.db
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create new_main.db: %v", err)
	}
	defer destFile.Close()

	// Copy the content from main.db to new_main.db
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy main.db to new_main.db: %v", err)
	}

	fmt.Println("Successfully copied main.db to new_main.db")
	return nil
}

/// weather-main.go

type WeatherDataMain struct {
	CityName          string
	CountryCode       string
	Date              string
	Temperature       float64
	WPI               float64
	WeatherIconURL    string
	GoogleWeatherLink string
}

type CompiledWeather struct {
	City           string
	Country        string
	Date           string
	AvgDaytimeTemp float64
	WeatherIcon    string
	GoogleURL      string
	AvgDaytimeWPI  float64
}

func ProcessCompileMainWeather() {
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/weather/weather.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `
	SELECT city_name, country_code, date, AVG(temperature) AS avg_temp, AVG(wpi) AS avg_wpi, weather_icon_url, google_weather_link
	FROM current_weather
	WHERE strftime('%H:%M:%S', date) BETWEEN '10:00:00' AND '18:00:00'
	GROUP BY city_name, country_code, strftime('%Y-%m-%d', date)
	`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var weathers []CompiledWeather
	for rows.Next() {
		var wd WeatherDataMain
		err := rows.Scan(&wd.CityName, &wd.CountryCode, &wd.Date, &wd.Temperature, &wd.WPI, &wd.WeatherIconURL, &wd.GoogleWeatherLink)
		if err != nil {
			log.Fatal(err)
		}
		formattedTemp, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", wd.Temperature), 64)
		formattedWPI, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", wd.WPI), 64)
		weathers = append(weathers, CompiledWeather{
			City:           wd.CityName,
			Country:        wd.CountryCode,
			Date:           strings.Split(wd.Date, " ")[0],
			AvgDaytimeTemp: formattedTemp,
			WeatherIcon:    wd.WeatherIconURL,
			GoogleURL:      wd.GoogleWeatherLink,
			AvgDaytimeWPI:  formattedWPI,
		})
	}

	compiledDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer compiledDB.Close()

	// Clear the existing weather data
	_, err = compiledDB.Exec("DELETE FROM weather")
	if err != nil {
		log.Fatal("Failed to clear existing weather data:", err)
	}

	stmt, err := compiledDB.Prepare("INSERT INTO weather (city, country, date, avg_daytime_temp, weather_icon, google_url, avg_daytime_wpi) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	bar := progressbar.Default(int64(len(weathers)))
	for _, w := range weathers {
		_, err := stmt.Exec(w.City, w.Country, w.Date, w.AvgDaytimeTemp, w.WeatherIcon, w.GoogleURL, w.AvgDaytimeWPI)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1)
	}
	fmt.Println("Data successfully transferred to new_main.db")
}

// locations-main.go

type CityMain struct {
	City       string
	IncludeTF  int
	CityAscii  string
	Lat        float64
	Lon        float64
	Country    string
	Iso2       string
	Iso3       string
	AdminName  sql.NullString
	Capital    sql.NullString
	Population sql.NullInt64
	Id         int
	IATACodes  []string
}

func ProcessCompileMainLocations() {
	// Open locations.db
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/locations/locations.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Query cities where include_tf == 1
	rows, err := db.Query("SELECT city, include_tf, city_ascii, lat, lon, country, iso2, iso3, admin_name, capital, population, id FROM city WHERE include_tf = 1")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	var cities []CityMain

	// Iterate over the rows
	for rows.Next() {
		var c CityMain
		err = rows.Scan(&c.City, &c.IncludeTF, &c.CityAscii, &c.Lat, &c.Lon, &c.Country, &c.Iso2, &c.Iso3, &c.AdminName, &c.Capital, &c.Population, &c.Id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Fetch IATA codes from the "airport" table
		airports, err := db.Query("SELECT iata FROM airport WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", c.CityAscii, c.Iso2)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for airports.Next() {
			var iata string
			if err := airports.Scan(&iata); err != nil {
				fmt.Println("Error scanning IATA code:", err)
				continue
			}
			fmt.Printf("Fetched IATA: %s for city: %s\n", iata, c.CityAscii) // Debug output
			c.IATACodes = append(c.IATACodes, iata)
		}
		if len(c.IATACodes) == 0 {
			fmt.Printf("No IATA codes found for city: %s %s\n, ", c.CityAscii, c.Country) // Debug output
		}
		airports.Close()

		cities = append(cities, c)
	}

	// Close and open new database
	db.Close()
	db, err = sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// Initialize the progress bar
	bar := progressbar.NewOptions(len(cities),
		progressbar.OptionSetDescription("Inserting into new_main.db"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("cities"),
	)

	// Insert into the new database
	for _, c := range cities {
		query := "INSERT OR REPLACE INTO location (city, country, iata_1, iata_2, iata_3, iata_4, iata_5, iata_6, iata_7, avg_wpi) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		args := fillIATAs(c.CityAscii, c.Iso2, c.IATACodes)
		if _, err := db.Exec(query, args...); err != nil {
			fmt.Println(err)
			continue
		}
		bar.Add(1) // Increment the progress bar for each city processed
	}
	bar.Finish() // End the progress bar when loop is complete

	db, err = sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		fmt.Println("Error reopening database:", err)
		return
	}
	defer db.Close()

	// Update the avg_wpi based on weather data
	UpdateAvgWPI(db)

}

func fillIATAs(city, country string, codes []string) []interface{} {
	result := make([]interface{}, 10) // Total 10 placeholders: city, country, 7 iatas, avg_wpi
	result[0] = city
	result[1] = country
	for i, code := range codes {
		if i >= 7 {
			break
		}
		result[i+2] = code
	}
	for i := len(codes) + 2; i < 9; i++ {
		result[i] = nil
	}
	result[9] = nil // avg_wpi as nil, replace or remove as needed
	return result
}

// avg-wpi.go

func UpdateAvgWPI(db *sql.DB) {
	// Query unique city-country pairs from the 'location' table
	rows, err := db.Query("SELECT DISTINCT city, country FROM location")
	if err != nil {
		fmt.Println("Error fetching city-country pairs:", err)
		return
	}
	defer rows.Close()

	// Collect all city-country pairs for progress bar initialization
	cityCountryPairs := make([]struct{ city, country string }, 0)
	for rows.Next() {
		var city, country string
		if err := rows.Scan(&city, &country); err != nil {
			fmt.Println("Error scanning city-country pair:", err)
			continue
		}
		cityCountryPairs = append(cityCountryPairs, struct{ city, country string }{city, country})
	}

	// Initialize the progress bar
	bar := progressbar.NewOptions(len(cityCountryPairs),
		progressbar.OptionSetDescription("Updating average WPI"),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("pairs"),
	)

	// Loop over each city-country pair to calculate and update avg_wpi
	for _, pair := range cityCountryPairs {
		var avgResult sql.NullFloat64 // Use sql.NullFloat64 to handle NULL values
		err = db.QueryRow("SELECT AVG(avg_daytime_wpi) FROM weather WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", pair.city, pair.country).Scan(&avgResult)
		if err != nil {
			fmt.Printf("Failed to calculate average WPI for %s, %s: %v\n", pair.city, pair.country, err)
			bar.Add(1) // Increment the progress bar even in case of an error
			continue
		}

		if avgResult.Valid { // Check if the result is valid (not NULL)
			_, err = db.Exec("UPDATE location SET avg_wpi = ? WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", avgResult.Float64, pair.city, pair.country)
			if err != nil {
				fmt.Printf("Failed to update avg_wpi for %s, %s: %v\n", pair.city, pair.country, err)
			}
		} else {
			fmt.Printf("No valid average WPI found for %s, %s, setting avg_wpi to NULL\n", pair.city, pair.country)
			_, err = db.Exec("UPDATE location SET avg_wpi = NULL WHERE LOWER(city) = LOWER(?) AND LOWER(country) = LOWER(?)", pair.city, pair.country)
			if err != nil {
				fmt.Printf("Failed to set avg_wpi to NULL for %s, %s: %v\n", pair.city, pair.country, err)
			}
		}

		bar.Add(1) // Increment the progress bar after processing each pair
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Error during rows iteration:", err)
	}

	fmt.Println("Updated avg_wpi based on weather data successfully.")
}

/// flights-main.go

func ProcessCompileMainFlights() {
	fmt.Printf("Starting to Compile Flights Table")
	db, err := sql.Open("sqlite3", "../../../../../../data/raw/flights/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// todo: relate to database manager
	rows, err := db.Query("SELECT * FROM skyscannerprices")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var prices []db_manager.SkyScannerPrice

	for rows.Next() {
		var sp db_manager.SkyScannerPrice
		err = rows.Scan(&sp.OriginCity, &sp.OriginCountry, &sp.OriginIATA, &sp.OriginSkyScannerID, &sp.DestinationCity, &sp.DestinationCountry, &sp.DestinationIATA, &sp.DestinationSkyScannerID, &sp.ThisWeekend, &sp.NextWeekend)
		if err != nil {
			log.Fatal(err)
		}
		sp.OriginCountry = GetISOCode(sp.OriginCountry)
		sp.SkyScannerURL = fmt.Sprintf("https://www.skyscanner.de/transport/fluge/%s/%s/?adults=1&adultsv2=1&cabinclass=economy&children=0&inboundaltsenabled=false&infants=0&outboundaltsenabled=false&preferdirects=true&ref=home&rtn=1", sp.OriginIATA, sp.DestinationIATA)
		prices = append(prices, sp)
	}
	if rows.Err() != nil {
		log.Fatal(rows.Err())
	}

	// Open a new connection to the `new_main.db` database
	newDB, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer newDB.Close()

	// Delete all existing entries from the flight table
	_, err = newDB.Exec("DELETE FROM flight")
	if err != nil {
		log.Fatal("Failed to delete existing data: ", err)
	}
	fmt.Println("Existing data deleted.")

	// Create a new progress bar
	bar := progressbar.Default(int64(len(prices)), "Inserting records")

	// Insert data into the new table
	for _, price := range prices {
		_, err := newDB.Exec("INSERT INTO flight (origin_city_name, origin_country, origin_iata, origin_skyscanner_id, destination_city_name, destination_country, destination_iata, destination_skyscanner_id, price_this_week, skyscanner_url_this_week, price_next_week, skyscanner_url_next_week) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			price.OriginCity, price.OriginCountry, price.OriginIATA, price.OriginSkyScannerID, price.DestinationCity, price.DestinationCountry, price.DestinationIATA, price.DestinationSkyScannerID, price.ThisWeekend.Float64, price.SkyScannerURL, price.NextWeekend.Float64, price.SkyScannerURL)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1) // Update the progress bar
	}
	fmt.Println("\nData inserted successfully into new_main.db")
}

/// accommodation-main.go

// Accommodation represents the filtered data we will extract
type Accommodation struct {
	City       string
	Country    string
	GrossPrice float64
	Checkin    string
	Checkout   string
}

// LocationPrices holds prices for a specific location (city + country)
type LocationPrices struct {
	City    string
	Country string
	Prices  []float64
}

func ProcessCompileMainAccomodation() {
	// Step 1: Open (or create) "new_main.db"
	newDb, err := sql.Open("sqlite3", "../../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatalf("Failed to open new_main.db: %v", err)
	}
	defer newDb.Close()

	// Step 2: Ensure that the 'accommodation' table exists in new_main.db
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS accommodation (
		city TEXT NOT NULL,
		country TEXT NOT NULL,
		booking_url TEXT,
		booking_pppn REAL NOT NULL
	);`
	_, err = newDb.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create accommodation table: %v", err)
	}

	// Step 3: Open the "raw/booking.db"
	rawDb, err := sql.Open("sqlite3", "../../../../../../../data/raw/accommocation/booking-com/booking.db")
	if err != nil {
		log.Fatalf("Failed to open booking.db: %v", err)
	}
	defer rawDb.Close()

	// Step 4: Query the 'property' table for records where review_score > 7
	query := `SELECT city, country, gross_price, checkin_date, checkout_date FROM property WHERE review_score > 7`
	rows, err := rawDb.Query(query)
	if err != nil {
		log.Fatalf("Failed to query property table: %v", err)
	}
	defer rows.Close()

	// Variables to collect checkin_date and checkout_date once
	var checkinDate, checkoutDate string
	// Step 5: Collect prices for each unique location
	locationData := make(map[string]LocationPrices)

	for rows.Next() {
		var acc Accommodation
		err := rows.Scan(&acc.City, &acc.Country, &acc.GrossPrice, &acc.Checkin, &acc.Checkout)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		// Collect checkin_date and checkout_date only once
		if checkinDate == "" && checkoutDate == "" {
			checkinDate = acc.Checkin
			checkoutDate = acc.Checkout
		}

		locationKey := fmt.Sprintf("%s,%s", acc.City, acc.Country)
		if _, exists := locationData[locationKey]; !exists {
			locationData[locationKey] = LocationPrices{
				City:    acc.City,
				Country: acc.Country,
				Prices:  []float64{},
			}
		}

		location := locationData[locationKey]
		location.Prices = append(location.Prices, acc.GrossPrice)
		locationData[locationKey] = location
	}

	// Step 6: Set up the progress bar for processing the locations
	bar := progressbar.Default(int64(len(locationData)))

	// Step 7: Process each location's prices and insert into new_main.db
	for _, loc := range locationData {
		bar.Add(1)

		// Sort the prices (lowest to highest)
		sort.Float64s(loc.Prices)

		// Calculate the 10% drop count
		numEntries := len(loc.Prices)
		if numEntries < 10 {
			fmt.Printf("Not enough entries for %s, %s to drop 10%% outliers\n", loc.City, loc.Country)
			continue
		}

		dropCount := int(math.Floor(float64(numEntries) * 0.10))

		// Collect the remaining prices (middle 80%)
		remainingPrices := loc.Prices[dropCount : numEntries-dropCount]

		// Sort remaining prices (lowest to highest)
		sort.Float64s(remainingPrices)

		// Step 8: Calculate the median of remaining prices
		medianPrice := calculateMedian(remainingPrices)

		// Step 9: Calculate avg_pppn by dividing the median by 14 and rounding to 2 decimal places
		avgPppn := roundToTwoDecimalPlaces(medianPrice / 14)

		// Step 10: Create the booking URL for this location
		bookingURL := fmt.Sprintf("https://www.booking.com/searchresults.en-gb.html?ss=%s&group_adults=1&no_rooms=1&group_children=0&nflt=price%%3DEUR-min-110-1%%3Breview_score%%3D80&flex_window=2&checkin=%s&checkout=%s", loc.City, checkinDate, checkoutDate)

		// Step 11: Insert the data into the accommodation table
		insertQuery := `INSERT INTO accommodation (city, country, booking_url, booking_pppn) VALUES (?, ?, ?, ?)`
		_, err := newDb.Exec(insertQuery, loc.City, loc.Country, bookingURL, avgPppn)
		if err != nil {
			log.Printf("Failed to insert accommodation for %s, %s: %v", loc.City, loc.Country, err)
		}
	}

	fmt.Println("Data inserted into new_main.db successfully!")
}

// Function to calculate the median of a sorted list of prices
func calculateMedian(prices []float64) float64 {
	n := len(prices)
	if n%2 == 0 {
		return (prices[n/2-1] + prices[n/2]) / 2
	}
	return prices[n/2]
}

// Function to round to two decimal places
func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}

// backup.go
func BackupDatabase(dbPath string) error {
	// Check if results.db exists
	if _, err := os.Stat(dbPath); err == nil {
		// get parent directory of the database
		dbDir := filepath.Dir(dbPath)
		// Create /out/backups directory if it does not exist
		backupDir := filepath.Join(dbDir, "backups")
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			err := os.Mkdir(backupDir, 0755)
			if err != nil {
				log.Printf("Failed to create backup directory %s: %v", backupDir, err)
				return err
			}
		}

		// Create a timestamped backup file path
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, fmt.Sprintf("main_backup_%s.db", timestamp))

		// Copy the database to the backup file
		err := CopyFile(dbPath, backupPath)
		if err != nil {
			log.Printf("Failed to copy database to backup: %v", err)
			return err
		}
	}
	return nil
}
