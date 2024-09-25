package main


import (
	"fmt"
	"path/filepath"
)


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
