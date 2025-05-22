#!/bin/bash

# Script to manually fetch news from RSS feeds
# Usage: ./fetch_rss.sh

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
API_URL="http://localhost:9876/api/admin/news/fetch-rss"

# Get JWT token
echo "Authenticating..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:9876/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email": "'${ADMIN_EMAIL:-admin@example.com}'", "password": "'${ADMIN_PASSWORD:-admin123}'"}')

# Extract token
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "Failed to authenticate. Please check your credentials."
    exit 1
fi

echo "Fetching RSS news..."
curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"limit": 20}' | jq .

echo "Done!"
