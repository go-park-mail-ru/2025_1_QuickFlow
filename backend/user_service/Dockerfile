# Стейдж сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект в билд-контейнер
COPY . .

# Переходим в папку user_service
WORKDIR /app/user_service

# Собираем сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user_service ./internal/server.go

# Стейдж финальный
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/user_service/user_service .

COPY --from=builder /app/user_service/deploy/config ./config

EXPOSE 8083

ENTRYPOINT ["./user_service"]
