#!/bin/bash

# Set the base directory
BASE_DIR=$(pwd)

# Find all main.go files and build them
find "$BASE_DIR" -name "main.go" | while read -r MAIN_GO_FILE; do
  # Get the directory of the main.go file
  PROJECT_DIR=$(dirname "$MAIN_GO_FILE")
  
  echo "Building project in $PROJECT_DIR..."
  
  # Change to the project directory
  cd "$PROJECT_DIR" || continue
  
  # Build the main.go file
  if go build -o main_binary main.go; then
    echo "Build succeeded for $PROJECT_DIR"
  else
    echo "Build failed for $PROJECT_DIR"
  fi
  
  # Return to the base directory
  cd "$BASE_DIR" || exit
done

echo "All projects processed."
