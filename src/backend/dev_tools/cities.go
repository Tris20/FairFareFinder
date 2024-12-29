package dev_tools

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
)

type City struct {
	Name     string
	ImageURL string
}

var (
	cityChannel = make(chan []City)
	doneChannel = make(chan bool)
)

func LoadCitiesInBatches(db *sql.DB) {
	// fmt.Println("go routine LoadCitiesInBatches")
	offset := 0
	for {
		rows, err := db.Query("SELECT city, image_1 FROM location ORDER BY city LIMIT 10 OFFSET $1", offset)
		if err != nil {
			fmt.Println("error querying cities")
			close(cityChannel)
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
				close(cityChannel)
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
			close(cityChannel)
			return
		}

		cityChannel <- batch
		offset += 10
		// fmt.Println("added batch to channel")
	}
}

func GetNextCitiesBatch() []City {
	// fmt.Println("GetNextCitiesBatch")
	select {
	case batch := <-cityChannel:
		return batch
	case <-doneChannel:
		return nil
	}
}

func LoadMoreCities(tmpl *template.Template) http.HandlerFunc {
	// fmt.Println("LoadMoreCities")
	return func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println("LoadMoreCities serving request")
		cities := GetNextCitiesBatch()
		if cities == nil {
			http.Error(w, "No more cities to load", http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err := tmpl.ExecuteTemplate(w, "cities.html", cities)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// fmt.Println("LoadMoreCities served request")
	}
}

func AllCitiesHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		go LoadCitiesInBatches(db)                             // Start the Go routine here
		err := tmpl.ExecuteTemplate(w, "all-cities.html", nil) // Serve the all-cities.html template
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
