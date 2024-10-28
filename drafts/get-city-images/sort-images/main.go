
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func main() {
	// Set the directory containing the images
	imageDirectory := "../pixabay/images/"
	files, err := ioutil.ReadDir(imageDirectory)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// Initialize the progress bar
	bar := progressbar.Default(int64(len(files)))

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			// Get the first letter of the city name
			firstLetter := strings.ToUpper(string(fileName[0]))
			targetDirectory := fmt.Sprintf("%s/%s", imageDirectory, firstLetter)

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
			oldPath := fmt.Sprintf("%s/%s", imageDirectory, fileName)
			newPath := fmt.Sprintf("%s/%s", targetDirectory, fileName)
			err := os.Rename(oldPath, newPath)
			if err != nil {
				fmt.Println("Error moving file:", err)
			}
		}
		// Update the progress bar after each file operation
		bar.Add(1)
	}
}

