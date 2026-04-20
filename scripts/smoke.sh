#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-/tmp/article-sentinel-go-cache}"
mkdir -p "$GOCACHE"

CONFIG_PATH="configs/config.local.yaml"
BASE_URL="http://127.0.0.1:8080"
SERVER_LOG="/tmp/article-sentinel-smoke-server.log"

wait_for_mysql() {
  local retries=30
  while (( retries > 0 )); do
    if docker compose exec -T mysql sh -c 'mysqladmin ping -h127.0.0.1 -uroot -p"$MYSQL_ROOT_PASSWORD" --silent' >/dev/null 2>&1; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 1
  done
  echo "mysql is not ready" >&2
  return 1
}

wait_for_redis() {
  local retries=30
  while (( retries > 0 )); do
    if docker compose exec -T redis sh -c 'redis-cli -a "$REDIS_PASSWORD" ping' >/dev/null 2>&1; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 1
  done
  echo "redis is not ready" >&2
  return 1
}

wait_for_http() {
  local path="$1"
  local retries=30
  while (( retries > 0 )); do
    if curl -fsS "$BASE_URL$path" >/dev/null 2>&1; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 1
  done
  echo "http endpoint not ready: $path" >&2
  return 1
}

check_endpoint() {
  local path="$1"
  local expected_status="$2"
  local status

  status="$(curl -sS -o /tmp/article-sentinel-smoke-body.txt -w '%{http_code}' "$BASE_URL$path")"
  if [[ "$status" != "$expected_status" ]]; then
    echo "unexpected status for $path: $status (want $expected_status)" >&2
    cat /tmp/article-sentinel-smoke-body.txt >&2 || true
    return 1
  fi
}

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "$SERVER_PID" >/dev/null 2>&1 || true
    wait "$SERVER_PID" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

docker compose up -d mysql redis
wait_for_mysql
wait_for_redis

go run ./cmd/migrate -config "$CONFIG_PATH"

go run ./cmd/server -config "$CONFIG_PATH" >"$SERVER_LOG" 2>&1 &
SERVER_PID=$!
wait_for_http "/healthz"

check_endpoint "/healthz" "200"
check_endpoint "/readyz" "200"
check_endpoint "/openapi.json" "200"
check_endpoint "/api/v1/posts" "200"

echo "smoke check passed"
