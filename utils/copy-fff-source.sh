
#!/bin/bash

# Check if a name argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <name>"
  exit 1
fi

# Variables
NAME="$1"
TIMESTAMP=$(date +%Y%m%d%H%M%S)
ROOT_DIR="$(dirname "$(pwd)")"
DEST_DIR="$ROOT_DIR/../FFF-${NAME}-${TIMESTAMP}"

# Create the destination folder
mkdir -p "$DEST_DIR"

# Copy main.go if it exists
if [ -f "$ROOT_DIR/main.go" ]; then
  cp "$ROOT_DIR/main.go" "$DEST_DIR"
else
  echo "Warning: main.go not found."
fi

# Copy the /src folder
if [ -d "$ROOT_DIR/src" ]; then
  cp -r "$ROOT_DIR/src" "$DEST_DIR"
else
  echo "Warning: src folder not found."
fi

# Copy all .go files in /utils
if [ -d "$ROOT_DIR/utils" ]; then
  mkdir -p "$DEST_DIR/utils"
  find "$ROOT_DIR/utils" -maxdepth 1 -type f -name "*.go" -exec cp {} "$DEST_DIR/utils" \;
else
  echo "Warning: utils folder not found."
fi

echo "Files copied to $DEST_DIR"
