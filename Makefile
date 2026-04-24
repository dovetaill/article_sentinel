export GOCACHE ?= /tmp/article-sentinel-go-cache
CONFIG ?= configs/config.local.yaml
COMPOSE ?= docker compose
DEV_SCRIPT := bash scripts/dev.sh
DEV_GO_ENV := CONFIG=$(CONFIG) GOCACHE=$(GOCACHE)

.PHONY: up down stop dev dev-api dev-worker dev-scheduler dev-admin dev-check test verify smoke migrate

up:
	$(COMPOSE) up -d --wait mysql redis

down:
	$(COMPOSE) down

stop:
	$(DEV_SCRIPT) stop

dev:
	$(DEV_SCRIPT) stop; \
	session_id="$$($(DEV_SCRIPT) start-session)"; \
	trap '$(DEV_SCRIPT) stop-session "$$session_id"' INT TERM EXIT; \
	$(DEV_GO_ENV) setsid $(DEV_SCRIPT) api & \
	api_pid=$$!; \
	$(DEV_SCRIPT) register-dev-pid "$$session_id" api "$$api_pid"; \
	$(DEV_GO_ENV) setsid $(DEV_SCRIPT) worker & \
	worker_pid=$$!; \
	$(DEV_SCRIPT) register-dev-pid "$$session_id" worker "$$worker_pid"; \
	$(DEV_GO_ENV) setsid $(DEV_SCRIPT) scheduler & \
	scheduler_pid=$$!; \
	$(DEV_SCRIPT) register-dev-pid "$$session_id" scheduler "$$scheduler_pid"; \
	setsid $(DEV_SCRIPT) admin & \
	admin_pid=$$!; \
	$(DEV_SCRIPT) register-dev-pid "$$session_id" admin "$$admin_pid"; \
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
	bash scripts/dev_test.sh

test:
	go test ./...

verify:
	bash scripts/verify.sh

smoke:
	bash scripts/smoke.sh

migrate:
	go run ./cmd/migrate -config $(CONFIG)
