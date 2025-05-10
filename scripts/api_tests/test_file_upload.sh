#!/bin/bash

# Test script for file upload API
# Usage: ./test_file_upload.sh <token> <file_path>

# Check if token is provided
if [ -z "$1" ]; then
  echo "Error: No authentication token provided"
  echo "Usage: ./test_file_upload.sh <token> <file_path>"
  exit 1
fi

# Check if file path is provided
if [ -z "$2" ]; then
  echo "Error: No file path provided"
  echo "Usage: ./test_file_upload.sh <token> <file_path>"
  exit 1
fi

# Check if file exists
if [ ! -f "$2" ]; then
  echo "Error: File not found: $2"
  exit 1
fi

# Set variables
TOKEN="$1"
FILE_PATH="$2"
API_URL="http://localhost:9876/api/files/upload"

# Make the API call
echo "Uploading file: $FILE_PATH"
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@$FILE_PATH" \
  "$API_URL"

echo -e "\n"