#!/bin/bash

# Set the base directory
BASE_DIR=$(pwd)/utils

# Find all main.go files and build them
find "$BASE_DIR" -type f -name "main.go" | while read -r MAIN_GO_FILE; do
  # Get the directory of the main.go file
  PROJECT_DIR=$(dirname "$MAIN_GO_FILE")
  
  echo "Found main.go in $PROJECT_DIR. Building..."
  
  # Change to the project directory
  cd "$PROJECT_DIR" || continue
  
  # Run go build in the directory
  if go build; then
    echo "Build succeeded in $PROJECT_DIR"
  else
    echo "Build failed in $PROJECT_DIR"
  fi
  
  # Return to the base directory
  cd "$BASE_DIR" || exit
done

echo "All folders processed."
