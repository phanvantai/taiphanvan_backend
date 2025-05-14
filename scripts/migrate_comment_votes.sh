#!/bin/bash

# Script to migrate the database for the Comment upvote/downvote feature

echo "Running migration for upvote functionality in Comment model..."

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

# SQL to check if the upvote_count column exists in comments table
CHECK_COLUMN_SQL="SELECT column_name FROM information_schema.columns WHERE table_name='comments' AND column_name='upvote_count';"

# SQL to add the upvote_count column if it doesn't exist
ADD_COLUMN_SQL="
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name='comments' AND column_name='upvote_count') THEN
        ALTER TABLE comments ADD COLUMN upvote_count INTEGER NOT NULL DEFAULT 0;
        RAISE NOTICE 'Added upvote_count column to comments table';
    ELSE
        RAISE NOTICE 'upvote_count column already exists in comments table';
    END IF;
END \$\$;
"

# SQL to check if the comment_votes table exists
CHECK_TABLE_SQL="SELECT tablename FROM pg_tables WHERE tablename='comment_votes';"

# SQL to create the comment_votes table if it doesn't exist
CREATE_TABLE_SQL="
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename='comment_votes') THEN
        CREATE TABLE comment_votes (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            comment_id INTEGER NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
            vote_type SMALLINT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(user_id, comment_id)
        );
        CREATE INDEX comment_votes_comment_id_idx ON comment_votes(comment_id);
        CREATE INDEX comment_votes_user_id_idx ON comment_votes(user_id);
        RAISE NOTICE 'Created comment_votes table';
    ELSE
        RAISE NOTICE 'comment_votes table already exists';
    END IF;
END \$\$;
"

# Execute the SQL queries
echo "Checking for upvote_count column in comments table..."
$PSQL_CMD -c "$CHECK_COLUMN_SQL" -t | grep -q upvote_count

if [ $? -eq 0 ]; then
  echo "upvote_count column already exists in comments table"
else
  echo "Adding upvote_count column to comments table..."
  $PSQL_CMD -c "$ADD_COLUMN_SQL"
  
  if [ $? -eq 0 ]; then
    echo "Successfully added upvote_count column to comments table"
  else
    echo "Failed to add upvote_count column to comments table"
    exit 1
  fi
fi

echo "Checking for comment_votes table..."
$PSQL_CMD -c "$CHECK_TABLE_SQL" -t | grep -q comment_votes

if [ $? -eq 0 ]; then
  echo "comment_votes table already exists"
else
  echo "Creating comment_votes table..."
  $PSQL_CMD -c "$CREATE_TABLE_SQL"
  
  if [ $? -eq 0 ]; then
    echo "Successfully created comment_votes table"
  else
    echo "Failed to create comment_votes table"
    exit 1
  fi
fi

echo "Migration completed successfully"
