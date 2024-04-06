package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
"github.com/schollz/progressbar/v3"
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
	Flights []FlightConfig `yaml:"flights"`
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

	db, err := sql.Open("sqlite3", "../../../../../data/longterm_db/flights.db")
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

	// Initialize the progress bar
	bar := progressbar.NewOptions(len(configs.Flights),
		progressbar.OptionSetDescription("Processing flights"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	for _, flight := range configs.Flights {
		// Convert start and end dates to time.Time
		startDate, err := time.Parse("02-01-2006", flight.StartDate)
		if err != nil {
			log.Fatalf("Error parsing start date: %v", err)
		}
		endDate, err := time.Parse("02-01-2006", flight.EndDate)
		if err != nil {
			log.Fatalf("Error parsing end date: %v", err)
		}

		// Iterate over each day in the date range
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			dateFormatted := d.Format("2006-01-02")
			//	dateParts := strings.Split(*date, "-")
			//	if len(dateParts) != 3 {
			//		fmt.Println("Invalid date format. Please use DD-MM-YYYY.")
			//		return
			//	}
			//	dateFormatted := fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])

			intervals := []struct {
				urlSuffix  string
				fileSuffix string
			}{
				{"T00:00/%sT11:59", "AM"},
				{"T12:00/%sT23:59", "PM"},
			}

			for _, interval := range intervals {
				url := fmt.Sprintf("https://aerodatabox.p.rapidapi.com/flights/airports/iata/%s/%s"+interval.urlSuffix+"?withLeg=true&direction=%s&withCancelled=true&withCodeshared=true&withLocation=false", flight.Airport, dateFormatted, dateFormatted, flight.Direction)
				data, err := fetchFlightData(url, apiKey)
				if err != nil {
					log.Fatalf("Error fetching flight data: %v", err)
				}

				if flight.Direction == "Arrival" {
					var arrivals ArrivalData
					err = json.Unmarshal(data, &arrivals)
					if err != nil {
						log.Fatalf("Error unmarshaling arrivals data: %v", err)
					}

					for _, arrival := range arrivals.Arrivals {
						_, err = db.Exec("INSERT INTO schedule (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
							arrival.Number, arrival.Departure.Airport.IATA, flight.Airport, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local, "Arrival")
						if err != nil {
							log.Printf("Error inserting arrival into database: %v", err)
						}
					}
				} else {
					var departures DepartureData
					err = json.Unmarshal(data, &departures)
					if err != nil {
						log.Fatalf("Error unmarshaling departures data: %v", err)
					}

					for _, departure := range departures.Departures {
						_, err = db.Exec("INSERT INTO schedule (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
							departure.Number, flight.Airport, departure.Arrival.Airport.IATA, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local, "Departure")
						if err != nil {
							log.Printf("Error inserting departure into database: %v", err)
						}
					}
				}
			}
	  }
	bar.Add(1)
  }
	fmt.Println("Flight data successfully fetched and stored.")
}
