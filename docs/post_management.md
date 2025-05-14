# Post Management

This feature provides complete functionality for managing blog posts, including creation, reading, updating, deletion, and status management.

## Database Models

### Post

- `id` - Unique identifier
- `title` - Post title
- `slug` - URL-friendly version of the title
- `content` - Main content of the post
- `excerpt` - Short summary or preview of the post
- `cover` - URL to the post's cover image
- `status` - Publication status (draft, published, archived, scheduled)
- `view_count` - Number of times the post has been viewed
- `user_id` - ID of the post author
- `created_at` - When the post was created
- `updated_at` - When the post was last updated

## API Endpoints

### Get Posts

```bash
GET /api/posts
```

Retrieves a list of published posts with optional pagination and filtering.

#### Query Parameters

- `page` - Page number for pagination (default: 1)
- `limit` - Number of posts per page (default: 10)
- `tag` - Filter posts by tag name
- `search` - Search posts by title or content

#### Response

```json
{
  "status": "success",
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "My First Blog Post",
        "slug": "my-first-blog-post",
        "excerpt": "A short summary of the post",
        "cover": "https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg",
        "status": "published",
        "view_count": 42,
        "user": {
          "id": 1,
          "username": "johndoe"
        },
        "tags": [
          {
            "id": 1,
            "name": "technology"
          }
        ],
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 42,
      "page_count": 5
    }
  },
  "message": "Posts retrieved successfully"
}
```

### Get Post by Slug

```bash
GET /api/posts/slug/:slug
```

Retrieves a single post by its slug.

#### Get Post by Slug Response

```json
{
  "status": "success",
  "data": {
    "id": 1,
    "title": "My First Blog Post",
    "slug": "my-first-blog-post",
    "content": "This is the content of my blog post...",
    "excerpt": "A short summary of the post",
    "cover": "https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg",
    "status": "published",
    "view_count": 42,
    "user": {
      "id": 1,
      "username": "johndoe",
      "first_name": "John",
      "last_name": "Doe",
      "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg"
    },
    "tags": [
      {
        "id": 1,
        "name": "technology"
      }
    ],
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-02T12:00:00Z"
  },
  "message": "Post retrieved successfully"
}
```

### Increment Post View Count

```bash
POST /api/posts/:id/view
```

Increments the view count for a post.

#### Increment Post View Count Response

```json
{
  "status": "success",
  "data": {
    "view_count": 43
  },
  "message": "View count incremented"
}
```

### Create Post

```bash
POST /api/posts
```

Creates a new blog post (requires authentication).

#### Request Body

```json
{
  "title": "My New Post",
  "content": "This is the content of my new post",
  "excerpt": "A short excerpt",
  "cover": "https://example.com/image.jpg",
  "tags": ["technology", "programming"],
  "status": "draft"
}
```

#### Create Post Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "title": "My New Post",
    "slug": "my-new-post",
    "content": "This is the content of my new post",
    "excerpt": "A short excerpt",
    "cover": "https://example.com/image.jpg",
    "status": "draft",
    "view_count": 0,
    "user_id": 1,
    "tags": [
      {
        "id": 1,
        "name": "technology"
      },
      {
        "id": 2,
        "name": "programming"
      }
    ],
    "created_at": "2023-01-03T12:00:00Z",
    "updated_at": "2023-01-03T12:00:00Z"
  },
  "message": "Post created successfully"
}
```

### Update Post

```bash
PUT /api/posts/:id
```

Updates an existing blog post (requires authentication and ownership or admin role).

#### Update Post Request Body

```json
{
  "title": "Updated Post Title",
  "content": "Updated content",
  "excerpt": "Updated excerpt",
  "tags": ["technology", "programming", "golang"]
}
```

#### Update Post Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "title": "Updated Post Title",
    "slug": "updated-post-title",
    "content": "Updated content",
    "excerpt": "Updated excerpt",
    "cover": "https://example.com/image.jpg",
    "status": "draft",
    "view_count": 0,
    "user_id": 1,
    "tags": [
      {
        "id": 1,
        "name": "technology"
      },
      {
        "id": 2,
        "name": "programming"
      },
      {
        "id": 3,
        "name": "golang"
      }
    ],
    "created_at": "2023-01-03T12:00:00Z",
    "updated_at": "2023-01-04T12:00:00Z"
  },
  "message": "Post updated successfully"
}
```

### Delete Post

```bash
DELETE /api/posts/:id
```

Deletes a blog post (requires authentication and ownership or admin role).

#### Delete Post Response

```json
{
  "status": "success",
  "data": null,
  "message": "Post deleted successfully"
}
```

### Get My Posts

```bash
GET /api/posts/me
```

Retrieves a list of the authenticated user's posts for the dashboard (requires authentication).

#### Get My Posts Query Parameters

- `page` - Page number for pagination (default: 1)
- `limit` - Number of posts per page (default: 10)
- `status` - Filter by post status

#### Get My Posts Response

```json
{
  "status": "success",
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "My First Blog Post",
        "slug": "my-first-blog-post",
        "excerpt": "A short summary of the post",
        "status": "published",
        "view_count": 42,
        "created_at": "2023-01-01T12:00:00Z",
        "updated_at": "2023-01-02T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 5,
      "page_count": 1
    }
  },
  "message": "Posts retrieved successfully"
}
```

### Upload Post Cover

```bash
POST /api/posts/:id/cover
```

Uploads a cover image for a post (requires authentication and ownership or admin role).

#### Upload Post Cover Request Body

- Multipart form data with `cover` file field

#### Upload Post Cover Response

```json
{
  "status": "success",
  "data": {
    "cover_url": "https://res.cloudinary.com/demo/image/upload/v1234567890/folder/post_1_1620000000.jpg"
  },
  "message": "Cover image uploaded successfully"
}
```

### Delete Post Cover

```bash
DELETE /api/posts/:id/cover
```

Removes the cover image from a post (requires authentication and ownership or admin role).

#### Delete Post Cover Response

```json
{
  "status": "success",
  "data": null,
  "message": "Cover image removed successfully"
}
```

### Publish Post

```bash
POST /api/posts/:id/publish
```

Changes a post's status to "published" (requires authentication and ownership or admin role).

#### Publish Post Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "status": "published",
    "updated_at": "2023-01-04T12:00:00Z"
  },
  "message": "Post published successfully"
}
```

### Unpublish Post

```bash
POST /api/posts/:id/unpublish
```

Changes a post's status to "draft" (requires authentication and ownership or admin role).

#### Unpublish Post Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "status": "draft",
    "updated_at": "2023-01-04T12:00:00Z"
  },
  "message": "Post unpublished successfully"
}
```

### Set Post Status

```bash
POST /api/posts/:id/status
```

Sets a post's status explicitly (requires authentication and ownership or admin role).

#### Set Post Status Request Body

```json
{
  "status": "archived"
}
```

#### Set Post Status Response

```json
{
  "status": "success",
  "data": {
    "id": 2,
    "status": "archived",
    "updated_at": "2023-01-04T12:00:00Z"
  },
  "message": "Post status updated successfully"
}
```

## Business Logic

- Posts can have four statuses: draft, published, archived, or scheduled
- Post slugs are automatically generated from titles and must be unique
- Only published posts are publicly visible
- Authors can see all their own posts regardless of status
- View counts are tracked for analytics
- Posts can be tagged for categorization
- Cover images are stored in Cloudinary

## Migration

Run the view count migration script to update existing posts:

```bash
./scripts/migrate_view_count.sh
```
