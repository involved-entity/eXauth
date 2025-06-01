FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth_service ./cmd/auth/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth_service .

EXPOSE 50051

RUN chmod +x auth_service

CMD ["./auth_service"]
