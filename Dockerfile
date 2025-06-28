# Stage 1: Build the Go binary
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o jollfi-gaming-api ./cmd/api

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS calls (e.g., Sui blockchain)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/jollfi-gaming-api .

# Expose port 8080
EXPOSE 8080

# Command to run the application
CMD ["./jollfi-gaming-api"]