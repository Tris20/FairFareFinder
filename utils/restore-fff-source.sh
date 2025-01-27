
#!/bin/bash

# Set the base directory for the search
BASE_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

# Search for FFF folders
FFF_FOLDERS=$(find "$BASE_DIR" -maxdepth 1 -type d -name "FFF-*")

# Check if any FFF folders are found
if [ -z "$FFF_FOLDERS" ]; then
  echo "No FFF folders found in $BASE_DIR."
  exit 1
fi

# Display the list of FFF folders
echo "Found the following FFF folders:"
select FOLDER in $FFF_FOLDERS; do
  if [ -n "$FOLDER" ]; then
    echo "You selected: $FOLDER"
    break
  else
    echo "Invalid selection. Please choose a valid folder."
  fi
done

# Define the project root directory
ROOT_DIR="$BASE_DIR/FairFareFinder"

# Copy contents to the respective locations
echo "Copying contents of $FOLDER to the project root directory at $ROOT_DIR..."

# Copy main.go
if [ -f "$FOLDER/main.go" ]; then
  cp "$FOLDER/main.go" "$ROOT_DIR/"
  echo "Copied main.go."
else
  echo "main.go not found in $FOLDER."
fi

# Copy /src folder
if [ -d "$FOLDER/src" ]; then
  rsync -a --delete "$FOLDER/src/" "$ROOT_DIR/src/"
  echo "Copied src folder."
else
  echo "src folder not found in $FOLDER."
fi

# Copy .go files in /utils
if [ -d "$FOLDER/utils" ]; then
  mkdir -p "$ROOT_DIR/utils"
  rsync -a "$FOLDER/utils/" "$ROOT_DIR/utils/"
  echo "Copied .go files from utils folder."
else
  echo "utils folder not found in $FOLDER."
fi

echo "Contents copied successfully to the project root directory."

