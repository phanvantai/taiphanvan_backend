# Comment Upvote/Downvote Feature

This feature implements Reddit-style upvoting and downvoting for comments on blog posts.

## Database Changes

1. Added `upvote_count` column to the `comments` table
2. Created a new `comment_votes` table to track user votes

## New API Endpoints

### Get Comment Votes

```bash
GET /api/comments/{commentID}/votes
```

Returns the vote counts for a specific comment, as well as the current user's vote if authenticated.

### Vote on a Comment

```bash
POST /api/comments/{commentID}/vote
```

Allows an authenticated user to upvote, downvote, or remove their vote from a comment.

#### Request Body

```json
{
  "vote_type": 1  // 1 for upvote, -1 for downvote, 0 to remove vote
}
```

#### Response

```json
{
  "comment_id": 123,
  "upvote_count": 42,
  "user_vote": 1
}
```

## Updates to Existing Endpoints

### Get Comments by Post ID

```bash
GET /api/posts/{postID}/comments
```

Now includes upvote counts and the current user's vote for each comment.

## Models

### CommentVote

Tracks a user's vote on a comment:

- `id` - Unique identifier
- `user_id` - ID of the user who voted
- `comment_id` - ID of the comment that was voted on
- `vote_type` - Type of vote (1 for upvote, -1 for downvote, 0 for none)
- `created_at` - When the vote was created
- `updated_at` - When the vote was last updated

## Business Logic

- Users can upvote, downvote, or remove their vote from a comment
- A user can only have one vote per comment
- Upvotes increase the comment's upvote count by 1
- Downvotes decrease the comment's upvote count by 1
- Removing a vote reverses the effect of the user's previous vote
- Changing from an upvote to a downvote decreases the upvote count by 2, and vice versa

## Migration

Run the migration script to update the database schema:

```bash
./scripts/migrate_comment_votes.sh
```
