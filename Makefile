SHELL := /bin/bash

# Container & volume names (override via CLI: e.g. PG_PORT=15432)
PG_CONTAINER ?= fjord-pg
PG_VOLUME ?= fjord-pg-data
PG_PORT ?= 5432
PG_USER ?= fjord
PG_PASSWORD ?= fjordpass
PG_DB ?= fjorddb
PG_IMAGE ?= postgres:15

REDIS_CONTAINER ?= fjord-redis
REDIS_VOLUME ?= fjord-redis-data
REDIS_PORT ?= 6379
REDIS_IMAGE ?= redis:7

.PHONY: help create create-db create-redis start start-db start-redis stop stop-db stop-redis delete delete-db delete-redis logs logs-db logs-redis status psql redis-cli

help:
	@printf "Makefile - Dev Docker helpers\n"
	@printf "\nTargets:\n"
	@printf "  create        - Create both Postgres and Redis containers\n"
	@printf "  create-db     - Create Postgres container\n"
	@printf "  create-redis  - Create Redis container\n"
	@printf "  start         - Start both services\n"
	@printf "  stop          - Stop both services\n"
	@printf "  delete        - Remove both containers and their volumes\n"
	@printf "  status        - List the containers\n"
	@printf "  logs          - Tail logs for both containers\n"
	@printf "  psql          - Open psql inside the Postgres container\n"
	@printf "  redis-cli     - Open redis-cli inside the Redis container\n"

#######################
# Create / Start / Stop
#######################

create: create-db create-redis

create-db:
	@echo "Creating Postgres container $(PG_CONTAINER)"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(PG_CONTAINER)$$"; then \
		echo "Container $(PG_CONTAINER) already exists"; \
	else \
		docker volume create $(PG_VOLUME) >/dev/null || true; \
		docker run -d --name $(PG_CONTAINER) \
			-e POSTGRES_USER=$(PG_USER) -e POSTGRES_PASSWORD=$(PG_PASSWORD) -e POSTGRES_DB=$(PG_DB) \
			-p $(PG_PORT):5432 -v $(PG_VOLUME):/var/lib/postgresql/data $(PG_IMAGE); \
		echo "Created $(PG_CONTAINER)"; \
	fi

create-redis:
	@echo "Creating Redis container $(REDIS_CONTAINER)"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(REDIS_CONTAINER)$$"; then \
		echo "Container $(REDIS_CONTAINER) already exists"; \
	else \
		docker volume create $(REDIS_VOLUME) >/dev/null || true; \
		docker run -d --name $(REDIS_CONTAINER) -p $(REDIS_PORT):6379 -v $(REDIS_VOLUME):/data $(REDIS_IMAGE) \
			--appendonly yes; \
		echo "Created $(REDIS_CONTAINER)"; \
	fi

start: start-db start-redis

start-db:
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(PG_CONTAINER)$$"; then \
		docker start $(PG_CONTAINER) >/dev/null && echo "Started $(PG_CONTAINER)" || true; \
	else \
		echo "Postgres container does not exist. Run 'make create-db' first."; \
	fi

start-redis:
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(REDIS_CONTAINER)$$"; then \
		docker start $(REDIS_CONTAINER) >/dev/null && echo "Started $(REDIS_CONTAINER)" || true; \
	else \
		echo "Redis container does not exist. Run 'make create-redis' first."; \
	fi

stop: stop-db stop-redis

stop-db:
	@if docker ps --format '{{.Names}}' | grep -Eq "^$(PG_CONTAINER)$$"; then \
		docker stop $(PG_CONTAINER) >/dev/null && echo "Stopped $(PG_CONTAINER)" || true; \
	else \
		echo "Postgres is not running"; \
	fi

stop-redis:
	@if docker ps --format '{{.Names}}' | grep -Eq "^$(REDIS_CONTAINER)$$"; then \
		docker stop $(REDIS_CONTAINER) >/dev/null && echo "Stopped $(REDIS_CONTAINER)" || true; \
	else \
		echo "Redis is not running"; \
	fi

#######################
# Delete / Remove
#######################

delete: delete-db delete-redis

delete-db:
	@echo "Removing Postgres container and volume"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(PG_CONTAINER)$$"; then \
		docker rm -f $(PG_CONTAINER) >/dev/null && echo "Removed container $(PG_CONTAINER)"; \
	else \
		echo "Postgres container does not exist"; \
	fi
	@-docker volume rm $(PG_VOLUME) >/dev/null 2>&1 || true

delete-redis:
	@echo "Removing Redis container and volume"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^$(REDIS_CONTAINER)$$"; then \
		docker rm -f $(REDIS_CONTAINER) >/dev/null && echo "Removed container $(REDIS_CONTAINER)"; \
	else \
		echo "Redis container does not exist"; \
	fi
	@-docker volume rm $(REDIS_VOLUME) >/dev/null 2>&1 || true

#######################
# Logs / Status / Utils
#######################

logs: logs-db logs-redis

logs-db:
	@docker logs -f $(PG_CONTAINER) || true

logs-redis:
	@docker logs -f $(REDIS_CONTAINER) || true

status:
	@docker ps -a --filter "name=$(PG_CONTAINER)" --filter "name=$(REDIS_CONTAINER)"

psql:
	@docker exec -it $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB)

redis-cli:
	@docker exec -it $(REDIS_CONTAINER) redis-cli
