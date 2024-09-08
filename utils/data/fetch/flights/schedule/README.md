# Flight Data Fetcher 

This Go program fetches flight data based on configurations specified in a YAML file and stores this information in a SQLite database. The application targets the Aerodatabox API to retrieve flight information for specified airports, dates, and directions (arrivals or departures). This README provides an overview of the application, including setup instructions, usage, and an explanation of its components.

## Setup Instructions

### Prerequisites

- Go (version 1.13 or higher)
- SQLite3
- An Aerodatabox API key

### Steps

1. **Clone the Repository**: Clone or download the source code to your local machine.

2. **Obtain an Aerodatabox API Key**: Sign up at Aerodatabox's website and obtain an API key. You will need this to fetch flight data.

3. **Configure API Key**: Store your Aerodatabox API key in a `secrets.yaml` file located outside your project directory to avoid accidentally committing it to version control. The file should follow this structure:

   ```yaml
   api_keys:
     aerodatabox: YOUR_AERODATABOX_API_KEY
   ```

4. **Install Dependencies**: Install the required Go packages by running `go get` inside your project directory.

5. **Build the Application**: Compile the application using `go build` to generate an executable.

## Usage

### Configuration File

Create a `fetch-these-flights.yaml` configuration file to specify the flights you want to fetch. The file should follow this format:

```yaml
flights:
  - direction: "Departure"
    airport: "BER"
    startDate: "27-03-2024"
    endDate: "29-03-2024"
  - direction: "Arrival"
    airport: "BER"
    startDate: "01-04-2024"
    endDate: "03-04-2024"
```

### Running the Application

Execute the compiled program. Ensure the configuration file `fetch-these-flights.yaml` is in the same directory as the executable or provide the path to it. The application reads the configuration, fetches the flight data for the specified periods and directions, and stores the data in a SQLite database named `flights.db`.

## Components

### Main Application (`main.go`)

- **API Key Reading**: Reads the Aerodatabox API key from the `secrets.yaml` file.
- **Flight Data Fetching**: Fetches flight data from the Aerodatabox API based on the configurations specified in the `fetch-these-flights.yaml` file.
- **Database Interaction**: Creates a SQLite database (if not existing) and stores fetched flight data for future reference.

### Configuration File (`fetch-these-flights.yaml`)

Specifies the flights for which data should be fetched, including direction, airport code, and date ranges.

### SQLite Database (`flights.db`)

Stores flight information retrieved from the Aerodatabox API, including flight numbers, departure and arrival airports, and times.

## Final Note

This application provides a straightforward way to fetch and store flight data for specified dates and airports. It can be modified or extended to include more detailed flight information or to interact with other APIs.njoy retrieving and analyzing flight data with ease and precision using the Flight Data Fetcher!
