#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEV_SCRIPT="$ROOT_DIR/scripts/dev.sh"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" != *"$needle"* ]]; then
    fail "expected output to contain: $needle"
  fi
}

test_make_dev_stops_previous_stack_first() {
  local output
  output="$(cd "$ROOT_DIR" && make -n dev)"

  assert_contains "$output" "scripts/dev.sh stop"
}

test_make_dev_prints_backend_endpoints() {
  local output
  output="$(cd "$ROOT_DIR" && make -n dev)"

  assert_contains "$output" "scripts/dev.sh print-endpoints"
}

test_print_endpoints_reports_backend_and_jump_login_urls() {
  local temp_config
  temp_config="$(mktemp)"
  cat >"$temp_config" <<'YAML'
app:
  host: 0.0.0.0
  port: 18080
YAML

  local output
  output="$(CONFIG="$temp_config" "$DEV_SCRIPT" print-endpoints)"

  assert_contains "$output" "Admin jump login: http://127.0.0.1:5173/auth/login?jwt=<legacy-jwt>"
  assert_contains "$output" "Backend API: http://127.0.0.1:18080"
  assert_contains "$output" "Jump login: http://127.0.0.1:18080/auth/login?jwt=<legacy-jwt>"

  rm -f "$temp_config"
}

test_admin_vite_dev_server_proxies_auth_routes() {
  local content
  content="$(cat "$ROOT_DIR/web/admin/vite.config.ts")"

  assert_contains "$content" "'/auth':"
}

test_admin_bootstraps_dependencies_when_vite_is_missing() {
  local admin_node_modules="$ROOT_DIR/web/admin/node_modules"
  local backup_dir=""
  if [[ -d "$admin_node_modules" ]]; then
    backup_dir="$(mktemp -d)"
    mv "$admin_node_modules" "$backup_dir/node_modules"
  fi

  local fake_bin
  fake_bin="$(mktemp -d)"
  local npm_log
  npm_log="$(mktemp)"
  cat >"$fake_bin/npm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"$NPM_LOG"

if [[ "${1:-}" == "ci" ]]; then
  mkdir -p node_modules/.bin
  cat >node_modules/.bin/vite <<'INNER'
#!/usr/bin/env bash
exit 0
INNER
  chmod +x node_modules/.bin/vite
  exit 0
fi

if [[ "${1:-}" == "run" && "${2:-}" == "dev" ]]; then
  exit 0
fi

echo "unexpected npm invocation: $*" >&2
exit 1
EOF
  chmod +x "$fake_bin/npm"

  cleanup() {
    rm -rf "$admin_node_modules" "$fake_bin"
    rm -f "$npm_log"
    if [[ -n "$backup_dir" && -d "$backup_dir/node_modules" ]]; then
      mv "$backup_dir/node_modules" "$admin_node_modules"
      rmdir "$backup_dir"
    fi
  }
  trap cleanup EXIT

  PATH="$fake_bin:$PATH" NPM_LOG="$npm_log" "$DEV_SCRIPT" admin >/dev/null 2>&1

  local log
  log="$(cat "$npm_log")"
  assert_contains "$log" "ci"
  assert_contains "$log" "run dev -- --host 0.0.0.0"

  trap - EXIT
  cleanup
}

test_config_example_uses_insecure_cookie_for_http_local_dev() {
  local content
  content="$(cat "$ROOT_DIR/configs/config.example.yaml")"

  assert_contains "$content" "secure_cookie: false"
}

test_stop_kills_registered_processes() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  local session_id
  session_id="$("$DEV_SCRIPT" start-session)"

  sleep 30 &
  local sleeper_pid=$!

  cleanup() {
    kill "$sleeper_pid" >/dev/null 2>&1 || true
    wait "$sleeper_pid" >/dev/null 2>&1 || true
    "$DEV_SCRIPT" stop >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  "$DEV_SCRIPT" register-dev-pid "$session_id" test "$sleeper_pid"
  "$DEV_SCRIPT" stop

  if kill -0 "$sleeper_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate pid $sleeper_pid"
  fi

  trap - EXIT
}

test_stop_kills_stale_server_process_without_session_file() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  bash -c 'exec -a article-sentinel-server sleep 30' &
  local stale_pid=$!

  cleanup() {
    kill "$stale_pid" >/dev/null 2>&1 || true
    wait "$stale_pid" >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  "$DEV_SCRIPT" stop

  if kill -0 "$stale_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate stale article-sentinel-server pid $stale_pid"
  fi

  trap - EXIT
}

test_stop_kills_stale_go_run_temp_server_without_session_file() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  (
    cd "$ROOT_DIR"
    exec bash -c 'exec -a /tmp/go-build123/b001/exe/server sleep 30'
  ) &
  local stale_pid=$!

  cleanup() {
    kill "$stale_pid" >/dev/null 2>&1 || true
    wait "$stale_pid" >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  "$DEV_SCRIPT" stop

  if kill -0 "$stale_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate stale temp go-run server pid $stale_pid"
  fi

  trap - EXIT
}

test_stop_kills_stale_go_run_cache_server_without_session_file() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  (
    cd "$ROOT_DIR"
    exec bash -c 'exec -a /tmp/article-sentinel-go-cache/fakehash/server sleep 30'
  ) &
  local stale_pid=$!

  cleanup() {
    kill "$stale_pid" >/dev/null 2>&1 || true
    wait "$stale_pid" >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  "$DEV_SCRIPT" stop

  if kill -0 "$stale_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate stale go-run cache server pid $stale_pid"
  fi

  trap - EXIT
}

test_make_dev_stops_previous_stack_first
test_stop_kills_registered_processes
test_stop_kills_stale_server_process_without_session_file
test_stop_kills_stale_go_run_temp_server_without_session_file
test_stop_kills_stale_go_run_cache_server_without_session_file
test_make_dev_prints_backend_endpoints
test_print_endpoints_reports_backend_and_jump_login_urls
test_admin_vite_dev_server_proxies_auth_routes
test_admin_bootstraps_dependencies_when_vite_is_missing
test_config_example_uses_insecure_cookie_for_http_local_dev

echo "dev script tests passed"
