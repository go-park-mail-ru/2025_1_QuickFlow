#!/bin/bash

if [[ "$MINIO_SCHEME" == "https" ]]; then
  export MINIO_SCHEME=https
  export MINIO_PUBLIC_ENDPOINT=$DOMAIN/minio
else
  export MINIO_SCHEME=http
  export MINIO_PUBLIC_ENDPOINT=localhost:9000
fi

echo "!!! scheme setup: $MINIO_SCHEME for domain $DOMAIN"
echo "!!! public endpoint: $MINIO_PUBLIC_ENDPOINT"