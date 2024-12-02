package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Base directory that contains alphabetically sorted folders
	baseDirectory := "../pixabay/images/"

	// Get all directories in the base directory
	alphabetFolders, err := ioutil.ReadDir(baseDirectory)
	if err != nil {
		fmt.Println("Error reading base directory:", err)
		return
	}

	totalFiles := 0

	// First pass to count total files for the progress bar
	for _, folder := range alphabetFolders {
		if folder.IsDir() {
			files, err := ioutil.ReadDir(filepath.Join(baseDirectory, folder.Name()))
			if err != nil {
				fmt.Println("Error reading directory:", folder.Name(), err)
				continue
			}
			for _, file := range files {
				if !file.IsDir() {
					totalFiles++
				}
			}
		}
	}

	// Initialize the progress bar
	bar := progressbar.Default(int64(totalFiles))

	// Process each directory
	for _, folder := range alphabetFolders {
		if folder.IsDir() {
			currentFolder := filepath.Join(baseDirectory, folder.Name())
			files, err := ioutil.ReadDir(currentFolder)
			if err != nil {
				fmt.Println("Error reading directory:", folder.Name(), err)
				continue
			}

			for _, file := range files {
				if !file.IsDir() {
					fileName := file.Name()
					underscoreIndex := strings.Index(fileName, "_")
					if underscoreIndex == -1 {
						fmt.Println("Skipping file with unexpected name format:", fileName)
						bar.Add(1)
						continue
					}
					cityName := fileName[:underscoreIndex]
					targetDirectory := fmt.Sprintf("%s/%s", baseDirectory, cityName)

					// Create the directory if it doesn't exist
					if _, err := os.Stat(targetDirectory); os.IsNotExist(err) {
						err := os.Mkdir(targetDirectory, 0755)
						if err != nil {
							fmt.Println("Error creating directory:", err)
							bar.Add(1)
							continue
						}
					}

					// Move the file to the appropriate directory
					oldPath := filepath.Join(currentFolder, fileName)
					newPath := filepath.Join(targetDirectory, fileName)
					err := os.Rename(oldPath, newPath)
					if err != nil {
						fmt.Println("Error moving file:", err)
					}

					// Update the progress bar
					bar.Add(1)
				}
			}
		}
	}
}
