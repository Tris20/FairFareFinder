# Steps to Reconstruct the database

0. get the locations.db from tristan

The database creation doesn't create the ascii_name column

1. Generate all the raw databases
   This has to be done manually for now, but should be automated.

```shell
cd ~/play/FairFareFinder/utils/data/process/generate/raw-dbs/flights
./flights

cd ~/play/FairFareFinder/utils/data/process/generate/raw-dbs/locations/init
./init

cd ~/play/FairFareFinder/utils/data/process/generate/raw-dbs/weather
./weather
```

Current structure of the utils directory:

```
utils/
├── code_analysis
│   ├── css-conflict-detector
│   │   ├── css-file-based-conflict-detector
│   │   │   └── main.go
│   │   └── css-urlbased-conflict-detector
│   │       └── main.go
│   └── get-all-functions-in-dir
│       ├── go.mod
│       └── main.go
├── data
│   ├── fetch
│   │   ├── accommocation
│   │   │   ├── airbnb
│   │   │   └── booking-com
│   │   │       ├── get-destination-ids
│   │   │       │   └── main.go
│   │   │       └── get-properties
│   │   │           └── main.go
│   │   ├── events
│   │   ├── flights
│   │   │   ├── prices
│   │   │   │   ├── main.go
│   │   │   │   └── schedule-for-next-14-days.go
│   │   │   └── schedule
│   │   │       ├── README.md
│   │   │       ├── aerodatabox
│   │   │       ├── go.mod
│   │   │       ├── go.sum
│   │   │       └── main.go
│   │   ├── locations
│   │   │   ├── airports
│   │   │   │   ├── README.md
│   │   │   │   ├── add-skyscanner-ids.go
│   │   │   │   ├── airports.csv
│   │   │   │   ├── go.mod
│   │   │   │   ├── go.sum
│   │   │   │   ├── main.go
│   │   │   │   └── other_data_sources
│   │   │   │       └── airports.csv
│   │   │   ├── cities
│   │   │   │   └── get-city-images
│   │   │   │       ├── pixabay
│   │   │   │       │   └── main.go
│   │   │   │       ├── sort-images
│   │   │   │       │   ├── main.go
│   │   │   │       │   └── main.go.bak
│   │   │   │       ├── wiki-media
│   │   │   │       │   └── main.go
│   │   │   │       └── zip-cities
│   │   │   │           └── main.go
│   │   │   └── train-stations
│   │   └── weather
│   │       ├── database.go
│   │       ├── go.mod
│   │       ├── go.sum
│   │       ├── main.go
│   │       └── weather.go
│   ├── process
│   │   ├── calculate
│   │   │   ├── flights
│   │   │   │   └── flight-duration
│   │   │   │       ├── go.mod
│   │   │   │       ├── go.sum
│   │   │   │       └── main.go
│   │   │   ├── main
│   │   │   │   └── five-nights-and-flights
│   │   │   │       └── main.go
│   │   │   └── weather
│   │   │       ├── main.go
│   │   │       └── utils.go
│   │   ├── compile
│   │   │   ├── accommodation
│   │   │   ├── events
│   │   │   ├── flights
│   │   │   │   └── schedule-for-next-14-days.go
│   │   │   ├── locations
│   │   │   │   └── location-images
│   │   │   │       ├── get_image_1_of_cities
│   │   │   │       │   └── main.go
│   │   │   │       └── main.go
│   │   │   └── main
│   │   │       ├── accommodation
│   │   │       │   └── booking-com
│   │   │       │       └── main.go
│   │   │       ├── backup.go
│   │   │       ├── database-setup.go
│   │   │       ├── flights
│   │   │       │   ├── get_iso_code_of_country.go
│   │   │       │   └── main.go
│   │   │       ├── go.mod
│   │   │       ├── go.sum
│   │   │       ├── locations
│   │   │       │   ├── avg_wpi.go
│   │   │       │   └── main.go
│   │   │       ├── main.go
│   │   │       ├── run-flags.go
│   │   │       └── weather
│   │   │           └── main.go
│   │   └── generate
│   │       ├── compiled-dbs
│   │       │   └── main
│   │       │       └── init-main-db.go
│   │       ├── raw-dbs
│   │       │   ├── accommodation
│   │       │   ├── events
│   │       │   ├── flights
│   │       │   │   └── init-flights.go
│   │       │   ├── locations
│   │       │   │   ├── init
│   │       │   │   │   └── init-locations-db.go
│   │       │   │   └── select-cities
│   │       │   │       └── set-iata-cities-to-true.go
│   │       │   └── weather
│   │       │       └── init-weather-db.go
│   │       └── urls
│   │           └── destination_urls.go
│   └── transfer
│       └── send-new-db-to-server
├── tests
│   └── mock-main-db-generator
│       ├── cpu.prof
│       ├── go.mod
│       ├── go.sum
│       ├── input-data
│       │   ├── accommodation.csv
│       │   ├── five_nights_and_flights.csv
│       │   ├── flight.csv
│       │   ├── location.csv
│       │   ├── sqlite_sequence.csv
│       │   └── weather.csv
│       ├── main.go
│       ├── mem.prof
│       └── mock-main-db-generator
└── time-and-date
    ├── go.mod
    ├── go.sum
    └── weekdays.go
```

in order to understand what happens in all of these, I plan on making a restructured version of it all. This will help me go over the code and understand what is happening in each of the files.
