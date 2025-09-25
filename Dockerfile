# ------------------------
# Build stage
# ------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files and download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application (static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/app

# ------------------------
# Final stage
# ------------------------
FROM alpine:latest

WORKDIR /app

# Install required packages: bash for scripts and PostgreSQL client for migrations/testing
RUN apk add --no-cache bash postgresql-client

# Copy the built Go binary from the builder stage
COPY --from=builder /app/main /app/main

# Copy the wait-for-it.sh script to wait for Postgres to be ready
COPY --from=builder /app/wait-for-it.sh /app/wait-for-it.sh
RUN sed -i 's/\r$//' /app/wait-for-it.sh && chmod +x /app/wait-for-it.sh

# Optionally copy the migrations directory into the container
COPY --from=builder /app/internal/infrastructure/adapters/outbound/postgres/migrations /app/migrations

# Start the application, waiting for Postgres to be ready first
# Assumes environment variables PG_HOST=postgres, PG_PORT=5432
CMD ["/app/wait-for-it.sh", "postgres:5432", "--timeout=30", "--strict", "--", "/app/main"]
