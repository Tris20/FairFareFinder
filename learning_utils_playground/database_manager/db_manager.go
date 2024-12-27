package db_manager

import (
	"database/sql"
	"fmt"
	"log"
)

type newthing struct {
	ID int64 `db:"id"`
}

// DBManager handles database operations and schema management
type DBManager struct {
	db *sql.DB
}

// NewDBManager creates a new database manager
func NewDBManager(dbPath string) (*DBManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &DBManager{db: db}, nil
}

// Close closes the database connection
func (m *DBManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// InitializeTables creates all necessary tables
func (m *DBManager) InitializeTables() error {
	// Create a slice of all models that implement table definitions
	tables := []DatabaseType{
		&User{},
		// Add other models here
	}

	for _, table := range tables {
		query := table.CreateTableQuery()
		if _, err := m.db.Exec(query); err != nil {
			return fmt.Errorf("error creating table: %w", err)
		}
	}

	return nil
}

// ////////////////////////
// user related functions
// ////////////////////////
func (m *DBManager) GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := m.db.QueryRow(`
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE id = ?`, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return user, nil
}

// utils

func CreateTable(db *sql.DB, table DatabaseType) error {
	// Create table
	_, err := db.Exec(table.CreateTableQuery())
	if err != nil {
		log.Fatalf("Failed to create table %s: %v", table.TableName(), err)
		return err
	}
	return nil
}

func DropTable(db *sql.DB, table DatabaseType) error {
	// Drop table if it exists
	_, err := db.Exec(DropTableQuery(table))
	if err != nil {
		log.Fatalf("Failed to drop table %s: %v", table.TableName(), err)
		return err
	}
	return nil
}

func RecreateTable(db *sql.DB, table DatabaseType) error {

	err := DropTable(db, table)
	if err != nil {
		return err
	}
	return CreateTable(db, table)
}
