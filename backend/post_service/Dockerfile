# Стейдж сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект в билд-контейнер
COPY . .

# Переходим в папку post_service
WORKDIR /app/post_service

# Собираем сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o post_service ./internal/server.go

# Стейдж финальный
FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/post_service/post_service .

COPY --from=builder /app/post_service/deploy/config ./config

EXPOSE 8082

ENTRYPOINT ["./post_service"]
