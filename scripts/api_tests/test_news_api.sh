#!/bin/bash
# Test script for news API endpoints

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Print colored message
print_message() {
  echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Base URL
BASE_URL=${1:-http://localhost:9876/api}
ADMIN_TOKEN=""

# Login as admin to get token
print_message "Logging in as admin to get token..."
ADMIN_TOKEN=$(curl -s -X POST ${BASE_URL}/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin", "password":"replace_with_secure_password"}' | jq -r '.access_token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  print_error "Failed to get admin token"
  exit 1
fi

print_message "Successfully acquired admin token"

# Test fetching news from external API
print_message "Testing news fetch from external API..."
curl -s -X POST ${BASE_URL}/admin/news/fetch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d '{"categories":["technology", "business"], "limit":10}' | jq

# Test getting news articles
print_message "Testing get news articles..."
curl -s -X GET "${BASE_URL}/news?page=1&per_page=5" | jq

# Test getting news categories
print_message "Testing get news categories..."
curl -s -X GET ${BASE_URL}/news/categories | jq

# Test manually creating a news article
print_message "Testing manual news article creation..."
NEWS_ID=$(curl -s -X POST ${BASE_URL}/admin/news \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d '{
    "title": "Test News Article",
    "content": "This is a test news article created via API",
    "summary": "A brief summary of the test article",
    "source": "Test Script",
    "category": "technology",
    "tags": ["test", "api"]
  }' | jq -r '.id')

if [ "$NEWS_ID" == "null" ] || [ -z "$NEWS_ID" ]; then
  print_error "Failed to create test news article"
else
  print_message "Successfully created test news article with ID: ${NEWS_ID}"
  
  # Test getting a specific news article
  print_message "Testing get news article by ID..."
  curl -s -X GET ${BASE_URL}/news/${NEWS_ID} | jq
  
  # Test updating a news article
  print_message "Testing update news article..."
  curl -s -X PUT ${BASE_URL}/admin/news/${NEWS_ID} \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -d '{
      "title": "Updated Test News Article",
      "content": "This is an updated test news article",
      "tags": ["test", "api", "updated"]
    }' | jq
  
  # Test setting news status
  print_message "Testing set news status..."
  curl -s -X POST ${BASE_URL}/admin/news/${NEWS_ID}/status \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -d '{
      "status": "published"
    }' | jq
fi

print_message "All tests completed!"
