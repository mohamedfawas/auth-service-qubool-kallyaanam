FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/server

# Use a small alpine image for the final stage
FROM alpine:latest

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/auth-service .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./auth-service"]