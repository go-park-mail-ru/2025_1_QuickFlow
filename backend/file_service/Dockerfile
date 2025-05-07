# Стейдж сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект в билд-контейнер
COPY . .

# Переходим в папку file_service
WORKDIR /app/file_service

# Собираем сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o file_service ./internal/server.go

# Стейдж финальный
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/file_service/file_service .

COPY --from=builder /app/file_service/deploy/config ./config

EXPOSE 8081

ENTRYPOINT ["./file_service"]
