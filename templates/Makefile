# ==============================================================================
# HELIX : {{ .ProjectName }} MAKEFILE
# ==============================================================================

.PHONY: help up down logs shell tidy ent-gen proto migrate-diff migrate-apply clean init swagger docs up-infra down-infra db-create wait-db

# Variables
DB_USER=dev
DB_NAME={{ .EntityPluralLower }}
CONTAINER_DB=helix-shared-db
CONTAINER_ATLAS=helix-shared-db-atlas
SHARED_PROJECT_NAME=helix-shared

# Default target: Help
help: ## Show this help message
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# INFRASTRUCTURE OPERATIONS (SHARED)
# ==============================================================================

up-infra: ## Start shared infrastructure (DB, Kafka, Redis)
	@echo "Starting Shared Infrastructure..."
	@# FIX: Use explicit project name (-p) to prevent conflicts between services
	@docker compose -p $(SHARED_PROJECT_NAME) -f docker-compose.infra.yml up -d
	@echo "Infrastructure is up! (Network: helix-shared-net)"

down-infra: ## Stop shared infrastructure
	@echo "Stopping Shared Infrastructure..."
	@docker compose -p $(SHARED_PROJECT_NAME) -f docker-compose.infra.yml down
	@echo "Infrastructure stopped."
clean-infra: ## Stop shared infrastructure
	@echo "Cleaning Shared Infrastructure..."
	@docker compose -p $(SHARED_PROJECT_NAME) -f docker-compose.infra.yml down -v
	@echo "Infrastructure cleaned."
reset:
	@echo "Resetting environment..."
	@make db-clean
	@make clean-infra
	@echo "Reset success"

# ==============================================================================
# DATABASE UTILS (The Fix)
# ==============================================================================

wait-db: ## Wait for DB to be healthy (No more sleep 5)
	@echo "Waiting for $(CONTAINER_DB) to be ready..."
	@until docker exec $(CONTAINER_DB) pg_isready -U $(DB_USER); do \
		echo "DB is unavailable - sleeping..."; \
		sleep 3; \
	done
	@echo "DB is up!"

db-create: wait-db ## DESTRUCTIVE DB Creation (Main & Shadow) - FORCE RESET
	@echo "------------------------------------------------------------------"
	@echo "NUKE & RECREATE DATABASES..."
	@echo "------------------------------------------------------------------"
	
	@# Main DB: FORCE RECREATE
	@echo "Resetting Main DB: $(DB_NAME)..."
	@# 1. Kill active connections to the DB so we can drop it
	@docker exec $(CONTAINER_DB) psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_NAME)';" > /dev/null 2>&1 || true
	@# 2. Drop if exists
	@docker exec $(CONTAINER_DB) psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS \"$(DB_NAME)\"" || true
	@# 3. Create fresh
	@docker exec $(CONTAINER_DB) psql -U $(DB_USER) -d postgres -c "CREATE DATABASE \"$(DB_NAME)\""
	
	@# Atlas Shadow DB: FORCE RECREATE
	@echo "Resetting Atlas Shadow DB: $(DB_NAME)..."
	@docker exec $(CONTAINER_ATLAS) psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS \"$(DB_NAME)\"" || true
	@docker exec $(CONTAINER_ATLAS) psql -U $(DB_USER) -d postgres -c "CREATE DATABASE \"$(DB_NAME)\""
	
	@echo "Databases nuked and ready."

fix-deps: ## Hard reset go.mod/sum and tidy
	@echo "Nuking go.sum and forcing tidy..."
	@rm -f go.sum
	@docker compose run --rm {{ .ProjectName }} go mod tidy
	@echo "Dependencies fixed."

force-migrate: fix-deps db-create ## Nuke migrations and re-generate from Ent Schema
	@echo "Regenerating Migrations..."
	@rm -f migrations/*.sql
	@$(MAKE) migrate-diff name=init_force
	@$(MAKE) migrate-apply
	@echo "Migration forced."

# ==============================================================================
# BOOTSTRAP (One-Command Setup)
# ==============================================================================

init: ## Full Setup: Infra -> DB Init -> Deps -> Gen -> Migrate -> Run
	@echo "------------------------------------------------------------------"
	@echo "Ensuring Shared Infrastructure is Up..."
	@echo "------------------------------------------------------------------"
	$(MAKE) up-infra
	
	@# INI YANG BARU: Create DB dulu sebelum connect
	$(MAKE) db-create

	@echo "------------------------------------------------------------------"
	@echo "Downloading Dependencies..."
	@echo "------------------------------------------------------------------"
	# 'go mod download' inside container requires the app container to run, 
	# but we use 'run --rm' which mounts the code.
	@make up
	docker compose run --rm {{ .ProjectName }} go mod download
	
	@echo "------------------------------------------------------------------"
	@echo "Generating Code (Ent & Proto)..."
	@echo "------------------------------------------------------------------"
	$(MAKE) ent-gen
	$(MAKE) proto

	@echo "------------------------------------------------------------------"
	@echo "Cleaning Modules (Tidy)..."
	@echo "------------------------------------------------------------------"
	$(MAKE) tidy

	@echo "------------------------------------------------------------------"
	@echo "Creating Initial Migration..."
	@echo "------------------------------------------------------------------"
	# Check if migration exists, if not create 'init'
	@if [ ! -f "migrations/*_init.sql" ]; then \
		$(MAKE) migrate-diff name=init; \
	fi
	
	@echo "------------------------------------------------------------------"
	@echo "Applying Migrations..."
	@echo "------------------------------------------------------------------"
	$(MAKE) migrate-apply
	
	@echo "------------------------------------------------------------------"
	@echo "Generating docs..."
	@echo "------------------------------------------------------------------"
	$(MAKE) swagger

	@echo "------------------------------------------------------------------"
	@echo "Launching Application..."
	@echo "------------------------------------------------------------------"
	$(MAKE) up
	$(MAKE) logs

# ==============================================================================
# INFRASTRUCTURE (Docker Compose)
# ==============================================================================

up: ## Start environment (App, DB, Atlas-DB)
	docker compose up -d
	@echo "Environment is up!"

down: ## Stop environment and remove orphans
	docker compose down --remove-orphans
	@echo "Environment stopped."

restart: ## Restart the app container
	docker compose restart {{ .ProjectName }}
	@echo "App restarted."

logs: ## Follow app logs
	docker compose logs -f {{ .ProjectName }}

# ==============================================================================
# GO TOOLING
# ==============================================================================

tidy: ## Run go mod tidy inside container
	docker compose run --rm {{ .ProjectName }} go mod tidy

test: ## Run tests (Unit & Integration)
	docker compose run --rm {{ .ProjectName }} go test -v ./...

shell: ## Enter the app container shell (ephemeral)
	docker compose run --rm {{ .ProjectName }} /bin/sh

# ==============================================================================
# CODE GENERATION
# ==============================================================================

ent-gen: ## Generate Ent assets from schema
	docker compose run --rm {{ .ProjectName }} go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/modifier,sql/upsert ./ent/schema

proto: ## Generate Protobuf (Containerized)
	# Runs buf inside the container so you don't need it locally
	docker compose run --rm {{ .ProjectName }} buf generate

# ==============================================================================
# MIGRATIONS (Atlas) - OPS READY
# ==============================================================================

# Variables for Atlas
ATLAS_DIR="file://migrations"
ENT_SCHEMA="ent://ent/schema"
# Dev URL pointing to the '{{ .ProjectName }}-db-dev' service
DEV_URL="postgres://dev:dev@helix-shared-db-atlas:5432/{{ .EntityPluralLower }}?sslmode=disable"
DB_URL="postgres://dev:dev@helix-shared-db:5432/{{ .EntityPluralLower }}?sslmode=disable"

migrate-diff: ## Create migration file (Dev only)
	@if [ -z "$(name)" ]; then echo "Error: name required. Usage: make migrate-diff name=add_users"; exit 1; fi
	docker compose run --rm {{ .ProjectName }} atlas migrate diff $(name) \
	  --dir $(ATLAS_DIR) \
	  --to $(ENT_SCHEMA) \
	  --dev-url $(DEV_URL)

migrate-apply: ## Apply pending migrations (CI/CD Safe)
	docker compose run --rm {{ .ProjectName }} atlas migrate apply \
	  --dir $(ATLAS_DIR) \
	  --url $(DB_URL)

migrate-down: ## Rollback 1 migration version (Use with Caution)
	@echo "Rolling back the last migration..."
	docker compose run --rm {{ .ProjectName }} atlas migrate down \
	  --dir $(ATLAS_DIR) \
	  --url $(DB_URL) \
	  1

migrate-hash: 
	docker compose run --rm {{ .ProjectName }} atlas migrate hash \
	  --dir $(ATLAS_DIR)
	  
migrate-status: ## Check migration status
	docker compose run --rm {{ .ProjectName }} atlas migrate status \
	  --dir $(ATLAS_DIR) \
	  --url $(DB_URL)

# ==============================================================================
# CLEANUP
# ==============================================================================

clean: ## Remove temp files
	rm -rf tmp bin

db-clean: ## Stop env & remove ALL Data Volumes (Postgres & Redpanda) - HARD RESET
	@echo "WARNING: This will delete all database and message broker data!"
	@echo "Stopping services and removing volumes..."
	docker compose down --volumes --remove-orphans
	@echo "Environment cleaned. Run 'make init' or 'make up' to start fresh."

swagger: ## Generate Swagger/OpenAPI spec from code annotations (Containerized)
	@echo "Generating API Documentation..."
	docker compose run --rm {{ .ProjectName }} swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal --parseDepth 3

docs: swagger ## Alias for swagger

# ==============================================================================
# BUILD OPERATIONS
# ==============================================================================

docker-build: ## Build Docker Image (Force Rebuild)
	docker compose build --no-cache {{ .ProjectName }}

rebuild: docker-build ## Alias for docker-build

docker-build-migrate: ## Build Migration Runner Image (For Production)
	@echo "Building Migration Image..."
	docker build -f Dockerfile.migrate -t {{ .ProjectName }}-migrate:latest .
	@echo "Migration image built: {{ .ProjectName }}-migrate:latest"
	@echo "Run in prod: docker run --rm -e DB_DSN=$DB_URL {{ .ProjectName }}-migrate:latest"