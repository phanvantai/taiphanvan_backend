#!/bin/bash

# Set base URL - change this if your server runs on a different URL/port
BASE_URL="http://localhost:8080/api"
EMAIL="test@example.com"
PASSWORD="password123"
USERNAME="testuser"
FIRST_NAME="Test"
LAST_NAME="User"

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Starting Logout API Test${NC}"
echo "=================================================="

# Step 1: Register a new user (if needed)
echo -e "\n${YELLOW}Step 1: Registering a new user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"username\": \"$USERNAME\",
    \"first_name\": \"$FIRST_NAME\",
    \"last_name\": \"$LAST_NAME\"
  }")

echo "Register Response: $REGISTER_RESPONSE"

# Step 2: Log in to get a token
echo -e "\n${YELLOW}Step 2: Logging in to get token...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "Login Response: $LOGIN_RESPONSE"

# Extract token from login response
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | grep -o '[^"]*$')

if [ -z "$TOKEN" ]; then
  echo -e "${RED}Failed to get token. Cannot continue test.${NC}"
  exit 1
fi

echo -e "${GREEN}Successfully obtained token.${NC}"

# Step 3: Test a protected endpoint to verify token works
echo -e "\n${YELLOW}Step 3: Testing protected endpoint with token...${NC}"
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $TOKEN")

echo "Profile Response: $PROFILE_RESPONSE"

# Step 4: Logout
echo -e "\n${YELLOW}Step 4: Testing logout endpoint...${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $TOKEN")

echo "Logout Response: $LOGOUT_RESPONSE"

# Step 5: Try to use the token again after logout
echo -e "\n${YELLOW}Step 5: Testing protected endpoint after logout...${NC}"
AFTER_LOGOUT_RESPONSE=$(curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $TOKEN")

echo "After Logout Response: $AFTER_LOGOUT_RESPONSE"

# Check if token was properly invalidated
if echo "$AFTER_LOGOUT_RESPONSE" | grep -q "revoked\|invalid\|expired\|unauthorized"; then
  echo -e "\n${GREEN}✅ Test Passed: Token was properly invalidated after logout${NC}"
else
  echo -e "\n${RED}❌ Test Failed: Token still works after logout${NC}"
fi

echo -e "\n${YELLOW}Logout API test completed${NC}"
echo "=================================================="