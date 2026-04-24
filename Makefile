export GOCACHE ?= /tmp/article-sentinel-go-cache
CONFIG ?= configs/config.local.yaml
COMPOSE ?= docker compose
DEV_SCRIPT := bash scripts/dev.sh
DEV_GO_ENV := CONFIG=$(CONFIG) GOCACHE=$(GOCACHE)

.PHONY: up down dev dev-api dev-worker dev-scheduler dev-admin dev-check test verify smoke migrate

up:
	$(COMPOSE) up -d --wait mysql redis

down:
	$(COMPOSE) down

dev:
	trap 'kill 0' INT TERM EXIT; \
	$(DEV_GO_ENV) $(DEV_SCRIPT) api & \
	api_pid=$$!; \
	$(DEV_GO_ENV) $(DEV_SCRIPT) worker & \
	worker_pid=$$!; \
	$(DEV_GO_ENV) $(DEV_SCRIPT) scheduler & \
	scheduler_pid=$$!; \
	$(DEV_SCRIPT) admin & \
	admin_pid=$$!; \
	wait $$api_pid $$worker_pid $$scheduler_pid $$admin_pid

dev-api:
	$(DEV_GO_ENV) $(DEV_SCRIPT) api

dev-worker:
	$(DEV_GO_ENV) $(DEV_SCRIPT) worker

dev-scheduler:
	$(DEV_GO_ENV) $(DEV_SCRIPT) scheduler

dev-admin:
	$(DEV_SCRIPT) admin

dev-check:
	$(DEV_SCRIPT) print-plan
	$(DEV_SCRIPT) assert-make-dev

test:
	go test ./...

verify:
	bash scripts/verify.sh

smoke:
	bash scripts/smoke.sh

migrate:
	go run ./cmd/migrate -config $(CONFIG)
