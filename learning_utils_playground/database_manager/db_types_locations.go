package db_manager

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// ///////////////////////////////////////////

// type MainDBAccomodation struct {
// 	City        string  `db:"city"`
// 	Country     string  `db:"country"`
// 	BookingURL  string  `db:"booking_url"`
// 	BookingPPPN float64 `db:"booking_pppn"`
// }

// func (a *MainDBAccomodation) TableName() string {
// 	return "accomodation"
// }

// func (a *MainDBAccomodation) CreateTableQuery() string {
// 	return fmt.Sprintf(`
//     CREATE TABLE IF NOT EXISTS %s (
//         city TEXT NOT NULL,
//         country TEXT NOT NULL,
//         booking_url TEXT,
//         booking_pppn TEXT NOT NULL
//     )`, a.TableName())
// }

// func (a *MainDBAccomodation) InsertEntry(m *DBManager) error {
// 	return genericInsert(a, m)
// }

///////////////////////////////////////////

type MainDBLocation struct {
	City    string  `db:"city"`
	Country string  `db:"country"`
	Iata1   string  `db:"iata_1"`
	Iata2   string  `db:"iata_2"`
	Iata3   string  `db:"iata_3"`
	Iata4   string  `db:"iata_4"`
	Iata5   string  `db:"iata_5"`
	Iata6   string  `db:"iata_6"`
	Iata7   string  `db:"iata_7"`
	AvgWpi  float64 `db:"avg_wpi"`
	Image1  string  `db:"image_1"`
	Image2  string  `db:"image_2"`
	Image3  string  `db:"image_3"`
	Image4  string  `db:"image_4"`
	Image5  string  `db:"image_5"`
}

func (l *MainDBLocation) TableName() string {
	return "location"
}

func (l *MainDBLocation) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		city VARCHAR(255) NOT NULL,
		country CHAR(2) NOT NULL,
		iata_1 CHAR(3) NOT NULL,
		iata_2 CHAR(3),
		iata_3 CHAR(3),
		iata_4 CHAR(3),
		iata_5 CHAR(3),
		iata_6 CHAR(3),
		iata_7 CHAR(3),
		avg_wpi FLOAT(10,1),
		image_1 TEXT,
		image_2 TEXT,
		image_3 TEXT,
		image_4 TEXT,
		image_5 TEXT
	)`, l.TableName())
}

func (l *MainDBLocation) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

///////////////////////////////////////////

// for managing the properties for the hotel API
type PropertyFetch struct {
	HotelID            int      `json:"hotel_id" db:"hotel_id"`
	AccessibilityLabel string   `json:"accessibilityLabel" db:"accessibility_label"`
	CountryCode        string   `json:"countryCode" db:"country_code"`
	PhotoUrls          []string `json:"photoUrls" db:"photo_urls"`
	IsPreferred        bool     `json:"isPreferred" db:"is_preferred"`
	Longitude          float64  `json:"longitude" db:"longitude"`
	Latitude           float64  `json:"latitude" db:"latitude"`
	Name               string   `json:"name" db:"name"`
	GrossPrice         float64  `json:"gross_price" db:"gross_price"`
	Currency           string   `json:"currency" db:"currency"`
	ReviewScore        float64  `json:"review_score" db:"review_score"`
	ReviewCount        int      `json:"review_count" db:"review_count"`
	CheckinDate        string   `json:"checkin_date" db:"checkin_date"`
	CheckoutDate       string   `json:"checkout_date" db:"checkout_date"`
	City               string   `json:"-" db:"city"`    // Assuming you want to keep this field private in JSON
	Country            string   `json:"-" db:"country"` // Assuming you want to keep this field private in JSON
}

func (p *PropertyFetch) TableName() string {
	return "property"
}

func (p *PropertyFetch) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		hotel_id INTEGER,
		accessibility_label TEXT,
		country_code TEXT,
		photo_urls TEXT,
		is_preferred INTEGER,
		longitude REAL,
		latitude REAL,
		name TEXT,
		gross_price REAL,
		currency TEXT,
		review_score REAL,
		review_count INTEGER,
		checkin_date TEXT,
		checkout_date TEXT
	)`, p.TableName())
}

func (p *PropertyFetch) InsertEntry(m *DBManager) error {
	return genericInsert(p, m)
}

// insertProperties remains unchanged as it inserts data into the database
func InsertProperties(db *sql.DB, properties []PropertyFetch, city CityFetch) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO property (hotel_id, accessibility_label, country_code, photo_urls, is_preferred, longitude, latitude, name, gross_price, currency, review_score, review_count, checkin_date, checkout_date, city, country)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, property := range properties {
		photoUrls := strings.Join(property.PhotoUrls, ",")

		fmt.Printf("Inserting Property: %s, City: %s, Price: %.2f %s\n",
			property.Name, city.CityName, property.GrossPrice, property.Currency)

		_, err := stmt.Exec(
			property.HotelID, property.AccessibilityLabel, property.CountryCode, photoUrls, property.IsPreferred,
			property.Longitude, property.Latitude, property.Name, property.GrossPrice, property.Currency, property.ReviewScore,
			property.ReviewCount, property.CheckinDate, property.CheckoutDate, city.CityName, city.CountryCode,
		)
		if err != nil {
			log.Printf("Error inserting property %s: %v", property.Name, err)
			continue
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

///////////////////////////////////////////

// City represents the city data to be inserted into the accommodation.db
type CityFetch struct {
	CityName      string `db:"city"`
	CountryCode   string `db:"country"`
	DestinationID string `db:"destination_id"`
}

func (c *CityFetch) TableName() string {
	return "city"
}

func (c *CityFetch) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		city TEXT,
		country TEXT,
		destination_id TEXT
	)`, c.TableName())
}

func (c *CityFetch) InsertEntry(m *DBManager) error {
	return genericInsert(c, m)
}

// /////////////////////////////////////////
// init-locations-db.go
type LocationsCity struct {
	ID          int     `db:"id"`
	IncludeTF   bool    `db:"include_tf"`
	City        string  `db:"city"`
	CountryCode string  `db:"countrycode"`
	Population  int     `db:"population"`
	Elevation   float64 `db:"elevation"`
	Lat         float64 `db:"lat"`
	Long        float64 `db:"long"`
}

func (l *LocationsCity) TableName() string {
	return "city"
}

func (l *LocationsCity) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		include_tf BOOLEAN,
		city TEXT,
		countrycode TEXT,
		population INTEGER,
		elevation REAL,
		lat REAL,
		long REAL
	)`, l.TableName())
}

func (l *LocationsCity) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

// /////////////////////////////////////////
type LocationsAirport struct {
	ICAO         string  `db:"icao"`
	IATA         string  `db:"iata"`
	Name         string  `db:"name"`
	City         string  `db:"city"`
	Subd         string  `db:"subd"`
	Country      string  `db:"country"`
	Elevation    int     `db:"elevation"`
	Lat          float64 `db:"lat"`
	Lon          float64 `db:"lon"`
	TZ           string  `db:"tz"`
	LID          string  `db:"lid"`
	SkyscannerID string  `db:"skyscannerid"`
}

func (l *LocationsAirport) TableName() string {
	return "airport"
}

func (l *LocationsAirport) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		icao TEXT PRIMARY KEY,
		iata TEXT,
		name TEXT,
		city TEXT,
		subd BLOB,
		country TEXT,
		elevation INTEGER,
		lat REAL,
		lon REAL,
		tz TEXT,
		lid TEXT,
		skyscannerid TEXT
	)`, l.TableName())
}

func (l *LocationsAirport) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

// /////////////////////////////////////////
type LocationsMarina struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Location   string  `db:"location"`
	Capacity   int     `db:"capacity"`
	Facilities string  `db:"facilities"`
	Lat        float64 `db:"lat"`
	Lon        float64 `db:"lon"`
}

func (l *LocationsMarina) TableName() string {
	return "marina"
}

func (l *LocationsMarina) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		capacity INTEGER,
		facilities TEXT,
		lat REAL,
		lon REAL
	)`, l.TableName())
}

func (l *LocationsMarina) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

///////////////////////////////////////////

type LocationsBeach struct {
	ID            int     `db:"id"`
	Name          string  `db:"name"`
	Location      string  `db:"location"`
	Accessibility string  `db:"accessibility"`
	Facilities    string  `db:"facilities"`
	WaterQuality  string  `db:"water_quality"`
	Lat           float64 `db:"lat"`
	Lon           float64 `db:"lon"`
}

func (l *LocationsBeach) TableName() string {
	return "beach"
}

func (l *LocationsBeach) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		accessibility TEXT,
		facilities TEXT,
		water_quality TEXT,
		lat REAL,
		lon REAL
	)`, l.TableName())
}

func (l *LocationsBeach) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

///////////////////////////////////////////

type LocationsSkiResort struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Location   string  `db:"location"`
	NumTrails  int     `db:"num_trails"`
	Difficulty string  `db:"difficulty"`
	LiftCount  int     `db:"lift_count"`
	Lat        float64 `db:"lat"`
	Lon        float64 `db:"lon"`
}

func (l *LocationsSkiResort) TableName() string {
	return "ski_resort"
}

func (l *LocationsSkiResort) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		num_trails INTEGER,
		difficulty TEXT,
		lift_count INTEGER,
		lat REAL,
		lon REAL
	)`, l.TableName())
}

func (l *LocationsSkiResort) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}

///////////////////////////////////////////

type LocationsNationalPark struct {
	ID              int     `db:"id"`
	Name            string  `db:"name"`
	Location        string  `db:"location"`
	AreaSqKm        float64 `db:"area_sq_km"`
	VisitorsPerYear int     `db:"visitors_per_year"`
	EstablishedYear int     `db:"established_year"`
	Lat             float64 `db:"lat"`
	Lon             float64 `db:"lon"`
}

func (l *LocationsNationalPark) TableName() string {
	return "national_park"
}

func (l *LocationsNationalPark) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY,
		name TEXT,
		location TEXT,
		area_sq_km REAL,
		visitors_per_year INTEGER,
		established_year INTEGER,
		lat REAL,
		lon REAL
	)`, l.TableName())
}

func (l *LocationsNationalPark) InsertEntry(m *DBManager) error {
	return genericInsert(l, m)
}
