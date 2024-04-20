package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Tris20/FairFareFinder/src/go_files/timeutils"
)

// Secrets struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
}

// FlightConfig and Configs structs to handle a date range
type FlightConfig struct {
	Direction string `yaml:"direction"`
	Airport   string `yaml:"airport"`
	StartDate string `yaml:"startDate"`
	EndDate   string `yaml:"endDate"`
}

type Configs struct {
	Airports []string `yaml:"airports"`
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

func readAPIKey(filepath string) (string, error) {
	var secrets Secrets
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(file, &secrets)
	if err != nil {
		return "", err
	}
	return secrets.APIKeys.Aerodatabox, nil
}

func fetchFlightData(url, apiKey string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", "aerodatabox.p.rapidapi.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func main() {
	//direction := flag.String("direction", "Departure", "Flight direction: Departure or Arrival")
	//airport := flag.String("airport", "EDI", "IATA airport code")
	//	date := flag.String("date", "27-02-2024", "Date in DD-MM-YYYY format")
	flag.Parse()

	// Load configurations from YAML
	var configs Configs
	configFile, err := ioutil.ReadFile("fetch-these-flights.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
	err = yaml.Unmarshal(configFile, &configs)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	apiKey, err := readAPIKey("../../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
	}

	//db, err := sql.Open("sqlite3", "../../../../../data/longterm_db/flights.db")
	db, err := sql.Open("sqlite3", "../../../../../data/flights.db")
  //db, err := sql.Open("sqlite3", "./flights.db")

	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS schedule (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        flightNumber TEXT NOT NULL,
        departureAirport TEXT,
        arrivalAirport TEXT,
        departureTime TEXT,
        arrivalTime TEXT,
        direction TEXT NOT NULL
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	bar := progressbar.NewOptions(len(configs.Airports)*2, // Assuming two operations (arrival and departure) per airport
		progressbar.OptionSetDescription("Processing flights"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	// Generate dates using previously discussed CalculateWeekendRange function
	departureStartDate, departureEndDate, arrivalStartDate, arrivalEndDate := timeutils.CalculateWeekendRange(1)
    // Print the generated dates
    fmt.Println("Departure Start Date:", departureStartDate)
    fmt.Println("Departure End Date:", departureEndDate)
    fmt.Println("Arrival Start Date:", arrivalStartDate)
    fmt.Println("Arrival End Date:", arrivalEndDate)

	for _, airport := range configs.Airports {
		// Handle departure data
		if err := processFlightData(db, airport, "Departure", departureStartDate, departureEndDate, apiKey); err != nil {
			log.Printf("Error processing departure data for airport %s: %v", airport, err)
			// Decide on error handling: halt or continue
		}
		bar.Add(1)

		// Handle arrival data
		if err := processFlightData(db, airport, "Arrival", arrivalStartDate, arrivalEndDate, apiKey); err != nil {
			log.Printf("Error processing arrival data for airport %s: %v", airport, err)
			// Decide on error handling: halt or continue
		}
		bar.Add(1)
	}
	fmt.Println("Flight data successfully fetched and stored.")
}



func processFlightData(db *sql.DB, airport, direction, startDate, endDate, apiKey string) error {
    // Parse the start and end dates into time.Time objects
    startDateTime, err := time.Parse("2006-01-02", startDate)
    if err != nil {
        log.Printf("Error parsing start date: %v", err)
        return err
    }
    endDateTime, err := time.Parse("2006-01-02", endDate)
    if err != nil {
        log.Printf("Error parsing end date: %v", err)
        return err
    }

    // Define two 12-hour intervals for each day: AM and PM
    intervals := []struct {
        start string
        end   string
    }{
        {"T00:00", "T11:59"},
        {"T12:00", "T23:59"},
    }


   // Set up a ticker for rate limiting: 30 calls/minute, one tick every 2 seconds
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for d := startDateTime; !d.After(endDateTime); d = d.AddDate(0, 0, 1) {
        for _, interval := range intervals {

            // Wait on each ticker's tick before proceeding with the API call
            <-ticker.C
            // Format start and end times for each interval
            startTime := fmt.Sprintf("%s%s", d.Format("2006-01-02"), interval.start)
            endTime := fmt.Sprintf("%s%s", d.Format("2006-01-02"), interval.end)

            // Construct the API URL
            url := fmt.Sprintf("https://aerodatabox.p.rapidapi.com/flights/airports/iata/%s/%s/%s?withLeg=true&direction=%s&withCancelled=true&withCodeshared=true&withLocation=false", 
                airport, startTime, endTime, direction)

            fmt.Println("Fetching URL:", url) // Print the API request URL

          // Fetch the flight data
        data, err := fetchFlightData(url, apiKey)
        if err != nil {
            log.Printf("Error fetching flight data for %s: %v", airport, err)
            return err
        }

        if direction == "Arrival" {
            var arrivals ArrivalData
            if err := json.Unmarshal(data, &arrivals); err != nil {
                log.Printf("Error unmarshaling arrivals data: %v", err)
                return err
            }

            for _, arrival := range arrivals.Arrivals {
                // Check if the entry already exists
                var exists bool
                err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM flights WHERE flightNumber = ? AND departureTime = ? AND arrivalTime = ?)",
                    arrival.Number, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local).Scan(&exists)
                if err != nil {
                    log.Printf("Error checking for existing record: %v", err)
                }
                if !exists {
                    _, err := db.Exec("INSERT INTO flights(flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
                        arrival.Number, arrival.Departure.Airport.IATA, airport, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local, direction)
                    if err != nil {
                        log.Printf("Error inserting arrival into database: %v", err)
                    }
                }
            }
        } else { // Assume direction == "Departure"
            var departures DepartureData
            if err := json.Unmarshal(data, &departures); err != nil {
                log.Printf("Error unmarshaling departures data: %v", err)
                return err
            }

            for _, departure := range departures.Departures {
                // Check if the entry already exists
                var exists bool
                err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM flights WHERE flightNumber = ? AND departureTime = ? AND arrivalTime = ?)",
                    departure.Number, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local).Scan(&exists)
                if err != nil {
                    log.Printf("Error checking for existing record: %v", err)
                }
                if !exists {
                    _, err := db.Exec("INSERT INTO flights (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
                        departure.Number, airport, departure.Arrival.Airport.IATA, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local, direction)
                    if err != nil {
                        log.Printf("Error inserting departure into database: %v", err)
                    }
                }
            }
        }
    }
  }
    return nil
}
