# Comment System

This feature allows users to comment on blog posts, creating interactive discussions.

## Database Models

### Comment

- `id` - Unique identifier
- `content` - Comment content
- `user_id` - ID of the comment author
- `post_id` - ID of the post being commented on
- `upvote_count` - Number of upvotes on the comment
- `created_at` - When the comment was created
- `updated_at` - When the comment was last updated

## API Endpoints

### Get Comments by Post ID

```bash
GET /api/posts/:id/comments
```

Retrieves all comments for a specific post, ordered by creation date (newest first).

#### Response

```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "content": "Great post!",
      "user": {
        "id": 2,
        "username": "janedoe",
        "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_2_1620000000.jpg"
      },
      "upvote_count": 5,
      "user_vote": 1,
      "created_at": "2023-01-02T12:00:00Z",
      "updated_at": "2023-01-02T12:00:00Z"
    }
  ],
  "message": "Comments retrieved successfully"
}
```

### Create Comment

```bash
POST /api/posts/:id/comments
```

Creates a new comment on a post (requires authentication).

#### Request Body

```json
{
  "content": "This is my comment on this post."
}
```

#### Create Comment Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "content": "This is my comment on this post.",
    "user": {
      "id": 1,
      "username": "johndoe",
      "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg"
    },
    "upvote_count": 0,
    "user_vote": 0,
    "created_at": "2023-01-03T12:00:00Z",
    "updated_at": "2023-01-03T12:00:00Z"
  },
  "message": "Comment created successfully"
}
```

### Update Comment

```bash
PUT /api/comments/:commentID
```

Updates an existing comment (requires authentication and ownership).

#### Update Comment Request Body

```json
{
  "content": "Updated comment content."
}
```

#### Update Comment Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "content": "Updated comment content.",
    "user": {
      "id": 1,
      "username": "johndoe",
      "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg"
    },
    "upvote_count": 0,
    "user_vote": 0,
    "created_at": "2023-01-03T12:00:00Z",
    "updated_at": "2023-01-03T12:15:00Z"
  },
  "message": "Comment updated successfully"
}
```

### Delete Comment

```bash
DELETE /api/comments/:commentID
```

Deletes a comment (requires authentication and ownership or admin role).

#### Delete Comment Response

```json
{
  "status": "success",
  "data": null,
  "message": "Comment deleted successfully"
}
```

## Business Logic

- Users must be authenticated to create, update, or delete comments
- Users can only update their own comments (admins can update any comment)
- Users can delete their own comments, admins can delete any comment, and post authors can delete any comments on their posts
- Comments are associated with both a user and a post
- Comments support upvoting and downvoting (see [Comment Votes](comment_votes.md) for details)
- Comments are soft-deleted to maintain referential integrity
- Comments are displayed in chronological order (newest first)

## Related Features

The comment system also includes a voting system that allows users to upvote or downvote comments. For details on this feature, see the [Comment Votes documentation](comment_votes.md).
