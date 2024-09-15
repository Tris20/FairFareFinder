
#!/bin/bash

# Stop on error
set -e

# Function to run executables in their respective directories
run_in_directory() {
    # $1 is the directory path
    # $2 is the executable name
    echo "Navigating to directory: $1"
    
    cd "$1"

    echo "Current directory: $PWD"
    echo "Executable to run: $2"
    if [[ -x "$2" ]]; then
        echo "Executing: $2"
        ./"$2"
        if [ $? -eq 0 ]; then
            echo "Execution of $2 succeeded."
        else
            echo "Execution of $2 failed."
            exit 1
        fi
    else
        echo "Executable $2 not found or not executable."
        exit 1
    fi
    # Return to the original directory, stored at script start
    cd "$ORIGINAL_DIR"
}

# Store the original directory
ORIGINAL_DIR=$(pwd)

# List of directories and executables
# Format: directory, then executable
commands=(
    "generate/compiled-dbs/main" "main"
    "generate/compiled-dbs/main/weather" "weather"
)

# Loop through commands array in pairs
for ((i=0; i<${#commands[@]}; i+=2)); do
    run_in_directory "${commands[$i]}" "${commands[$i+1]}"
done

echo "All executables have been run successfully."

