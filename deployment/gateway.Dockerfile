# Multi-stage build for api-gateway
# Stage 1: Builder
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY services/api-gateway ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/api-gateway .

EXPOSE 8000

# Run the application
CMD ["./api-gateway"]
