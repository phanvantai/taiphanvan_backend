# File Upload System

This feature enables users to upload and manage files, particularly images, for use in blog posts and user profiles. It integrates with Cloudinary for cloud storage.

## Cloudinary Integration

The system uses Cloudinary as a cloud storage provider for uploaded files, which offers:

- Automatic image optimization
- CDN delivery
- Image transformations
- Secure storage

## API Endpoints

### Upload File

```bash
POST /api/files/upload
```

Uploads a file to Cloudinary for use in the editor (requires authentication).

#### Request Body

- Multipart form data with `file` field

#### Response

```json
{
  "status": "success",
  "data": {
    "file_url": "https://res.cloudinary.com/demo/image/upload/v1234567890/folder/image_1_1620000000.jpg"
  },
  "message": "File uploaded successfully"
}
```

### Delete File

```bash
POST /api/files/delete
```

Deletes a file from Cloudinary (requires authentication).

#### Request Body

```json
{
  "file_url": "https://res.cloudinary.com/demo/image/upload/v1234567890/folder/image_1_1620000000.jpg"
}
```

#### Response

```json
{
  "status": "success",
  "data": null,
  "message": "File deleted successfully"
}
```

### Upload Avatar

```bash
POST /api/profile/avatar
```

Uploads a user profile avatar (requires authentication).

#### Request Body

- Multipart form data with `avatar` file field

#### Response

```json
{
  "status": "success",
  "data": {
    "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg"
  },
  "message": "Avatar uploaded successfully"
}
```

### Upload Post Cover

```bash
POST /api/posts/:id/cover
```

Uploads a cover image for a post (requires authentication and ownership or admin role).

#### Request Body

- Multipart form data with `cover` file field

#### Response

```json
{
  "status": "success",
  "data": {
    "cover": "https://res.cloudinary.com/demo/image/upload/v1234567890/covers/post_1_1620000000.jpg"
  },
  "message": "Cover uploaded successfully"
}
```

## File Validation

### File Size Limits

- General files: Maximum 5MB
- Avatar images: Maximum 2MB
- Post cover images: Maximum 5MB

### Allowed File Types

- General files: JPG, JPEG, PNG, WEBP, GIF, SVG, PDF
- Avatar images: JPG, JPEG, PNG
- Post cover images: JPG, JPEG, PNG, WEBP

## Business Logic

- File uploads are validated for allowed file types (typically images)
- File size is limited to prevent abuse
- Files are stored in Cloudinary with organized folder structures:
  - User avatars in `/avatars/user_{user_id}_{timestamp}`
  - Post covers in `/covers/post_{post_id}_{timestamp}`
  - General uploads in `/uploads/{user_id}/{timestamp}_{filename}`
- Old images are automatically deleted when replaced
- URLs are stored in the database, not the actual files
- Secure URLs (HTTPS) are used for all file access

## Configuration

The Cloudinary service requires the following environment variables:

```bash
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
CLOUDINARY_UPLOAD_FOLDER=folder_name
```

## Testing

You can test the file upload endpoints using the included test scripts:

```bash
# Test avatar upload
./scripts/api_tests/test_avatar_upload.sh

# Test general file upload
./scripts/api_tests/test_file_upload.sh

# Test file deletion
./scripts/api_tests/test_file_delete.sh
```
