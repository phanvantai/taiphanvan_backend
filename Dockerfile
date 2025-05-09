FROM golang:1.24.3-alpine AS builder

# Install dependencies and apply security updates
RUN apk update && apk upgrade && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# No need to create .env files as the application now detects Docker environments
# and uses environment variables directly

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/api ./cmd/api

# Create a minimal production image
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our binary
COPY --from=builder /app/api /app/api

# No need to copy configs directory as we don't use .env files in Docker

# Use non-root user
USER appuser:appuser

# Expose API port
EXPOSE 9876

# Set the entrypoint
ENTRYPOINT ["/app/api"]