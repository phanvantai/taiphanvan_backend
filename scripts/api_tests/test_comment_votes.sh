#!/bin/bash

# Test script for the comment vote API

# Set API_URL
API_URL="http://localhost:9876/api"

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
  echo -e "\n${YELLOW}$1${NC}\n"
}

print_success() {
  echo -e "${GREEN}$1${NC}"
}

print_error() {
  echo -e "${RED}$1${NC}"
}

# Register a test user
print_header "1. Registering a test user..."
register_response=$(curl -s -X POST "${API_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser_vote",
    "email": "testuser_vote@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }')

echo "Register response: $register_response"

# Login as the test user
print_header "2. Logging in as test user..."
login_response=$(curl -s -X POST "${API_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser_vote@example.com",
    "password": "password123"
  }')

echo "Login response: $login_response"

# Extract the token from the login response
token=$(echo $login_response | grep -o '"access_token":"[^"]*' | sed 's/"access_token":"//')

if [ -z "$token" ]; then
  print_error "Failed to get access token"
  exit 1
fi

print_success "Successfully logged in"

# Create a test post
print_header "3. Creating a test post..."
post_response=$(curl -s -X POST "${API_URL}/posts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "title": "Test Post for Comment Votes",
    "content": "This is a test post to test comment voting.",
    "status": "published"
  }')

echo "Post response: $post_response"

# Extract the post ID from the response
post_id=$(echo $post_response | grep -o '"id":[0-9]*' | head -1 | sed 's/"id"://')

if [ -z "$post_id" ]; then
  print_error "Failed to get post ID"
  exit 1
fi

print_success "Successfully created post with ID: $post_id"

# Create a test comment
print_header "4. Creating a test comment..."
comment_response=$(curl -s -X POST "${API_URL}/posts/$post_id/comments" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "content": "This is a test comment to test voting."
  }')

echo "Comment response: $comment_response"

# Extract the comment ID from the response
comment_id=$(echo $comment_response | grep -o '"id":[0-9]*' | head -1 | sed 's/"id"://')

if [ -z "$comment_id" ]; then
  print_error "Failed to get comment ID"
  exit 1
fi

print_success "Successfully created comment with ID: $comment_id"

# Get comment votes (should be 0)
print_header "5. Getting comment votes (initial)..."
votes_response=$(curl -s -X GET "${API_URL}/comments/$comment_id/votes" \
  -H "Authorization: Bearer $token")

echo "Votes response (initial): $votes_response"

# Upvote the comment
print_header "6. Upvoting the comment..."
upvote_response=$(curl -s -X POST "${API_URL}/comments/$comment_id/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "vote_type": 1
  }')

echo "Upvote response: $upvote_response"

# Get comment votes again (should be 1)
print_header "7. Getting comment votes after upvote..."
votes_response=$(curl -s -X GET "${API_URL}/comments/$comment_id/votes" \
  -H "Authorization: Bearer $token")

echo "Votes response (after upvote): $votes_response"

# Downvote the comment
print_header "8. Downvoting the comment..."
downvote_response=$(curl -s -X POST "${API_URL}/comments/$comment_id/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "vote_type": -1
  }')

echo "Downvote response: $downvote_response"

# Get comment votes again (should be -1)
print_header "9. Getting comment votes after downvote..."
votes_response=$(curl -s -X GET "${API_URL}/comments/$comment_id/votes" \
  -H "Authorization: Bearer $token")

echo "Votes response (after downvote): $votes_response"

# Remove vote from the comment
print_header "10. Removing vote from the comment..."
remove_vote_response=$(curl -s -X POST "${API_URL}/comments/$comment_id/vote" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "vote_type": 0
  }')

echo "Remove vote response: $remove_vote_response"

# Get comment votes one last time (should be 0)
print_header "11. Getting comment votes after removing vote..."
votes_response=$(curl -s -X GET "${API_URL}/comments/$comment_id/votes" \
  -H "Authorization: Bearer $token")

echo "Votes response (after removing vote): $votes_response"

# Get all comments for the post to see if they include vote info
print_header "12. Getting all comments for the post..."
comments_response=$(curl -s -X GET "${API_URL}/posts/$post_id/comments" \
  -H "Authorization: Bearer $token")

echo "Comments response: $comments_response"

print_header "Test completed!"
