# 1. Используем официальный образ Golang для сборки приложения
FROM golang:1.24 AS builder

# 2. Устанавливаем рабочую директорию внутри контейнера
WORKDIR /quickflow_app

# 3. Копируем go.mod и go.sum и загружаем зависимости
COPY backend/go.mod backend/go.sum backend/
RUN cd backend && go mod download

# 4. Копируем исходный код с сохранением структуры
COPY backend /quickflow_app/backend
COPY deploy/config /quickflow_app/deploy/config

# 5. Собираем Go-приложение
WORKDIR /quickflow_app/backend
RUN go build -o main main.go

# 6. Создаём минимальный образ для продакшена
FROM debian:bookworm-slim

# 7. Устанавливаем рабочую директорию
WORKDIR /quickflow_app/backend

# 8. Копируем бинарник и конфиги
COPY --from=builder /quickflow_app/backend/main .
COPY --from=builder /quickflow_app/deploy /quickflow_app/deploy

# 9. Указываем команду для запуска
CMD ["./main", "-config", "/quickflow_app/deploy/config/feeder/config.toml"]
