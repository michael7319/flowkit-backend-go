# Build stage
FROM golang:1.21-alpine AS builder

# Install git (required for go mod download)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o flowkit-api main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/flowkit-api .

# Copy .env file (optional, better to use environment variables)
# COPY .env .

# Expose port
EXPOSE 5000

# Run the application
CMD ["./flowkit-api"]
