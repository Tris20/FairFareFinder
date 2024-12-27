package workflows

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/data_management"
)

// collect the flight schedule data from Aerodatabox
func RunAerodatabox(configFilePath, secretsFilePath, flightsDBPath string) error {
	apiClient := data_management.NewRealFlightAPIClient()
	err := data_management.FetchFlightSchedule(apiClient, configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		log.Printf("Failed to fetch flight schedule: %v", err)
		return err
	}
	return nil
}

// collect the flight prices
func RunFetchFlightPrices(originsYamlPath, flightsDBPath, airpotsDBPath string) {
	// todo: find out if this is actually broken or not
	// the origins.yaml file hasn't been updated since march
	err := data_management.FetchFlightPrices(originsYamlPath, flightsDBPath, airpotsDBPath)
	if err != nil {
		log.Fatalf("Failed to fetch flight prices: %v", err)
	}
}

// original files: utils/data/fetch/weather/main.go, weather.go, database.go
// new file: learning_utils_playground/data_management/fetch_weather.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "fetch/weather"), "update-weather-db")
// fmt.Printf("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)
func RunFetchWeather_UpdateWeatherDB() {
	// Fetch weather data
	weatherDBPath := "./testdata/weather.db"
	locationsDBPath := "./testdata/locations.db"
	err := data_management.FetchWeatherMain_UpdateWeatherDB(weatherDBPath, locationsDBPath)
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v", err)
	}
}

// original files: utils/data/fetch/accommodation/booking-com/get-properties/main.go
// new file: learning_utils_playground/data_management/fetch_accommodation_info.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "fetch/accommocation/booking-com/get-properties"), "get-properties")
// fmt.Printf("%sCOMPLETED: get-properties (properties update)%s\n", green, reset)
func RunBookingComGetProperties() {
	bookingDBPath := "./testdata/booking.db"
	secretsFilePath := "../ignore/secrets.yaml"
	data_management.GetProperties_BookingCom("", bookingDBPath, secretsFilePath)
}

// original files: utils/data/process/calculate/weather/main.go, utils.go
// new file: learning_utils_playground/data_management/process_calculate_weather.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/calculate/weather"), "weather")
// fmt.Printf("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)
func RunCalculateWeather() {
	weatherDBPath := "./testdata/weather.db"
	weatherPleasantnessYamlPath := "../config/weatherPleasantness.yaml"
	data_management.ProcessDatabaseEntries(weatherDBPath, weatherPleasantnessYamlPath)
}

// original files: utils/data/process/compile/main/flights/main.go, get_iso_code_of_country.go
// new file: learning_utils_playground/data_management/process_compile_main.go, data_mapping.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/flights"), "flights")
// fmt.Printf("%sCOMPLETED: flights (process compile)%s\n", green, reset)
func RunCompileMainFlights() {
	// todo: in progress
	data_management.ProcessCompileMainFlights()
}

// original files: utils/data/process/compile/main/weather/main.go
// new file: learning_utils_playground/data_management/process_compile_main.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/weather"), "weather")
// fmt.Printf("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)
func RunCompileMainWeather() {
	// todo: in progress
	data_management.ProcessCompileMainWeather()
}

// original files: utils/data/process/compile/main/locations/main.go, avg-wpi.go
// new file: learning_utils_playground/data_management/process_compile_main.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/locations"), "locations")
// fmt.Printf("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)
func RunCompileMainLocations() {
	// todo: in progress
	// UpdateAvgWPI()
	data_management.ProcessCompileMainLocations()
}

// original files: utils/data/process/compile/main/accommodation/booking-com/main.go
// new file: learning_utils_playground/data_management/process_compile_main.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/accommodation/booking-com"), "booking-com")
// fmt.Printf("%sCOMPLETED:  process/compile/main/accommodation/booking-com%s\n", green, reset)
func RunCompileMainBookingCom() {
	// todo: in progress
	data_management.ProcessCompileMainAccomodation()
}

// original files: utils/data/process/calculate/main/five-nights-and-flights/main.go
// new file: learning_utils_playground/data_management/process_calculate_fnf.go
//
// old code:
// runExecutableInDir(filepath.Join(relativeBase, "process/calculate/main/five-nights-and-flights"), "five-nights-and-flights")
// fmt.Printf("%sCOMPLETED:  process/calculate/main/five-nights-and-flights%s\n", green, reset)
func RunProcessCalculateFNF() {
	// todo: in progress
	data_management.ProcessCalculateFNF()
}

// original files: utils/data/process/calculate/flights/flight-duration/main.go
// new file: learning_utils_playground/data_management/process_calculate_flights.go
//
// old code:
// runExecutableInDir(filepath.Join(absoluteBase, "process/calculate/flights/flight-duration"), "flight-duration")
// log.Println("%sCOMPLETED: process/calculate/flights/flight-duration%s\n", green, reset)
func RunCalculateFlightDuration() {
	// todo: in progress
	data_management.CalculateFlightDurations()
}

// original files: utils/data/process/compile/locations/location-images/main.go
// new file: learning_utils_playground/data_management/manage_images.go
//
// old code:
// runExecutableInDir(filepath.Join(absoluteBase, "process/compile/locations/location-images"), "location-images")
// log.Println("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)
func RunProcessLocationImages() {

}

// run-flags.go

// Function to run all tasks in sequence
func runAllTasks(mainDbPath, configFilePath, secretsFilePath, flightsDBPath string,
	originsYamlPath, airpotsDBPath string) error {
	// Backup existing database if it exists
	err := data_management.BackupDatabase(mainDbPath)
	if err != nil {
		log.Printf("Failed to backup database: %v", err)
		return err
	}
	// Initialize the new database and create tables
	err = data_management.InitializeDatabase(mainDbPath)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return err
	}

	// FETCH
	err = RunAerodatabox(configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		log.Printf("Failed to run Aerodatabox: %v", err)
		return err
	}
	// possibly broken
	RunFetchFlightPrices(originsYamlPath, flightsDBPath, airpotsDBPath)
	// need to setup the database first, "no such table: airport"
	RunFetchWeather_UpdateWeatherDB()
	RunBookingComGetProperties()
	//Calculate
	// needs the all_weather table to be  setup first
	RunCalculateWeather()
	//Compile
	RunCompileMainFlights()
	RunCompileMainWeather()
	RunCompileMainLocations()
	RunCompileMainBookingCom()

	// Calculate again
	RunProcessCalculateFNF()
	return nil
}

// Function to run only compile tasks
func runCompileTasks() {
	RunCalculateWeather()
	RunCompileMainFlights()
	RunCompileMainWeather()
	RunCompileMainLocations()
	RunCompileMainBookingCom()
	RunProcessCalculateFNF()
}

// Function to run only weather-related tasks
func runWeatherTasks() {
	RunFetchWeather_UpdateWeatherDB()
	RunCalculateWeather()
	//included beacuse we always create a new completely new main.db, so need to rebuild the flights table
	RunCompileMainFlights()

	RunCompileMainWeather()
	//included beacuse we always create a new completely new main.db, so need to rebuild the locations table
	RunCompileMainLocations()
}

func transferFlightsDB(absoluteNewMainDbPath string) error {
	// Get the user's home directory
	//homeDir, err := os.UserHomeDir()
	homeDir := "/home/tristan" // os.userhomedir returns root/ which is incorrect, so we hard code here
	/*	if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
	*/
	// Build the full path to the SSH key
	sshKeyPath := filepath.Join(homeDir, ".ssh", "fff_server")

	// Define the maximum number of retries
	maxRetries := 13

	for i := 0; i <= maxRetries; i++ {
		cmd := exec.Command("scp", "-i", sshKeyPath, absoluteNewMainDbPath, "root@fairfarefinder.com:~/FairFareFinder/data/compiled/new_main.db")

		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		// Run the command and capture any error

		err := cmd.Run()

		if err == nil {
			log.Printf("Operations completed successfully")
			return nil
		}

		log.Printf("Attempt %d: SCP failed with error: %v", i+1, err)
		log.Printf("Attempt %d: SCP stdout: %s", i+1, outBuf.String())
		log.Printf("Attempt %d: SCP stderr: %s", i+1, errBuf.String())

		if i < maxRetries {
			log.Printf("Request failed: %v. Retrying...", err)
			time.Sleep(time.Duration(2^(i+1)) * time.Second) // Exponential backoff
		}
	}

	return fmt.Errorf("failed to run scp command after %d attempts", maxRetries)
}

func CreateNewMainDB(mainDbPath, configFilePath, secretsFilePath, flightsDBPath string,
	originsYamlPath, airpotsDBPath string) error {
	// Backup existing database if it exists
	err := data_management.BackupDatabase(mainDbPath)
	if err != nil {
		log.Printf("Failed to backup database: %v", err)
		return err
	}

	// Delete the database
	err = data_management.DeleteDatabase(mainDbPath)
	if err != nil {
		log.Printf("Failed to delete database: %v", err)
		return err
	}

	// Initialize the new database and create tables
	err = data_management.InitializeDatabase(mainDbPath)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return err
	}

	//Fetch
	err = RunAerodatabox(configFilePath, secretsFilePath, flightsDBPath)
	if err != nil {
		log.Printf("Failed to run Aerodatabox: %v", err)
		return err
	}

	RunFetchFlightPrices(originsYamlPath, flightsDBPath, airpotsDBPath)
	RunFetchWeather_UpdateWeatherDB()
	RunBookingComGetProperties()

	//Calculate
	RunCalculateWeather()

	// Compile
	RunCompileMainFlights()
	RunCompileMainWeather()

	RunCompileMainLocations()

	RunCalculateFlightDuration()

	// add location image paths to table
	RunProcessLocationImages()

	RunCompileMainBookingCom()
	// 5 nights and flights
	RunProcessCalculateFNF()
	return nil
}

func PeriodicDatabaseUpdate(absoluteNewMainDbPath, absoluteOutputDir string) {
	// Backup existing database if it exists
	data_management.BackupDatabase(absoluteNewMainDbPath)

	RunFetchWeather_UpdateWeatherDB()

	// Run weather calculation after weather update completes
	RunCalculateWeather()

	// Run process/compile/main/weather after calculation
	RunCompileMainWeather()

	// Calcualte and Compile WPI for Locations
	RunCompileMainLocations()

	// add location image paths to table
	RunProcessLocationImages()
}

func NewProcessCompileMain(mainDbPath, absoluteOutputDir, relativeBase, absoluteBase string,
	runAll, runCompile, runWeather, daemonMode, transferDB, newDB bool,
	configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath string) {

	// Initial log setup
	if err := data_management.UpdateLogFile(); err != nil {
		log.Fatalf("Failed to initialize log file: %v", err)
	}

	if newDB {
		CreateNewMainDB(mainDbPath, configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath)
	}

	// If the --all flag is set, run all tasks sequentially
	if runAll {
		err := runAllTasks(mainDbPath, configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath)
		log.Fatalf("Failed to run all tasks: %v", err)
		return
	}

	// If the --compile flag is set, run only compile tasks
	if runCompile {
		runCompileTasks()
		return
	}

	// If the --weather flag is set, run only weather-related tasks
	if runWeather {
		runWeatherTasks()
		return
	}
	if transferDB {
		transferFlightsDB(mainDbPath)
		return
	}

	newDaemonMode(daemonMode, mainDbPath, absoluteOutputDir, configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath)

	//	 If no flags are set, print a message
	log.Println("No flags set. Use --all, --compile, --weather, or --daemon.")

	data_management.CloseLogFile()
}

func newDaemonMode(daemonMode bool, absoluteNewMainDbPath, absoluteOutputDir string,
	configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath string) {
	// If --daemon flag is set, run in an infinite loop
	if daemonMode {
		log.Println("Daemon mode is enabled. Running tasks in loop...")
		// Infinite loop for daemon mode
		for {

			if err := data_management.UpdateLogFile(); err != nil {
				log.Printf("Error updating log file: %v", err)
			}
			// Get current day and time
			currentDay, currentHour := data_management.GetCurrentTime()
			transfer := false
			shouldLongSleep := false
			hourInterval := 6 // every 6 hours
			wakeUpHour := 3   // 3am
			// Generate new db every 6 hours: 3 = 3am; 9am; 3pm; 9pm.
			// currentHour%6==0 and currentHour==3 can never happen at the same time
			if (currentHour+wakeUpHour)%hourInterval == 0 {
				shouldLongSleep = true
				// Monday, 3am, Start a completely new new_main.db
				if currentDay == time.Monday && currentHour == wakeUpHour {
					CreateNewMainDB(absoluteNewMainDbPath, configFilePath, secretsFilePath, flightsDBPath, originsYamlPath, locationsDBPath)
					transfer = true
				} else {
					PeriodicDatabaseUpdate(absoluteNewMainDbPath, absoluteOutputDir)
					transfer = true
				}

				if transfer {
					err := transferFlightsDB(absoluteNewMainDbPath)
					if err != nil {
						log.Println("Error occurred during transfer:", err)
						// You may exit or handle the error as needed
					}
					transfer = false
				}
			}
			// sleep should be in the main loop, otherwise it will only sleep when the condition is met
			// Sleep for a specified time interval before checking again
			if shouldLongSleep {
				if hourInterval >= 2 {
					time.Sleep(time.Duration(hourInterval-2) * time.Hour)
				} else {
					time.Sleep(10 * time.Minute)
				}
				shouldLongSleep = false
			} else {
				time.Sleep(10 * time.Minute) // Check every 10 minutes in daemon mode
			}
		}
	}
}
