#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG_PATH="${CONFIG:-configs/config.local.yaml}"
ADMIN_DIR="$ROOT_DIR/web/admin"
# 用仓库路径做哈希，避免同一台机器上多个 worktree 互相抢占 dev 状态目录。
DEV_STATE_DIR="${TMPDIR:-/tmp}/article-sentinel-dev-$(printf '%s' "$ROOT_DIR" | cksum | awk '{print $1}')"
CURRENT_SESSION_FILE="$DEV_STATE_DIR/current-session"

usage() {
  cat <<'USAGE'
Usage: scripts/dev.sh <command>

Commands:
  api              Run the backend API server
  worker           Run the async worker
  scheduler        Run the scheduler
  admin            Run the admin Vite dev server
  print-endpoints  Print backend/admin dev URLs for the current config
  stop             Stop the current make dev stack
  print-plan       Print the expected make dev process layout
  assert-make-dev  Verify that `make -n dev` includes all four processes

Internal commands:
  start-session                 Create a dev session id for make dev
  register-dev-pid <id> <name> <pid>
  stop-session <id>
USAGE
}

ensure_state_dir() {
  mkdir -p "$DEV_STATE_DIR"
}

read_config_scalar() {
  local section="$1"
  local key="$2"
  local default_value="${3:-}"

  local value
  value="$(
    awk -v target_section="$section" -v target_key="$key" '
      BEGIN {
        in_section = 0
      }
      /^[[:space:]]*#/ || /^[[:space:]]*$/ {
        next
      }
      {
        if ($0 ~ ("^" target_section ":[[:space:]]*$")) {
          in_section = 1
          next
        }
        if (in_section && $0 ~ /^[^[:space:]]/) {
          in_section = 0
        }
        if (!in_section) {
          next
        }
        pattern = "^[[:space:]]+" target_key ":[[:space:]]*"
        if ($0 ~ pattern) {
          line = $0
          sub(pattern, "", line)
          gsub(/^[\"\047]|[\"\047]$/, "", line)
          print line
          exit
        }
      }
    ' "$CONFIG_PATH" 2>/dev/null
  )"

  if [[ -n "$value" ]]; then
    printf '%s\n' "$value"
    return
  fi

  printf '%s\n' "$default_value"
}

backend_port() {
  printf '%s\n' "${APP_PORT:-$(read_config_scalar app port 8080)}"
}

backend_bind_host() {
  printf '%s\n' "${APP_HOST:-$(read_config_scalar app host 0.0.0.0)}"
}

backend_display_host() {
  local host
  host="$(backend_bind_host)"
  case "$host" in
    ""|"0.0.0.0"|"::")
      printf '%s\n' "127.0.0.1"
      ;;
    *)
      printf '%s\n' "$host"
      ;;
  esac
}

backend_network_host() {
  local bind_host
  bind_host="$(backend_bind_host)"
  case "$bind_host" in
    ""|"0.0.0.0"|"::")
      hostname -I 2>/dev/null | awk '{for (i = 1; i <= NF; i++) if ($i !~ /^127\./) {print $i; exit}}'
      ;;
    *)
      printf '%s\n' "$bind_host"
      ;;
  esac
}

print_endpoints() {
  local port backend_host backend_base admin_base admin_jump_login network_host
  port="$(backend_port)"
  backend_host="$(backend_display_host)"
  backend_base="http://${backend_host}:${port}"
  admin_base="http://127.0.0.1:5173"
  admin_jump_login="${admin_base}/auth/login?jwt=<legacy-jwt>"

  echo "Admin UI: ${admin_base}"
  echo "Admin jump login: ${admin_jump_login}"
  echo "Backend API: ${backend_base}"
  echo "Session API: ${backend_base}/api/v1/auth/session"
  echo "Jump login: ${backend_base}/auth/login?jwt=<legacy-jwt>"

  network_host="$(backend_network_host)"
  if [[ -n "$network_host" && "$network_host" != "$backend_host" ]]; then
    echo "Admin UI (network): http://${network_host}:5173"
    echo "Admin jump login (network): http://${network_host}:5173/auth/login?jwt=<legacy-jwt>"
    echo "Backend API (network): http://${network_host}:${port}"
    echo "Jump login (network): http://${network_host}:${port}/auth/login?jwt=<legacy-jwt>"
  fi
}

session_pid_file() {
  local session_id="$1"
  printf '%s/%s.pids\n' "$DEV_STATE_DIR" "$session_id"
}

current_session_id() {
  if [[ ! -f "$CURRENT_SESSION_FILE" ]]; then
    return 1
  fi

  local session_id
  session_id="$(<"$CURRENT_SESSION_FILE")"
  if [[ -z "$session_id" ]]; then
    return 1
  fi

  printf '%s\n' "$session_id"
}

start_session() {
  ensure_state_dir

  local session_id
  session_id="$(date +%s)-$$-$RANDOM"
  : >"$(session_pid_file "$session_id")"
  printf '%s\n' "$session_id" >"$CURRENT_SESSION_FILE"
  printf '%s\n' "$session_id"
}

register_dev_pid() {
  local session_id="${1:-}"
  local process_name="${2:-}"
  local process_pid="${3:-}"

  if [[ -z "$session_id" || -z "$process_name" || -z "$process_pid" ]]; then
    echo "usage: scripts/dev.sh register-dev-pid <session-id> <name> <pid>" >&2
    return 1
  fi

  ensure_state_dir
  printf '%s %s\n' "$process_name" "$process_pid" >>"$(session_pid_file "$session_id")"
}

stop_pid() {
  local process_pid="$1"

  if ! kill -0 "$process_pid" >/dev/null 2>&1; then
    return 0
  fi

  kill -TERM -- "-$process_pid" >/dev/null 2>&1 || true
  kill -TERM "$process_pid" >/dev/null 2>&1 || true

  local retries=20
  while (( retries > 0 )); do
    if ! kill -0 "$process_pid" >/dev/null 2>&1; then
      return 0
    fi
    retries=$((retries - 1))
    sleep 0.25
  done

  kill -KILL -- "-$process_pid" >/dev/null 2>&1 || true
  kill -KILL "$process_pid" >/dev/null 2>&1 || true
}

stale_process_matches() {
  local process_pid="$1"
  local process_args="$2"

  case "$process_args" in
    *article-sentinel-server*|*article-sentinel-worker*|*article-sentinel-scheduler*)
      return 0
      ;;
  esac

  local process_cwd=""
  if [[ -L "/proc/$process_pid/cwd" ]]; then
    process_cwd="$(readlink -f "/proc/$process_pid/cwd" 2>/dev/null || true)"
  fi

  if [[ "$process_cwd" == "$ROOT_DIR" ]]; then
    case "$process_args" in
      *"go run ./cmd/server"*|*"go run ./cmd/worker"*|*"go run ./cmd/scheduler"*|\
      */go-build*/exe/server*|*/go-build*/exe/worker*|*/go-build*/exe/scheduler*|\
      */article-sentinel-go-cache/*/server*|*/article-sentinel-go-cache/*/worker*|*/article-sentinel-go-cache/*/scheduler*)
        return 0
        ;;
    esac
  fi

  if [[ "$process_cwd" == "$ADMIN_DIR" ]]; then
    case "$process_args" in
      *vite*|*"npm run dev -- --host 0.0.0.0"*)
        return 0
        ;;
    esac
  fi

  return 1
}

stop_stale_dev_processes() {
  local line process_pid process_args
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue

    read -r process_pid process_args <<<"$line"

    [[ "$process_pid" =~ ^[0-9]+$ ]] || continue
    [[ "$process_pid" != "$$" ]] || continue

    if stale_process_matches "$process_pid" "$process_args"; then
      stop_pid "$process_pid"
    fi
  done < <(ps -eo pid=,args=)
}

stop_session() {
  local session_id="${1:-}"
  if [[ -z "$session_id" ]]; then
    echo "usage: scripts/dev.sh stop-session <session-id>" >&2
    return 1
  fi

  local pid_file
  pid_file="$(session_pid_file "$session_id")"
  if [[ -f "$pid_file" ]]; then
    while read -r _process_name process_pid; do
      [[ -n "${process_pid:-}" ]] || continue
      stop_pid "$process_pid"
    done <"$pid_file"
    rm -f "$pid_file"
  fi

  if [[ -f "$CURRENT_SESSION_FILE" ]] && [[ "$(<"$CURRENT_SESSION_FILE")" == "$session_id" ]]; then
    rm -f "$CURRENT_SESSION_FILE"
  fi

  rmdir "$DEV_STATE_DIR" >/dev/null 2>&1 || true
}

stop_current_session() {
  local session_id
  if ! session_id="$(current_session_id)"; then
    stop_stale_dev_processes
    return 0
  fi

  stop_session "$session_id"
  stop_stale_dev_processes
}

prepare_go_env() {
  # 把 go build 缓存固定到稳定目录，便于 stop 逻辑识别临时编译产物。
  export GOCACHE="${GOCACHE:-/tmp/article-sentinel-go-cache}"
  mkdir -p "$GOCACHE"
}

print_plan() {
  cat <<PLAN
make dev must start:
- stop the previous dev stack first: bash scripts/dev.sh stop
- print backend/admin endpoints first: bash scripts/dev.sh print-endpoints
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
    "scripts/dev.sh stop" \
    "scripts/dev.sh print-endpoints" \
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

ensure_admin_deps() {
  if [[ -x "$ADMIN_DIR/node_modules/.bin/vite" ]]; then
    return 0
  fi

  echo "[dev] admin dependencies missing, running npm ci..." >&2
  npm ci
}

run_admin() {
  cd "$ADMIN_DIR"
  ensure_admin_deps
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
  stop)
    stop_current_session
    ;;
  print-endpoints)
    print_endpoints
    ;;
  print-plan)
    print_plan
    ;;
  assert-make-dev)
    assert_make_dev
    ;;
  start-session)
    start_session
    ;;
  register-dev-pid)
    register_dev_pid "${2:-}" "${3:-}" "${4:-}"
    ;;
  stop-session)
    stop_session "${2:-}"
    ;;
  *)
    usage >&2
    exit 1
    ;;
esac
