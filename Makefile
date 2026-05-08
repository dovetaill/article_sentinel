export GOCACHE ?= /tmp/article-sentinel-go-cache
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
TARGET_OS ?= linux
TARGET_ARCH ?= amd64
BUILD_DIR := build
RELEASE_DIR := release
RELEASE_BASENAME = article-sentinel_$(VERSION)_$(TARGET_OS)_$(TARGET_ARCH)
BIN_DIR := $(BUILD_DIR)/bin/$(TARGET_OS)-$(TARGET_ARCH)
ADMIN_DIR := web/admin
ADMIN_DIST_DIR := $(BUILD_DIR)/admin-dist
PACKAGE_ROOT = $(BUILD_DIR)/package/$(RELEASE_BASENAME)
RELEASE_TARBALL = $(RELEASE_DIR)/$(RELEASE_BASENAME).tar.gz
RELEASE_TARBALL_SHA256 = $(RELEASE_TARBALL).sha256
CONFIG ?= configs/config.local.yaml
COMPOSE ?= docker compose
DEV_SCRIPT := bash scripts/dev.sh
DEV_GO_ENV := CONFIG=$(CONFIG) GOCACHE=$(GOCACHE)

.PHONY: up down stop dev dev-api dev-worker dev-scheduler dev-admin dev-check test verify smoke migrate print-version build-server build-worker build-scheduler build-migrate build-go build-admin build package release

up:
	$(COMPOSE) up -d --wait mysql redis

down:
	$(COMPOSE) down

stop:
	$(DEV_SCRIPT) stop

dev:
	$(DEV_SCRIPT) stop; \
	$(DEV_SCRIPT) print-endpoints; \
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

print-version:
	@echo "VERSION=$(VERSION)"
	@echo "COMMIT=$(COMMIT)"
	@echo "BUILD_TIME=$(BUILD_TIME)"
	@echo "TARGET_OS=$(TARGET_OS)"
	@echo "TARGET_ARCH=$(TARGET_ARCH)"

build-server:
	@mkdir -p $(BIN_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -trimpath -o $(BIN_DIR)/article-sentinel-server ./cmd/server

build-worker:
	@mkdir -p $(BIN_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -trimpath -o $(BIN_DIR)/article-sentinel-worker ./cmd/worker

build-scheduler:
	@mkdir -p $(BIN_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -trimpath -o $(BIN_DIR)/article-sentinel-scheduler ./cmd/scheduler

build-migrate:
	@mkdir -p $(BIN_DIR)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) CGO_ENABLED=0 go build -trimpath -o $(BIN_DIR)/article-sentinel-migrate ./cmd/migrate

build-go: build-server build-worker build-scheduler build-migrate

build-admin:
	npm --prefix $(ADMIN_DIR) ci
	npm --prefix $(ADMIN_DIR) run build
	rm -rf $(ADMIN_DIST_DIR)
	@mkdir -p $(ADMIN_DIST_DIR)
	cp -R $(ADMIN_DIR)/dist/. $(ADMIN_DIST_DIR)/

build: build-go build-admin

package: build scripts/package_release.sh scripts/write_release_manifest.sh
	bash scripts/package_release.sh \
		--package-root "$(PACKAGE_ROOT)" \
		--bin-dir "$(BIN_DIR)" \
		--admin-dir "$(ADMIN_DIST_DIR)" \
		--migrations-dir "migrations" \
		--config-file "configs/config.example.yaml" \
		--deploy-dir "deploy"
	bash scripts/write_release_manifest.sh \
		--package-root "$(PACKAGE_ROOT)" \
		--app "article-sentinel" \
		--version "$(VERSION)" \
		--git-sha "$(COMMIT)" \
		--build-time "$(BUILD_TIME)" \
		--target-os "$(TARGET_OS)" \
		--target-arch "$(TARGET_ARCH)"

release: package
	@mkdir -p $(RELEASE_DIR)
	rm -f "$(RELEASE_TARBALL)" "$(RELEASE_TARBALL_SHA256)"
	tar --sort=name --mtime='UTC 1970-01-01' --owner=0 --group=0 --numeric-owner -cf - \
		-C "$(BUILD_DIR)/package" "$(RELEASE_BASENAME)" | gzip -n >"$(RELEASE_TARBALL)"
	@(cd "$(RELEASE_DIR)" && sha256sum "$(RELEASE_BASENAME).tar.gz" >"$(RELEASE_BASENAME).tar.gz.sha256")
