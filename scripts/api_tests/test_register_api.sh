#!/bin/bash

# Test the registration API endpoint
set -e

# Set color codes for better readability
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Generate a unique username and email
TIMESTAMP=$(date +%s)
USERNAME="testuser_$TIMESTAMP"
EMAIL="testuser_$TIMESTAMP@example.com"

# API URL - Change this to match your server
API_URL="http://localhost:9876/api/auth/register"

echo -e "${BLUE}===== Testing Register API =====${NC}"
echo -e "${BLUE}Sending request to: ${API_URL}${NC}"

# Test Case 1: Successful Registration
echo -e "\n${GREEN}Test Case 1: Successful Registration${NC}"
echo "Request data:"
cat << EOF | tee /dev/tty | curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @- \
  ${API_URL}
{
  "username": "testuser_$(date +%s)",
  "email": "testuser_$(date +%s)@example.com",
  "password": "securePassword123",
  "first_name": "Test",
  "last_name": "User"
}
EOF

echo -e "\nResponse:"
echo
sleep 1

# Test Case 2: Invalid Email Format
echo -e "\n${GREEN}Test Case 2: Invalid Email Format${NC}"
echo "Request data:"
cat << EOF | tee /dev/tty | curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @- \
  ${API_URL}
{
  "username": "testuser_invalid",
  "email": "invalid_email",
  "password": "securePassword123",
  "first_name": "Test",
  "last_name": "User"
}
EOF

echo -e "\nResponse:"
echo
sleep 1

# Test Case 3: Password Too Short
echo -e "\n${GREEN}Test Case 3: Password Too Short${NC}"
echo "Request data:"
cat << EOF | tee /dev/tty | curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @- \
  ${API_URL}
{
  "username": "testuser_short",
  "email": "shortpw@example.com",
  "password": "short",
  "first_name": "Test",
  "last_name": "User"
}
EOF

echo -e "\nResponse:"
echo
sleep 1

# Test Case 4: Missing Required Fields
echo -e "\n${GREEN}Test Case 4: Missing Required Fields${NC}"
echo "Request data:"
cat << EOF | tee /dev/tty | curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @- \
  ${API_URL}
{
  "first_name": "Test",
  "last_name": "User"
}
EOF

echo -e "\nResponse:"
echo

echo -e "\n${BLUE}===== Test Complete =====${NC}"