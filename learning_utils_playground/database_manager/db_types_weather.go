package db_manager

import "fmt"

///////////////////////////////////////////

type MainDBWeather struct {
	City           string  `db:"city"`
	Country        string  `db:"country"`
	Date           string  `db:"date"`
	AvgDaytimeTemp float64 `db:"avg_daytime_temp"`
	WeatherIcon    string  `db:"weather_icon"`
	GoogleURL      string  `db:"google_url"`
	AvgDaytimeWpi  float64 `db:"avg_daytime_wpi"`
}

func (w *MainDBWeather) TableName() string {
	return "weather"
}

func (w *MainDBWeather) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		city VARCHAR(255) NOT NULL,
		country CHAR(2) NOT NULL,
		date DATE NOT NULL,
		avg_daytime_temp FLOAT(10,1),
		weather_icon VARCHAR(255),
		google_url VARCHAR(255),
		avg_daytime_wpi FLOAT(10,1)
	)`, w.TableName())
}

func (w *MainDBWeather) InsertEntry(m *DBManager) error {
	return genericInsert(w, m)
}

///////////////////////////////////////////

// taken from the backend model
// WeatherRecord holds weather data
type WeatherRecord struct {
	WeatherID         int     `db:"weather_id" noInsert:"true"`
	CityName          string  `db:"city_name"`
	CountryCode       string  `db:"country_code"`
	IATA              string  `db:"iata"`
	Date              string  `db:"date"`
	WeatherType       string  `db:"weather_type"`
	Temperature       float64 `db:"temperature"`
	WeatherIconURL    string  `db:"weather_icon_url"`
	GoogleWeatherLink string  `db:"google_weather_link"`
	WindSpeed         float64 `db:"wind_speed"`
	WPI               float64 `db:"wpi"`
}

func (w *WeatherRecord) TableName() string {
	return "weather"
}

func (w *WeatherRecord) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name TEXT,
		country_code TEXT,
		iata TEXT,
		date TEXT,
		weather_type TEXT,
		temperature REAL,
		weather_icon_url TEXT,
		google_weather_link TEXT,
		wind_speed REAL,
		wpi REAL
	)`, w.TableName())
}

func (w *WeatherRecord) InsertEntry(m *DBManager) error {
	return genericInsert(w, m)
}

// ///////////////////////////////////////////
// CurrentWeather holds weather data
type CurrentWeather struct {
	CityName          string  `db:"city_name"`
	CountryCode       string  `db:"country_code"`
	IATA              string  `db:"iata"`
	Date              string  `db:"date"`
	WeatherType       string  `db:"weather_type"`
	Temperature       float64 `db:"temperature"`
	WeatherIconURL    string  `db:"weather_icon_url"`
	GoogleWeatherLink string  `db:"google_weather_link"`
	WindSpeed         float64 `db:"wind_speed"`
	WPI               float64 `db:"wpi"`
}

func (w *CurrentWeather) TableName() string {
	return "current_weather"
}

func (w *CurrentWeather) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		city_name TEXT,
		country_code TEXT,
		iata TEXT,
		date TEXT,
		weather_type TEXT,
		temperature REAL,
		weather_icon_url TEXT,
		google_weather_link TEXT,
		wind_speed REAL,
		wpi REAL
	)`, w.TableName())
}

func (w *CurrentWeather) InsertEntry(m *DBManager) error {
	return genericInsert(w, m)
}

// ///////////////////////////////////////////
// AllWeather holds weather data
type AllWeather struct {
	WeatherID         int     `db:"weather_id" noInsert:"true"`
	CityName          string  `db:"city_name"`
	CountryCode       string  `db:"country_code"`
	IATA              string  `db:"iata"`
	Date              string  `db:"date"`
	WeatherType       string  `db:"weather_type"`
	Temperature       float64 `db:"temperature"`
	WeatherIconURL    string  `db:"weather_icon_url"`
	GoogleWeatherLink string  `db:"google_weather_link"`
	WindSpeed         float64 `db:"wind_speed"`
	WPI               float64 `db:"wpi"`
}

func (w *AllWeather) TableName() string {
	return "all_weather"
}

func (w *AllWeather) CreateTableQuery() string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		weather_id INTEGER PRIMARY KEY AUTOINCREMENT,
		city_name TEXT NOT NULL,
		country_code TEXT NOT NULL,
    	iata TEXT NOT NULL,
		date TEXT NOT NULL,
		weather_type TEXT NOT NULL,
		temperature REAL NOT NULL,
		weather_icon_url TEXT NOT NULL,
		google_weather_link TEXT NOT NULL,
    	wind_speed REAL NOT NULL,
    	wpi FLOAT(10,1)
	)`, w.TableName())
}

func (w *AllWeather) InsertEntry(m *DBManager) error {
	return genericInsert(w, m)
}
