#!/bin/bash
# Production deployment script for Personal Blog Backend
set -e

# Configuration
APP_NAME="taiphanvan_backend"
DEPLOY_DIR="/opt/taiphanvan_backend"
BACKUP_DIR="/opt/backups/taiphanvan_backend"
GIT_REPO="https://github.com/phanvantai/taiphanvan_backend.git"
BRANCH="main"

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

# Check if docker and docker-compose are installed
if ! [ -x "$(command -v docker)" ]; then
  print_error "Docker is not installed. Please install Docker first."
  exit 1
fi

if ! [ -x "$(command -v docker-compose)" ]; then
  print_error "Docker Compose is not installed. Please install Docker Compose first."
  exit 1
fi

# Create deployment directory if it doesn't exist
if [ ! -d "$DEPLOY_DIR" ]; then
  print_message "Creating deployment directory: $DEPLOY_DIR"
  mkdir -p "$DEPLOY_DIR"
fi

# Navigate to the deployment directory
cd "$DEPLOY_DIR"

# Check if this is first deployment or update
if [ ! -d ".git" ]; then
  print_message "First deployment - cloning repository"
  git clone -b "$BRANCH" "$GIT_REPO" .
else
  print_message "Updating existing deployment"
  git fetch
  git reset --hard origin/"$BRANCH"
fi

# Check if .env file exists
if [ ! -f "configs/.env" ]; then
  print_warning "No .env file found. Creating from example."
  cp configs/.env.example configs/.env
  print_warning "Please edit configs/.env with your production settings!"
  exit 1
fi

# Set JWT secret if not already set
if ! grep -q "JWT_SECRET=" configs/.env || grep -q "JWT_SECRET=$" configs/.env; then
  print_message "Setting JWT_SECRET environment variable"
  JWT_SECRET=$(openssl rand -base64 32)
  sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" configs/.env
fi

# Make sure the script is executable
chmod +x scripts/backup.sh

# Build and start containers
print_message "Building and starting containers"
docker-compose -f docker-compose.yml build
docker-compose -f docker-compose.yml up -d

# Check if containers are running
sleep 5
if [ "$(docker ps -q -f name=${APP_NAME})" ]; then
  print_message "Deployment successful! Application is running."
  echo -e "${GREEN}---------------------------------------------${NC}"
  echo -e "API is available at: http://localhost:9876"
  echo -e "Health check endpoint: http://localhost:9876/health"
  echo -e "${GREEN}---------------------------------------------${NC}"
else
  print_error "Deployment failed. Containers are not running."
  docker-compose -f docker-compose.yml logs
  exit 1
fi