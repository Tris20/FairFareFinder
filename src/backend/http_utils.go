package backend

import (
	"log"
	"net/http"
)

func HandleHTTPError(w http.ResponseWriter, message string, code int) {
	log.Printf("Error: %s", message)
	http.Error(w, message, code)
}
