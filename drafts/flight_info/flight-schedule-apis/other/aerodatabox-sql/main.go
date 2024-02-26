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
	"strings"
)

// Secrets struct to match the YAML structure
type Secrets struct {
	APIKeys struct {
		Aerodatabox string `yaml:"aerodatabox"`
	} `yaml:"api_keys"`
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
	direction := flag.String("direction", "Departure", "Flight direction: Departure or Arrival")
	airport := flag.String("airport", "EDI", "IATA airport code")
	date := flag.String("date", "27-02-2024", "Date in DD-MM-YYYY format")
	flag.Parse()

	apiKey, err := readAPIKey("../../../../../ignore/secrets.yaml")
	if err != nil {
		log.Fatalf("Error reading API key: %v", err)
	}

	db, err := sql.Open("sqlite3", "flights.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	createTableSQL := `CREATE TABLE IF NOT EXISTS flights (
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

	dateParts := strings.Split(*date, "-")
	if len(dateParts) != 3 {
		fmt.Println("Invalid date format. Please use DD-MM-YYYY.")
		return
	}
	dateFormatted := fmt.Sprintf("%s-%s-%s", dateParts[2], dateParts[1], dateParts[0])

	intervals := []struct {
		urlSuffix  string
		fileSuffix string
	}{
		{"T00:00/%sT11:59", "AM"},
		{"T12:00/%sT23:59", "PM"},
	}

	for _, interval := range intervals {
		url := fmt.Sprintf("https://aerodatabox.p.rapidapi.com/flights/airports/iata/%s/%s"+interval.urlSuffix+"?withLeg=true&direction=%s&withCancelled=true&withCodeshared=true&withLocation=false", *airport, dateFormatted, dateFormatted, *direction)
		data, err := fetchFlightData(url, apiKey)
		if err != nil {
			log.Fatalf("Error fetching flight data: %v", err)
		}

		if *direction == "Arrival" {
			var arrivals ArrivalData
			err = json.Unmarshal(data, &arrivals)
			if err != nil {
				log.Fatalf("Error unmarshaling arrivals data: %v", err)
			}

			for _, arrival := range arrivals.Arrivals {
				_, err = db.Exec("INSERT INTO flights (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
					arrival.Number, arrival.Departure.Airport.IATA, *airport, arrival.Departure.ScheduledTime.Local, arrival.Arrival.ScheduledTime.Local, "Arrival")
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
				_, err = db.Exec("INSERT INTO flights (flightNumber, departureAirport, arrivalAirport, departureTime, arrivalTime, direction) VALUES (?, ?, ?, ?, ?, ?)",
					departure.Number, *airport, departure.Arrival.Airport.IATA, departure.Departure.ScheduledTime.Local, departure.Arrival.ScheduledTime.Local, "Departure")
				if err != nil {
					log.Printf("Error inserting departure into database: %v", err)
				}
			}
		}
	}
	fmt.Println("Flight data successfully fetched and stored.")
}
