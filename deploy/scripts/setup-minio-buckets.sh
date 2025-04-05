#!/bin/bash

# Чтение переменных окружения
source /etc/environment

MINIO_ALIAS="myminio"
MINIO_HOST="http://$MINIO_INTERNAL_ENDPOINT"
MINIO_ACCESS_KEY="$MINIO_ROOT_USER"
MINIO_SECRET_KEY="$MINIO_ROOT_PASSWORD"

# Логирование переменных
echo "MinIO settings:"
echo "MINIO_ROOT_USER: $MINIO_ROOT_USER"
echo "MINIO_INTERNAL_ENDPOINT: $MINIO_INTERNAL_ENDPOINT"

# Ожидание запуска MinIO
until mc alias set $MINIO_ALIAS "$MINIO_HOST" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY"; do
  echo "Waiting for MinIO..."
  sleep 2
done

create_and_set_public_policy() {
  local bucket_name=$1
  if ! mc ls $MINIO_ALIAS/"$bucket_name" > /dev/null 2>&1; then
    echo "Creating bucket: $bucket_name"
    mc mb $MINIO_ALIAS/"$bucket_name"
  fi

  echo "Setting public policy for bucket: $bucket_name"
  mc anonymous set public $MINIO_ALIAS/"$bucket_name" || echo "Failed to set policy for $bucket_name"
}

create_and_set_public_policy "$MINIO_POSTS_BUCKET_NAME"
create_and_set_public_policy "$MINIO_PROFILE_BUCKET_NAME"
create_and_set_public_policy "$MINIO_ATTACHMENTS_BUCKET_NAME"
