package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

// populateTables inserts the provided data into each table.
func populateTables(db *sql.DB) error {
	// Begin a transaction so that all inserts are done atomically.
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Use a deferred function to commit or rollback.
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 1. Populate airline_multipliers table.
	airlineData := []struct {
		Airline    string
		Multiplier float64
	}{
		{"Ryanair", 0.6},
		{"Wizz Air", 0.7},
		{"EasyJet", 0.8},
		{"Allegiant Air", 0.8},
		{"Frontier Airlines", 0.8},
		{"Spirit Airlines", 0.8},
		{"Southwest Airlines", 0.9},
		{"JetBlue", 1.0},
		{"Alaska Airlines", 1.0},
		{"AirAsia", 1.0},
		{"IndiGo", 1.0},
		{"Vueling", 1.0},
		{"Norwegian", 1.0},
		{"Scoot", 1.0},
		{"Delta Air Lines", 1.2},
		{"United Airlines", 1.2},
		{"American Airlines", 1.2},
		{"Aer Lingus", 1.2},
		{"Turkish Airlines", 1.3},
		{"Lufthansa", 1.3},
		{"British Airways", 1.4},
		{"Air France", 1.4},
		{"KLM Royal Dutch Airlines", 1.4},
		{"TAP Air Portugal", 1.4},
		{"Iberia", 1.4},
		{"Virgin Atlantic", 1.4},
		{"Air Canada", 1.4},
		{"Emirates", 1.5},
		{"Qatar Airways", 1.5},
		{"Thai Airways", 1.5},
		{"Vietnam Airlines", 1.5},
		{"Malaysia Airlines", 1.5},
		{"Japan Airlines (JAL)", 1.6},
		{"ANA (All Nippon Airways)", 1.6},
		{"Qantas", 1.7},
		{"Cathay Pacific", 1.7},
		{"Finnair", 1.7},
		{"Swiss International Air Lines", 1.8},
		{"Austrian Airlines", 1.8},
		{"Etihad Airways", 1.8},
		{"Korean Air", 1.9},
		{"EVA Air", 1.9},
		{"SAS (Scandinavian Airlines)", 1.9},
		{"Singapore Airlines", 2.0},
		{"Hawaiian Airlines", 2.0},
		{"China Airlines", 2.0},
		{"LATAM Airlines", 2.0},
		{"Avianca", 2.0},
		{"South African Airways", 2.0},
		{"Philippine Airlines", 2.0},
		{"Asiana Airlines", 2.1},
		{"Garuda Indonesia", 2.1},
		{"Air New Zealand", 2.1},
		{"Oman Air", 2.1},
		{"Royal Air Maroc", 2.2},
		{"Saudi Arabian Airlines (Saudia)", 2.2},
		{"SriLankan Airlines", 2.2},
		{"Air India", 2.3},
		{"Hainan Airlines", 2.3},
		{"China Southern Airlines", 2.3},
		{"China Eastern Airlines", 2.3},
		{"Gulf Air", 2.3},
		{"Azul Brazilian Airlines", 2.5},
		{"Singapore Airlines Suites", 3.0},
	}

	for _, a := range airlineData {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO airline_multipliers (airline, multiplier) VALUES (?, ?)`,
			a.Airline, a.Multiplier,
		)
		if err != nil {
			return err
		}
	}

	// 2. Populate date_modifiers table with some sample rows.
	// Full date_modifiers data based on the provided table
	dateData := []struct {
		StartDate  string
		EndDate    string
		Multiplier float64
		Reason     string
		Countries  string
	}{
		{"2025-01-01", "2025-01-01", 2.0, "New Year's Day surge pricing", "Global"},
		{"2025-01-02", "2025-01-05", 1.6, "Post-New Year return travel", "Global"},
		{"2025-01-06", "2025-01-06", 1.4, "Epiphany Holiday", "ES,IT,DE,AT"},
		{"2025-01-07", "2025-01-15", 1.1, "Post-holiday lull", "Global"},
		{"2025-01-16", "2025-01-31", 1.0, "Off-peak winter pricing", "Global"},
		{"2025-02-01", "2025-02-10", 1.0, "Low season, cheap fares", "Global"},
		{"2025-02-11", "2025-02-14", 1.3, "Valentine’s Day trips", "Global"},
		{"2025-02-15", "2025-02-25", 1.5, "School winter break", "DE,FR,UK,CH,NL"},
		{"2025-02-26", "2025-02-28", 1.0, "Low season resumes", "Global"},
		{"2025-03-01", "2025-03-15", 1.0, "Late winter shoulder season", "Global"},
		{"2025-03-16", "2025-03-19", 1.2, "Early spring travel picks up", "Global"},
		{"2025-03-17", "2025-03-17", 1.5, "St. Patrick’s Day", "IE,US"},
		{"2025-03-20", "2025-03-24", 1.3, "Spring travel picks up", "Global"},
		{"2025-03-25", "2025-03-31", 1.5, "Easter/Spring Break starts", "Global"},
		{"2025-04-01", "2025-04-10", 1.8, "Peak Easter holiday travel", "Global"},
		{"2025-04-11", "2025-04-20", 1.3, "Easter return traffic", "Global"},
		{"2025-04-21", "2025-04-23", 1.5, "Eid al-Fitr", "SA,AE,IN,ID,MY,PK"},
		{"2025-04-24", "2025-04-30", 1.0, "Spring shoulder season", "Global"},
		{"2025-05-01", "2025-05-01", 1.4, "May Day (Labor Day)", "EU,CN,RU,BR"},
		{"2025-05-02", "2025-05-05", 1.4, "Early summer trips", "EU,CN,RU"},
		{"2025-05-06", "2025-05-20", 1.0, "Pre-summer lower demand", "Global"},
		{"2025-05-21", "2025-05-31", 1.3, "Memorial Day (USA), early travel", "US"},
		{"2025-06-01", "2025-06-10", 1.4, "Start of summer travel", "Global"},
		{"2025-06-11", "2025-06-20", 1.5, "Peak pre-holiday travel", "Global"},
		{"2025-06-21", "2025-06-30", 1.7, "Schools close, summer peak", "EU,US,CA,UK"},
		{"2025-07-01", "2025-07-10", 2.0, "Peak summer vacation season", "Global"},
		{"2025-07-04", "2025-07-04", 1.7, "Independence Day", "US"},
		{"2025-07-11", "2025-07-20", 2.0, "High summer pricing continues", "Global"},
		{"2025-07-14", "2025-07-14", 1.5, "Bastille Day", "FR"},
		{"2025-07-21", "2025-07-31", 1.9, "Mid-summer, still expensive", "Global"},
		{"2025-08-01", "2025-08-01", 1.5, "Swiss National Day", "CH"},
		{"2025-08-01", "2025-08-10", 1.8, "Late summer vacations", "Global"},
		{"2025-08-15", "2025-08-15", 1.6, "Assumption Day", "IT,FR,ES,DE"},
		{"2025-08-11", "2025-08-20", 1.5, "Summer winding down", "Global"},
		{"2025-08-21", "2025-08-31", 1.2, "Back-to-school, demand drops", "EU,US,UK"},
		{"2025-09-01", "2025-09-10", 1.0, "Shoulder season, cheaper fares", "Global"},
		{"2025-09-11", "2025-09-30", 0.9, "Low demand, cheap flights", "Global"},
		{"2025-10-01", "2025-10-10", 0.9, "Off-season continues", "Global"},
		{"2025-10-03", "2025-10-03", 1.4, "German Unity Day", "DE"},
		{"2025-10-11", "2025-10-20", 1.0, "Fall travel picks up", "Global"},
		{"2025-10-31", "2025-10-31", 1.3, "Halloween", "US,UK,CA"},
		{"2025-11-01", "2025-11-10", 1.1, "Pre-holiday travel starts", "Global"},
		{"2025-11-11", "2025-11-11", 1.2, "Veterans Day / Armistice Day", "US,FR,DE"},
		{"2025-11-20", "2025-11-22", 1.4, "Thanksgiving travel begins", "US"},
		{"2025-11-23", "2025-11-26", 1.8, "Thanksgiving peak travel", "US"},
		{"2025-11-27", "2025-11-30", 1.5, "Black Friday, Cyber Monday", "US,CA,UK"},
		{"2025-12-01", "2025-12-10", 1.3, "Christmas travel begins", "Global"},
		{"2025-12-11", "2025-12-20", 1.6, "Pre-Christmas peak travel", "Global"},
		{"2025-12-21", "2025-12-24", 2.2, "Christmas holiday peak", "Global"},
		{"2025-12-25", "2025-12-25", 1.3, "Cheaper day to fly (low demand)", "Global"},
		{"2025-12-26", "2025-12-30", 1.8, "Post-Christmas return travel", "Global"},
		{"2025-12-31", "2025-12-31", 2.0, "New Year's Eve surge pricing", "Global"},
	}

	for _, d := range dateData {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO date_modifiers (start_date, end_date, multiplier, reason, countries) VALUES (?, ?, ?, ?, ?)`,
			d.StartDate, d.EndDate, d.Multiplier, d.Reason, d.Countries,
		)
		if err != nil {
			return err
		}
	}

	// 3. Populate population_modifiers table.
	populationData := []struct {
		MinPopulation int
		MaxPopulation int
		Multiplier    float64
		Description   string
	}{
		{0, 9999, 2.6, "Remote villages/islands"},
		{10000, 49999, 2.4, "Many small towns/cities"},
		{50000, 99999, 2.2, "Small cities (e.g., St. Gallen, Truro)"},
		{100000, 499999, 2.0, "Mid-sized cities (e.g., Aberdeen, Brest)"},
		{500000, 999999, 1.8, "Larger cities (e.g., Glasgow, Leeds)"},
		{1000000, 1999999, 1.6, "Major cities (e.g., Vienna, Budapest)"},
		{2000000, 2999999, 1.5, "Large cities (e.g., Warsaw, Bucharest)"},
		{3000000, 4999999, 1.4, "Very large cities (e.g., Berlin, Prague)"},
		{5000000, 6999999, 1.3, "Mega cities (e.g., Barcelona, Dallas metro)"},
		{7000000, 9999999, 1.2, "Ultra mega cities (e.g., Hong Kong, London)"},
		{10000000, 1000000000, 1.0, "Largest cities (e.g., Tokyo, Shanghai)"},
	}

	for _, p := range populationData {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO population_modifiers (min_population, max_population, multiplier, description) VALUES (?, ?, ?, ?)`,
			p.MinPopulation, p.MaxPopulation, p.Multiplier, p.Description,
		)
		if err != nil {
			return err
		}
	}

	// 4. Populate flight_frequency_modifiers table.
	// We use sql.NullInt64 for possible NULL values.
	flightFrequencyData := []struct {
		MinFlights int
		MaxFlights sql.NullInt64
		Multiplier float64
		Notes      string
	}{
		// For "100+" flights, we insert NULL for max_flights.
		{100, sql.NullInt64{Valid: false}, 0.8, "Ultra-high frequency routes (100+)"},
		{50, sql.NullInt64{Int64: 99, Valid: true}, 0.9, "Very frequent routes"},
		{20, sql.NullInt64{Int64: 49, Valid: true}, 1.1, "Frequent but not oversaturated"},
		{10, sql.NullInt64{Int64: 19, Valid: true}, 1.3, "Limited direct flight options"},
		{5, sql.NullInt64{Int64: 9, Valid: true}, 1.5, "Few flights, higher price due to scarcity"},
		{2, sql.NullInt64{Int64: 4, Valid: true}, 1.8, "Very rare direct flights"},
		{1, sql.NullInt64{Int64: 1, Valid: true}, 2.0, "Only one flight available over 5 days"},
		{0, sql.NullInt64{Int64: 0, Valid: true}, 2.5, "No direct flights (layovers required)"},
	}

	for _, f := range flightFrequencyData {
		if !f.MaxFlights.Valid {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO flight_frequency_modifiers (min_flights, max_flights, multiplier, notes) VALUES (?, NULL, ?, ?)`,
				f.MinFlights, f.Multiplier, f.Notes,
			)
		} else {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO flight_frequency_modifiers (min_flights, max_flights, multiplier, notes) VALUES (?, ?, ?, ?)`,
				f.MinFlights, f.MaxFlights.Int64, f.Multiplier, f.Notes,
			)
		}
		if err != nil {
			return err
		}
	}

	// 5. Populate short_notice_modifiers table.
	shortNoticeData := []struct {
		MinDays     int
		MaxDays     sql.NullInt64
		Multiplier  float64
		Explanation string
	}{
		// For the "90+ days" row, we set a lower bound and no upper bound.
		{90, sql.NullInt64{Valid: false}, 0.9, "Early bird deals; low demand"},
		{60, sql.NullInt64{Int64: 89, Valid: true}, 1.0, "Typical booking window"},
		{30, sql.NullInt64{Int64: 59, Valid: true}, 1.1, "Moderate demand"},
		{14, sql.NullInt64{Int64: 29, Valid: true}, 1.2, "Increased demand"},
		{7, sql.NullInt64{Int64: 13, Valid: true}, 1.4, "Higher demand as departure nears"},
		{3, sql.NullInt64{Int64: 6, Valid: true}, 1.6, "Last-minute bookings, higher prices"},
		{1, sql.NullInt64{Int64: 2, Valid: true}, 1.8, "Very last minute; premium fares"},
		{0, sql.NullInt64{Int64: 0, Valid: true}, 2.0, "Same day departure; urgent travel"},
	}

	for _, s := range shortNoticeData {
		if !s.MaxDays.Valid {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO short_notice_modifiers (min_days, max_days, multiplier, explanation) VALUES (?, NULL, ?, ?)`,
				s.MinDays, s.Multiplier, s.Explanation,
			)
		} else {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO short_notice_modifiers (min_days, max_days, multiplier, explanation) VALUES (?, ?, ?, ?)`,
				s.MinDays, s.MaxDays.Int64, s.Multiplier, s.Explanation,
			)
		}
		if err != nil {
			return err
		}
	}

	// 6. Populate aircraft_capacity_modifiers table.
	aircraftCapacityData := []struct {
		MinCapacity int
		MaxCapacity sql.NullInt64
		Multiplier  float64
		Description string
	}{
		{0, sql.NullInt64{Int64: 49, Valid: true}, 1.3, "Small turboprops (e.g., DHC-6 Twin Otter)"},
		{50, sql.NullInt64{Int64: 100, Valid: true}, 1.1, "Regional jets (e.g., CRJ-900, Embraer E175)"},
		{100, sql.NullInt64{Int64: 200, Valid: true}, 1.0, "Narrow-body jets (e.g., A319, B737)"},
		{200, sql.NullInt64{Int64: 300, Valid: true}, 0.9, "Larger narrow-bodies (e.g., B737-800/Max, A321)"},
		// For "300+" seats, we use NULL for the max_capacity.
		{300, sql.NullInt64{Valid: false}, 0.8, "Wide-bodies (e.g., B777, A350)"},
	}

	for _, a := range aircraftCapacityData {
		if !a.MaxCapacity.Valid {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO aircraft_capacity_modifiers (min_capacity, max_capacity, multiplier, description) VALUES (?, NULL, ?, ?)`,
				a.MinCapacity, a.Multiplier, a.Description,
			)
		} else {
			_, err = tx.Exec(
				`INSERT OR IGNORE INTO aircraft_capacity_modifiers (min_capacity, max_capacity, multiplier, description) VALUES (?, ?, ?, ?)`,
				a.MinCapacity, a.MaxCapacity.Int64, a.Multiplier, a.Description,
			)
		}
		if err != nil {
			return err
		}
	}

	// 7. Populate route_classification_modifiers table.
	routeClassificationData := []struct {
		Classification string
		Multiplier     float64
		Description    string
	}{
		{"Pure Business", 1.5, "High corporate travel; strong demand"},
		{"Mixed Business/Leisure", 1.2, "Both business and leisure travelers"},
		{"Pure Leisure", 1.1, "Vacation destinations; seasonal demand"},
		{"Essential/Remote", 1.4, "Lifeline routes to remote areas"},
		{"Low-Cost Tourist", 1.0, "Ultra-budget routes with lower base fares"},
		{"Hub-to-Hub", 1.3, "Major airline hubs with frequent connections"},
		{"Seasonal Charter", 1.2, "Seasonal demand with charter pricing"},
	}

	for _, r := range routeClassificationData {
		_, err := tx.Exec(
			`INSERT INTO route_classification_modifiers (classification, multiplier, description) VALUES (?, ?, ?)`,
			r.Classification, r.Multiplier, r.Description,
		)
		if err != nil {
			return err
		}
	}

	// 8. Populate aircraft_capacity_lookup table.
	aircraftLookupTable := `
	CREATE TABLE IF NOT EXISTS aircraft_capacity_lookup (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		aircraft_model TEXT NOT NULL,
		seating_capacity TEXT NOT NULL
	);
	`
	_, err = tx.Exec(aircraftLookupTable)
	if err != nil {
		return err
	}

	aircraftLookupData := []struct {
		Model    string
		Capacity string
	}{
		{"ATR 42-300", "42"},
		{"ATR 42/72 Freighter", "10"},
		{"ATR 72", "70"},
		{"Airbus A220-100", "100"},
		{"Airbus A220-300", "130"},
		{"Airbus A318", "107"},
		{"Airbus A319", "124"},
		{"Airbus A320", "150"},
		{"Airbus A320 NEO", "150"},
		{"Airbus A321", "185"},
		{"Airbus A330", "250"},
		{"Airbus A330-200", "247"},
		{"Airbus A330-300", "277"},
		{"Airbus A330-900", "287"},
		{"Airbus A340-300", "295"},
		{"Airbus A340-600", "380"},
		{"Airbus A350-900", "325"},
		{"Airbus A350-1000", "366"},
		{"Airbus A380-800", "525"},
		{"Boeing 737", "130"},
		{"Boeing 737-600", "108"},
		{"Boeing 737-700", "126"},
		{"Boeing 737-800", "162"},
		{"Boeing 737-900", "180"},
		{"Boeing 747-400", "416"},
		{"Boeing 747-8", "467"},
		{"Boeing 747-8f (freighter)", "10"},
		{"Boeing 757-200", "200"},
		{"Boeing 757-300", "243"},
		{"Boeing 767-300", "218"},
		{"Boeing 767-400", "245"},
		{"Boeing 777-200", "314"},
		{"Boeing 777-200LR", "301"},
		{"Boeing 777-300", "368"},
		{"Boeing 777-300ER", "365"},
		{"Boeing 787-8", "242"},
		{"Boeing 787-9", "280"},
		{"Bombardier CRJ1000", "100"},
		{"Bombardier CRJ900", "90"},
		{"Bombardier Dash 8 / DHC-8", "37"},
		{"Bombardier Dash 8 Q400 / DHC-8-400", "68"},
		{"De Havilland Canada DHC-3 Otter", "10"},
		{"De Havilland Canada DHC-6 Twin Otter", "19"},
		{"Embraer 170", "70"},
		{"Embraer 175", "78"},
		{"Embraer 190", "98"},
		{"Embraer 195", "104"},
		{"Embraer RJ145", "50"},
		{"Fairchild-Swearingen SA226", "19"},
		{"Saab 340", "34"},
	}

	for _, a := range aircraftLookupData {
		_, err = tx.Exec(
			`INSERT OR IGNORE INTO aircraft_capacity_lookup (aircraft_model, seating_capacity) VALUES (?, ?)`,
			a.Model, a.Capacity,
		)
		if err != nil {
			return err
		}
	}

	// 9. Create and populate route_classification_lookup table.
	// This table will be created in the current (flight_price_modifiers.db) database.
	createRouteLookupTable := `
	CREATE TABLE IF NOT EXISTS route_classification_lookup (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		departureAirport TEXT NOT NULL,
		arrivalAirport TEXT NOT NULL,
		route_classification TEXT
	);
	`
	_, err = tx.Exec(createRouteLookupTable)
	if err != nil {
		return err
	}

	// Attach the flights.db database (which contains the schedule table)
	attachStmt := `ATTACH DATABASE '../../../data/raw/flights/flights.db' AS flights_db;`
	_, err = tx.Exec(attachStmt)
	if err != nil {
		return fmt.Errorf("failed to attach flights_db: %w", err)
	}

	// Query unique routes from the schedule table in flights_db.
	rows, err := tx.Query(`
		SELECT DISTINCT departureAirport, arrivalAirport
		FROM flights_db.schedule
		ORDER BY departureAirport, arrivalAirport;
	`)
	if err != nil {
		return fmt.Errorf("failed to query unique routes: %w", err)
	}
	defer rows.Close()

	// Default classification value; adjust as needed.
	defaultClassification := "Mixed Business/Leisure"

	for rows.Next() {
		var dep, arr string
		if err := rows.Scan(&dep, &arr); err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT OR IGNORE INTO route_classification_lookup (departureAirport, arrivalAirport, route_classification)
			VALUES (?, ?, ?)
		`, dep, arr, defaultClassification)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return err
}
