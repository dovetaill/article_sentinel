#!/usr/bin/env bash
set -euo pipefail

APP_NAME="${APP_NAME:-article-sentinel}"
DEPLOY_ROOT="${ARTICLE_SENTINEL_DEPLOY_ROOT:-/srv/article-sentinel}"
SMOKE_BASE_URL="${ARTICLE_SENTINEL_SMOKE_BASE_URL:-http://127.0.0.1}"
RELEASES_DIR="$DEPLOY_ROOT/releases"
CURRENT_LINK="$DEPLOY_ROOT/current"
PREVIOUS_LINK="$DEPLOY_ROOT/previous"

usage() {
  cat <<'EOF'
Usage: rollback-release.sh [--version <version>]

Rollback /srv/article-sentinel/current to an explicit release version or, if
omitted, to the version currently referenced by /srv/article-sentinel/previous.

Options:
  --version <value>   Release version to rollback to
  --help              Show this help text

Environment:
  ARTICLE_SENTINEL_DEPLOY_ROOT    Override the deploy root (default: /srv/article-sentinel)
  ARTICLE_SENTINEL_SMOKE_BASE_URL Smoke-check base URL (default: http://127.0.0.1)
EOF
}

log() {
  printf '[rollback-release] %s\n' "$*"
}

fail() {
  printf '[rollback-release] error: %s\n' "$*" >&2
  exit 1
}

require_root() {
  if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
    fail "please run as root"
  fi
}

require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || fail "required command not found: $cmd"
}

set_symlink() {
  local target="$1"
  local link_path="$2"
  local tmp_link="${link_path}.tmp"

  ln -sfn "$target" "$tmp_link"
  mv -Tf "$tmp_link" "$link_path"
}

restart_services() {
  systemctl daemon-reload
  systemctl restart \
    article-sentinel-server.service \
    article-sentinel-worker.service \
    article-sentinel-scheduler.service
}

smoke_check() {
  local base_url="${SMOKE_BASE_URL%/}"
  local attempt=1
  local max_attempts=15

  while (( attempt <= max_attempts )); do
    if curl -fsS --max-time 5 -o /dev/null "${base_url}/healthz" \
      && curl -fsS --max-time 5 -o /dev/null "${base_url}/readyz" \
      && curl -fsS --max-time 5 -o /dev/null "${base_url}/"; then
      log "smoke checks passed against ${base_url}"
      return 0
    fi

    sleep 2
    attempt=$((attempt + 1))
  done

  fail "smoke checks failed against ${base_url}"
}

main() {
  local version=""
  local target_path=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --version)
        version="${2:-}"
        shift 2
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        fail "unknown argument: $1"
        ;;
    esac
  done

  if [[ -n "$version" ]]; then
    target_path="$RELEASES_DIR/$version"
  elif [[ -L "$PREVIOUS_LINK" ]]; then
    target_path="$(readlink -f "$PREVIOUS_LINK")"
  else
    fail "no rollback target found; provide --version or create $PREVIOUS_LINK"
  fi

  [[ -d "$target_path" ]] || fail "rollback target not found: $target_path"

  require_root
  require_cmd curl
  require_cmd mv
  require_cmd readlink
  require_cmd systemctl

  set_symlink "$target_path" "$CURRENT_LINK"
  restart_services
  smoke_check

  log "rolled back current -> $target_path"
  log "previous remains unchanged to avoid pointing back at a failed release"
}

main "$@"
