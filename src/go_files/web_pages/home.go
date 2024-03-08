package fffwebpages

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// HomeHandler handles requests to the home page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		pageContent, err := ioutil.ReadFile("src/html/landingPage.html")
		if err != nil {
			log.Printf("Error reading landing page file: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(pageContent)
	} else if strings.HasSuffix(r.URL.Path, ".css") {
		cssPath := "src/css" + r.URL.Path
		cssContent, err := ioutil.ReadFile(cssPath)
		if err != nil {
			log.Printf("Error reading CSS file: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/css")
		w.Write(cssContent)
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
}
