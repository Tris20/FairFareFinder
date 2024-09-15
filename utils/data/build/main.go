
package main

import (
    "fmt"
    "os"
    "os/exec"
    "log"
)

func main() {
    // Define the directory and the executable
    dir := "../process/calculate/weather"       // Change this to the directory where the executable is
    executable := "./weather"      // The executable name, e.g., "main"

    // Change to the specified directory
    err := os.Chdir(dir)
    if err != nil {
        log.Fatalf("Failed to change directory: %v", err)
    }

    fmt.Printf("Successfully changed to directory: %s\n", dir)

    // Prepare the command to run the executable
    cmd := exec.Command(executable)

    // Set the output of the command to standard output and standard error
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // Run the command (this will execute the Go executable)
    err = cmd.Run()
    if err != nil {
        log.Fatalf("Failed to run executable: %v", err)
    }

    fmt.Printf("Successfully executed: %s\n", executable)
}

