
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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
	// Paths for database and directories
	dbPath := "../../../data/raw/locations/locations.db"
	imageDir := "images"
	landscapeDir := "images/highres-landscapes"

	// Create necessary directories if they don't exist
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating image directory: %v", err)
	}
	if err := os.MkdirAll(landscapeDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating landscape directory: %v", err)
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Query for cities where include_tf == 1
	cities, err := getCitiesToInclude(db)
	if err != nil {
		log.Fatalf("Failed to get cities: %v", err)
	}

	// Initialize the progress bar for downloading images for all cities
	bar := progressbar.NewOptions(len(cities),
		progressbar.OptionSetDescription("Downloading images"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "#", SaucerPadding: "-", BarStart: "[", BarEnd: "]"}),
	)

	// Download images for all cities
	for _, city := range cities {
		err := downloadCityImages(city, imageDir)
		if err != nil {
			fmt.Printf("Error downloading images for %s: %v\n", city, err)
		}
		_ = bar.Add(1) // Update progress bar after each city is processed
	}

	// Cleanup small files and then filter high-res landscapes
	err = cleanupSmallFiles(imageDir, 92160) // 90 KB in bytes
	if err != nil {
		log.Fatalf("Error during cleanup: %v", err)
	}

	err = filterHighResLandscapes(imageDir, landscapeDir, 1920, 1080, 1.78)
	if err != nil {
		log.Fatalf("Error filtering high-res landscapes: %v", err)
	}
}

// Function to retrieve cities where include_tf = 1
func getCitiesToInclude(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT city_ascii FROM city WHERE include_tf = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}
	return cities, nil
}

// Function to download images for a given city
func downloadCityImages(city, imageDir string) error {
	apiURL := fmt.Sprintf(
		"https://commons.wikimedia.org/w/api.php?action=query&generator=images&prop=imageinfo&gimlimit=5&redirects=1&titles=%s&iiprop=url&format=json",
		strings.ReplaceAll(city, " ", "_"),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	imageNumber := 1
	for _, page := range result.Query.Pages {
		for _, image := range page.ImageInfo {
			fileName := fmt.Sprintf("%s_%d.jpg", strings.ReplaceAll(city, " ", "_"), imageNumber)
			fullPath := filepath.Join(imageDir, fileName)

			err := downloadImage(image.URL, fullPath)
			if err != nil {
				fmt.Printf("Failed to download %s: %v\n", fileName, err)
			} else {
				fmt.Printf("Downloaded %s for %s\n", fileName, city)
			}
			imageNumber++
		}
	}
	return nil
}

// Helper function to download an image from a URL
func downloadImage(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Cleanup function to delete files smaller than a specified size (in bytes)
func cleanupSmallFiles(dir string, minSize int64) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Size() < minSize {
			fmt.Printf("Deleting small file: %s (size: %d bytes)\n", path, info.Size())
			return os.Remove(path)
		}
		return nil
	})
}

// Filter high-resolution landscape images
func filterHighResLandscapes(srcDir, dstDir string, minWidth, minHeight int, minAspectRatio float64) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Open the image file to read EXIF data
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		x, err := exif.Decode(file)
		if err != nil {
			// Skip files without EXIF data
			return nil
		}

		// Get dimensions from EXIF metadata
		width, errW := x.Get(exif.PixelXDimension)
		height, errH := x.Get(exif.PixelYDimension)
		if errW != nil || errH != nil {
			return nil
		}

		w, _ := width.Int(0)
		h, _ := height.Int(0)
		aspectRatio := float64(w) / float64(h)

		// Check if the image meets high-resolution and wide aspect ratio criteria
		if w >= minWidth && h >= minHeight && aspectRatio >= minAspectRatio {
			dstPath := filepath.Join(dstDir, info.Name())
			fmt.Printf("Copying high-res landscape: %s\n", dstPath)
			return copyFile(path, dstPath)
		}

		return nil
	})
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}

