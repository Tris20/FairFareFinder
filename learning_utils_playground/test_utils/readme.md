# Profiling Data

Profiling data is generated when you run the `FetchFlightSchedule` function with profiling enabled. The profiling data includes CPU and memory profiles.

## Accessing Profiling Data

1. **CPU Profile**: The CPU profile is saved to `testdata/FetchFlightSchedule_cpu.prof`.
2. **Memory Profile**: The memory profile is saved to `testdata/FetchFlightSchedule_mem.prof`.

## Viewing Profiling Data

To view the profiling data, you can use the `go tool pprof` command.

all this is from the starting point of `learning_utils_playground` directory.

### CPU Profile

```sh
go tool pprof ./data_management/testdata/FetchFlightSchedule_cpu.prof
```

after prof starts, you can type `web` to get a visualization of the profiling data.

You can also generate a web-based visualization of the profiling data:

```sh
go tool pprof -http=:8080 ./data_management/testdata/FetchFlightSchedule_cpu.prof
```

### Memory Profile

```sh
go tool pprof ./data_management/testdata/FetchFlightSchedule_mem.prof
```

You can also generate a web-based visualization of the profiling data:

```sh
go tool pprof -http=:8080 ./data_management/testdata/FetchFlightSchedule_cpu.prof
```

This will start a web server on port 8080 where you can interactively explore the profiling data.
