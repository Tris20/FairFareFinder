package dev_tools

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"sync"
)

type City struct {
	Name        string // Format: "City, Country"
	ImageURL    string
	FlightCount int
	GoogleLink  string
}

var (
	mu           sync.Mutex
	cityChannels = make(map[string]chan []City)
	doneChannels = make(map[string]chan bool)
)

func LoadCitiesInBatches(db *sql.DB, clientID string) {
	offset := 0
	for {
		// New query: selects city, country, image_1 and counts flights for each city-country combination,
		// filtering on non-NULL image_1 and ordering by flight count descending.
		rows, err := db.Query(`
			SELECT 
				l.city, 
				l.country,
				l.image_1,
				COUNT(f.id) AS flight_count
			FROM location l
			LEFT JOIN flight f
				ON l.city = f.destination_city_name 
			   AND l.country = f.destination_country
			WHERE l.image_1 IS NOT NULL
			GROUP BY l.city, l.country
			ORDER BY flight_count DESC
			LIMIT 50 OFFSET ?`, offset)
		if err != nil {
			fmt.Println("error querying cities:", err)
			close(cityChannels[clientID])
			return
		}

		var batch []City
		for rows.Next() {
			var cityName, country string
			var imageURL sql.NullString
			var flightCount int
			if err := rows.Scan(&cityName, &country, &imageURL, &flightCount); err != nil {
				fmt.Println("error scanning city:", err)
				close(cityChannels[clientID])
				return
			}

			city := City{
				Name:        fmt.Sprintf("%s, %s", cityName, country),
				FlightCount: flightCount,
				GoogleLink:  "https://www.google.com/search?&q=" + url.QueryEscape(cityName),
			}
			if imageURL.Valid {
				city.ImageURL = imageURL.String
			} else {
				city.ImageURL = "" // or set a default image URL if preferred
			}

			batch = append(batch, city)
		}
		rows.Close()

		if len(batch) == 0 {
			close(cityChannels[clientID])
			return
		}

		cityChannels[clientID] <- batch
		offset += 50
	}
}

func GetNextCitiesBatch(clientID string) []City {
	select {
	case batch := <-cityChannels[clientID]:
		return batch
	case <-doneChannels[clientID]:
		return nil
	}
}

func LoadMoreCities(tmpl *template.Template, clientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cities := GetNextCitiesBatch(clientID)
		if cities == nil {
			http.Error(w, "No more cities to load", http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err := tmpl.ExecuteTemplate(w, "cities.html", cities)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func AllCitiesHandler(db *sql.DB, tmpl *template.Template, clientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		cityChannels[clientID] = make(chan []City)
		doneChannels[clientID] = make(chan bool)
		mu.Unlock()

		go LoadCitiesInBatches(db, clientID)
		err := tmpl.ExecuteTemplate(w, "all-cities.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// OpenImageFolderHandler handles the request to open an image folder.
func OpenImageFolderHandler(w http.ResponseWriter, r *http.Request) {
	image := r.URL.Query().Get("image")
	if image == "" {
		http.Error(w, "missing image parameter", http.StatusBadRequest)
		return
	}

	// Base folder for images
	basePath := "/home/tristan/Documents/Workspace/SVN/SVN_BASE/Software/Shared_Projects/FairFareFinder/ignore/"

	// Construct the full image path
	fullImagePath := basePath + image

	// Compute the folder path by extracting the directory from the image path
	folder := filepath.Dir(image)
	fullFolderPath := basePath + folder

	// Open the image file
	cmdImage := exec.Command("xdg-open", fullImagePath)
	if err := cmdImage.Start(); err != nil {
		http.Error(w, fmt.Sprintf("failed to open image: %v", err), http.StatusInternalServerError)
		return
	}

	// Open the folder containing the image
	cmdFolder := exec.Command("xdg-open", fullFolderPath)
	if err := cmdFolder.Start(); err != nil {
		http.Error(w, fmt.Sprintf("failed to open image folder: %v", err), http.StatusInternalServerError)
		return
	}

	// Optionally, send a success message
	fmt.Fprintf(w, "Opened image: %s and folder: %s", fullImagePath, fullFolderPath)
}
