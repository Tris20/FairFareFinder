
package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type Flight struct {
	DestinationCityName string
	PriceCity1          sql.NullFloat64
	PriceCity2          sql.NullFloat64
	CombinedPrice       sql.NullFloat64
	AvgWpi              sql.NullFloat64
}

type FlightsData struct {
	SelectedCity1 string
	SelectedCity2 string
	Flights       []Flight
}

var (
	tmpl     *template.Template
	db       *sql.DB
	store    *sessions.CookieStore = sessions.NewCookieStore([]byte("your-secret-key"))
)

func main() {
	var err error

	db, err = sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tmpl = template.Must(template.ParseFiles("index.html", "table.html"))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/filter", filterHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT DISTINCT origin_city_name FROM flight")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cities = append(cities, city)
	}

	err = tmpl.ExecuteTemplate(w, "index.html", cities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	city1 := r.URL.Query().Get("city1")
	city2 := r.URL.Query().Get("city2")
	sortOption := r.URL.Query().Get("sort")

	session.Values["city1"] = city1
	session.Values["city2"] = city2
	session.Save(r, w)

	orderClause := "ORDER BY combined_price ASC"
	switch sortOption {
	case "low_price":
		orderClause = "ORDER BY combined_price ASC"
	case "high_price":
		orderClause = "ORDER BY combined_price DESC"
	case "best_weather":
		orderClause = "ORDER BY avg_wpi DESC"
	case "worst_weather":
		orderClause = "ORDER BY avg_wpi ASC"
	}

	query := `
	SELECT f1.destination_city_name, MIN(f1.price_this_week), MIN(f2.price_this_week),
	(MIN(f1.price_this_week) + MIN(f2.price_this_week)) AS combined_price, l.avg_wpi
	FROM flight f1
	INNER JOIN flight f2 ON f1.destination_city_name = f2.destination_city_name
	INNER JOIN location l ON f1.destination_city_name = l.city
	WHERE f1.origin_city_name = ? AND f2.origin_city_name = ?
	GROUP BY f1.destination_city_name
	` + orderClause
	rows, err := db.Query(query, city1, city2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var flights []Flight
	for rows.Next() {
		var flight Flight
		if err := rows.Scan(&flight.DestinationCityName, &flight.PriceCity1, &flight.PriceCity2, &flight.CombinedPrice, &flight.AvgWpi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flights = append(flights, flight)
	}

	data := FlightsData{
		SelectedCity1: city1,
		SelectedCity2: city2,
		Flights:       flights,
	}

	err = tmpl.ExecuteTemplate(w, "table.html", data)
	if err != nil {
		http.Error(w, "Error rendering results", http.StatusInternalServerError)
	}
}

