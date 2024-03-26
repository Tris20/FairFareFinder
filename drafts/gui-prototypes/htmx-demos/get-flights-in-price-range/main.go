
package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type SkyscannerPrice struct {
	Origin      string
	Destination string
	ThisWeekend sql.NullFloat64
	NextWeekend sql.NullFloat64
}

func main() {
	http.HandleFunc("/", servePage)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func servePage(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "../../../../data/flights.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// HTMX request to update table based on slider value
	if r.Method == "GET" && r.Header.Get("HX-Request") != "" {
		updateTable(db, w, r)
		return
	}

	// Initial page load
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	tmpl.Execute(w, nil)
}

func updateTable(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	minPriceStr, maxPriceStr := r.URL.Query().Get("minPrice"), r.URL.Query().Get("maxPrice")
	minPrice, _ := strconv.ParseFloat(minPriceStr, 64)
	maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)

	var prices []SkyscannerPrice
	query := `SELECT origin, destination, this_weekend, next_weekend FROM skyscannerprices WHERE (this_weekend BETWEEN ? AND ? OR next_weekend BETWEEN ? AND ?)`
	rows, err := db.Query(query, minPrice, maxPrice, minPrice, maxPrice)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var price SkyscannerPrice
		err := rows.Scan(&price.Origin, &price.Destination, &price.ThisWeekend, &price.NextWeekend)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}
	tmpl.Execute(w, prices)
}

