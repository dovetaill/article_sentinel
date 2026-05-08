#!/usr/bin/env bash
set -euo pipefail

APP_NAME="${APP_NAME:-article-sentinel}"
DEPLOY_ROOT="${ARTICLE_SENTINEL_DEPLOY_ROOT:-/srv/article-sentinel}"
RELEASES_DIR="$DEPLOY_ROOT/releases"

usage() {
  cat <<'EOF'
Usage: install-release.sh --tarball <path> --version <version> [--checksum <path>]

Install a release tarball into /srv/article-sentinel/releases/<version> without
switching the live current symlink.

Options:
  --tarball <path>    Release tarball produced by make release
  --version <value>   Release version to install under releases/<version>
  --checksum <path>   Tarball checksum file (default: <tarball>.sha256)
  --help              Show this help text

Environment:
  ARTICLE_SENTINEL_DEPLOY_ROOT   Override the deploy root (default: /srv/article-sentinel)
EOF
}

log() {
  printf '[install-release] %s\n' "$*"
}

fail() {
  printf '[install-release] error: %s\n' "$*" >&2
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

abs_path() {
  local path="$1"
  python3 - "$path" <<'PY'
import os
import sys

print(os.path.abspath(sys.argv[1]))
PY
}

verify_checksum() {
  local tarball="$1"
  local checksum_file="$2"
  local expected actual

  [[ -f "$checksum_file" ]] || fail "checksum file not found: $checksum_file"
  expected="$(awk 'NR == 1 { print $1 }' "$checksum_file")"
  [[ -n "$expected" ]] || fail "checksum file is empty: $checksum_file"

  actual="$(sha256sum "$tarball" | awk '{ print $1 }')"
  [[ "$expected" == "$actual" ]] || fail "checksum mismatch for $tarball"
}

extract_release() {
  local tarball="$1"
  local version="$2"
  local final_release="$RELEASES_DIR/$version"
  local stage_root extracted_root

  [[ ! -e "$final_release" ]] || fail "release already exists: $final_release"

  mkdir -p "$RELEASES_DIR"
  stage_root="$(mktemp -d "${TMPDIR:-/tmp}/${APP_NAME}-install-${version}.XXXXXX")"
  trap 'rm -rf "$stage_root"' EXIT

  tar -xzf "$tarball" -C "$stage_root" --no-same-owner

  mapfile -t extracted_dirs < <(find "$stage_root" -mindepth 1 -maxdepth 1 -type d | sort)
  [[ "${#extracted_dirs[@]}" -eq 1 ]] || fail "expected exactly one top-level directory in archive"

  extracted_root="${extracted_dirs[0]}"
  mv "$extracted_root" "$final_release"
  trap - EXIT
  rm -rf "$stage_root"
}

main() {
  local tarball=""
  local version=""
  local checksum_file=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --tarball)
        tarball="${2:-}"
        shift 2
        ;;
      --version)
        version="${2:-}"
        shift 2
        ;;
      --checksum)
        checksum_file="${2:-}"
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

  [[ -n "$tarball" ]] || fail "--tarball is required"
  [[ -n "$version" ]] || fail "--version is required"

  tarball="$(abs_path "$tarball")"
  [[ -f "$tarball" ]] || fail "tarball not found: $tarball"

  if [[ -z "$checksum_file" ]]; then
    checksum_file="${tarball}.sha256"
  fi
  checksum_file="$(abs_path "$checksum_file")"

  require_root
  require_cmd awk
  require_cmd find
  require_cmd mktemp
  require_cmd python3
  require_cmd sha256sum
  require_cmd tar

  verify_checksum "$tarball" "$checksum_file"
  extract_release "$tarball" "$version"

  log "installed release at $RELEASES_DIR/$version"
  log "current and previous remain unchanged until activate-release.sh succeeds"
}

main "$@"
