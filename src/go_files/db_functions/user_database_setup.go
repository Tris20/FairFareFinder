package user_db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// SetupDatabase is now exported by starting with an uppercase letter
func Setup_database() {
	// fmt.Println("Setting up the database...")
	hello_myself()
	fmt.Println("Setting up the database...")
	// Add your database setup logic here
}

func hello_myself() {
	fmt.Println("Hello, myself")
}

func CreateDatabase(dbPath string) {
	// Connect to the SQLite database
	// Since the database file does not exist, the driver creates it automatically
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Example: Creating a table
	createUserTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER PRIMARY KEY AUTOINCREMENT,  
		"username" TEXT NOT NULL,
		"email" TEXT UNIQUE
	);`

	createPreferencesTableSQL := `CREATE TABLE IF NOT EXISTS preferences (
		"user_id" INTEGER PRIMARY KEY,
		"preference" TEXT,
		FOREIGN KEY (user_id) REFERENCES USERS(user_id)
	);`

	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(createPreferencesTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Database and table created.")
}

// AddNewUserWithPreferences adds a new user and creates a corresponding preferences entry
func AddNewUserWithPreferences(dbPath, username, email, preference string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to start transaction: %v", err)
	}

	// Insert into users table
	res, err := tx.Exec("INSERT INTO users(username, email) VALUES (?, ?)", username, email)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		log.Fatalf("Failed to insert user: %v", err)
	}

	// Get the last inserted ID
	userID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback() // Rollback in case of error
		log.Fatalf("Failed to retrieve last insert ID: %v", err)
	}

	// Insert into preferences table with the new user's ID
	_, err = tx.Exec("INSERT INTO preferences(user_id, preference) VALUES (?, ?)", userID, preference)
	if err != nil {
		tx.Rollback() // Rollback in case of error
		log.Fatalf("Failed to insert preferences: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Printf("New user added with ID %d and corresponding preferences entry created.\n", userID)
}

// AddColumnToPreferencesTable adds a new column to the preferences table
// func AddColumnToPreferencesTable(dbPath, columnName, columnType string) {
// 	// print formated string with the column name and type
// 	fmt.Printf("Adding column '%s' of type '%s' to preferences table...\n", columnName, columnType)

// 	db, err := sql.Open("sqlite3", dbPath)
// 	if err != nil {
// 		log.Fatalf("Failed to open database: %v", err)
// 	}
// 	defer db.Close()

// 	// Prepare the SQL statement to add a new column
// 	alterTableSQL := fmt.Sprintf("ALTER TABLE preferences ADD COLUMN IF NOT EXISTS %s %s;", columnName, columnType)

// 	// Execute the SQL statement
// 	_, err = db.Exec(alterTableSQL)
// 	if err != nil {
// 		log.Fatalf("Failed to add new column to preferences table: %v", err)
// 	}

// 	log.Printf("Column '%s' added to preferences table.\n", columnName)
// }
