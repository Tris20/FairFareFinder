This folder holds the long term databases on the main data server.

Because these may end up very big, and may have strange extensibility requirements, we'll separate them by purpose, and the tables can be particular suppliers or purposes etc. This gives us a bit of decoupling and flexibility should we ever need to change API.

These databases and tables are compiled periodically into a smaller results.db which is then sent to webserver's incoming_db folder. The webserver checks that folder periodically for new data, and does an atomic move as soon as it sees a new db. At the time of writing that was once every 6 hours (which has no doubt changed by the time anyone reads this). Look for a count or ticker in main to see how often we do this.

Note the following is true/intended at the time of writing. For this document, software supercedes documentation in case of conflicts.
________



Database: flights.db 

Table: airports
  Populated by:
    utils/.../add-airports-to-table
    utils/.../add-skyscanner-id-to-airports-table
  Fields:
    CREATE TABLE "airport_info" (
	"icao"	TEXT,
	"iata"	TEXT,
	"name"	TEXT,
	"city"	TEXT,
	"subd"	BLOB,
	"country"	TEXT,
	"elevation"	INTEGER,
	"lat"	REAL,
	"lon"	REAL,
	"tz"	TEXT,
	"lid"	TEXT,
	"skyscannerid"	TEXT,
	PRIMARY KEY("icao")
)
    

Table: scheduled_flights
  Populated By:
    utils/.../add-flights-to-table  
  Fields:
    CREATE TABLE flights (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        flightNumber TEXT NOT NULL,
        departureAirport TEXT,
        arrivalAirport TEXT,
        departureTime TEXT,
        arrivalTime TEXT,
        direction TEXT NOT NULL,
    )
  Fields to add:
        distance REAL, (calculate from long and lat of airports)
        duration_mins INTEGER, (Included in skyscannerAPI result)
        price REAL (Included in SkyscannerAPI result, match itineraries:0:legs:0:departure&arrival with existing flights schedule)

Table: flight_best_prices
  Populated by:
    ./FairFareFinder updateFlightPrices
  Fields:   
    CREATE TABLE skyscannerprices (
	  	origin TEXT,
		destination TEXT,
		this_weekend REAL,
		next_weekend REAL
	)


__________

Database: weather.db
Table: openweatherapi

  Populated by:
    TODO
  
  Fields: 
    CREATE TABLE Weather (
      WeatherID INT AUTO_INCREMENT PRIMARY KEY,
      CityName VARCHAR(255),
      CountryCode CHAR(2),
      Date DATE,
      WeatherType VARCHAR(50),
      Temperature DECIMAL(5,2),
      WeatherIconURL VARCHAR(255),
      GoogleWeatherLink VARCHAR(255)
    );

__________

Database: Accomodation
Table: Airbnb
  Populated by: 
    TODO

   Fields: TODO

   CREATE TABLE airbnb_availability (
      id INTEGER PRIMARY KEY,
      listing_id INTEGER,
      location TEXT,
      date_available DATE,
      price REAL,
      availability_status TEXT,
      max_guests INTEGER,
      UNIQUE(listing_id, date_available)
    );



________
Database daily_Expenses.db

Table: food_etc
  - beer 
  - meal for 2 
