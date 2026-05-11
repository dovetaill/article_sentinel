#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEV_SCRIPT="$ROOT_DIR/scripts/dev.sh"
ADMIN_DIR="$ROOT_DIR/web/admin"
TEST_TMPDIR="$ROOT_DIR/.tmp/dev-test"

mkdir -p "$TEST_TMPDIR"
export TMPDIR="$TEST_TMPDIR"

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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" == *"$needle"* ]]; then
    fail "expected output to NOT contain: $needle"
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

test_print_endpoints_prefers_running_admin_port() {
  local temp_config
  temp_config="$(mktemp)"
  cat >"$temp_config" <<'YAML'
app:
  host: 0.0.0.0
  port: 18080
YAML

  local admin_port
  admin_port="$(
    python - <<'PY'
import socket

sock = socket.socket()
sock.bind(("127.0.0.1", 0))
print(sock.getsockname()[1])
sock.close()
PY
  )"

  (
    cd "$ADMIN_DIR"
    exec bash -c "exec -a '$ADMIN_DIR/node_modules/@umijs/max/bin/max.js dev --port 5173 --host 0.0.0.0' python -m http.server $admin_port --bind 127.0.0.1 >/dev/null 2>&1"
  ) &
  local admin_pid=$!

  cleanup() {
    kill "$admin_pid" >/dev/null 2>&1 || true
    wait "$admin_pid" >/dev/null 2>&1 || true
    rm -f "$temp_config"
  }
  trap cleanup EXIT

  local tries=50
  while (( tries > 0 )); do
    if ss -ltnp | rg -q ":${admin_port}\\b"; then
      break
    fi
    tries=$((tries - 1))
    sleep 0.1
  done

  local output
  output="$(CONFIG="$temp_config" "$DEV_SCRIPT" print-endpoints)"

  assert_contains "$output" "Admin UI: http://127.0.0.1:${admin_port}"
  assert_contains "$output" "Admin jump login: http://127.0.0.1:${admin_port}/auth/login?jwt=<legacy-jwt>"
  assert_contains "$output" "Backend API: http://127.0.0.1:18080"

  trap - EXIT
  cleanup
}

test_readme_describes_umi_pro_admin_shell() {
  local content
  content="$(cat "$ROOT_DIR/README.md")"

  assert_contains "$content" "React + Umi Max + ant-design-pro"
  assert_not_contains "$content" "React + Vite"
}

test_admin_uses_umi_max_dev_server_settings() {
  local content
  content="$(cat "$ROOT_DIR/web/admin/package.json")"

  assert_contains "$content" "@umijs/max/bin/max.js dev --port 5173 --host 0.0.0.0"
}

test_admin_proxy_configuration_covers_auth_routes() {
  local content
  content="$(cat "$ROOT_DIR/web/admin/config/proxy.ts")"

  assert_contains "$content" "ADMIN_API_BASE_URL"
  assert_contains "$content" "'/api'"
  assert_contains "$content" "'/auth':"
}

test_admin_dev_script_drops_vite_specific_wording() {
  local content
  content="$(cat "$ROOT_DIR/scripts/dev.sh")"

  assert_not_contains "$content" "Vite dev server"
  assert_contains "$content" "admin Umi Max dev server"
}

test_admin_bootstraps_dependencies_when_max_runtime_is_missing() {
  local admin_max_runtime_dir="$ROOT_DIR/web/admin/node_modules/@umijs/max/bin"
  local admin_max_runtime="$admin_max_runtime_dir/max.js"
  local backup_dir=""
  if [[ -e "$admin_max_runtime" ]]; then
    backup_dir="$(mktemp -d)"
    mv "$admin_max_runtime" "$backup_dir/max.js"
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
  mkdir -p node_modules/@umijs/max/bin
  cat >node_modules/@umijs/max/bin/max.js <<'INNER'
#!/usr/bin/env bash
exit 0
INNER
  chmod +x node_modules/@umijs/max/bin/max.js
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
    rm -f "$admin_max_runtime"
    rm -rf "$fake_bin"
    rm -f "$npm_log"
    if [[ -n "$backup_dir" && -e "$backup_dir/max.js" ]]; then
      mkdir -p "$admin_max_runtime_dir"
      mv "$backup_dir/max.js" "$admin_max_runtime"
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

  setsid bash -c 'exec sleep 30' &
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

test_stop_kills_registered_processes_when_tmpdir_changes() {
  mkdir -p "$TEST_TMPDIR/a" "$TEST_TMPDIR/b"
  TMPDIR="$TEST_TMPDIR/a" "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  local session_id
  session_id="$(TMPDIR="$TEST_TMPDIR/a" "$DEV_SCRIPT" start-session)"

  setsid bash -c 'exec sleep 30' &
  local sleeper_pid=$!

  cleanup() {
    kill "$sleeper_pid" >/dev/null 2>&1 || true
    wait "$sleeper_pid" >/dev/null 2>&1 || true
    TMPDIR="$TEST_TMPDIR/a" "$DEV_SCRIPT" stop >/dev/null 2>&1 || true
    TMPDIR="$TEST_TMPDIR/b" "$DEV_SCRIPT" stop >/dev/null 2>&1 || true
  }
  trap cleanup EXIT

  TMPDIR="$TEST_TMPDIR/a" "$DEV_SCRIPT" register-dev-pid "$session_id" test "$sleeper_pid"
  TMPDIR="$TEST_TMPDIR/b" "$DEV_SCRIPT" stop

  if kill -0 "$sleeper_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate pid $sleeper_pid even if TMPDIR changed"
  fi

  trap - EXIT
}

test_stop_kills_stale_server_process_without_session_file() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  setsid bash -c 'exec -a article-sentinel-server sleep 30' &
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
    exec setsid bash -c 'exec -a /tmp/go-build123/b001/exe/server sleep 30'
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
    exec setsid bash -c 'exec -a /tmp/article-sentinel-go-cache/fakehash/server sleep 30'
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

test_stop_kills_orphaned_admin_child_process_without_session_file() {
  "$DEV_SCRIPT" stop >/dev/null 2>&1 || true

  local child_pid_file
  child_pid_file="$(mktemp)"

  (
    cd "$ADMIN_DIR"
    CHILD_PID_FILE="$child_pid_file" setsid bash -c '
      bash -c '"'"'
        trap "" TERM
        sleep 30 &
        echo $! >"$CHILD_PID_FILE"
        wait
      '"'"' "@umijs/max/bin/max.js dev --port 5173 --host 0.0.0.0" &
    ' >/dev/null 2>&1
  ) || true

  local child_pid=""
  local tries=50
  while (( tries > 0 )); do
    if [[ -s "$child_pid_file" ]]; then
      child_pid="$(<"$child_pid_file")"
      break
    fi
    tries=$((tries - 1))
    sleep 0.1
  done

  if [[ -z "$child_pid" ]]; then
    rm -f "$child_pid_file"
    fail "expected orphaned admin child pid to be recorded"
  fi

  cleanup() {
    kill "$child_pid" >/dev/null 2>&1 || true
    wait "$child_pid" >/dev/null 2>&1 || true
    rm -f "$child_pid_file"
  }
  trap cleanup EXIT

  "$DEV_SCRIPT" stop

  if kill -0 "$child_pid" >/dev/null 2>&1; then
    fail "expected stop to terminate orphaned admin child pid $child_pid"
  fi

  trap - EXIT
  cleanup
}

test_make_dev_stops_previous_stack_first
test_stop_kills_registered_processes
test_stop_kills_registered_processes_when_tmpdir_changes
test_stop_kills_stale_server_process_without_session_file
test_stop_kills_stale_go_run_temp_server_without_session_file
test_stop_kills_stale_go_run_cache_server_without_session_file
test_stop_kills_orphaned_admin_child_process_without_session_file
test_make_dev_prints_backend_endpoints
test_print_endpoints_reports_backend_and_jump_login_urls
test_print_endpoints_prefers_running_admin_port
test_readme_describes_umi_pro_admin_shell
test_admin_uses_umi_max_dev_server_settings
test_admin_proxy_configuration_covers_auth_routes
test_admin_dev_script_drops_vite_specific_wording
test_admin_bootstraps_dependencies_when_max_runtime_is_missing
test_config_example_uses_insecure_cookie_for_http_local_dev

echo "dev script tests passed"
