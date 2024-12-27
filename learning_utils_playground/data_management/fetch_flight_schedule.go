package data_management

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	db_manager "github.com/Tris20/FairFareFinder/learning_utils_playground/database_manager"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/test_utils"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/time_utils"
	"gopkg.in/yaml.v2"
)

// FlightQuery struct to organize the query parameters
type FlightQuery struct {
	Direction    string             `yaml:"direction"`
	Airport      string             `yaml:"airport"`
	StartTime    string             `yaml:"startDate"`
	EndTime      string             `yaml:"endDate"`
	apiUrl       string             `yaml:"-"`
	responseData FlightResponseData `yaml:"-"`
}

type Configs struct {
	Airports []string `yaml:"airports"`
}

type FlightResponseData interface {
	InsertIntoDB(db *sql.DB, airport string) error
}

// FlightData structs to match the JSON structure
type ArrivalData struct {
	Arrivals []struct {
		Departure struct {
			Airport struct {
				ICAO string `json:"icao"`
				IATA string `json:"iata"`
				Name string `json:"name"`
			} `json:"airport"`
			ScheduledTime struct {
				UTC   string `json:"utc"`
				Local string `json:"local"`
			} `json:"scheduledTime"`
		} `json:"departure"`
		Arrival struct {
			ScheduledTime struct {
				UTC   string `json:"utc"`
				Local string `json:"local"`
			} `json:"scheduledTime"`
		} `json:"arrival"`
		Number string `json:"number"`
	} `json:"arrivals"`
}

func (a ArrivalData) InsertIntoDB(db *sql.DB, airport string) error {
	// direction := "Arrival"
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, arrival := range a.Arrivals {
		// Check if the entry already exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM schedule WHERE flightNumber = ? AND departureTime = ? AND arrivalTime = ?)",
			arrival.Number, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local).Scan(&exists)
		if err != nil {
			log.Printf("Error checking for existing record: %v", err)
		}
		if !exists {
			_, err := tx.Exec("INSERT INTO schedule(flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
				arrival.Number, arrival.Departure.Airport.IATA, airport, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local, "Arrival")
			if err != nil {
				log.Printf("Error inserting arrival into database: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// todo: move this to database manager so that the table name is consistent
// and the query / insert can be reused

type DepartureData struct {
	Departures []struct {
		Departure struct {
			ScheduledTime struct {
				UTC   string `json:"utc"`
				Local string `json:"local"`
			} `json:"scheduledTime"`
		} `json:"departure"`
		Arrival struct {
			Airport struct {
				ICAO string `json:"icao"`
				IATA string `json:"iata"`
				Name string `json:"name"`
			} `json:"airport"`
			ScheduledTime struct {
				UTC   string `json:"utc"`
				Local string `json:"local"`
			} `json:"scheduledTime"`
		} `json:"arrival"`
		Number string `json:"number"`
	} `json:"departures"`
}

func (d DepartureData) InsertIntoDB(db *sql.DB, airport string) error {
	// direction := "Departure"
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, departure := range d.Departures {
		// Check if the entry already exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM schedule WHERE flightNumber = ? AND departureTime = ? AND arrivalTime = ?)",
			departure.Number, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local).Scan(&exists)
		if err != nil {
			log.Printf("Error checking for existing record: %v", err)
		}
		if !exists {
			_, err := tx.Exec("INSERT INTO schedule (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
				departure.Number, airport, departure.Arrival.Airport.IATA, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local, "Departure")
			if err != nil {
				log.Printf("Error inserting departure into database: %v", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

type FlightAPIClient interface {
	FetchFlightData(url string) ([]byte, error)
	SetAPIKey(apiKey string)
}

type RealFlightAPIClient struct {
	ticker  *time.Ticker
	apiKey  string
	logFile *os.File
}

func NewRealFlightAPIClient() *RealFlightAPIClient {
	return &RealFlightAPIClient{
		ticker: time.NewTicker(2 * time.Second),
	}
}

func (c *RealFlightAPIClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}

func (c *RealFlightAPIClient) SetLogFile(logFilePath string) error {
	var err error
	c.logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

func (c *RealFlightAPIClient) FetchFlightData(url string) ([]byte, error) {
	// Wait on each ticker's tick before proceeding with the API call
	<-c.ticker.C

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", c.apiKey)
	req.Header.Add("X-RapidAPI-Host", "aerodatabox.p.rapidapi.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log the response to the file
	if c.logFile != nil {
		c.logFile.WriteString(fmt.Sprintf("URL: %s\nResponse: %s\n\n", url, string(body)))
	}

	return body, nil
}

func (c *RealFlightAPIClient) StopTicker() {
	c.ticker.Stop()
	if c.logFile != nil {
		c.logFile.Close()
	}
}

func getFileDependencies(configFilePath, secretsFilePath, flightsDBPath string) (Configs, string, *sql.DB, error) {

	// Load configurations from YAML
	var configs Configs
	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Printf("Error reading config file %s: %v", configFilePath, err)
		return Configs{}, "", nil, err
	}
	err = yaml.Unmarshal(configFile, &configs)
	if err != nil {
		log.Printf("Error parsing config file %s: %v", configFilePath, err)
		return Configs{}, "", nil, err
	}

	apiKey, err := readAPIKey(secretsFilePath)
	if err != nil {
		log.Printf("Error reading API key: %v", err)
		return Configs{}, "", nil, err
	}

	db, err := sql.Open("sqlite3", flightsDBPath)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return Configs{}, "", nil, err
	}
	// must close the database connection after the function returns

	return configs, apiKey, db, nil
}

func FetchFlightSchedule(apiClient FlightAPIClient,
	configFilePath, secretsFilePath, flightsDBPath string) error {
	// start := time.Now()
	profile := true
	if profile {
		// Profiling setup
		cleanup, err := test_utils.SetupProfiling("testdata/FetchFlightSchedule_cpu.prof", "testdata/FetchFlightSchedule_mem.prof")
		if err != nil {
			log.Fatalf("Failed to setup profiling: %v", err)
		}
		defer cleanup()
	}
	// configFilePath := "../config/config.yaml"
	// secretsFilePath := "../ignore/secrets.yaml"
	// flightsDBPath := "testdata/flights.db"

	configs, apiKey, db, err := getFileDependencies(configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		log.Printf("Error getting file dependencies: %v", err)
		return err
	}
	defer func() {
		if realClient, ok := apiClient.(*RealFlightAPIClient); ok {
			realClient.StopTicker()
		}
		db.Close()
	}()

	apiClient.SetAPIKey(apiKey)

	rawDBFlight := db_manager.RawDBFlight{}
	_, err = db.Exec(rawDBFlight.CreateTableQuery())
	if err != nil {
		log.Printf("Error creating table %s: %v", rawDBFlight.TableName(), err)
		return err
	}

	// Generate dates using previously discussed CalculateWeekendRange function
	departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate := time_utils.CalculateWeekendRange(1)
	// Print the generated dates
	// fmt.Println("Departure Start Date:", departureStartDate)
	// fmt.Println("Departure End Date:", departureEndDate)
	// fmt.Println("Arrival Start Date:", arrivalStartDate)
	// fmt.Println("Arrival End Date:", arrivalEndDate)

	queryList_depature, err := constructQueryList(configs.Airports, "Departure", departureStartDate, departureEndDate)
	if err != nil {
		return err
	}
	queryList_arrival, err := constructQueryList(configs.Airports, "Arrival", arrivalStartDate, arrivalEndDate)
	if err != nil {
		return err
	}

	queryList := append(queryList_depature, queryList_arrival...)

	for _, query := range queryList {
		// fmt.Println(i)
		err = processFlightData(db, query, apiClient)
		if err != nil {
			log.Printf("Error processing flight data: %v", err)
			return err
		}
	}

	// log.Printf("FetchFlightSchedule took %v", time.Since(start))
	return nil
}

func constructQueryList(airports []string, direction, startDate, endDate string) ([]FlightQuery, error) {
	var queryList []FlightQuery
	// Parse the start and end dates into time.Time objects
	startDateTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		log.Printf("Error parsing start date: %v", err)
		return queryList, err
	}
	endDateTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		log.Printf("Error parsing end date: %v", err)
		return queryList, err
	}

	// Define two 12-hour intervals for each day: AM and PM
	time_intervals := []struct {
		start string
		end   string
	}{
		{"T00:00", "T11:59"},
		{"T12:00", "T23:59"},
	}

	for d := startDateTime; !d.After(endDateTime); d = d.AddDate(0, 0, 1) {
		for _, interval := range time_intervals {
			// Format start and end times for each interval
			startTime := fmt.Sprintf("%s%s", d.Format("2006-01-02"), interval.start)
			endTime := fmt.Sprintf("%s%s", d.Format("2006-01-02"), interval.end)

			for _, airport := range airports {
				query := FlightQuery{
					Direction: direction,
					Airport:   airport,
					StartTime: startTime,
					EndTime:   endTime,
				}

				query.apiUrl = constructAPIURL(query)

				if direction == "Arrival" {
					query.responseData = &ArrivalData{}
				} else { // Assume direction == "Departure"
					query.responseData = &DepartureData{}
				}

				queryList = append(queryList, query)
			}
		}
	}
	return queryList, nil
}

// produces a list of URLs to fetch flight data from
func constructAPIURL(query FlightQuery) string {
	base := "https://aerodatabox.p.rapidapi.com/flights/airports/iata/"
	url := fmt.Sprintf("%s%s/%s/%s?withLeg=true&direction=%s&withCancelled=true&withCodeshared=true&withLocation=false",
		base, query.Airport, query.StartTime, query.EndTime, query.Direction)
	return url
}

func processFlightData(db *sql.DB, query FlightQuery, apiClient FlightAPIClient) error {
	fmt.Println(query.apiUrl)

	// todo: before fetching data, check if the data already exists in the database
	// todo: add in data to register which requests have been made
	// 	so that we can check if the data has been fetched before (or an equivalent request)

	// Fetch the flight data
	data, err := apiClient.FetchFlightData(query.apiUrl)
	if err != nil {
		log.Printf("Error fetching flight data for %s: %v", query.Airport, err)
		return err
	}
	// Unmarshal the data into the interface
	if err := json.Unmarshal(data, query.responseData); err != nil {
		log.Printf("Error unmarshaling arrivals data: %v", err)
		return err
	}

	err = query.responseData.InsertIntoDB(db, query.Airport)
	if err != nil {
		log.Printf("Error inserting data into database: %v", err)
		return err
	}
	return nil
}
