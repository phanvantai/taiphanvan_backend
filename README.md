# TaiPhanVan Blog Backend API

This repository contains the backend API service for a personal blog built with Go. The API provides endpoints for blog post management, user authentication, comments, and tags.

## Project Structure

```bash
taiphanvan_backend/
├── cmd/api/           # Application entrypoint and API documentation
├── configs/           # Configuration files
├── docs/              # Swagger documentation
├── internal/          # Private application code
│   ├── config/        # Application configuration
│   ├── database/      # Database connection and management
│   ├── handlers/      # HTTP request handlers
│   ├── logger/        # Logging configuration
│   ├── middleware/    # HTTP middleware components
│   ├── models/        # Data models and business logic
│   ├── services/      # External service integrations
│   └── testutil/      # Testing utilities
└── pkg/               # Reusable packages
    └── utils/         # Utility functions
```

## Features

- RESTful API endpoints for blog content management
- JWT-based authentication with access and refresh tokens
- Database interactions with connection pooling using GORM
- Input validation and error handling
- Structured logging with zerolog
- API documentation with Swagger
- Security features (rate limiting, input sanitization, CORS support)
- Cloudinary integration for image uploads
- News integration with external API providers
- Automatic news fetching and categorization
- Containerization with Docker
- Support for multiple deployment environments (local, Docker, Railway)

## Requirements

- Go 1.24+
- PostgreSQL
- Environment configuration (see Configuration section)

## Getting Started

### Setup

1. Clone the repository:

    ```bash
    git clone https://github.com/phanvantai/taiphanvan_backend.git
    cd taiphanvan_backend
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

3. Configure the environment variables (see Configuration section)

4. Set up the database:

    ```bash
    # The application will automatically create tables on first run
    # For local development, you can use Docker Compose to set up PostgreSQL
    docker-compose up -d postgres
    ```

### Running the Server

```bash
# Run in development mode
go run cmd/api/main.go

# Or build and run
go build -o blog-api cmd/api/main.go
./blog-api
```

The server will start on `http://localhost:9876` by default (configurable in config file).

## Configuration

The application uses environment variables for configuration. For local development, create a `.env` file in the project root directory (not in any subdirectory). You can use the provided `.env.example` as a template:

> **Note:** The `.env` file must be placed in the root directory of the project. When running in Docker or cloud environments (like Railway), the application will automatically use environment variables directly and ignore the `.env` file.

### Environment Detection

The application includes an intelligent environment detection system that automatically:

1. Detects when it's running in a Docker container
2. Detects when it's running on Railway.app
3. Applies appropriate configuration defaults based on the environment
4. Skips loading the `.env` file in containerized environments

This makes deployment seamless across different environments without requiring manual configuration changes.

```bash
# API Configuration
API_PORT=9876
GIN_MODE=debug # Use 'release' for production

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=bloguser
DB_PASS=your_secure_password_here
DB_NAME=blog_db
DB_SSL_MODE=disable # Use 'require' for production

# JWT Configuration
JWT_SECRET=replace_with_secure_random_string
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Logging Configuration
LOG_LEVEL=debug # Use 'info' for production
LOG_FORMAT=console # Use 'json' for production

# Admin User Configuration
CREATE_DEFAULT_ADMIN=true
DEFAULT_ADMIN_USERNAME=admin
DEFAULT_ADMIN_EMAIL=admin@example.com
DEFAULT_ADMIN_PASSWORD=replace_with_secure_password

# PostgreSQL Configuration
POSTGRES_USER=bloguser
POSTGRES_PASSWORD=your_secure_password_here
POSTGRES_DB=blog_db

# Cloudinary Configuration
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
CLOUDINARY_UPLOAD_FOLDER=blog_images

# NewsAPI Configuration
NEWS_API_KEY=your_newsapi_key
NEWS_API_BASE_URL=https://newsapi.org/v2
NEWS_API_DEFAULT_LIMIT=10
NEWS_API_FETCH_INTERVAL=1h
NEWS_API_ENABLE_AUTO_FETCH=false
```

## API Documentation

API documentation is available via Swagger UI when the application is running:

```url
http://localhost:9876/swagger/index.html
```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT tokens
- `POST /api/auth/refresh` - Refresh JWT token
- `POST /api/auth/revoke` - Revoke a refresh token (requires auth)
- `POST /api/auth/logout` - Logout and invalidate tokens (requires auth)

### User Profile

- `GET /api/profile` - Get user profile (requires auth)
- `PUT /api/profile` - Update user profile (requires auth)
- `POST /api/profile/avatar` - Upload user avatar using Cloudinary (requires auth)

### Blog Posts

- `GET /api/posts` - Get all posts (with pagination, tag filtering, and status filtering)
- `GET /api/posts/slug/:slug` - Get a specific post by slug
- `GET /api/posts/me` - Get the current user's posts (requires auth)
- `POST /api/posts` - Create a new post (requires auth)
- `PUT /api/posts/:id` - Update a post (requires auth)
- `DELETE /api/posts/:id` - Delete a post (requires auth)
- `POST /api/posts/:id/cover` - Upload post cover image (requires auth)
- `DELETE /api/posts/:id/cover` - Delete post cover image (requires auth)
- `POST /api/posts/:id/publish` - Publish a post (requires auth)
- `POST /api/posts/:id/unpublish` - Unpublish a post (requires auth)
- `POST /api/posts/:id/status` - Change post status (requires auth)

### Comments

- `GET /api/posts/:id/comments` - Get comments for a post
- `POST /api/posts/:id/comments` - Add a comment (requires auth)
- `PUT /api/comments/:commentID` - Update a comment (requires auth)
- `DELETE /api/comments/:commentID` - Delete a comment (requires auth)

### Tags

- `GET /api/tags` - Get all tags
- `GET /api/tags/popular` - Get popular tags

### Health Check

- `GET /health` - Check API health status

## Post Status Feature

The blog platform supports a comprehensive post status system that allows for flexible content management:

### Available Post Statuses

- `draft` - Unpublished posts that are still being edited or reviewed
- `published` - Live posts that are publicly visible to all users
- `archived` - Previously published posts that are now archived and no longer actively displayed
- `scheduled` - Posts scheduled to be published at a future date

### Managing Post Status

Post status can be managed through several endpoints:

- When creating a post (`POST /api/posts`), you can set the initial status
- Update the status when editing a post (`PUT /api/posts/:id`)
- Use the dedicated status endpoint (`POST /api/posts/:id/status`) for status-specific updates
- Use convenience endpoints for common transitions:
  - `POST /api/posts/:id/publish` to quickly publish a post
  - `POST /api/posts/:id/unpublish` to quickly unpublish a post

### Filtering Posts by Status

When retrieving posts, you can filter by status:

```bash
GET /api/posts?status=published
GET /api/posts?status=draft
```

By default, the public posts endpoint only returns published posts.

### Scheduled Posts

For scheduled posts, you must provide a future publication date in the `publish_at` field. The system will validate that the date is in the future.

Example request body for scheduling a post:

```json
{
  "status": "scheduled",
  "publish_at": "2025-06-01T12:00:00Z"
}
```

## Development

### Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

### API Testing Scripts

The project includes shell scripts for testing API endpoints:

```bash
# Run specific API test
./scripts/api_tests/test_login_api.sh

# Or run all API tests
./scripts/api_tests/run_all_tests.sh
```

### Code Formatting

Format your code before committing:

```bash
go fmt ./...
```

### Linting

Run linting checks:

```bash
# Requires golangci-lint to be installed
golangci-lint run
```

## Deployment

The application can be deployed as a standalone binary or containerized with Docker.

### Docker Deployment

#### Prerequisites

- Docker and Docker Compose installed on your machine
- Basic understanding of Docker concepts

#### Running Locally with Docker

1. **Build and run using Docker directly:**

   ```bash
   # Build the Docker image
   docker build -t taiphanvan-api .

   # Run the container (replace .env with your actual env file if needed)
   docker run -p 9876:9876 --env-file .env taiphanvan-api
   ```

2. **Using Docker Compose (recommended for local development):**

   The project includes a `docker-compose.yml` file that sets up both the API service and a PostgreSQL database:

   ```bash
   # Start all services in the background
   docker-compose up -d

   # To see logs in real-time
   docker-compose logs -f

   # To see logs for a specific service
   docker-compose logs -f api
   docker-compose logs -f postgres
   ```

   This will start:
   - The API service on port 9876
   - PostgreSQL database on port 5433 (mapped from 5432 inside the container)

#### Environment Variables in Docker

When using Docker Compose, environment variables are passed from your local environment or .env file to the containers as defined in the `docker-compose.yml` file:

```yaml
environment:
  - API_PORT=${API_PORT}
  - GIN_MODE=${GIN_MODE}
  - DB_HOST=${DB_HOST}
  - DB_PORT=${DB_PORT}
  - DB_USER=${DB_USER}
  - DB_PASS=${DB_PASS}
  - DB_NAME=${DB_NAME}
  # ... other configurations
```

The application is designed to detect when it's running in a Docker container and will automatically:

1. Skip loading the `.env` file (using environment variables directly)
2. Apply sensible defaults for database connection (using `postgres` as the host)
3. Set appropriate fallback values for missing configuration

#### Managing Docker Containers

```bash
# Check running containers
docker-compose ps

# Stop all containers
docker-compose down

# Stop and remove volumes (will delete database data)
docker-compose down -v

# Rebuild containers after code changes
docker-compose up -d --build
```

#### Accessing the Application

- API: <http://localhost:9876>
- Swagger documentation: <http://localhost:9876/swagger/index.html>
- Database is accessible on port 5433 with credentials from docker-compose.yml

#### Troubleshooting Docker Setup

1. **Check container status:**

   ```bash
   docker-compose ps
   ```

2. **View detailed logs:**

   ```bash
   docker-compose logs -f
   ```

3. **Verify database connection:**

   ```bash
   # Connect to the database container
   docker exec -it blog_postgres psql -U bloguser -d blog_db
   ```

4. **Check API health endpoint:**

   ```bash
   curl http://localhost:9876/health
   ```

5. **Restart services:**

   ```bash
   docker-compose restart
   ```

### Cloud Deployment

#### Railway Deployment

The application is configured to run on Railway.app with minimal configuration:

1. **Connect your GitHub repository to Railway**

2. **Set up environment variables in Railway**
   - Use the provided `railway.env.example` file as a reference
   - Railway automatically provides `PORT` and `DATABASE_URL` environment variables
   - Make sure to set at least the following variables:
     - `JWT_SECRET` (important for security)
     - `GIN_MODE=release` (for production)
     - `CORS_ALLOWED_ORIGINS` (your frontend domain)
     - Cloudinary credentials if you're using image uploads

3. **No need for a .env file**
   - The application is designed to run without a `.env` file in Railway
   - All configuration is done through Railway's environment variables
   - The Dockerfile has been modified to handle the absence of a `.env` file

4. **Deploy your application**
   - Railway will automatically build and deploy your application
   - The application will detect it's running on Railway and apply appropriate settings

5. **Verify deployment**
   - Check the deployment logs for any issues
   - Test the API endpoints using the provided URL
   - Verify the health endpoint: `https://your-railway-url.up.railway.app/health`

The application automatically detects when it's running in a cloud environment and will use environment variables provided by the platform. For Railway, it will:

1. Detect the Railway environment automatically
2. Use the `PORT` environment variable for the server port
3. Use the `DATABASE_URL` environment variable for database connection
4. Generate a secure JWT secret if none is provided
5. Apply appropriate production settings

## Security Considerations

- All endpoints requiring authentication are protected with JWT tokens
- Separate access and refresh token mechanism for better security
- Passwords are hashed using bcrypt with proper salting
- Input sanitization and validation using gin-validator
- Rate limiting is applied to all API endpoints (stricter limits for auth endpoints)
- CORS protection with configurable allowed origins
- HTTPS is required for all communications in production
- Database queries use prepared statements to prevent SQL injection
- Request ID tracking for better debugging and audit trails

## License

[MIT License](LICENSE)

## Contact

For questions or support, please contact [taipv.swe@gmail.com](mailto:taipv.swe@gmail.com)
