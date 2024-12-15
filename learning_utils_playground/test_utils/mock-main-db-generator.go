package test_utils

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
)

type insertFunc func([]string) error

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
	"duration_in_minutes"	DECIMAL,
  "duration_in_hours"	DECIMAL,
  "duration_hour_dot_mins" REAL
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

func SetupMockDatabase(inputDataDir string, outputDir string) {
	// Profiling setup
	cleanup := setupProfiling()
	defer cleanup()

	// Ensure the output path is formatted correctly
	outputDB := filepath.Join(outputDir, "test.db")

	// Attempt to create the directory
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

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
		csvFile := filepath.Join(inputDataDir, tableName+".csv")
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

// loadWeatherData loads weather data from a CSV file into the database
// It adjusts the date based on the offset between the oldest date and today
// It uses a transaction for batch inserts and parallel processing for performance
func loadWeatherData(db *sql.DB, csvFile string) error {
	rows, headers, err := readCSVFile(csvFile)
	if err != nil {
		return err
	}

	dateIndex := findColumnIndex(headers, "date")
	if dateIndex == -1 {
		return fmt.Errorf("no date column found in %s", csvFile)
	}

	dates, err := parseDates(rows, dateIndex)
	if err != nil {
		return err
	}

	offset := calculateDateOffset(dates)
	insertQuery := buildInsertQuery("weather", headers)

	// Use a transaction for batch inserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	// setup insertion function
	insertionFunc := insertWeatherRowFunc(tx, dateIndex, offset, insertQuery)

	err = genericParallelProcessing(rows, insertionFunc, "weather")
	if err != nil {
		return fmt.Errorf("could not insert weather data: %w", err)
	}

	// Commit the transaction after all rows have been processed
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

// loadCSVToTable loads generic CSV data into a specified table with a progress bar
// It uses a transaction for batch inserts and parallel processing for performance
func loadCSVToTable(db *sql.DB, csvFile, tableName string) error {
	rows, headers, err := readCSVFile(csvFile)
	if err != nil {
		return err
	}
	insertQuery := buildInsertQuery(tableName, headers)

	// Use a transaction for batch inserts
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	// setup insertion function
	insertionFunc := insertRowFunc(tx, insertQuery)

	err = genericParallelProcessing(rows, insertionFunc, tableName)
	if err != nil {
		return fmt.Errorf("could not insert data: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// genericParallelProcessing processes csv rows in worker groups
// using a generic insert function. It returns an error if any row fails to insert.
func genericParallelProcessing(rows [][]string, insertionFunc insertFunc, tableName string) error {
	// Prepare progress bar
	bar := progressbar.Default(int64(len(rows)-1), fmt.Sprintf("Inserting %s data", tableName))

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	numWorkers := 4
	rowCh := make(chan []string, len(rows)-1)

	// start multiple worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// get rows from the channel, process them, and increment the progress bar
			for row := range rowCh {
				err := insertionFunc(row)
				if err != nil {
					log.Printf("error processing row: %v", err)
				}
				bar.Add(1)
			}
		}()
	}

	// Send rows to workers
	for _, row := range rows[1:] {
		rowCh <- row
	}
	close(rowCh)
	// Wait for all workers to finish
	wg.Wait()
	return nil
}

// creates a function to insert a weather row, used for generic concurrent processing
func insertWeatherRowFunc(tx *sql.Tx, dateIndex int, offset float64, insertQuery string) insertFunc {
	return func(row []string) error {
		date, _ := time.Parse("2006-01-02", row[dateIndex])
		newDate := date.Add(time.Duration(offset) * 24 * time.Hour).Format("2006-01-02")
		row[dateIndex] = newDate

		placeholders := make([]interface{}, len(row))
		for i, value := range row {
			if value == "" {
				placeholders[i] = nil
			} else {
				placeholders[i] = value
			}
		}

		_, err := tx.Exec(insertQuery, placeholders...)
		if err != nil {
			log.Printf("could not insert row into weather table: %v", err)
		}
		return err
	}
}

// creates a function to insert a generic row, used for generic concurrent processing
func insertRowFunc(tx *sql.Tx, insertQuery string) insertFunc {
	return func(row []string) error {
		placeholders := make([]interface{}, len(row))
		for i, value := range row {
			if value == "" {
				placeholders[i] = nil
			} else {
				placeholders[i] = value
			}
		}

		_, err := tx.Exec(insertQuery, placeholders...)
		if err != nil {
			log.Printf("could not insert row into table: %v", err)
		}
		return err
	}
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

// readCSVFile reads a CSV file and returns the rows and headers
func readCSVFile(csvFile string) ([][]string, []string, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open file %s: %w", csvFile, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read CSV data: %w", err)
	}

	if len(rows) < 1 {
		return nil, nil, fmt.Errorf("CSV file %s is empty", csvFile)
	}

	headers := rows[0]
	return rows, headers, nil
}

// parseDates converts date strings to time.Time objects
func parseDates(rows [][]string, dateIndex int) ([]time.Time, error) {
	dates := make([]time.Time, len(rows)-1)
	for i, row := range rows[1:] {
		date, err := time.Parse("2006-01-02", row[dateIndex])
		if err != nil {
			return nil, fmt.Errorf("invalid date format in row %d: %w", i+2, err)
		}
		dates[i] = date
	}
	return dates, nil
}

// calculateDateOffset calculates the offset in days between the oldest date and today
func calculateDateOffset(dates []time.Time) float64 {
	today := time.Now()
	oldestDate := findOldestDate(dates)
	return today.Sub(oldestDate).Hours() / 24
}

// setupProfiling sets up CPU and memory profiling and returns a cleanup function
func setupProfiling() func() {
	// Profiling setup
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}

	// Memory profiling
	memProf, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}

	// Return a cleanup function
	return func() {
		pprof.StopCPUProfile()
		f.Close()
		pprof.WriteHeapProfile(memProf)
		memProf.Close()
	}
}
