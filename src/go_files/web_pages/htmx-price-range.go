
package fffwebpages

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
"fmt"
	_ "github.com/mattn/go-sqlite3"
    "os"
"io/ioutil"
)

type SkyscannerPrice struct {
	Origin      string
	Destination string
	ThisWeekend sql.NullFloat64
	NextWeekend sql.NullFloat64
}

func HtmxPriceRange(w http.ResponseWriter, r *http.Request) {
	// serving a static file
	pageContent, err := ioutil.ReadFile("src/html/htmx.html")
	if err != nil {
		log.Printf("Error reading forecast page file: %v", err)
		http.Error(w, "Internal server error", 500)
		return
	}
	w.Write(pageContent)
}

func HtmxPriceRange2(w http.ResponseWriter, r *http.Request) {
  fmt.Printf("\nHTMX\n\n")
    wd, err := os.Getwd()
log.Println("Current working directory:", wd)
	db, err := sql.Open("sqlite3", "data/flights.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		log.Printf("Failed to open database connection: %v", err)
		return
	}
	defer db.Close()

	// HTMX request to update table based on slider value
	if r.Method == "GET" && r.Header.Get("HX-Request") != "" {
		updateTable(db, w, r)
		return
	}

	// Initial page load
	tmpl, err := template.ParseFiles("src/html/htmx.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Failed to parse template: %v", err)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		log.Printf("Failed to execute template: %v", err)
	}
}

func updateTable(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	minPriceStr, maxPriceStr := r.URL.Query().Get("minPrice"), r.URL.Query().Get("maxPrice")
	minPrice, err := strconv.ParseFloat(minPriceStr, 64)
	if err != nil {
		http.Error(w, "Invalid minimum price format", http.StatusBadRequest)
		return
	}
	maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
	if err != nil {
		http.Error(w, "Invalid maximum price format", http.StatusBadRequest)
		return
	}

	var prices []SkyscannerPrice
//	query := `SELECT origin, destination, this_weekend, next_weekend FROM skyscannerprices WHERE (this_weekend BETWEEN ? AND ? OR next_weekend BETWEEN ? AND ?)`
  query:=  	
`SELECT DISTINCT
    ai1.city AS origin_city, 
    ai2.city AS destination_city, 
    sp.this_weekend, 
    sp.next_weekend 
FROM 
    skyscannerprices sp
JOIN 
    airport_info ai1 ON sp.origin = ai1.skyscannerid
JOIN 
    airport_info ai2 ON sp.destination = ai2.skyscannerid
WHERE 
    (sp.this_weekend BETWEEN ? AND ? OR sp.next_weekend BETWEEN ? AND ?)
ORDER BY
    ai1.city, ai2.city, sp.this_weekend, sp.next_weekend;`
  rows, err := db.Query(query, minPrice, maxPrice, minPrice, maxPrice)
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var price SkyscannerPrice
		if err := rows.Scan(&price.Origin, &price.Destination, &price.ThisWeekend, &price.NextWeekend); err != nil {
			http.Error(w, "Failed to read query results", http.StatusInternalServerError)
			log.Printf("Failed to scan row: %v", err)
			return
		}
		prices = append(prices, price)
	}

	tmpl, err := template.New("table").Parse(`
<table>
	<thead>
		<tr>
			<th>Origin</th>
			<th>Destination</th>
			<th>This Weekend</th>
			<th>Next Weekend</th>
		</tr>
	</thead>
	<tbody>
		{{range .}}
		<tr>
			<td>{{.Origin}}</td>
			<td>{{.Destination}}</td>
			<td>{{if .ThisWeekend.Valid}}{{printf "%.2f" .ThisWeekend.Float64}}{{else}}N/A{{end}}</td>
			<td>{{if .NextWeekend.Valid}}{{printf "%.2f" .NextWeekend.Float64}}{{else}}N/A{{end}}</td>
		</tr>
		{{end}}
	</tbody>
</table>
`)
	if err != nil {
		http.Error(w, "Failed to parse inline template", http.StatusInternalServerError)
		log.Printf("Failed to parse inline template: %v", err)
		return
	}
	if err := tmpl.Execute(w, prices); err != nil {
		http.Error(w, "Failed to execute inline template", http.StatusInternalServerError)
		log.Printf("Failed to execute inline template: %v", err)
	}
}

