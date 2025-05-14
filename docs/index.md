# TaiPhanVan Blog Backend Documentation

This documentation provides detailed information about the features and functionality of the TaiPhanVan Blog Backend API.

## Features

The backend API includes the following core features:

1. [Authentication & User Management](authentication.md)
   - User registration and login
   - JWT-based authentication
   - Profile management
   - Token refreshing and revocation

2. [Post Management](post_management.md)
   - Create, read, update, delete blog posts
   - Post status management (draft, published, archived, scheduled)
   - View count tracking
   - Post cover image management

3. [Comment System](comments.md)
   - Create, read, update, delete comments
   - Comment threading and organization

4. [Comment Voting](comment_votes.md)
   - Upvote and downvote functionality
   - Vote tracking and counting

5. [Tags](tags.md)
   - Tag management for posts
   - Popular tags retrieval

6. [File Upload System](file_upload.md)
   - File uploads for the editor
   - Cloudinary integration for image storage
   - Avatar and cover image management

7. [Security Features](security.md)
   - Rate limiting
   - Token management
   - CORS protection
   - Input validation
   - Password security
   - Role-based access control

8. [Health Check](health_check.md)
   - Application health monitoring
   - Dependency status checking

## API Overview

The API follows RESTful conventions and uses JSON for request and response bodies. All endpoints are prefixed with `/api`.

### Authentication

Most endpoints require authentication using a JWT token. Include the token in the `Authorization` header:

```bash
Authorization: Bearer your.jwt.token
```

### Response Format

All API responses follow a consistent format:

```json
{
  "status": "success",
  "data": {
    // Response data here
  },
  "message": "Operation successful"
}
```

Or for errors:

```json
{
  "status": "error",
  "error": "Error type",
  "message": "Detailed error message"
}
```

### HTTP Status Codes

The API uses standard HTTP status codes:

- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Testing

The project includes a suite of test scripts located in the `scripts/api_tests` directory. These scripts can be used to test individual API endpoints.

## Deployment

The project includes Docker and Docker Compose configurations for easy deployment. See the [main README](../README.md) for deployment instructions.
