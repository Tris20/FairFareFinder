package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

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
func getCurrentTime() (time.Weekday, int) {
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
func updateLogFile() error {
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

func main() {

	//Colours for CLI comments
	green := "\033[32m"
	reset := "\033[0m"

	// Add flags for running all tasks or just compile tasks
	runAll := flag.Bool("all", false, "Run all tasks in sequence regardless of time")
	runCompile := flag.Bool("compile", false, "Run only compile tasks")
	runWeather := flag.Bool("weather", false, "Run only weather-related tasks")
	daemonMode := flag.Bool("daemon", false, "Run the program indefinitely as a daemon")
	transferDB := flag.Bool("transfer", false, "Performing transfer of new_main to webserver")

	//Create output directory if not exists
	outputDir := "../../../../../data/compiled/"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", outputDir, err)
		}
	}

	// Get the current directory of the script
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Set Database file paths
	newMainDBPath := filepath.Join(outputDir, "new_main.db")
	// Define the relative path to go up three directories
	relativeBase := filepath.Join(baseDir, "../../../")

	// Get the absolute path from the relative path
	absoluteNewMainDbPath, err := filepath.Abs(newMainDBPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	log.Printf("Absolute path of new_main.db: %s", absoluteNewMainDbPath)

	// Get the absolute path from the relative path
	absoluteOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	log.Printf("Absolute path of outputdir: %s", absoluteOutputDir)

	// Get the absolute path from the relative path
	absoluteBase, err := filepath.Abs(relativeBase)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	log.Printf("Absolute path of base directory utils/data: %s", absoluteBase)

	flag.Parse()

	// Initial log setup
	if err := updateLogFile(); err != nil {
		log.Fatalf("Failed to initialize log file: %v", err)
	}

	// If the --all flag is set, run all tasks sequentially
	if *runAll {
		// Backup existing database if it exists
		backupDatabase(absoluteNewMainDbPath, absoluteOutputDir)
		// Initialize the new database and create tables
		initializeDatabase(absoluteNewMainDbPath)

		runAllTasks(relativeBase)
		return
	}

	// If the --compile flag is set, run only compile tasks
	if *runCompile {
		runCompileTasks(relativeBase)
		return
	}

	// If the --weather flag is set, run only weather-related tasks
	if *runWeather {
		runWeatherTasks(relativeBase)
		return
	}
	if *transferDB {
		transferFlightsDB(absoluteNewMainDbPath)
		return
	}

	// If --daemon flag is set, run in an infinite loop
	if *daemonMode {
		log.Println("Daemon mode is enabled. Running tasks in loop...")
		// Infinite loop for daemon mode
		for {
			if err := updateLogFile(); err != nil {
				log.Printf("Error updating log file: %v", err)
			}
			// Get current day and time
			currentDay, currentHour := getCurrentTime()
			transfer := false

			// Generate new db every 6 hours: 3 = 3am; 9am; 3pm; 9pm.
			if currentHour%6 == 3 {

				// Monday, 3am, Start a completely new new_main.db
				if currentDay == time.Monday && currentHour == 3 {

					// Backup existing database if it exists
					backupDatabase(absoluteNewMainDbPath, absoluteOutputDir)
					log.Println("%sCOMPLETED: Backup of existing database%s\n", green, reset)

					// Delete existing new_main.db if it exists
					deleteNewMainDB(absoluteNewMainDbPath)
					log.Println("%sCOMPLETED: Deletion of new_main.db%s\n", green, reset)

					// Initialize the new database and create tables
					initializeDatabase(absoluteNewMainDbPath)
					log.Println("%sCOMPLETED: Initialization of new database%s\n", green, reset)

					//Fetch
					runExecutableInDir(filepath.Join(absoluteBase, "fetch/flights/schedule"), "aerodatabox")
					log.Println("%sCOMPLETED: aerodatabox (flight schedule)%s\n", green, reset)
					runExecutableInDir(filepath.Join(absoluteBase, "fetch/flights/prices"), "prices")
					log.Println("%sCOMPLETED: prices (flight prices)%s\n", green, reset)
					runExecutableInDir(filepath.Join(absoluteBase, "fetch/weather"), "update-weather-db")
					log.Println("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)

					// Temporarily paused due to high API cost. Existing values of accomodation raw db
					// are used by by the locations compiler stage instead
					//runExecutableInDir(filepath.Join(absoluteBase, "fetch/accommocation/booking-com/get-properties"), "get-properties")
					//log.Println("%sCOMPLETED: get-properties (properties update)%s\n", green, reset)

					//Calculate
					runExecutableInDir(filepath.Join(absoluteBase, "process/calculate/weather"), "weather")
					log.Println("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)

					// Compile
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/flights"), "flights")
					log.Println("%sCOMPLETED: flights (process compile)%s\n", green, reset)
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/weather"), "weather")
					log.Println("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)

					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/locations"), "locations")
					log.Println("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)

					runExecutableInDir(filepath.Join(absoluteBase, "process/calculate/flights/flight-duration"), "flight-duration")
					log.Println("%sCOMPLETED: process/calculate/flights/flight-duration%s\n", green, reset)

					// add location image paths to table
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/locations/location-images"), "location-images")
					log.Println("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)

					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/accommodation/booking-com"), "booking-com")
					log.Println("%sCOMPLETED:  process/compile/main/accommodation/booking-com%s\n", green, reset)
					// 5 nights and flights
					runExecutableInDir(filepath.Join(absoluteBase, "process/calculate/main/five-nights-and-flights"), "five-nights-and-flights")
					log.Println("%sCOMPLETED:  process/calculate/main/five-nights-and-flights%s\n", green, reset)
					transfer = true
				} else {
					// Backup existing database if it exists
					backupDatabase(absoluteNewMainDbPath, absoluteOutputDir)

					runExecutableInDir(filepath.Join(absoluteBase, "fetch/weather"), "update-weather-db")
					log.Println("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)

					// Run weather calculation after weather update completes
					runExecutableInDir(filepath.Join(absoluteBase, "process/calculate/weather"), "weather")
					log.Println("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)

					// Run process/compile/main/weather after calculation
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/weather"), "weather")
					log.Println("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)

					// Calcualte and Compile WPI for Locations
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/main/locations"), "locations")
					log.Println("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)

					// add location image paths to table
					runExecutableInDir(filepath.Join(absoluteBase, "process/compile/locations/location-images"), "location-images")
					log.Println("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)
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

				// Sleep for a specified time interval before checking again
				time.Sleep(50 * time.Minute) // Check every 10 minutes in daemon mode
			}
		}
	}
	//	 If no flags are set, print a message
	log.Println("No flags set. Use --all, --compile, --weather, or --daemon.")

	// Clean up: close the log file on program exit
	if currentLogFile != nil {
		currentLogFile.Close()
	}
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
