package backend

import (
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

func GetUserSession(store *sessions.CookieStore, r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Error retrieving user session: %v", err)
		return nil, err
	}
	return session, nil
}
