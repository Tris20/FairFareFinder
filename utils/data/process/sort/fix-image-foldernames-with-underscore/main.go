package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	rootDir := "../../../../../ignore/location-images/" // Update this to your images root directory

	// Read all folders in the root directory
	folders, err := ioutil.ReadDir(rootDir)
	if err != nil {
		fmt.Printf("Error reading root directory: %v\n", err)
		return
	}

	for _, folder := range folders {
		if !folder.IsDir() {
			continue
		}

		folderName := folder.Name()
		folderPath := filepath.Join(rootDir, folderName)

		// Read all files in the current folder
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			fmt.Printf("Error reading folder %s: %v\n", folderName, err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fileName := file.Name()
			locationName := extractLocationName(fileName)
			if locationName == "" {
				continue
			}

			correctFolderName := strings.ReplaceAll(locationName, " ", "_")
			if correctFolderName != folderName {
				newFolderPath := filepath.Join(rootDir, correctFolderName)

				// Create the new folder if it doesn't exist
				if _, err := os.Stat(newFolderPath); os.IsNotExist(err) {
					err = os.Mkdir(newFolderPath, 0755)
					if err != nil {
						fmt.Printf("Error creating folder %s: %v\n", newFolderPath, err)
						continue
					}
				}

				// Move the file to the correct folder
				oldFilePath := filepath.Join(folderPath, fileName)
				newFilePath := filepath.Join(newFolderPath, fileName)
				err = os.Rename(oldFilePath, newFilePath)
				if err != nil {
					fmt.Printf("Error moving file %s to %s: %v\n", oldFilePath, newFilePath, err)
					continue
				}

				fmt.Printf("Moved %s to %s\n", oldFilePath, newFilePath)
			}
		}
	}
}

// extractLocationName extracts the location name from the filename using a regex.
func extractLocationName(fileName string) string {
	// Regex to match the location part of the filename (e.g., Las_Vegas from Las_Vegas_1.jpg)
	re := regexp.MustCompile(`^([a-zA-Z_]+)_\d+\.(jpg|jpeg|png)$`)
	matches := re.FindStringSubmatch(fileName)
	if len(matches) > 1 {
		return strings.ReplaceAll(matches[1], "_", " ")
	}
	return ""
}
