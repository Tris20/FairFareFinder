
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

// Helper function to run a command in a specific directory
func runExecutableInDir(dir string, executable string) {
    // Change to the specified directory
    err := os.Chdir(dir)
    if err != nil {
        log.Fatalf("Failed to change directory to %s: %v", dir, err)
    }
    fmt.Printf("Changed to directory: %s\n", dir)

    // Execute the executable
    cmd := exec.Command("./" + executable)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    fmt.Printf("Running executable: %s\n", executable)
    err = cmd.Run()
    if err != nil {
        log.Fatalf("Failed to run executable %s in directory %s: %v", executable, dir, err)
    }

    fmt.Printf("Successfully executed: %s\n", executable)
}

// Helper function to get the current weekday and hour
func getCurrentTime() (time.Weekday, int) {
    now := time.Now()
    return now.Weekday(), now.Hour()
}

// Function to run all tasks in sequence
func runAllTasks(relativeBase string) {
    green := "\033[32m"
    reset := "\033[0m"
 

    // FETCH
    runExecutableInDir(filepath.Join(relativeBase, "fetch/flights/schedule"), "aerodatabox")
    fmt.Printf("%sCOMPLETED: aerodatabox (flight schedule)%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "fetch/flights/prices"), "prices")
    fmt.Printf("%sCOMPLETED: prices (flight prices)%s\n", green, reset)
    runExecutableInDir(filepath.Join(relativeBase, "fetch/weather"), "update-weather-db")
    fmt.Printf("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)
    runExecutableInDir(filepath.Join(relativeBase, "fetch/accommocation/booking-com/get-properties"), "get-properties")
    fmt.Printf("%sCOMPLETED: get-properties (properties update)%s\n", green, reset)



    //Calculate
    runExecutableInDir(filepath.Join(relativeBase, "process/calculate/weather"), "weather")
    fmt.Printf("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)
 
   
    //Compile
    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/flights"), "flights")
    fmt.Printf("%sCOMPLETED: flights (process compile)%s\n", green, reset)
    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/weather"), "weather")
    fmt.Printf("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/locations"), "locations")
    fmt.Printf("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)

   runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/accommodation/booking-com"), "booking-com")
   fmt.Printf("%sCOMPLETED:  process/compile/main/accommodation/booking-com%s\n", green, reset)


   // 5 nights and flights
   runExecutableInDir(filepath.Join(relativeBase, "process/calculate/main/five-nights-and-flights"), "five-nights-and-flights")
   fmt.Printf("%sCOMPLETED:  process/calculate/main/five-nights-and-flights%s\n", green, reset)



 }

// Function to run only compile tasks
func runCompileTasks(relativeBase string) {
    green := "\033[32m"
    reset := "\033[0m"

    runExecutableInDir(filepath.Join(relativeBase, "process/calculate/weather"), "weather")
    fmt.Printf("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/flights"), "flights")
    fmt.Printf("%sCOMPLETED: flights (process compile)%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/weather"), "weather")
    fmt.Printf("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/locations"), "locations")
    fmt.Printf("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)

   runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/accommodation/booking-com"), "booking-com")
   fmt.Printf("%sCOMPLETED:  process/compile/main/accommodation/booking-com%s\n", green, reset)

  // 5 nights and flights
   runExecutableInDir(filepath.Join(relativeBase, "process/calculate/main/five-nights-and-flights"), "five-nights-and-flights")
   fmt.Printf("%sCOMPLETED:  process/calculate/main/five-nights-and-flights%s\n", green, reset)



 }



// Function to run only weather-related tasks
func runWeatherTasks(relativeBase string) {
    green := "\033[32m"
    reset := "\033[0m"

    runExecutableInDir(filepath.Join(relativeBase, "fetch/weather"), "update-weather-db")
    fmt.Printf("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/calculate/weather"), "weather")
    fmt.Printf("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)
//included beacuse we always create a new completely new main.db, so need to rebuild the flights table
    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/flights"), "flights")
    fmt.Printf("%sCOMPLETED: flights (process compile)%s\n", green, reset)

    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/weather"), "weather")
    fmt.Printf("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)
//included beacuse we always create a new completely new main.db, so need to rebuild the locations table
    runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/locations"), "locations")
    fmt.Printf("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)
}


func main() {
    // Add flags for running all tasks or just compile tasks
    runAll := flag.Bool("all", false, "Run all tasks in sequence regardless of time")
    runCompile := flag.Bool("compile", false, "Run only compile tasks")
 runWeather := flag.Bool("weather", false, "Run only weather-related tasks")

    flag.Parse()

    // Create /out directory if it does not exist
    outputDir := "../../../../../data/compiled/"
    if _, err := os.Stat(outputDir); os.IsNotExist(err) {
        err := os.Mkdir(outputDir, 0755)
        if err != nil {
            log.Fatalf("Failed to create directory %s: %v", outputDir, err)
        }
    }

    // Database file paths
    dbPath := filepath.Join(outputDir, "new_main.db")

    // Backup existing database if it exists
    backupDatabase(dbPath, outputDir)
    // Initialize the new database and create tables
    initializeDatabase(dbPath)

    // Get the current directory of the script
    baseDir, err := os.Getwd()
    if err != nil {
        log.Fatalf("Failed to get current working directory: %v", err)
    }

    // Define the relative path to go up three directories
    relativeBase := filepath.Join(baseDir, "../../../")

    green := "\033[32m"
    reset := "\033[0m"

    // If the --all flag is set, run all tasks sequentially
    if *runAll {
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

    // Get current day and time
    currentDay, currentHour := getCurrentTime()

    // Task logic based on time and task completion
    if currentDay == time.Monday && currentHour == 9 {
        runExecutableInDir(filepath.Join(relativeBase, "fetch/flights/schedule"), "aerodatabox")
        fmt.Printf("%sCOMPLETED: aerodatabox (flight schedule)%s\n", green, reset)
    }

    if currentDay == time.Monday && currentHour == 10 {
        runExecutableInDir(filepath.Join(relativeBase, "fetch/flights/prices"), "prices")
        fmt.Printf("%sCOMPLETED: prices (flight prices)%s\n", green, reset)

        // Run next task after prices completes
        runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/flights"), "flights")
        fmt.Printf("%sCOMPLETED: flights (process compile)%s\n", green, reset)
    }

    // Update weather every 6 hours
    if currentHour%6 == 0 {
        runExecutableInDir(filepath.Join(relativeBase, "fetch/weather"), "update-weather-db")
        fmt.Printf("%sCOMPLETED: update-weather-db (weather update)%s\n", green, reset)

        // Run weather calculation after weather update completes
        runExecutableInDir(filepath.Join(relativeBase, "process/calculate/weather"), "weather")
        fmt.Printf("%sCOMPLETED: weather (weather calculation)%s\n", green, reset)

        // Run process/compile/main/weather after calculation
        runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/weather"), "weather")
        fmt.Printf("%sCOMPLETED: process/compile/main/weather%s\n", green, reset)

        // Run process/compile/main/locations after weather compile
        runExecutableInDir(filepath.Join(relativeBase, "process/compile/main/locations"), "locations")
        fmt.Printf("%sCOMPLETED: process/compile/main/locations%s\n", green, reset)
    }

    fmt.Println("All tasks executed successfully.")
}

