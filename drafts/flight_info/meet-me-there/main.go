
package main

import (
    "database/sql"
    "fmt"
    "io/ioutil"
    "log"
    "strings"

    _ "github.com/mattn/go-sqlite3"
    "gopkg.in/yaml.v2"
)

type FlightCriteria struct {
    Direction string `yaml:"direction"`
    Airport   string `yaml:"airport"`
    StartDate string `yaml:"startDate"`
    EndDate   string `yaml:"endDate"`
}

type Config struct {
    Flights []FlightCriteria `yaml:"flights"`
}

type Flight struct {
    ID               int
    FlightNumber     string
    DepartureAirport string
    ArrivalAirport   string
    DepartureTime    string
    ArrivalTime      string
    Direction        string
}

func loadConfig(file string) (*Config, error) {
    cfg := &Config{}
    yamlFile, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }
    err = yaml.Unmarshal(yamlFile, cfg)
    if err != nil {
        return nil, err
    }
    return cfg, nil
}

func main() {
    cfg, err := loadConfig("input/places-and-dates-to-match.yaml")
    if err != nil {
        log.Fatalf("Failed to load YAML config: %v", err)
    }

    db, err := sql.Open("sqlite3", "input/flights.db")
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    arrivalAirports := make(map[string]struct{})
    departureAirports := make(map[string]struct{})

    for _, criteria := range cfg.Flights {
        startDate, endDate := convertDateToISO(criteria.StartDate), convertDateToISO(criteria.EndDate)
        if strings.ToLower(criteria.Direction) == "departure" {
            query := `SELECT DISTINCT ArrivalAirport FROM flights WHERE DepartureAirport = ? AND DepartureTime BETWEEN ? AND ?`
            rows, err := db.Query(query, criteria.Airport, startDate, endDate)
            if err != nil {
                log.Printf("Failed to query departure flights: %v", err)
                continue
            }
            defer rows.Close()
            var airport string
            for rows.Next() {
                err := rows.Scan(&airport)
                if err != nil {
                    log.Printf("Failed to scan airport: %v", err)
                    continue
                }
                arrivalAirports[airport] = struct{}{}
            }
        } else if strings.ToLower(criteria.Direction) == "arrival" {
            query := `SELECT DISTINCT DepartureAirport FROM flights WHERE ArrivalAirport = ? AND ArrivalTime BETWEEN ? AND ?`
            rows, err := db.Query(query, criteria.Airport, startDate, endDate)
            if err != nil {
                log.Printf("Failed to query arrival flights: %v", err)
                continue
            }
            defer rows.Close()
            var airport string
            for rows.Next() {
                err := rows.Scan(&airport)
                if err != nil {
                    log.Printf("Failed to scan airport: %v", err)
                    continue
                }
                departureAirports[airport] = struct{}{}
            }
        }
    }

    commonAirports := findCommonAirports(arrivalAirports, departureAirports)
    fmt.Println("Common Airports:", commonAirports)
}

func convertDateToISO(date string) string {
    parts := strings.Split(date, "-")
    return fmt.Sprintf("%s-%s-%s", parts[2], parts[1], parts[0])
}

func findCommonAirports(arrivals, departures map[string]struct{}) []string {
    common := []string{}
    for arrival := range arrivals {
        if _, exists := departures[arrival]; exists {
            common = append(common, arrival)
        }
    }
    return common
}

