package db_manager

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// DatabaseType is an interface that all database models must implement
type DatabaseType interface {
	CreateTableQuery() string
	InsertEntry(*DBManager) error
	TableName() string
}

func DropTableQuery(dt DatabaseType) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", dt.TableName())
}

// buildInsertQuery builds an INSERT query based on struct fields tagged with `db`
func buildGenericInsertQuery(dt DatabaseType) (string, []interface{}) {
	// t := reflect.TypeOf(dt)
	v := reflect.ValueOf(dt)

	// Use reflect.Indirect to handle both pointer and non-pointer types
	t := reflect.Indirect(v).Type()
	v = reflect.Indirect(v)

	var columns []string
	var values []interface{}
	var placeholders []string

	// Iterate over struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get the db tag
		dbTag := field.Tag.Get("db")
		noInsertTag := field.Tag.Get("noInsert")

		// Skip if tag is missing or explicitly set to "-"
		if (dbTag == "" || dbTag == "-") || noInsertTag == "true" {
			continue
		}

		columns = append(columns, dbTag)

		// Handle time.Time fields
		if field.Type == reflect.TypeOf(time.Time{}) {
			values = append(values, v.Field(i).Interface().(time.Time).Format(time.RFC3339))
		} else {
			values = append(values, v.Field(i).Interface())
		}

		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		dt.TableName(),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, values
}

func genericInsert(dt DatabaseType, m *DBManager) error {
	query, values := buildGenericInsertQuery(dt)
	_, err := m.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("error creating %s: %w", dt.TableName(), err)
	}
	return nil
}

///////////////////////////////////////////

// User represents a user in the system and maps to the users table
type User struct {
	ID        int64     `db:"id" noInsert:"true"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}

// CreateTableQuery returns the SQL for creating the users table
func (u User) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`, u.TableName())
}

func (user *User) InsertEntry(m *DBManager) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := fmt.Sprintf(`
	INSERT INTO %s (email, name, created_at, updated_at)
	VALUES (?, ?, ?, ?)`, user.TableName())

	result, err := m.db.Exec(query,
		user.Email, user.Name, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert id: %w", err)
	}

	user.ID = id
	return nil
}

///////////////////////////////////////////

// RawDBFlight represents a row from the flights table.
type RawDBFlight struct {
	ID               int    `db:"id" noInsert:"true"`
	FlightNumber     string `db:"flightNumber"`
	DepartureAirport string `db:"departureAirport"`
	ArrivalAirport   string `db:"arrivalAirport"`
	DepartureTime    string `db:"departureTime"`
	ArrivalTime      string `db:"arrivalTime"`
	Direction        string `db:"direction"`
}

func (f *RawDBFlight) TableName() string {
	return "schedule"
}

// CreateTableQuery returns the SQL for creating the flights table
func (f RawDBFlight) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		flightNumber TEXT NOT NULL,
		departureAirport TEXT,
		arrivalAirport TEXT,
		departureTime TEXT,
		arrivalTime TEXT,
		direction TEXT NOT NULL
	)`, f.TableName())
}

func (flight *RawDBFlight) InsertEntry(m *DBManager) error {
	return genericInsert(flight, m)
}

///////////////////////////////////////////

type MainDBAccommodation struct {
	City        string  `db:"city"`
	Country     string  `db:"country"`
	BookingURL  string  `db:"booking_url"`
	BookingPPPN float64 `db:"booking_pppn"`
}

func (a *MainDBAccommodation) TableName() string {
	return "accommodation"
}

func (a *MainDBAccommodation) CreateTableQuery() string {
	return fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
        city TEXT NOT NULL,
        country TEXT NOT NULL,
        booking_url TEXT,
        booking_pppn TEXT NOT NULL
    )`, a.TableName())
}

func (a *MainDBAccommodation) InsertEntry(m *DBManager) error {
	return genericInsert(a, m)
}

///////////////////////////////////////////

type MainDBFiveNightsAndFlights struct {
	OriginCity         string  `db:"origin_city"`
	OriginCountry      string  `db:"origin_country"`
	DestinationCity    string  `db:"destination_city"`
	DestinationCountry string  `db:"destination_country"`
	PriceFnaf          float64 `db:"price_fnaf"`
}

func (fnaf *MainDBFiveNightsAndFlights) TableName() string {
	return "five_nights_and_flights"
}

func (fnaf *MainDBFiveNightsAndFlights) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		origin_city TEXT,
		origin_country TEXT,
		destination_city TEXT,
		destination_country TEXT,
		price_fnaf REAL
	)`, fnaf.TableName())
}

func (fnaf *MainDBFiveNightsAndFlights) InsertEntry(m *DBManager) error {
	return genericInsert(fnaf, m)
}

///////////////////////////////////////////

type MainDBFlight struct {
	ID                      int     `db:"id" noInsert:"true"`
	OriginCityName          string  `db:"origin_city_name"`
	OriginCountry           string  `db:"origin_country"`
	OriginIata              string  `db:"origin_iata"`
	OriginSkyscannerID      string  `db:"origin_skyscanner_id"`
	DestinationCityName     string  `db:"destination_city_name"`
	DestinationCountry      string  `db:"destination_country"`
	DestinationIata         string  `db:"destination_iata"`
	DestinationSkyscannerID string  `db:"destination_skyscanner_id"`
	PriceThisWeek           float64 `db:"price_this_week"`
	SkyscannerURLThisWeek   string  `db:"skyscanner_url_this_week"`
	PriceNextWeek           float64 `db:"price_next_week"`
	SkyscannerURLNextWeek   string  `db:"skyscanner_url_next_week"`
	DurationInMinutes       float64 `db:"duration_in_minutes"`
	DurationInHours         float64 `db:"duration_in_hours"`
	DurationHourDotMins     float64 `db:"duration_hour_dot_mins"`
}

func (f *MainDBFlight) TableName() string {
	return "flight"
}

func (f *MainDBFlight) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		origin_city_name TEXT,
		origin_country TEXT,
		origin_iata TEXT,
		origin_skyscanner_id TEXT,
		destination_city_name TEXT,
		destination_country TEXT,
		destination_iata TEXT,
		destination_skyscanner_id TEXT,
		price_this_week DECIMAL,
		skyscanner_url_this_week VARCHAR(255),
		price_next_week DECIMAL,
		skyscanner_url_next_week VARCHAR(255),
		duration_in_minutes DECIMAL,
		duration_in_hours DECIMAL,
		duration_hour_dot_mins REAL
	)`, f.TableName())
}

func (f *MainDBFlight) InsertEntry(m *DBManager) error {
	return genericInsert(f, m)
}

// /////////////////////////////////////////
// Data for skyscanner
type SkyScannerPrice struct {
	OriginCity              string          `db:"origin_city"`
	OriginCountry           string          `db:"origin_country"`
	OriginIATA              string          `db:"origin_iata"`
	OriginSkyScannerID      string          `db:"origin_skyscanner_id"`
	DestinationCity         string          `db:"destination_city"`
	DestinationCountry      string          `db:"destination_country"`
	DestinationIATA         string          `db:"destination_iata"`
	DestinationSkyScannerID string          `db:"destination_skyscanner_id"`
	ThisWeekend             sql.NullFloat64 `db:"this_weekend"`
	NextWeekend             sql.NullFloat64 `db:"next_weekend"`
	SkyScannerURL           string
	Duration                int `db:"duration"`
}

func (s *SkyScannerPrice) TableName() string {
	return "skyscannerprices"
}

func (s *SkyScannerPrice) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		origin_city TEXT,
		origin_country TEXT,
		origin_iata TEXT,
		origin_skyscanner_id TEXT,
		destination_city TEXT,
		destination_country TEXT,
		destination_iata TEXT,
		destination_skyscanner_id TEXT,
		this_weekend REAL,
		next_weekend REAL,
		duration INTEGER
	)`, s.TableName())
}

func (s *SkyScannerPrice) InsertEntry(m *DBManager) error {
	return genericInsert(s, m)
}
