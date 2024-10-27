
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Structs to parse JSON response from Wikimedia Commons API
type ImageInfo struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type QueryResult struct {
	Query struct {
		Pages map[string]struct {
			ImageInfo []ImageInfo `json:"imageinfo"`
		} `json:"pages"`
	} `json:"query"`
}

func main() {
	// List of cities to search for
	cities := []string{"New York", "Tokyo", "Paris", "London", "Berlin"}

	// Create images directory if it doesn't exist
	imageDir := "images"
	err := os.MkdirAll(imageDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	for _, city := range cities {
		err := downloadCityImages(city, imageDir)
		if err != nil {
			fmt.Printf("Error downloading images for %s: %v\n", city, err)
		}
	}
}

func downloadCityImages(city, imageDir string) error {
	// Construct Wikimedia Commons API URL to get images related to the city
	apiURL := fmt.Sprintf(
		"https://commons.wikimedia.org/w/api.php?action=query&generator=images&prop=imageinfo&gimlimit=5&redirects=1&titles=%s&iiprop=url&format=json",
		strings.ReplaceAll(city, " ", "_"),
	)

	// Make the HTTP request
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Track the image number for naming
	imageNumber := 1

	// Iterate through image metadata and download each image
	for _, page := range result.Query.Pages {
		for _, image := range page.ImageInfo {
			// Set the filename format as cityname_number.jpg
			fileName := fmt.Sprintf("%s_%d.jpg", strings.ReplaceAll(city, " ", "_"), imageNumber)
			fullPath := filepath.Join(imageDir, fileName)

			err := downloadImage(image.URL, fullPath)
			if err != nil {
				fmt.Printf("Failed to download %s: %v\n", fileName, err)
			} else {
				fmt.Printf("Downloaded %s for %s\n", fileName, city)
			}
			imageNumber++ // Increment the image number for each new image
		}
	}
	return nil
}

func downloadImage(url, filePath string) error {
	// Make a request to download the image
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file to save the image
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Save the image content to the file
	_, err = io.Copy(file, resp.Body)
	return err
}

