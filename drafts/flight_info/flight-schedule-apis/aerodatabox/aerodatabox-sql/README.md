# Flight Data Fetcher

The Flight Data Fetcher is a powerful command-line tool designed to retrieve and store flight information based on user-specified parameters such as flight direction, airport, and date. Utilizing data from Aerodatabox via RapidAPI, this application efficiently captures both arrival and departure details, storing them in a local SQLite database for easy access and analysis.

## Features

- **Flight Direction**: Users can specify the flight direction to fetch data for either arrivals or departures.
- **Airport Selection**: Allows users to input IATA airport codes to target specific airports.
- **Date Specification**: Users can define the date for which they wish to retrieve flight data, using the DD-MM-YYYY format.
- **Local Database Storage**: Retrieved flight information is stored in a SQLite database, enabling easy data management and retrieval.
- **Flexible Time Intervals**: The program is designed to fetch data in AM and PM intervals, ensuring comprehensive coverage of daily flights.

## Requirements

- Go programming language setup
- SQLite3 for database operations
- Access to Aerodatabox API via RapidAPI (requires an API key)

## Setup and Installation

1. **Install Go**: Ensure that Go is installed on your system. Visit the [official Go website](https://golang.org/dl/) for installation instructions.
2. **SQLite3**: Verify that SQLite3 is installed. If not, follow the [SQLite3 installation guide](https://www.sqlite.org/download.html).
3. **API Key**: Obtain an Aerodatabox API key by creating an account on [RapidAPI](https://rapidapi.com/) and subscribing to the Aerodatabox API.

## Configuration

Before running the application, you must configure your API key:
1. Create a YAML file named `secrets.yaml` in an accessible location.
2. Add your Aerodatabox API key to the file as follows:

```yaml
api_keys:
  aerodatabox: YOUR_API_KEY_HERE
```

## Usage

To use the Flight Data Fetcher, navigate to the directory containing the program and run it using the command line. The following flags are available:

- `-direction`: Specify "Arrival" or "Departure" to fetch respective flight data.
- `-airport`: Enter the IATA code of the target airport.
- `-date`: Specify the date for fetching flight data in the DD-MM-YYYY format.

Example command:

```sh
go run main.go -direction Departure -airport EDI -date 27-02-2024
```

This command fetches and stores departure data for Edinburgh Airport (EDI) on February 27, 2024.

## Database Schema

The SQLite database, `flights.db`, contains a single table named `flights` with the following columns:
- `id`: Primary key.
- `flightNumber`: Flight number.
- `departureAirport`: Departure airport IATA code.
- `arrivalAirport`: Arrival airport IATA code.
- `departureTime`: Scheduled departure time.
- `arrivalTime`: Scheduled arrival time.
- `direction`: Flight direction ("Arrival" or "Departure").

## License

This project is open-source and available under the [MIT License](https://opensource.org/licenses/MIT).

## Disclaimer

This tool is for educational and informational purposes only. Please adhere to the Aerodatabox API's terms of use when fetching flight data.

---

Enjoy retrieving and analyzing flight data with ease and precision using the Flight Data Fetcher!
