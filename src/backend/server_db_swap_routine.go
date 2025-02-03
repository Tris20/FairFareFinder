package backend

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"os"
	"time"
)

// Function to check for new_main.db and swap it with main.db

func StartFileCheckRoutine(db **sql.DB, tmpl **template.Template) {
	fmt.Println("Entered the db monitoring loop")
	for {
		fmt.Printf("In the db loop\n")
		newDBPath := "./data/compiled/new_main.db"
		mainDBPath := "./data/compiled/main.db"

		if _, err := os.Stat(newDBPath); err == nil {
			fmt.Println("new_main.db exists")

			// Close the current database connection before swapping
			err := (*db).Close()
			if err != nil {
				log.Printf("Failed to close the database connection: %v", err)
				continue
			}

			// Perform atomic swap: rename new_main.db to main.db
			err = os.Rename(newDBPath, mainDBPath)
			if err != nil {
				log.Printf("Failed to swap new_main.db with main.db: %v", err)
			} else {
				log.Println("Successfully swapped new_main.db with main.db")

				// Re-open the database connection after the swap
				*db, err = sql.Open("sqlite3", mainDBPath)
				if err != nil {
					log.Printf("Failed to re-open the database after swap: %v", err)
				} else {
					log.Println("Successfully reconnected to the new main.db")

					// Reinitialize templates
					*tmpl, err = InitializeTemplates()
					if err != nil {
						log.Printf("Failed to reinitialize templates: %v", err)
						continue
					}

					Init(*db, *tmpl)
				}
			}
		} else if !os.IsNotExist(err) {
			log.Printf("Error checking for new_main.db: %v", err)
		}

		// Check every 2 hours
		time.Sleep(2 * time.Hour)
	}
}
