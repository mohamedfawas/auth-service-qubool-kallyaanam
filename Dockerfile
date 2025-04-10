# auth-service-qubool-kallyaanam/Dockerfile
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o authservice ./cmd/api


FROM alpine:latest

WORKDIR /app

# Install required packages
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/authservice .

# Copy the migrations directory
COPY --from=builder /app/migrations ./migrations

# Set environment variables
ENV PORT=8081

# Expose the port
EXPOSE 8081

# Use the correct binary name that matches the build command above
CMD ["./authservice"]