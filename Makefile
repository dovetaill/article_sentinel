export GOCACHE ?= /tmp/article-sentinel-go-cache
CONFIG ?= configs/config.local.yaml
COMPOSE ?= docker compose

.PHONY: up down dev test verify smoke migrate

up:
	$(COMPOSE) up -d --wait mysql redis

down:
	$(COMPOSE) down

dev:
	go run ./cmd/server -config $(CONFIG)

test:
	go test ./...

verify:
	bash scripts/verify.sh

smoke:
	bash scripts/smoke.sh

migrate:
	go run ./cmd/migrate -config $(CONFIG)
