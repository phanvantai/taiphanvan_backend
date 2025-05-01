#!/bin/bash

# Test script for the login API endpoint
# This script sends a login request and displays the response

# Configuration
API_URL="http://localhost:8080/api/auth/login"
EMAIL="testuser_1746097040@example.com"  # Using the email from the registration test
PASSWORD="securePassword123"    # Using the password from the registration test

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing Login API${NC}"
echo -e "${BLUE}URL:${NC} $API_URL"
echo -e "${BLUE}Email:${NC} $EMAIL"
echo -e "${BLUE}Password:${NC} ********"

# Send the login request
echo -e "\n${BLUE}Sending request...${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  $API_URL)

# Extract the HTTP status code
HTTP_STATUS=$(echo "$RESPONSE" | tail -n 1)
# Extract the response body
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

echo -e "\n${BLUE}Response Status:${NC} $HTTP_STATUS"
echo -e "${BLUE}Response Body:${NC}"
echo "$RESPONSE_BODY" | jq . 2>/dev/null || echo "$RESPONSE_BODY"

# Check if the request was successful
if [ $HTTP_STATUS -eq 200 ]; then
  echo -e "\n${GREEN}✅ Login successful!${NC}"
  
  # Extract the token
  TOKEN=$(echo "$RESPONSE_BODY" | jq -r .access_token 2>/dev/null)
  
  if [ ! -z "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "\n${BLUE}Access Token:${NC}"
    echo "$TOKEN"
    
    # Save token to a file for use with other API calls
    echo "$TOKEN" > .auth_token
    echo -e "\nToken saved to .auth_token file for use with other API calls"
  else
    echo -e "\n${RED}⚠️ Token extraction failed${NC}"
  fi
else
  echo -e "\n${RED}❌ Login failed with status code: $HTTP_STATUS${NC}"
fi