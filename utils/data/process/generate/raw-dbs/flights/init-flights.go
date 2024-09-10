
package main

import (
    "bufio"
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

func main() {
    // Open the SQLite database.
    db, err := sql.Open("sqlite3", "../../../../../../data/raw/flights/flights.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create the "schedule" table if it does not already exist.
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS "schedule" (
        "id"   INTEGER,
        "flightNumber" TEXT NOT NULL,
        "departureAirport" TEXT,
        "arrivalAirport" TEXT,
        "departureTime" TEXT,
        "arrivalTime" TEXT,
        "direction" TEXT NOT NULL,
        PRIMARY KEY("id" AUTOINCREMENT)
    );`)
    if err != nil {
        log.Fatal(err)
    } else {
        log.Println("Checked/created 'schedule' table successfully.")
    }

    // Prompt to continue
    fmt.Println("Press 'Enter' to continue with the next table...")
    bufio.NewReader(os.Stdin).ReadBytes('\n')

    // Create the "skyscannerprices" table if it does not already exist.
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS "skyscannerprices" (
        "origin_city" TEXT,
        "origin_country" TEXT,
        "origin_iata" TEXT,
        "origin_skyscanner_id" TEXT,
        "destination_city" TEXT,
        "destination_country" TEXT,
        "destination_iata" TEXT,
        "destination_skyscanner_id" TEXT,
        "this_weekend" REAL,    
        "next_weekend" REAL
    );
    `)
    if err != nil {
        log.Fatal(err)
    } else {
        log.Println("Checked/created 'skyscannerprices' table successfully.")
    }
}

