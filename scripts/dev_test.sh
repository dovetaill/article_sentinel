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

test_make_dev_stops_previous_stack_first
test_stop_kills_registered_processes

echo "dev script tests passed"
