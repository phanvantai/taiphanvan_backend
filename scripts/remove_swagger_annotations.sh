#!/bin/bash

# Script to remove Swagger annotations from Go files
# This script will find all Go files in the handlers directory and remove Swagger annotations

HANDLERS_DIR="/Users/november/Developer/taiphanvan_project/taiphanvan_backend/internal/handlers"

# Find all Go files in the handlers directory
GO_FILES=$(find "$HANDLERS_DIR" -name "*.go")

for file in $GO_FILES; do
  echo "Processing $file..."
  
  # Create a temporary file
  temp_file=$(mktemp)
  
  # Process the file to remove Swagger annotations
  awk '
    # Skip blocks of Swagger annotations
    /\/\/ @/ {
      if (!in_block) {
        in_block = 1
        # Store the line before the block starts (usually contains "godoc")
        prev_line = prev
      }
      next
    }
    
    # If we were in a block and now we are not, replace the godoc line with a simple comment
    !/\/\/ @/ && in_block {
      in_block = 0
      # Extract the function name from the godoc line
      if (prev_line ~ /\/\/ [A-Za-z]+ godoc/) {
        func_name = prev_line
        gsub(/\/\/ /, "", func_name)
        gsub(/ godoc/, "", func_name)
        # Print a simple comment with the function description
        print "// " func_name " handles the request"
      }
    }
    
    # If not in a block, print the line
    !in_block {
      print $0
    }
    
    # Store the previous line
    {
      prev = $0
    }
  ' "$file" > "$temp_file"
  
  # Replace the original file with the processed one
  mv "$temp_file" "$file"
done

echo "Swagger annotations removed from all handler files."