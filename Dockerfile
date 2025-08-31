# Build stage
FROM golang:1.23-alpine AS builder

# Install git (required for git operations)
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o git-remote-mcp .

# Runtime stage
FROM alpine:latest

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S git-mcp && \
    adduser -u 1001 -S git-mcp -G git-mcp

# Create workspace directory
RUN mkdir -p /workspace && \
    chown -R git-mcp:git-mcp /workspace

WORKDIR /app

# Copy the binary
COPY --from=builder /app/git-remote-mcp .

# Change ownership
RUN chown git-mcp:git-mcp git-remote-mcp

# Switch to non-root user
USER git-mcp

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
CMD ["./git-remote-mcp", "mcp", "--transport", "http", "--port", "8080", "--host", "0.0.0.0", "--workspace", "/workspace"]