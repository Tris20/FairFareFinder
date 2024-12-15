package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/test_utils"
)

func main() {
	// Get the path of the executable
	execDir, err := getExecutablePath()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	fmt.Println(execDir)

	// Setup paths
	inputDataDir := filepath.Join(execDir, "test_utils/input-data")
	outputDir := filepath.Join(execDir, "../testdata")

	fmt.Println(inputDataDir)
	fmt.Println(outputDir)

	test_utils.SetMutePrints(true)
	test_utils.SetupMockDatabase(execDir, inputDataDir, outputDir, false)
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
