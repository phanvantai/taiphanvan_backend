# Personal Blog Backend API

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
│   └── testutil/      # Testing utilities
└── pkg/               # Reusable packages
    └── utils/         # Utility functions
```

## Features

- RESTful API endpoints for blog content management
- JWT-based authentication and authorization
- Database interactions with connection pooling using GORM
- Input validation and error handling
- Structured logging with zerolog
- API documentation with Swagger
- Security features (rate limiting, input sanitization, CORS support)
- Containerization with Docker

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
    # Database setup commands or reference to migration scripts
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

The application uses environment variables for configuration. Create a `.env` file in the project root or configure these environment variables in your deployment environment:

```bash
# Server
PORT=9876
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=personal_blog

# JWT
JWT_SECRET=your_jwt_secret
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=info

# Security
RATE_LIMIT=100
CORS_ALLOWED_ORIGINS=*
```

## API Documentation

API documentation is available via Swagger UI when the application is running:

```url
http://localhost:9876/swagger/index.html
```

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token

### Blog Posts

- `GET /api/posts` - Get all posts (with pagination)
- `GET /api/posts/:id` - Get a specific post
- `POST /api/posts` - Create a new post (requires auth)
- `PUT /api/posts/:id` - Update a post (requires auth)
- `DELETE /api/posts/:id` - Delete a post (requires auth)

### Comments

- `GET /api/posts/:id/comments` - Get comments for a post
- `POST /api/posts/:id/comments` - Add a comment (requires auth)
- `PUT /api/comments/:id` - Update a comment (requires auth)
- `DELETE /api/comments/:id` - Delete a comment (requires auth)

### Tags

- `GET /api/tags` - Get all tags
- `POST /api/tags` - Create a new tag (requires auth)
- `DELETE /api/tags/:id` - Delete a tag (requires auth)

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

When using Docker Compose, environment variables are already defined in the `docker-compose.yml` file:

```yaml
environment:
  - API_PORT=9876
  - DB_HOST=postgres  # Uses the service name as hostname
  - DB_USER=bloguser
  - DB_PASS=blogpassword
  - DB_NAME=blog_db
  # ... other configurations
```

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

## Security Considerations

- All endpoints requiring authentication are protected with JWT tokens
- Passwords are hashed using bcrypt with proper salting
- Input sanitization and validation using gin-validator
- Rate limiting is applied to public endpoints
- CORS protection implemented
- HTTPS is required for all communications in production
- Database queries use prepared statements to prevent SQL injection

## License

[MIT License](LICENSE)

## Contact

For questions or support, please contact [taipv.swe@gmail.com](mailto:taipv.swe@gmail.com)
