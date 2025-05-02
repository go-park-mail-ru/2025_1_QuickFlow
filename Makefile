DOMAIN=quickflowapp.ru
CERT_PATH=/etc/letsencrypt/live/$(DOMAIN)/fullchain.pem
SCHEME?=http
MINIO_PUBLIC_ENDPOINT?=localhost:9000
COMPOSE_FILE=./deploy/micro-docker-compose.yml
MODE?=daemon
COMPOSE=docker compose

ifeq ($(shell [ -f $(CERT_PATH) ] && echo yes),yes)
    SCHEME=https
endif

ifeq (${SCHEME}, https)
    MINIO_PUBLIC_ENDPOINT=${DOMAIN}/minio
endif

build:
	$(COMPOSE) -f $(COMPOSE_FILE) build --build-arg SCHEME=$(SCHEME) --build-arg MINIO_PUBLIC_ENDPOINT=${MINIO_PUBLIC_ENDPOINT}

up: build
ifeq ($(MODE),daemon)
	$(COMPOSE) -f $(COMPOSE_FILE) up -d
else
	$(COMPOSE) -f $(COMPOSE_FILE) up
endif

down:
ifeq ($(ERASE),yes)
	$(COMPOSE) -f $(COMPOSE_FILE) down -v
else
	$(COMPOSE) -f $(COMPOSE_FILE) down
endif


restart: down up

logs:
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f

ps:
	$(COMPOSE) -f $(COMPOSE_FILE) ps

lint:
	cd backend && golangci-lint run ./...
