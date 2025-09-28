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

# Install bash and PostgreSQL client
RUN apk add --no-cache bash postgresql-client

# Copy the built Go binary
COPY --from=builder /app/main /app/main

# Copy scripts from ./scripts
COPY --from=builder /app/scripts/wait-for-it.sh /app/wait-for-it.sh
COPY --from=builder /app/scripts/entrypoint.sh /app/entrypoint.sh

# Make scripts executable and fix line endings
RUN sed -i 's/\r$//' /app/wait-for-it.sh && chmod +x /app/wait-for-it.sh
RUN sed -i 's/\r$//' /app/entrypoint.sh && chmod +x /app/entrypoint.sh

# Use the wrapper script as entrypoint
ENTRYPOINT ["/app/entrypoint.sh"]
