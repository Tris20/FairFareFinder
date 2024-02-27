# Flight Data Fetcher(Aerodatabox to JSON)

The Flight Data Fetcher is a Go program designed to fetch flight information based on user-defined parameters for airport IATA code, flight direction (Departure or Arrival), and date. It utilizes the Aerodatabox API to retrieve flight details and saves the fetched data into JSON files organized by specified parameters.

## Features

- Fetch flight information from Aerodatabox API.
- Filter fetched data by airport IATA code, flight direction, and specific date.
- Save fetched data into structured JSON files within a results directory.
- Support for fetching data in defined time intervals (AM/PM) due to API limitations.

## Prerequisites

- Go (Golang) installed on your system.
- Access to Aerodatabox API and a valid API key.
- The `secrets.yaml` file containing your API key stored securely.

## Setup

1. **API Key**: Store your Aerodatabox API key in a `secrets.yaml` file located at a secure and accessible path. The file structure should be as follows:

    ```yaml
    api_keys:
      aerodatabox: "YOUR_API_KEY_HERE"
    ```

2. **Build the Program**: Navigate to the program directory and build the program using the Go compiler.

    ```bash
    go build
    ```

3. **Run the Program**: Execute the program with required flags for direction, airport, and date.

    ```bash
    ./flightdatafetcher --direction=Departure --airport=EDI --date=27-02-2024
    ```

### Flags

- `--direction`: Specifies the flight direction to fetch. Acceptable values are `Departure` or `Arrival`.
- `--airport`: The IATA code of the airport for which to fetch flight data.
- `--date`: The date for which to fetch flight data, in `DD-MM-YYYY` format.

## Output

The program saves fetched flight information into JSON files within a `results` directory. Filenames are structured as follows: `[IATA]-[DIRECTION]-[DD]-[MM]-[YYYY]-[AM/PM].json`, where:

- `[IATA]` is the airport IATA code.
- `[DIRECTION]` is either `DEP` for departures or `ARR` for arrivals.
- `[DD]-[MM]-[YYYY]` represents the date.
- `[AM/PM]` indicates the time interval of the flight data.

## Notes

- Ensure the `results` directory is writable.
- The program automatically creates the `results` directory if it does not exist.
- Internet connection is required to fetch data from the Aerodatabox API.

## Troubleshooting

- **API Key Errors**: Ensure your `secrets.yaml` file is correctly formatted and the path in the program points accurately to its location.
- **Date Format**: Verify the date is correctly formatted as `DD-MM-YYYY`.
- **Permission Issues**: Ensure the program has permission to create directories and write files in the execution environment.

## License

Specify your license or usage rights.

---

This README provides a comprehensive guide to running and understanding the Flight Data Fetcher program. Adjust the paths, API keys, and other specific details as necessary to fit your setup.
