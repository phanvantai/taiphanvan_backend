#!/bin/bash

# Test script for avatar upload API
# Usage: ./test_avatar_upload.sh <access_token> <image_path>

# Check if access token and image path are provided
if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: $0 <access_token> <image_path>"
  exit 1
fi

ACCESS_TOKEN=$1
IMAGE_PATH=$2

# Check if the image file exists
if [ ! -f "$IMAGE_PATH" ]; then
  echo "Error: Image file not found: $IMAGE_PATH"
  exit 1
fi

# API endpoint
API_URL="http://localhost:9876/api/profile/avatar"

# Upload the avatar
echo "Uploading avatar from $IMAGE_PATH..."
curl -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -F "avatar=@$IMAGE_PATH" \
  $API_URL

echo -e "\n\nDone!"