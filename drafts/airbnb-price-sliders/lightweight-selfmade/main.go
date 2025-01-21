// main.go
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Serve static files from current dir
func serveStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "."+r.URL.Path)
}

// Example filter handler
func filterHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	r.ParseForm()
	currentValueStr := r.Form.Get("currentValue") // from slider
	currentValue, err := strconv.ParseFloat(currentValueStr, 64)
	if err != nil {
		http.Error(w, "Invalid slider value", http.StatusBadRequest)
		return
	}

	// Read the CSV file (dummy.csv) from disk
	f, err := os.Open("dummy.csv")
	if err != nil {
		http.Error(w, "Cannot open CSV", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	rows, err := csvReader.ReadAll()
	if err != nil {
		http.Error(w, "Error reading CSV", http.StatusInternalServerError)
		return
	}

	// rows[0] is the header row ["value"]
	var filteredValues []float64
	for i := 1; i < len(rows); i++ {
		v, _ := strconv.ParseFloat(rows[i][0], 64)
		// Filter condition: value <= currentValue
		if v <= currentValue {
			filteredValues = append(filteredValues, v)
		}
	}

	// For simplicity, let’s just return the count of each bin as a new HTML snippet
	// In reality, you’d want to replicate the bin calculation logic on the server side
	// and produce a set of <div class="bar"> elements. We'll show a simplified version:

	// Hard-code 10 bins for demonstration
	binCount := 10
	minVal, maxVal := 0.0, 100.0 // or compute from data
	rangeVal := maxVal - minVal
	binSize := rangeVal / float64(binCount)
	bins := make([]int, binCount)

	for _, v := range filteredValues {
		idx := int((v - minVal) / binSize)
		if idx >= binCount {
			idx = binCount - 1
		} else if idx < 0 {
			idx = 0
		}
		bins[idx]++
	}

	maxBin := 0
	for _, count := range bins {
		if count > maxBin {
			maxBin = count
		}
	}

	// Build the HTML snippet for the bars
	// Return only the inner HTML to be swapped into #chart
	fmt.Fprintln(w, `<div style="display: flex; align-items: flex-end; height: 100%; width: 100%;">`)
	for _, count := range bins {
		heightPercent := 0
		if maxBin > 0 {
			heightPercent = int(float64(count) / float64(maxBin) * 100)
		}
		label := ""
		if count > 0 {
			label = fmt.Sprintf("%d", count)
		}
		// Each bin becomes a bar
		fmt.Fprintf(w,
			`<div class="bar" style="flex:1; margin:0 1px; background-color:steelblue; height:%d%%; display:flex; align-items:flex-end; justify-content:center;">
                <div class="bar-label" style="writing-mode: vertical-rl; transform: rotate(180deg); font-size:12px; color:#fff; margin-bottom: 2px;">%s</div>
             </div>`,
			heightPercent, label)
	}
	fmt.Fprintln(w, `</div>`)
}

func main() {
	http.HandleFunc("/", serveStatic) // to serve index.html, dummy.csv
	http.HandleFunc("/filter", filterHandler)

	log.Println("Serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
