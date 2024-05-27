package main

import (
	"database/sql"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// FetchAllSkyScannerIDs fetches the skyscannerid for all unique IATA codes.
func FetchAllSkyScannerIDs(iataCodes []string) (map[string]string, error) {
	dbPath := "../../../data/longterm_db/flights.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Prepare the query
	query := "SELECT iata, skyscannerid FROM airport_info WHERE iata IN (?" + strings.Repeat(",?", len(iataCodes)-1) + ")"
	args := make([]interface{}, len(iataCodes))
	for i, iata := range iataCodes {
		args[i] = iata
	}

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Process the results
	skyscannerIDs := make(map[string]string)
	for rows.Next() {
		var iata string
		var skyscannerID sql.NullString
		if err := rows.Scan(&iata, &skyscannerID); err != nil {
			return nil, err
		}
		if skyscannerID.Valid {
			skyscannerIDs[iata] = skyscannerID.String
		} else {
			skyscannerIDs[iata] = ""
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return skyscannerIDs, nil
}

