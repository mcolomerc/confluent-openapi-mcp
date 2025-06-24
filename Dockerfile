# Multi-stage build for optimal image size
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates (needed for downloading dependencies)
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o mcp-server ./cmd/main.go

# Final stage - minimal alpine image with health check tools
FROM alpine:3.19

# Install ca-certificates and basic networking tools for health checks
RUN apk --no-cache add ca-certificates curl

# Create appuser for security
RUN adduser -D -g '' appuser

# Import user/group files
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary from builder stage
COPY --from=builder /build/mcp-server /app/mcp-server

# Copy prompts directory
COPY --from=builder /build/prompts /app/prompts

# Copy API specification
COPY --from=builder /build/api-spec /app/api-spec

# Set working directory
WORKDIR /app

# Use non-root user
USER appuser

# Expose port (optional, mainly for documentation)
EXPOSE 8080

# Command to run the binary
ENTRYPOINT ["/app/mcp-server"]
