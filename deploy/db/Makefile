# === CONFIGURATION ===
DB_HOST=localhost
DB_PORT=5432
DB_NAME=quickflow_db
DB_USER=quickflow_admin
DB_PASS=SuperSecurePassword1
SSL_MODE=disable

MIGRATIONS_DIR=./migrations
DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSL_MODE)

MIGRATE_BIN=migrate

# pgbadger
LOG_DIR=./infra/postgres/log
REPORT_DIR=./infra/postgres/reports
PG_BADGER=pgbadger
LOG_FORMAT=stderr

.PHONY: up down force goto create status drop help report clean

up:
	@echo "Running up $(if $(N),$(N),all)..."
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up $(if $(N),$(N),)

down:
	@echo "Rolling down $(if $(N),1)..."
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down $(if $(N),$(N),1)

goto:
	@echo "Migrating to version $(V)..."
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" goto $(V)

force:
	@echo "Forcing version to $(V)..."
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(V)

create:
	@read -p "Enter migration name: " name; \
	$(MIGRATE_BIN) create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

status:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

drop:
	@echo "Dropping all tables..."
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f

help:
	@echo "Available targets:"
	@echo "  up [N=1]        - Apply up migrations (default: all)"
	@echo "  down [N=1]      - Roll back down N migrations (default: 1)"
	@echo "  force V=2       - Set database version to V without running migrations"
	@echo "  goto V=2        - Migrate to a specific version"
	@echo "  create          - Create a new migration file"
	@echo "  status          - Print current DB version"
	@echo "  drop            - Drop everything"

report:
	mkdir -p $(REPORT_DIR)
	$(PG_BADGER) -f $(LOG_FORMAT) $(LOG_DIR)/*.log -o $(REPORT_DIR)/pgbadger-report.html

clean:
	rm -f $(REPORT_DIR)/*.html