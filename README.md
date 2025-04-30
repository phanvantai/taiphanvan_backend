# Personal Blog Backend API

This repository contains the backend API service for a personal blog built with Go. The API provides endpoints for blog post management, user authentication, comments, and tags.

## Project Structure

```bash
personal_blog_backend/
├── cmd/api/           # Application entrypoint
├── configs/           # Configuration files
├── internal/          # Private application code
│   ├── database/      # Database connection and management
│   ├── handlers/      # HTTP request handlers
│   ├── middleware/    # HTTP middleware components
│   └── models/        # Data models and business logic
└── pkg/               # Reusable packages
    └── utils/         # Utility functions
```

## Features

- RESTful API endpoints for blog content management
- JWT-based authentication and authorization
- Database interactions with connection pooling
- Input validation and error handling
- Structured logging
- Security features (rate limiting, input sanitization)

## Requirements

- Go 1.20+
- PostgreSQL
- Environment configuration (see Configuration section)

## Getting Started

### Setup

1. Clone the repository:

    ```bash
    git clone https://github.com/yourusername/personal_blog_backend.git
    cd personal_blog_backend
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

The server will start on `http://localhost:8080` by default (configurable in config file).

## Configuration

The application uses environment variables for configuration. Create a `.env` file in the project root or configure these environment variables in your deployment environment:

```bash
# Server
PORT=8080
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

```bash
# Build the Docker image
docker build -t personal-blog-api .

# Run the container
docker run -p 8080:8080 --env-file .env personal-blog-api
```

## Security Considerations

- All endpoints requiring authentication are protected with JWT tokens
- Passwords are hashed using bcrypt
- Input sanitization is implemented for all user inputs
- Rate limiting is applied to public endpoints
- HTTPS is recommended for all communications in production

## License

[MIT License](LICENSE)

## Contact

For questions or support, please contact [taipv.swe@gmail.com](mailto:taipv.swe@gmail.com)
