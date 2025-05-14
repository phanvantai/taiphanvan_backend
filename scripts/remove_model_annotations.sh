#!/bin/bash

# Script to remove Swagger annotations from model files
# This script will find all Go files in the models directory and remove Swagger annotations

MODELS_DIR="/Users/november/Developer/taiphanvan_project/taiphanvan_backend/internal/models"

# Find all Go files in the models directory
GO_FILES=$(find "$MODELS_DIR" -name "*.go")

for file in $GO_FILES; do
  echo "Processing $file..."
  
  # Create a temporary file
  temp_file=$(mktemp)
  
  # Process the file to remove Swagger annotations
  sed -E 's/\/\/ @Description.*$//' "$file" > "$temp_file"
  
  # Replace the original file with the processed one
  mv "$temp_file" "$file"
done

echo "Swagger annotations removed from all model files."