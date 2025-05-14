# Tags

This feature provides functionality for managing tags that categorize blog posts.

## Database Models

### Tag

- `id` - Unique identifier
- `name` - Tag name (unique)

### PostTags (Join Table)

- `post_id` - References Post model
- `tag_id` - References Tag model

## API Endpoints

### Get All Tags

```bash
GET /api/tags
```

Retrieves a list of all tags in the system.

#### Response

```json
{
  "status": "success",
  "data": [
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
  "message": "Tags retrieved successfully"
}
```

### Get Popular Tags

```bash
GET /api/tags/popular
```

Retrieves a list of the most popular tags based on usage count.

#### Query Parameters

- `limit` - Number of tags to return (default: 10)

#### Popular Tags Response

```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "name": "technology",
      "post_count": 42
    },
    {
      "id": 2,
      "name": "programming",
      "post_count": 35
    },
    {
      "id": 3,
      "name": "golang",
      "post_count": 28
    }
  ],
  "message": "Popular tags retrieved successfully"
}
```

## Business Logic

- Tags are automatically created when assigning them to posts if they don't already exist
- Tags are unique and case-sensitive
- Tags are used to categorize blog posts for easier navigation
- Tags are maintained in the many-to-many relationship with posts
- Popular tags are calculated based on the number of posts using each tag

## Usage in Post Endpoints

Tags are used in the following Post Management endpoints:

### Create Post

When creating a post, you can specify tags as an array of strings:

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

### Update Post

When updating a post, you can update the tags as an array of strings:

```json
{
  "title": "Updated Post Title",
  "content": "Updated content",
  "excerpt": "Updated excerpt",
  "tags": ["technology", "programming", "golang"]
}
```

This will replace the existing tags with the new tags. Any tags that don't exist in the system will be created automatically.
