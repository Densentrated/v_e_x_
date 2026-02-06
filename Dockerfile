# Multi-stage build for Go server
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY backend/go.mod backend/go.sum ./backend/

# Download dependencies
WORKDIR /app/backend
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o vex-server .

# Stage for handling optional .env file
FROM alpine:latest AS env-stage

# Copy .env file if it exists, otherwise create empty placeholder
COPY .env* /tmp/
RUN if [ -f /tmp/.env ]; then \
    cp /tmp/.env /.env; \
    else \
    touch /.env; \
    fi

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/backend/vex-server .

# Copy .env file from env-stage (will be either actual .env or empty file)
COPY --from=env-stage /.env /.env

# Create directories for data and set proper ownership
RUN mkdir -p /app/clone /app/vectors && \
    chown -R appuser:appgroup /app && \
    chown appuser:appgroup /.env

# Switch to non-root user
USER appuser

# Expose port (will be overridden by environment variable)
EXPOSE 22010

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:${SERVER_PORT:-22010}/health || exit 1

# Run the server
CMD ["./vex-server"]
