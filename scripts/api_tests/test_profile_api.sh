#!/bin/bash

# Test script for the profile API endpoints
# This script tests both GET and PUT operations on the profile endpoint

# Configuration
API_BASE_URL="http://localhost:9876/api"
TOKEN_FILE=".auth_token"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check if token file exists
if [ ! -f "$TOKEN_FILE" ]; then
  echo -e "${RED}Error: Token file not found. Please run test_login_api.sh first.${NC}"
  exit 1
fi

# Read token from file
TOKEN=$(cat $TOKEN_FILE)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}Error: Token is empty. Please run test_login_api.sh to get a valid token.${NC}"
  exit 1
fi

echo -e "${BLUE}=== Testing Profile API ===${NC}"
echo -e "${BLUE}Using token:${NC} ${TOKEN:0:20}..."

# Test 1: Get user profile
echo -e "\n${YELLOW}Test 1: GET Profile${NC}"
echo -e "${BLUE}URL:${NC} $API_BASE_URL/profile"

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X GET \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  $API_BASE_URL/profile)

# Extract the HTTP status code
HTTP_STATUS=$(echo "$RESPONSE" | tail -n 1)
# Extract the response body
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

echo -e "\n${BLUE}Response Status:${NC} $HTTP_STATUS"
echo -e "${BLUE}Response Body:${NC}"
echo "$RESPONSE_BODY" | jq . 2>/dev/null || echo "$RESPONSE_BODY"

if [ $HTTP_STATUS -eq 200 ]; then
  echo -e "\n${GREEN}✅ Get profile successful!${NC}"
  
  # Extract user ID for later use
  USER_ID=$(echo "$RESPONSE_BODY" | jq -r .id 2>/dev/null)
  echo -e "User ID: $USER_ID"
else
  echo -e "\n${RED}❌ Get profile failed with status code: $HTTP_STATUS${NC}"
fi

# Test 2: Update user profile
echo -e "\n${YELLOW}Test 2: UPDATE Profile${NC}"
echo -e "${BLUE}URL:${NC} $API_BASE_URL/profile"

# Generate a random bio to ensure we see a change
RANDOM_BIO="Updated bio $(date +%s)"

UPDATE_DATA=$(cat <<EOF
{
  "first_name": "Updated",
  "last_name": "User",
  "bio": "$RANDOM_BIO"
}
EOF
)

echo -e "${BLUE}Request Body:${NC}"
echo "$UPDATE_DATA" | jq . 2>/dev/null || echo "$UPDATE_DATA"

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$UPDATE_DATA" \
  $API_BASE_URL/profile)

# Extract the HTTP status code
HTTP_STATUS=$(echo "$RESPONSE" | tail -n 1)
# Extract the response body
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

echo -e "\n${BLUE}Response Status:${NC} $HTTP_STATUS"
echo -e "${BLUE}Response Body:${NC}"
echo "$RESPONSE_BODY" | jq . 2>/dev/null || echo "$RESPONSE_BODY"

if [ $HTTP_STATUS -eq 200 ]; then
  echo -e "\n${GREEN}✅ Update profile successful!${NC}"
  
  # Verify the bio was updated
  UPDATED_BIO=$(echo "$RESPONSE_BODY" | jq -r .user.bio 2>/dev/null)
  if [ "$UPDATED_BIO" = "$RANDOM_BIO" ]; then
    echo -e "${GREEN}✅ Bio was correctly updated to: $UPDATED_BIO${NC}"
  else
    echo -e "${RED}❌ Bio was not updated correctly${NC}"
    echo -e "Expected: $RANDOM_BIO"
    echo -e "Got: $UPDATED_BIO"
  fi
else
  echo -e "\n${RED}❌ Update profile failed with status code: $HTTP_STATUS${NC}"
fi

# Test 3: Verify profile was updated by getting it again
echo -e "\n${YELLOW}Test 3: Verify Profile Update${NC}"
echo -e "${BLUE}URL:${NC} $API_BASE_URL/profile"

RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X GET \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  $API_BASE_URL/profile)

# Extract the HTTP status code
HTTP_STATUS=$(echo "$RESPONSE" | tail -n 1)
# Extract the response body
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

echo -e "\n${BLUE}Response Status:${NC} $HTTP_STATUS"
echo -e "${BLUE}Response Body:${NC}"
echo "$RESPONSE_BODY" | jq . 2>/dev/null || echo "$RESPONSE_BODY"

if [ $HTTP_STATUS -eq 200 ]; then
  echo -e "\n${GREEN}✅ Get updated profile successful!${NC}"
  
  # Verify the bio was updated
  CURRENT_BIO=$(echo "$RESPONSE_BODY" | jq -r .bio 2>/dev/null)
  if [ "$CURRENT_BIO" = "$RANDOM_BIO" ]; then
    echo -e "${GREEN}✅ Bio verification successful: $CURRENT_BIO${NC}"
  else
    echo -e "${RED}❌ Bio verification failed${NC}"
    echo -e "Expected: $RANDOM_BIO"
    echo -e "Got: $CURRENT_BIO"
  fi
else
  echo -e "\n${RED}❌ Get updated profile failed with status code: $HTTP_STATUS${NC}"
fi

echo -e "\n${BLUE}=== Profile API Testing Complete ===${NC}"