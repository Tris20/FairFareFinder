package backend

import (
	"database/sql"
	"log"

	"github.com/Tris20/FairFareFinder/src/backend/model"
)

// SetFlightDurationInt handles setting integer duration fields and logging.
func SetFlightDurationInt(flight *model.Flight, duration sql.NullInt64, field *sql.NullInt64, logFormat string) {
	if duration.Valid {
		*field = duration
		log.Printf(logFormat, duration.Int64, flight.DestinationCityName)
	} else {
		log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
	}
}

// SetFlightDurationFloat handles setting float duration fields and logging.
func SetFlightDurationFloat(flight *model.Flight, duration sql.NullFloat64, field *sql.NullFloat64, logFormat string) {
	if duration.Valid {
		*field = duration
		log.Printf(logFormat, duration.Float64, flight.DestinationCityName)
	} else {
		log.Printf("No valid duration found for flight to %s", flight.DestinationCityName)
	}
}
