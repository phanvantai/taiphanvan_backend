#!/bin/bash

# Script to migrate the database for the ViewCount field in Post model

echo "Running migration for ViewCount field in Post model..."

# Get the database connection details from environment variables or use defaults
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"postgres"}
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"taiphanvandb"}

# Check if DATABASE_URL exists
if [ -n "$DATABASE_URL" ]; then
  echo "Using DATABASE_URL environment variable"
  PSQL_CMD="psql $DATABASE_URL"
else
  echo "Using individual database connection parameters"
  PSQL_CMD="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
fi

# SQL to check if the view_count column exists
CHECK_COLUMN_SQL="SELECT column_name FROM information_schema.columns WHERE table_name='posts' AND column_name='view_count';"

# SQL to add the view_count column if it doesn't exist
ADD_COLUMN_SQL="
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='posts' AND column_name='view_count') THEN
        ALTER TABLE posts ADD COLUMN view_count INTEGER NOT NULL DEFAULT 0;
        RAISE NOTICE 'Added view_count column to posts table';
    ELSE
        RAISE NOTICE 'view_count column already exists in posts table';
    END IF;
END \$\$;
"

# Execute the SQL commands
echo "Checking for view_count column..."
if $PSQL_CMD -t -c "$CHECK_COLUMN_SQL" | grep -q "view_count"; then
  echo "Column view_count already exists in posts table"
else
  echo "Adding view_count column to posts table..."
  $PSQL_CMD -c "$ADD_COLUMN_SQL"
  
  # Check if the column was added successfully
  if $PSQL_CMD -t -c "$CHECK_COLUMN_SQL" | grep -q "view_count"; then
    echo "Successfully added view_count column to posts table"
  else
    echo "Failed to add view_count column to posts table"
    exit 1
  fi
fi

echo "Migration completed successfully"
exit 0
