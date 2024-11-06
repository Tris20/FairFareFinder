
package main

import (
	"fmt"
	"io"
  "io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	// Get the directory where the script is located
	scriptDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Failed to get script directory: %v", err)
	}

	// Set the destination directory to a "location-images" folder in the same directory as the script
	destinationDir := filepath.Join(scriptDir, "location-images")
	sourceDir := "../../../../../../../ignore/location-images"

	// Create the destination directory if it doesn't exist
	if _, err := os.Stat(destinationDir); os.IsNotExist(err) {
		err := os.Mkdir(destinationDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create destination directory: %v", err)
		}
		fmt.Printf("Created destination directory: %s\n", destinationDir)
	} else {
		fmt.Printf("Destination directory already exists: %s\n", destinationDir)
	}

	// Step 1: Iterate through each city folder in the source directory
	fmt.Println("Starting to process city folders...")

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a directory (and not the source directory itself)
		if info.IsDir() && path != sourceDir {
			cityFolder := info.Name()
			fmt.Printf("\nProcessing city: %s\n", cityFolder)

			// Get the files inside this city folder
			files, err := ioutil.ReadDir(filepath.Join(sourceDir, cityFolder))
			if err != nil {
				log.Printf("Error reading directory for %s: %v", cityFolder, err)
				return nil
			}

			fmt.Printf("Found %d files in %s folder.\n", len(files), cityFolder)

			// Filter only image files and sort them alphabetically
			var imageFiles []string
			for _, file := range files {
				if !file.IsDir() && isImage(file.Name()) {
					imageFiles = append(imageFiles, file.Name())
				}
			}

			// Sort the image files alphabetically
			sort.Strings(imageFiles)

			// If there is at least one image file, copy the first one
			if len(imageFiles) > 0 {
				firstImage := imageFiles[0]
				fmt.Printf("First image (alphabetically) for %s: %s\n", cityFolder, firstImage)

				// Create the destination city folder if it doesn't exist
				destCityFolder := filepath.Join(destinationDir, cityFolder)
				if _, err := os.Stat(destCityFolder); os.IsNotExist(err) {
					err := os.Mkdir(destCityFolder, 0755)
					if err != nil {
						log.Printf("Failed to create directory for %s: %v", cityFolder, err)
						return nil
					}
					fmt.Printf("Created directory for city: %s\n", destCityFolder)
				}

				// Copy the first image to the new folder
				sourcePath := filepath.Join(sourceDir, cityFolder, firstImage)
				destPath := filepath.Join(destCityFolder, firstImage)
				err := copyFile(sourcePath, destPath)
				if err != nil {
					log.Printf("Failed to copy image %s: %v", firstImage, err)
				} else {
					fmt.Printf("Successfully copied %s to %s\n", firstImage, destPath)
				}
			} else {
				fmt.Printf("No image files found for %s.\n", cityFolder)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error processing folders: %v", err)
	}

	fmt.Println("Process complete.")
}

// isImage checks if a file is an image based on its extension
func isImage(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// copyFile copies a file from source to destination
func copyFile(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err == nil {
		fmt.Printf("Copied file from %s to %s\n", source, dest)
	}

	return err
}

