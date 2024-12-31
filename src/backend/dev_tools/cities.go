package dev_tools

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

type City struct {
	Name     string
	ImageURL string
}

var (
	mu           sync.Mutex
	cityChannels = make(map[string]chan []City)
	doneChannels = make(map[string]chan bool)
)

func LoadCitiesInBatches(db *sql.DB, clientID string) {
	offset := 0
	for {
		rows, err := db.Query("SELECT city, image_1 FROM location ORDER BY city LIMIT 10 OFFSET $1", offset)
		if err != nil {
			fmt.Println("error querying cities")
			close(cityChannels[clientID])
			return
		}
		defer rows.Close()

		var batch []City
		for rows.Next() {
			var city City
			var imageURL sql.NullString
			if err := rows.Scan(&city.Name, &imageURL); err != nil {
				fmt.Println(err)
				fmt.Println("error scanning city")
				close(cityChannels[clientID])
				return
			}
			if imageURL.Valid {
				city.ImageURL = imageURL.String
			} else {
				city.ImageURL = "" // or any default value you prefer
			}
			batch = append(batch, city)
		}

		if len(batch) == 0 {
			close(cityChannels[clientID])
			return
		}

		cityChannels[clientID] <- batch
		offset += 10
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

func removeDuplicateCities(db *sql.DB) error {
	query := `
		DELETE FROM location
		WHERE rowid NOT IN (
			SELECT MAX(rowid)
			FROM location
			GROUP BY city
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error removing duplicate cities: %v", err)
	}
	return nil
}

func RemoveDuplicateCitiesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := removeDuplicateCities(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Duplicate cities removed"))
	}
}
