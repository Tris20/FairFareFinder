package fffwebserver 


import (
	"fmt"
	"os"
	"os/exec"
)

// updateDatabaseWithIncoming checks for an incoming flights.db at a specified location,
// and if found, moves it to replace the existing flights.db. If not found, it quietly does nothing.
func UpdateDatabaseWithIncoming() {
	srcPath := "data/incoming_db/flights.db" 
	targetPath := "data/flights.db"          

	// Check if the source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// The file does not exist, do nothing
		fmt.Println("No incoming database found. No update performed.")
		return // Exit the function quietly
	} else if err != nil {
		// Some other error occurred when trying to access the file
		fmt.Printf("Error checking for source file: %v\n", err)
		return // Exit the function, considering handling this as per your need
	}

	// Perform an atomic move using the 'mv' command
	cmd := exec.Command("mv", "-T", srcPath, targetPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to move file: %v\n", err)
		return // Exit the function, considering handling this as per your need
	}

	fmt.Println("Successfully updated database with incoming file.")
}
