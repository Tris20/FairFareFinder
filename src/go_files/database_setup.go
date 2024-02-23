package go_files

import "fmt"

// SetupDatabase is now exported by starting with an uppercase letter
func Setup_database() {
	// fmt.Println("Setting up the database...")
	hello_myself()
	fmt.Println("Setting up the database...")
	// Add your database setup logic here
}

func hello_myself() {
	fmt.Println("Hello, myself")
}
