# Multi-stage build for Go API

# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Stage 2: Minimal runtime image
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/api .

# Copy migrations (if you want to run them from the container)
COPY migrations ./migrations

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./api"]
