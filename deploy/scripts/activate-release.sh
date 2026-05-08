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
Usage: activate-release.sh --version <version>

Run the release-bound migration for <version>, switch /srv/article-sentinel/current
to that release, restart long-running services, then run smoke checks.

Options:
  --version <value>   Release version to activate
  --help              Show this help text

Environment:
  ARTICLE_SENTINEL_DEPLOY_ROOT   Override the deploy root (default: /srv/article-sentinel)
  ARTICLE_SENTINEL_SMOKE_BASE_URL  Smoke-check base URL (default: http://127.0.0.1)
EOF
}

log() {
  printf '[activate-release] %s\n' "$*"
}

fail() {
  printf '[activate-release] error: %s\n' "$*" >&2
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

read_link_target() {
  local link_path="$1"
  if [[ -L "$link_path" ]]; then
    readlink -f "$link_path"
  fi
}

set_symlink() {
  local target="$1"
  local link_path="$2"
  local tmp_link="${link_path}.tmp"

  ln -sfn "$target" "$tmp_link"
  mv -Tf "$tmp_link" "$link_path"
}

run_migration_if_needed() {
  local version="$1"
  local release_path="$2"
  local stamp_file="$release_path/.migration-complete"

  if [[ -f "$stamp_file" ]]; then
    log "migration already recorded for $version; skipping"
    return
  fi

  systemctl start "article-sentinel-migrate@${version}.service"
  touch "$stamp_file"
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
  local release_path=""
  local old_current=""

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

  [[ -n "$version" ]] || fail "--version is required"
  release_path="$RELEASES_DIR/$version"
  [[ -d "$release_path" ]] || fail "release not installed: $release_path"

  require_root
  require_cmd curl
  require_cmd mv
  require_cmd readlink
  require_cmd systemctl

  old_current="$(read_link_target "$CURRENT_LINK" || true)"

  run_migration_if_needed "$version" "$release_path"
  set_symlink "$release_path" "$CURRENT_LINK"
  restart_services
  smoke_check

  if [[ -n "$old_current" && "$old_current" != "$release_path" && -d "$old_current" ]]; then
    set_symlink "$old_current" "$PREVIOUS_LINK"
    log "updated previous -> $old_current"
  else
    log "no previous release update required"
  fi

  log "activated release $version"
}

main "$@"
