#!/bin/bash
# Run all API tests

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASE_URL=${1:-http://localhost:9876/api}

# Print section header
print_section() {
  echo -e "\n${BLUE}======================${NC}"
  echo -e "${BLUE}Running $1 tests${NC}"
  echo -e "${BLUE}======================${NC}\n"
}

# Print test start
print_test() {
  echo -e "${GREEN}>>> Starting test: $1${NC}"
}

# Check if server is running
echo "Checking if API server is running..."
if ! curl -s "${BASE_URL}/health" > /dev/null; then
  echo "Error: API server not running at ${BASE_URL}"
  echo "Please start the server before running tests."
  exit 1
fi

echo "API server is running. Starting tests..."

# Authentication tests
print_section "Authentication"
print_test "Register API"
$SCRIPT_DIR/test_register_api.sh $BASE_URL

print_test "Login API"
$SCRIPT_DIR/test_login_api.sh $BASE_URL

print_test "Logout API"
$SCRIPT_DIR/test_logout_api.sh $BASE_URL

# Profile tests
print_section "Profile"
print_test "Profile API"
$SCRIPT_DIR/test_profile_api.sh $BASE_URL

print_test "Avatar Upload"
$SCRIPT_DIR/test_avatar_upload.sh $BASE_URL

# File management tests
print_section "File Management"
print_test "File Upload"
$SCRIPT_DIR/test_file_upload.sh $BASE_URL

print_test "File Delete"
$SCRIPT_DIR/test_file_delete.sh $BASE_URL

# News tests
print_section "News"
print_test "News API"
$SCRIPT_DIR/test_news_api.sh $BASE_URL

print_test "RSS Feed API"
$SCRIPT_DIR/test_rss_feed_api.sh $BASE_URL

echo -e "\n${GREEN}All tests completed successfully!${NC}"
