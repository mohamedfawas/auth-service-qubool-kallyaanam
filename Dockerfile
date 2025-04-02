FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth-service .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./auth-service"]