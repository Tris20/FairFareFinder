
package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"io/ioutil"
	_ "github.com/mattn/go-sqlite3"
)

type CityImages struct {
	CityName string
	Country  string
	Images   [5]string
}

var cities []CityImages

func main() {
	log.Println("Starting the process...")

	// Step 1: Open the database and load cities from the 'location' table
	log.Println("Opening the database...")
	db, err := sql.Open("sqlite3", "../../../../../../data/compiled/new_main.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()
	log.Println("Database opened successfully.")

	// Step 2: Load cities from the 'location' table
	err = loadCitiesFromDatabase(db)
	if err != nil {
		log.Fatalf("Failed to load cities from database: %v", err)
	}

	// Step 3: Populate the city struct with image paths
	for i := range cities {
		cityFolder := fmt.Sprintf("../../../../../../ignore/location-images/%s", cities[i].CityName)
		log.Printf("Looking for images in folder: %s", cityFolder)

		images, err := getCityImages(cityFolder)
		if err != nil {
			log.Printf("Error getting images for %s: %v", cities[i].CityName, err)
			continue
		}

		// If images were found, log them
		log.Printf("Found images for %s: %v", cities[i].CityName, images)

		// Assign the selected images directly to the struct with the corrected path
		for j := 0; j < 5 && j < len(images); j++ {
			// Replace the unnecessary prefix
			correctedPath := strings.Replace(images[j], "../../../../../../ignore", "", 1)
			cities[i].Images[j] = correctedPath
		}
	}

	// Step 4: Update each city in the database
	for _, city := range cities {
		// Log the current city and its images
		log.Printf("Updating city: %s, Country: %s, Images: %v", city.CityName, city.Country, city.Images)

		if city.Images[0] == "" {
			log.Printf("No images found for %s. Skipping update.", city.CityName)
			continue // Skip cities with no images
		}

		updateCityImages(db, city)
	}
	log.Println("Process complete.")
}

// loadCitiesFromDatabase fetches the city names and country codes from the 'location' table
func loadCitiesFromDatabase(db *sql.DB) error {
	query := "SELECT city, country FROM location"
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cityName string
		var countryCode string
		err := rows.Scan(&cityName, &countryCode)
		if err != nil {
			return err
		}
		cities = append(cities, CityImages{
			CityName: cityName,
			Country:  countryCode,
		})
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d cities from the database", len(cities))
	return nil
}

// getCityImages fetches the first 5 images from the city's folder
func getCityImages(cityFolder string) ([5]string, error) {
	var images [5]string
	files, err := ioutil.ReadDir(cityFolder)
	if err != nil {
		return images, err
	}

	log.Printf("Found %d files in folder: %s", len(files), cityFolder)

	var imageFiles []string
	for _, file := range files {
		if !file.IsDir() && isImage(file.Name()) {
			imageFiles = append(imageFiles, file.Name())
		}
	}

	// Sort files alphanumerically
	sort.Strings(imageFiles)

	log.Printf("Sorted image files: %v", imageFiles)

	// Select the first 5 images
	for i := 0; i < 5 && i < len(imageFiles); i++ {
		images[i] = filepath.Join(cityFolder, imageFiles[i])
	}

	log.Printf("Selected images: %v", images)

	return images, nil
}

// isImage checks if a file is an image based on its extension
func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// updateCityImages updates the image columns in the location table for a city
func updateCityImages(db *sql.DB, city CityImages) {
	log.Printf("Updating images in the database for city: %s", city.CityName)

	query := `UPDATE location SET image_1 = ?, image_2 = ?, image_3 = ?, image_4 = ?, image_5 = ?
			  WHERE city = ? AND country = ?`
	result, err := db.Exec(query, city.Images[0], city.Images[1], city.Images[2], city.Images[3], city.Images[4], city.CityName, city.Country)
	if err != nil {
		log.Printf("Failed to update city %s: %v", city.CityName, err)
		return
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected for %s: %v", city.CityName, err)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No rows were updated for city: %s. Check city and country values.", city.CityName)
	} else {
		log.Printf("Successfully updated %d row(s) for city: %s", rowsAffected, city.CityName)
	}
}

