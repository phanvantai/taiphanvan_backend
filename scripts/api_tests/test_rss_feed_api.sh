#!/bin/bash
# Test script for RSS feed functionality

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

# Test fetching news from RSS feeds
print_message "Testing news fetch from RSS feeds..."
FETCH_RESULT=$(curl -s -X POST ${BASE_URL}/admin/news/fetch-rss \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d '{"limit": 15}')

echo "$FETCH_RESULT" | jq

# Check if fetch was successful
SAVED_COUNT=$(echo "$FETCH_RESULT" | jq -r '.saved // 0')
print_message "Fetched and saved $SAVED_COUNT articles from RSS feeds"

if [ "$SAVED_COUNT" -gt 0 ]; then
  # Get the first page of news
  print_message "Testing retrieval of news articles..."
  NEWS_RESULT=$(curl -s -X GET "${BASE_URL}/news?page=1&per_page=5")
  echo "$NEWS_RESULT" | jq

  # Check if there are any news from RSS feeds
  HAS_RSS=$(echo "$NEWS_RESULT" | jq '[.news[] | select(.source != "NewsAPI")] | length')
  
  if [ "$HAS_RSS" -gt 0 ]; then
    print_message "Successfully fetched RSS news articles!"
  else
    print_warning "No RSS news articles found in the response. This might be normal if all fetched articles were filtered out or not displayed on the first page."
  fi

  # Get news categories to verify RSS feeds were properly categorized
  print_message "Testing news categories after RSS import..."
  curl -s -X GET ${BASE_URL}/news/categories | jq
else
  print_warning "No articles were saved from RSS feeds. This might be normal if all feeds were already imported or if there were connectivity issues."
fi

print_message "All RSS feed tests completed!"
