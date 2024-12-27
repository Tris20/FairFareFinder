package data_management

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tris20/FairFareFinder/learning_utils_playground/config_handlers"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/schollz/progressbar/v3"

	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
)

// pixabay

// Structs to parse JSON response from Pixabay API
type PixabayResponse struct {
	Hits []struct {
		LargeImageURL string `json:"largeImageURL"`
	} `json:"hits"`
}

func FetchLocationsCities_1() {
	// Load the Pixabay API key from secrets.yaml
	apiKey, err := config_handlers.LoadApiKey("../../../ignore/secrets.yaml", "pixabay")
	if err != nil {
		log.Fatalf("Error loading API key: %v", err)
	}

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
		err := downloadCityImagesFromPixabay(city, imageDir, apiKey)
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

// Function to download images for a given city using Pixabay API
func downloadCityImagesFromPixabay(city, imageDir, apiKey string) error {
	apiURL := fmt.Sprintf("https://pixabay.com/api/?key=%s&q=%s&image_type=photo&per_page=50", apiKey, strings.ReplaceAll(city, " ", "+"))

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result PixabayResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	imageNumber := 1
	for _, hit := range result.Hits {
		fileName := fmt.Sprintf("%s_%d.jpg", strings.ReplaceAll(city, " ", "_"), imageNumber)
		fullPath := filepath.Join(imageDir, fileName)

		err := downloadImage(hit.LargeImageURL, fullPath)
		if err != nil {
			fmt.Printf("Failed to download %s: %v\n", fileName, err)
		} else {
			fmt.Printf("Downloaded %s for %s\n", fileName, city)
		}
		imageNumber++
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
			return CopyFile(path, dstPath)
		}

		return nil
	})
}

// wiki-media

func FetchLocationsCities() {
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

// some break goes here

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

func FetchLocationsCities_2() {
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

// zip-cities

func FetchLocationsCities_3() {
	rootDir := "../pixabay/images/" // Set the source directory
	outputDir := "output"           // Set the output directory for zip files

	// Ensure the output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Println("Error creating output directory:", err)
			return
		}
	}

	// Walk through the directory
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory and not the root itself
		if info.IsDir() && path != rootDir {
			// Construct the filename for the zip file
			relativePath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			zipFileName := filepath.Join(outputDir, relativePath+".zip")

			// Create zip file
			err = zipFolder(path, zipFileName)
			if err != nil {
				fmt.Println("Error zipping directory:", err)
			} else {
				fmt.Println("Zipped directory:", zipFileName)
			}
			return filepath.SkipDir // Skip subdirectories
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking the directory:", err)
	}
}

func zipFolder(folder, zipFileName string) error {
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filepath.Base(folder), strings.TrimPrefix(path, folder))

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}

		return err
	})

	return nil
}
