#!/bin/bash
# Database backup script for Personal Blog Backend
set -e

# Configuration
APP_NAME="personal-blog-backend"
BACKUP_DIR="/opt/backups/$APP_NAME"
POSTGRES_CONTAINER="blog_postgres"
DB_NAME="blog_db"
DB_USER="bloguser"
RETENTION_DAYS=7

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

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  print_error "Docker is not installed. Please install Docker first."
  exit 1
fi

# Create backup directory if it doesn't exist
if [ ! -d "$BACKUP_DIR" ]; then
  print_message "Creating backup directory: $BACKUP_DIR"
  mkdir -p "$BACKUP_DIR"
fi

# Get current date and time for backup filename
DATE=$(date +"%Y-%m-%d_%H-%M-%S")
BACKUP_FILE="$BACKUP_DIR/${APP_NAME}_${DATE}.sql.gz"

# Check if the Postgres container is running
if ! docker ps | grep -q $POSTGRES_CONTAINER; then
  print_error "Postgres container ($POSTGRES_CONTAINER) is not running."
  exit 1
fi

# Perform the backup
print_message "Creating backup: $BACKUP_FILE"
docker exec $POSTGRES_CONTAINER pg_dump -U $DB_USER $DB_NAME | gzip > "$BACKUP_FILE"

# Check if the backup was successful
if [ $? -eq 0 ]; then
  print_message "Backup completed successfully."
  
  # Set proper permissions
  chmod 600 "$BACKUP_FILE"
  
  # Print backup info
  BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
  print_message "Backup size: $BACKUP_SIZE"
  
  # Delete old backups
  print_message "Cleaning up backups older than $RETENTION_DAYS days..."
  find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete
else
  print_error "Backup failed."
  exit 1
fi

# List available backups
print_message "Available backups:"
ls -lh "$BACKUP_DIR" | grep .sql.gz