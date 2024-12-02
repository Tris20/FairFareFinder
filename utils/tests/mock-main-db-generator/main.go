package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

// Define the schema for each table
var tableSchemas = map[string]string{
	"accommodation": `
	CREATE TABLE accommodation (
		city TEXT NOT NULL,
		country TEXT NOT NULL,
		booking_url TEXT,
		booking_pppn REAL NOT NULL
	)`,
	"five_nights_and_flights": `
	CREATE TABLE five_nights_and_flights (
		origin_city TEXT,
		origin_country TEXT,
		destination_city TEXT,
		destination_country TEXT,
		price_fnaf REAL
	)`,
	"flight": `
	CREATE TABLE flight (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		origin_city_name TEXT,
		origin_country TEXT,
		origin_iata TEXT,
		origin_skyscanner_id TEXT,
		destination_city_name TEXT,
		destination_country TEXT,
		destination_iata TEXT,
		destination_skyscanner_id TEXT,
		price_this_week DECIMAL,
		skyscanner_url_this_week VARCHAR(255),
		price_next_week DECIMAL,
		skyscanner_url_next_week VARCHAR(255),
		duration_in_minutes DECIMAL
	)`,
	"location": `
	CREATE TABLE location (
		city VARCHAR(255) NOT NULL,
		country CHAR(2) NOT NULL,
		iata_1 CHAR(3) NOT NULL,
		iata_2 CHAR(3),
		iata_3 CHAR(3),
		iata_4 CHAR(3),
		iata_5 CHAR(3),
		iata_6 CHAR(3),
		iata_7 CHAR(3),
		avg_wpi FLOAT(10,1),
		image_1 TEXT, image_2 TEXT, image_3 TEXT, image_4 TEXT, image_5 TEXT
	)`,
	"weather": `
	CREATE TABLE weather (
		city VARCHAR(255) NOT NULL,
		country CHAR(2) NOT NULL,
		date DATE NOT NULL,
		avg_daytime_temp FLOAT(10,1),
		weather_icon VARCHAR(255),
		google_url VARCHAR(255),
		avg_daytime_wpi FLOAT(10,1)
	)`,
}

func main() {
	// Paths
	inputFolder := "input-data"
	outputDB := "../../../data/compiled/main.db"

	// Open SQLite database
	db, err := sql.Open("sqlite3", outputDB)
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Iterate through the schemas and process CSV files
	for tableName, createStmt := range tableSchemas {
		// Create table
		_, err := db.Exec(createStmt)
		if err != nil {
			log.Fatalf("Failed to create table %s: %v", tableName, err)
		}
		fmt.Printf("Created table: %s\n", tableName)

		// Load data from CSV
		csvFile := filepath.Join(inputFolder, tableName+".csv")
		if tableName == "weather" {
			if err := loadWeatherData(db, csvFile); err != nil {
				log.Fatalf("Failed to load weather data: %v", err)
			}
		} else {
			if err := loadCSVToTable(db, csvFile, tableName); err != nil {
				log.Fatalf("Failed to load data for table %s: %v", tableName, err)
			}
		}
		fmt.Printf("Loaded data for table: %s\n", tableName)
	}

	fmt.Println("Database generation complete: main.db")
}

// loadWeatherData processes and inserts weather data with adjusted dates and a progress bar
func loadWeatherData(db *sql.DB, csvFile string) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", csvFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV data: %w", err)
	}

	if len(rows) < 1 {
		return fmt.Errorf("CSV file %s is empty", csvFile)
	}

	headers := rows[0]
	dateIndex := findColumnIndex(headers, "date")
	if dateIndex == -1 {
		return fmt.Errorf("no date column found in %s", csvFile)
	}

	// Parse dates and calculate offsets
	today := time.Now()
	dates := make([]time.Time, len(rows)-1)
	for i, row := range rows[1:] {
		date, err := time.Parse("2006-01-02", row[dateIndex])
		if err != nil {
			return fmt.Errorf("invalid date format in row %d: %w", i+2, err)
		}
		dates[i] = date
	}

	// Find the oldest date and calculate its offset
	oldestDate := findOldestDate(dates)
	offset := today.Sub(oldestDate).Hours() / 24

	// Prepare progress bar
	bar := progressbar.Default(int64(len(rows)-1), "Inserting weather data")

	// Adjust dates and insert data
	insertQuery := buildInsertQuery("weather", headers)
	for _, row := range rows[1:] {
		date, _ := time.Parse("2006-01-02", row[dateIndex])
		newDate := date.Add(time.Duration(offset) * 24 * time.Hour).Format("2006-01-02")
		row[dateIndex] = newDate

		placeholders := make([]interface{}, len(row))
		for i, value := range row {
			if value == "" { // Handle missing values
				placeholders[i] = nil
			} else {
				placeholders[i] = value
			}
		}

		_, err := db.Exec(insertQuery, placeholders...)
		if err != nil {
			return fmt.Errorf("could not insert row into weather table: %w", err)
		}

		// Increment progress bar
		bar.Add(1)
	}

	return nil
}

// loadCSVToTable loads generic CSV data into a specified table with a progress bar
func loadCSVToTable(db *sql.DB, csvFile, tableName string) error {
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", csvFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV data: %w", err)
	}

	if len(rows) < 1 {
		return fmt.Errorf("CSV file %s is empty", csvFile)
	}

	headers := rows[0]
	insertQuery := buildInsertQuery(tableName, headers)

	// Prepare progress bar
	bar := progressbar.Default(int64(len(rows)-1), fmt.Sprintf("Inserting %s data", tableName))

	for _, row := range rows[1:] {
		placeholders := make([]interface{}, len(row))
		for i, value := range row {
			if value == "" { // If the CSV entry is empty, set it to NULL
				placeholders[i] = nil
			} else {
				placeholders[i] = value
			}
		}

		_, err := db.Exec(insertQuery, placeholders...)
		if err != nil {
			return fmt.Errorf("could not insert row into table %s: %w", tableName, err)
		}

		// Increment progress bar
		bar.Add(1)
	}

	return nil
}

// findColumnIndex finds the index of a column in the CSV headers
func findColumnIndex(headers []string, column string) int {
	for i, header := range headers {
		if strings.EqualFold(header, column) {
			return i
		}
	}
	return -1
}

// findOldestDate returns the earliest date from a slice of time.Time
func findOldestDate(dates []time.Time) time.Time {
	oldest := dates[0]
	for _, date := range dates[1:] {
		if date.Before(oldest) {
			oldest = date
		}
	}
	return oldest
}

// buildInsertQuery generates an INSERT SQL query for the given table and columns
func buildInsertQuery(tableName string, columns []string) string {
	columnsList := strings.Join(columns, ", ")
	placeholders := strings.Repeat("?, ", len(columns))
	placeholders = strings.TrimSuffix(placeholders, ", ")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsList, placeholders)
}
