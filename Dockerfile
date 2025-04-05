# auth-service-qubool-kallyaanam/Dockerfile
FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o auth-service ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/auth-service .

EXPOSE 8081

CMD ["./auth-service"]