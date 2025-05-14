# Authentication & User Management

This feature provides a complete authentication system for the blog platform, including user registration, login, token management, and profile management.

## Database Models

### User

- `id` - Unique identifier
- `username` - Unique username
- `email` - Unique email address
- `password` - Hashed password (not returned in API responses)
- `first_name` - User's first name
- `last_name` - User's last name
- `bio` - User biography
- `role` - User role (admin, editor, user)
- `profile_image` - URL to profile image
- `created_at` - When the user account was created
- `updated_at` - When the user account was last updated

### BlacklistedToken

- `id` - Unique identifier
- `token` - The JWT token that has been blacklisted
- `expires_at` - When the token expires
- `created_at` - When the token was blacklisted

## API Endpoints

### Register

```bash
POST /api/auth/register
```

Creates a new user account.

#### Request Body

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "SecurePassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Response

```json
{
  "status": "success",
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com"
  },
  "message": "User registered successfully"
}
```

### Login

```bash
POST /api/auth/login
```

Authenticates a user and provides access and refresh tokens.

#### Login Request Body

```json
{
  "email": "john@example.com",
  "password": "SecurePassword123"
}
```

#### Login Response

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "role": "user"
    }
  },
  "message": "Login successful"
}
```

### Refresh Token

```bash
POST /api/auth/refresh
```

Refreshes an access token using a valid refresh token.

#### Refresh Token Request Body

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Refresh Token Response

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "message": "Token refreshed successfully"
}
```

### Revoke Token

```bash
POST /api/auth/revoke
```

Blacklists the current authentication token (requires authentication).

#### Revoke Token Response

```json
{
  "status": "success",
  "data": null,
  "message": "Token revoked successfully"
}
```

### Logout

```bash
POST /api/auth/logout
```

Logs the user out by blacklisting their current tokens (requires authentication).

#### Logout Response

```json
{
  "status": "success",
  "data": null,
  "message": "Logged out successfully"
}
```

### Get Profile

```bash
GET /api/profile
```

Retrieves the authenticated user's profile (requires authentication).

#### Get Profile Response

```json
{
  "status": "success",
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "I'm a software developer interested in web technologies.",
    "role": "user",
    "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-02T12:00:00Z"
  },
  "message": "Profile retrieved successfully"
}
```

### Update Profile

```bash
PUT /api/profile
```

Updates the authenticated user's profile (requires authentication).

#### Update Profile Request Body

```json
{
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Updated biography"
}
```

#### Update Profile Response

```json
{
  "status": "success",
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Updated biography",
    "role": "user",
    "profile_image": "https://res.cloudinary.com/demo/image/upload/v1234567890/avatars/user_1_1620000000.jpg",
    "created_at": "2023-01-01T12:00:00Z",
    "updated_at": "2023-01-02T12:00:00Z"
  },
  "message": "Profile updated successfully"
}
```

## Business Logic

- Passwords are hashed using bcrypt before being stored in the database
- Authentication uses JWT with separate access and refresh tokens
- Access tokens have a shorter lifespan (typically 1 hour)
- Refresh tokens have a longer lifespan (typically 7 days)
- Tokens can be revoked by adding them to the blacklist
- A background task periodically cleans up expired blacklisted tokens
- User roles control access to certain features (admin, editor, user)

## Security Features

- Rate limiting on authentication endpoints to prevent brute force attacks
- Token blacklisting for revocation
- Secure password hashing with bcrypt
- JWT-based stateless authentication
- Role-based access control

## Testing

You can test the authentication endpoints using the included test scripts:

```bash
# Test registration
./scripts/api_tests/test_register_api.sh

# Test login
./scripts/api_tests/test_login_api.sh

# Test logout
./scripts/api_tests/test_logout_api.sh

# Test profile management
./scripts/api_tests/test_profile_api.sh
```
