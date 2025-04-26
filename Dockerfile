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

# 7. Устанавливаем необходимые пакеты, включая mc (MinIO Client)
RUN apt-get update && apt-get install -y \
    curl \
    bash \
    && curl -sL https://dl.min.io/client/mc/release/linux-amd64/mc -o /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc \
    && apt-get clean

# 8. Устанавливаем рабочую директорию
WORKDIR /quickflow_app/backend

# 9. Копируем бинарник и конфиги
COPY --from=builder /quickflow_app/backend/main .
COPY --from=builder /quickflow_app/deploy /quickflow_app/deploy

# 10. Копируем скрипты
COPY deploy/scripts/setup-minio-buckets.sh /usr/local/bin/setup-minio-buckets.sh

# 11. Копируем .env файл
COPY deploy/.env /etc/environment

# 12. Делаем скрипты исполнимыми
RUN chmod +x /usr/local/bin/setup-minio-buckets.sh

# Добавь перед CMD для вывода путей
RUN ls -l /usr/local/bin/setup-minio-buckets.sh

RUN echo "SCHEME=$SCHEME" && \
    if [ "$SCHEME" == "https" ]; then \
        printf "\nMINIO_SCHEME=https" >> /etc/environment && \
        printf "\nMINIO_PUBLIC_ENDPOINT=$DOMAIN/minio" >> /etc/environment; \
    else \
        printf "\nMINIO_SCHEME=http" >> /etc/environment && \
        printf "\nMINIO_PUBLIC_ENDPOINT=localhost:9000" >> /etc/environment; \
    fi

# Загружаем переменные в окружение контейнера
ARG SCHEME
ARG MINIO_PUBLIC_ENDPOINT

ENV MINIO_SCHEME=${SCHEME}
ENV MINIO_PUBLIC_ENDPOINT=${MINIO_PUBLIC_ENDPOINT}


# 13. Указываем команду для запуска
CMD ["/bin/bash", "-c", "/usr/local/bin/setup-minio-buckets.sh && ./main -server-config /quickflow_app/deploy/config/feeder/config.toml -cors-config /quickflow_app/deploy/config/cors/config.toml -minio-config /quickflow_app/deploy/config/minio/config.toml -validation-config /quickflow_app/deploy/config/validation/config.toml"]
