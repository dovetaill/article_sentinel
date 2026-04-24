#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_PATH="${CONFIG:-configs/config.local.yaml}"
ADMIN_DIR="$ROOT_DIR/web/admin"

usage() {
  cat <<'USAGE'
Usage: scripts/dev.sh <command>

Commands:
  api              Run the backend API server
  worker           Run the async worker
  scheduler        Run the scheduler
  admin            Run the admin Vite dev server
  print-plan       Print the expected make dev process layout
  assert-make-dev  Verify that `make -n dev` includes all four processes
USAGE
}

prepare_go_env() {
  export GOCACHE="${GOCACHE:-/tmp/article-sentinel-go-cache}"
  mkdir -p "$GOCACHE"
}

print_plan() {
  cat <<PLAN
make dev must start:
- backend server: bash scripts/dev.sh api
- worker: bash scripts/dev.sh worker
- scheduler: bash scripts/dev.sh scheduler
- admin dev server: bash scripts/dev.sh admin
PLAN
}

assert_make_dev() {
  local output
  output="$(cd "$ROOT_DIR" && make -n dev)"

  local missing=0
  local expected
  for expected in \
    "scripts/dev.sh api" \
    "scripts/dev.sh worker" \
    "scripts/dev.sh scheduler" \
    "scripts/dev.sh admin"; do
    if [[ "$output" != *"$expected"* ]]; then
      echo "missing dev process in make -n dev: $expected" >&2
      missing=1
    fi
  done

  if (( missing != 0 )); then
    echo "current make -n dev output:" >&2
    echo "$output" >&2
    return 1
  fi
}

run_api() {
  prepare_go_env
  cd "$ROOT_DIR"
  exec go run ./cmd/server -config "$CONFIG_PATH"
}

run_worker() {
  prepare_go_env
  cd "$ROOT_DIR"
  exec go run ./cmd/worker -config "$CONFIG_PATH"
}

run_scheduler() {
  prepare_go_env
  cd "$ROOT_DIR"
  exec go run ./cmd/scheduler -config "$CONFIG_PATH"
}

run_admin() {
  cd "$ADMIN_DIR"
  exec npm run dev -- --host 0.0.0.0
}

command="${1:-}"
case "$command" in
  api)
    run_api
    ;;
  worker)
    run_worker
    ;;
  scheduler)
    run_scheduler
    ;;
  admin)
    run_admin
    ;;
  print-plan)
    print_plan
    ;;
  assert-make-dev)
    assert_make_dev
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
