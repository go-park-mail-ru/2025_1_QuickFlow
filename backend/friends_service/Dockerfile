# Стейдж сборки
FROM golang:1.24 AS builder
LABEL authors="Matvey"

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект в билд-контейнер
COPY . .

# Переходим в папку friends_service
WORKDIR /app/friends_service

# Собираем сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o friends_service ./internal/server.go

# Стейдж финальный
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/friends_service/friends_service .

# COPY --from=builder /app/friends_service/deploy/config ./config

EXPOSE 8086

ENTRYPOINT ["./friends_service"]
