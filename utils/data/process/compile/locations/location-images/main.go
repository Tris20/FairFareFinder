
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// CityImages stores information about a city, country, and its images
type CityImages struct {
	CityName string
	Country  string
	Images   [5]string
}


// cities input list with ISO country codes (placeholders, adjust as needed)
var cities = []CityImages{
	{"Alicante", "ES", [5]string{}},
	{"Amsterdam", "NL", [5]string{}},
	{"Ankara", "TR", [5]string{}},
	{"Antalya", "TR", [5]string{}},
	{"Arbil", "IQ", [5]string{}},
	{"Arrecife", "ES", [5]string{}},
	{"Athens", "GR", [5]string{}},
	{"Baku", "AZ", [5]string{}},
	{"Barcelona", "ES", [5]string{}},
	{"Basel", "CH", [5]string{}},
	{"Beijing", "CN", [5]string{}},
	{"Beirut", "LB", [5]string{}},
	{"Belgrad", "RS", [5]string{}},
	{"Bergamo", "IT", [5]string{}},
	{"Bergen", "NO", [5]string{}},
	{"Birmingham", "GB", [5]string{}},
	{"Bordeaux", "FR", [5]string{}},
	{"Bristol", "GB", [5]string{}},
	{"Brussels", "BE", [5]string{}},
	{"Bucharest", "RO", [5]string{}},
	{"Budapest", "HU", [5]string{}},
	{"Cairo", "EG", [5]string{}},
	{"Catania", "IT", [5]string{}},
	{"Cologne", "DE", [5]string{}},
	{"Copenhagen", "DK", [5]string{}},
	{"Corfu", "GR", [5]string{}},
	{"Doha", "QA", [5]string{}},
	{"Dubai", "AE", [5]string{}},
	{"Dublin", "IE", [5]string{}},
	{"Dubrovnik", "HR", [5]string{}},
	{"Dusseldorf", "DE", [5]string{}},
	{"Edinburgh", "GB", [5]string{}},
	{"Faro", "PT", [5]string{}},
	{"Frankfurt am Main", "DE", [5]string{}},
	{"Funchal", "PT", [5]string{}},
	{"Gaziantep", "TR", [5]string{}},
	{"Geneva", "CH", [5]string{}},
	{"Glasgow", "GB", [5]string{}},
	{"Gothenburg", "SE", [5]string{}},
	{"Graz", "AT", [5]string{}},
	{"Helsinki", "FI", [5]string{}},
	{"Heraklion", "GR", [5]string{}},
	{"Hurghada", "EG", [5]string{}},
	{"Iasi", "RO", [5]string{}},
	{"Ibiza Town", "ES", [5]string{}},
	{"Istanbul", "TR", [5]string{}},
	{"Izmir", "TR", [5]string{}},
	{"Kaunas", "LT", [5]string{}},
	{"Kos Island", "GR", [5]string{}},
	{"Kraków", "PL", [5]string{}},
	{"Kutaisi", "GE", [5]string{}},
	{"Larnaca", "CY", [5]string{}},
	{"Las Palmas", "ES", [5]string{}},
	{"Lisbon", "PT", [5]string{}},
	{"London", "GB", [5]string{}},
	{"Luqa", "MT", [5]string{}},
	{"Luxembourg", "LU", [5]string{}},
	{"Lyon", "FR", [5]string{}},
	{"Madrid", "ES", [5]string{}},
	{"Malaga", "ES", [5]string{}},
	{"Manchester", "GB", [5]string{}},
	{"Marrakech", "MA", [5]string{}},
	{"Marsa Alam", "EG", [5]string{}},
	{"Marseille", "FR", [5]string{}},
	{"Milan", "IT", [5]string{}},
	{"Monastir", "TN", [5]string{}},
	{"Munich", "DE", [5]string{}},
	{"Nantes", "FR", [5]string{}},
	{"Napoli", "IT", [5]string{}},
	{"New York", "US", [5]string{}},
	{"Newark", "US", [5]string{}},
	{"Nice", "FR", [5]string{}},
	{"Nottingham", "GB", [5]string{}},
	{"Olbia", "IT", [5]string{}},
	{"Ortaca", "TR", [5]string{}},
	{"Oslo", "NO", [5]string{}},
	{"Palermo", "IT", [5]string{}},
	{"Palma De Mallorca", "ES", [5]string{}},
	{"Paphos", "CY", [5]string{}},
	{"Paris", "FR", [5]string{}},
	{"Pisa", "IT", [5]string{}},
	{"Podgorica", "ME", [5]string{}},
	{"Porto", "PT", [5]string{}},
	{"Prishtina", "XK", [5]string{}},
	{"Puerto Del Rosario", "ES", [5]string{}},
	{"Reggio Calabria", "IT", [5]string{}},
	{"Reykjavik", "IS", [5]string{}},
	{"Rhodes", "GR", [5]string{}},
	{"Riga", "LV", [5]string{}},
	{"Rome", "IT", [5]string{}},
	{"Saarbrucken", "DE", [5]string{}},
	{"Salzburg", "AT", [5]string{}},
	{"Sharm el-Sheikh", "EG", [5]string{}},
	{"Skopje", "MK", [5]string{}},
	{"Sofia", "BG", [5]string{}},
	{"Souda", "GR", [5]string{}},
	{"Split", "HR", [5]string{}},
	{"Stockholm", "SE", [5]string{}},
	{"Strasbourg", "FR", [5]string{}},
	{"Stuttgart", "DE", [5]string{}},
	{"Tallinn", "EE", [5]string{}},
	{"Tbilisi", "GE", [5]string{}},
	{"Tel Aviv", "IL", [5]string{}},
	{"Thessaloniki", "GR", [5]string{}},
	{"Tirana", "AL", [5]string{}},
	{"Tivat", "ME", [5]string{}},
	{"Toulouse/Blagnac", "FR", [5]string{}},
	{"Treviso", "IT", [5]string{}},
	{"Trieste", "IT", [5]string{}},
	{"Tromsø", "NO", [5]string{}},
	{"Trondheim", "NO", [5]string{}},
	{"Tunis", "TN", [5]string{}},
	{"Valencia", "ES", [5]string{}},
	{"Varna", "BG", [5]string{}},
	{"Venice", "IT", [5]string{}},
	{"Verona", "IT", [5]string{}},
	{"Vienna", "AT", [5]string{}},
	{"Vilnius", "LT", [5]string{}},
	{"Warsaw", "PL", [5]string{}},
	{"Zadar", "HR", [5]string{}},
	{"Zagreb", "HR", [5]string{}},
	{"Zurich", "CH", [5]string{}},
}




func main() {
	log.Println("Starting the process...")

	// Step 1: Populate the city struct with image paths
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

	// Step 2: Open the database and update the location table
	log.Println("Opening the database...")
	db, err := sql.Open("sqlite3", "../../../../../../data/compiled/main.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()
	log.Println("Database opened successfully.")

	// Step 3: Update each city in the database
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
		log.Printf("No rows were updated for city: %s. Check city_ascii and country values.", city.CityName)
	} else {
		log.Printf("Successfully updated %d row(s) for city: %s", rowsAffected, city.CityName)
	}
}

