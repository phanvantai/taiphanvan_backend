# Security Features

This document outlines the security features implemented in the TaiPhanVan Blog Backend to protect user data, prevent abuse, and ensure secure operations.

## Rate Limiting

The application implements IP-based rate limiting to prevent abuse and protect against brute force attacks.

### Implementation

- Default rate limit: 100 requests per minute per IP address
- Stricter rate limits for authentication endpoints: 20 requests per minute per IP
- Background cleanup task to prevent memory leaks

### Rate Limited Endpoints

All API endpoints are protected by rate limiting, with stricter limits applied to sensitive endpoints like:

- `/api/auth/register`
- `/api/auth/login`
- `/api/auth/refresh`

### Rate Limit Response

When a client exceeds the rate limit, they receive a 429 Too Many Requests response:

```json
{
  "status": "error",
  "error": "Rate limit exceeded",
  "message": "Too many requests, please try again later"
}
```

## JWT Token Management

The application uses JSON Web Tokens (JWT) for authentication, with several security enhancements.

### Token Types

- **Access Tokens**: Short-lived tokens (typically 1 hour) used for API authentication
- **Refresh Tokens**: Longer-lived tokens (typically 7 days) used only to obtain new access tokens

### Token Security Measures

- Token signing with a secure secret key
- Token expiration times
- Token type verification to prevent misuse
- Token blacklisting for revocation

### Token Blacklisting

When a user logs out or explicitly revokes a token, the token is added to a blacklist in the database:

- The `blacklisted_tokens` table stores revoked tokens until they expire
- A background task periodically cleans up expired blacklisted tokens
- All authentication requests check against the blacklist

## CORS Protection

The application implements Cross-Origin Resource Sharing (CORS) protection to prevent unauthorized cross-origin requests.

### CORS Configuration

- Explicitly whitelisted origins
- Controlled HTTP methods
- Limited exposed headers
- Configurable per environment (development/production)

### Production CORS Settings

In production, only these origins are allowed:

- `https://api.taiphanvan.dev`
- `https://taiphanvan.dev`
- Any additional domains specified in configuration

### Development CORS Settings

In development mode, localhost is also allowed:

- `http://localhost:*`

## Input Validation

All user inputs are validated before processing:

- JSON request bodies are validated against defined schemas
- Path and query parameters are sanitized
- File uploads are validated for type and size

## Password Security

User passwords are secured using industry best practices:

- Passwords are hashed using bcrypt with appropriate cost factors
- Passwords are never stored in plain text
- Passwords are never returned in API responses

## HTTP Security Headers

The application sets appropriate security headers in responses:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Content-Security-Policy` with appropriate directives
- `X-XSS-Protection: 1; mode=block`

## TLS Support

The application supports TLS for secure HTTPS communication:

- Configurable TLS certificate and key paths
- Automatic HTTP to HTTPS redirection
- TLS version and cipher suite restrictions

## Request Tracing

For debugging and audit purposes, each request is assigned a unique ID:

- Request IDs are generated using UUIDs
- Request IDs are included in log entries
- Request IDs are returned in response headers as `X-Request-ID`
- Existing request IDs from load balancers or API gateways are preserved

## Role-Based Access Control

The application implements role-based access control (RBAC) to restrict access to certain endpoints:

- User roles include `admin`, `editor`, and `user`
- Certain operations (like deleting other users' content) are restricted to admins
- Middleware functions verify roles before allowing access to protected endpoints
