#!/bin/bash

# Test script for file delete API
# Usage: ./test_file_delete.sh <token> <file_url>

# Check if token is provided
if [ -z "$1" ]; then
  echo "Error: No authentication token provided"
  echo "Usage: ./test_file_delete.sh <token> <file_url>"
  exit 1
fi

# Check if file URL is provided
if [ -z "$2" ]; then
  echo "Error: No file URL provided"
  echo "Usage: ./test_file_delete.sh <token> <file_url>"
  exit 1
fi

# Set variables
TOKEN="$1"
FILE_URL="$2"
API_URL="http://localhost:9876/api/files/delete"

# Create JSON payload
JSON_PAYLOAD="{\"file_url\":\"$FILE_URL\"}"

# Make the API call
echo "Deleting file: $FILE_URL"
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$JSON_PAYLOAD" \
  "$API_URL"

echo -e "\n"