package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/code_analysis"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/test_utils"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/time_utils"
	"github.com/Tris20/FairFareFinder/learning_utils_playground/workflows"
)

// func initDatabase(dbPath string) (*sql.DB, error) {
// 	db, err := sql.Open("sqlite3", dbPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open SQLite database: %v", err)
// 	}
// 	return db, nil
// }

// constants
const (
	mainDBPath    = "../data/compiled/main.db"
	flightsDBPath = "./testdata/flights.db"
	// locationsDBPath = "./testdata/locations.db"
	airpotsDBPath   = "./testdata/airports.db"
	configFilePath  = "../config/config.yaml"
	secretsFilePath = "../ignore/secrets.yaml"
	originsYamlPath = "../config/origins.yaml"
)

func main() {
	// data_management.RunAerodatabox()
	// possibly broken
	// data_management.RunFetchFlightPrices()
	// need to setup the database first, "no such table: airport"
	// data_management.RunFetchWeather_UpdateWeatherDB()

	// data_management.RunBookingComGetProperties()
	// needs the all_weather table to be  setup first
	// data_management.RunCalculateWeather()

	// workflows.CreateNewMainDB(mainDBPath, configFilePath, secretsFilePath, flightsDBPath)
	workflows.RunFetchFlightPrices(originsYamlPath, flightsDBPath, airpotsDBPath)

	// db, err := initDatabase("test.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db.Close()

	// propertyFecth := db_manager.PropertyFetch{}
	// _, err = db.Exec(propertyFecth.CreateTableQuery())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// data_management.InitFlights(db)
	// data_management.InitLocationsDB(db)
	// data_management.SetIataCitiesToTrue(db)

	// db2, err := initDatabase("test_main.db")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer db2.Close()

	// setup_mock_database := true

	// if setup_mock_database {
	// 	// Get the path of the executable
	// 	execDir, err := getExecutablePath()
	// 	if err != nil {
	// 		fmt.Println("Error getting executable path:", err)
	// 		return
	// 	}
	// 	// Setup paths
	// 	inputDataDir := filepath.Join(execDir, "test_utils/input-data")
	// 	outputDir := filepath.Join(execDir, "../testdata")

	// 	fmt.Println(inputDataDir)
	// 	fmt.Println(outputDir)

	// 	fmt.Println("Setting up mock database")

	// 	test_utils.SetMutePrints(true)
	// 	test_utils.SetupMockDatabase(execDir, inputDataDir, outputDir, false)
	// 	fmt.Println("Finished setting up mock database")
	// }
}

func getExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return "", err
	}
	// Get the directory of the executable
	execDir := filepath.Dir(execPath)

	// Check if running with 'go run'
	if strings.Contains(execDir, "go-build") {
		// Fallback to current working directory
		execDir, err = os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			return "", err
		}
	}
	return execDir, nil
}

func forceLoadingPackage() {
	db, err := sql.Open("sqlite3", "test_force.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Get the path of the executable
	execDir, err := getExecutablePath()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	fmt.Println(execDir)

	setup_mock_database := false

	if setup_mock_database {
		// Setup paths
		inputDataDir := filepath.Join(execDir, "test_utils/input-data")
		outputDir := filepath.Join(execDir, "../testdata")

		fmt.Println(inputDataDir)
		fmt.Println(outputDir)

		fmt.Println("Setting up mock database")

		test_utils.SetMutePrints(true)
		test_utils.SetupMockDatabase(execDir, inputDataDir, outputDir, false)
		fmt.Println("Finished setting up mock database")
	}

	do_time_stuff := false

	if do_time_stuff {

		fmt.Println("Get date range based on current day")

		// Get the current day
		start_date, end_date := time_utils.DetermineRangeBasedOnCurrentDay(time.Now().Weekday())

		fmt.Println("Start date:", start_date)
		fmt.Println("End date:", end_date)

	}

	check_file_css := false

	if check_file_css {
		fmt.Println("Check for css conflicts")
		files := []string{"../src/frontend/css/styles.css", "../src/frontend/css/tableStyles.css"}
		code_analysis.DetectCSSConflict_FileBased(files)
	}

	server_on := false

	if server_on {
		fmt.Println("check url css conflicts")
		code_analysis.DetectCSSConflict_URLBased()
	}

	do_dir_analysis := false

	if do_dir_analysis {

		fmt.Println("Get functions in directory")
		code_analysis.GetFunctionsInDir("./code_analysis")
	}

	fmt.Println("doing data management")

	// data_management.GetProperties_BookingCom("test", db)
}
