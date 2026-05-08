#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: write_release_manifest.sh \
  --package-root <path> \
  --app <name> \
  --version <value> \
  --git-sha <sha> \
  --build-time <timestamp> \
  --target-os <os> \
  --target-arch <arch>
EOF
}

fail() {
  printf '[write-release-manifest] error: %s\n' "$*" >&2
  exit 1
}

main() {
  local package_root=""
  local app=""
  local version=""
  local git_sha=""
  local build_time=""
  local target_os=""
  local target_arch=""
  local manifest_path=""
  local checksum_path=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --package-root)
        package_root="${2:-}"
        shift 2
        ;;
      --app)
        app="${2:-}"
        shift 2
        ;;
      --version)
        version="${2:-}"
        shift 2
        ;;
      --git-sha)
        git_sha="${2:-}"
        shift 2
        ;;
      --build-time)
        build_time="${2:-}"
        shift 2
        ;;
      --target-os)
        target_os="${2:-}"
        shift 2
        ;;
      --target-arch)
        target_arch="${2:-}"
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

  [[ -d "$package_root" ]] || fail "package root not found: $package_root"
  [[ -n "$app" ]] || fail "--app is required"
  [[ -n "$version" ]] || fail "--version is required"
  [[ -n "$git_sha" ]] || fail "--git-sha is required"
  [[ -n "$build_time" ]] || fail "--build-time is required"
  [[ -n "$target_os" ]] || fail "--target-os is required"
  [[ -n "$target_arch" ]] || fail "--target-arch is required"

  manifest_path="$package_root/manifest.json"
  checksum_path="manifest.sha256"

  python3 - "$manifest_path" "$app" "$version" "$git_sha" "$build_time" "$target_os" "$target_arch" <<'PY'
import json
import sys

manifest_path, app, version, git_sha, build_time, target_os, target_arch = sys.argv[1:8]
payload = {
    "app": app,
    "version": version,
    "git_sha": git_sha,
    "build_time": build_time,
    "target_os": target_os,
    "target_arch": target_arch,
}
with open(manifest_path, "w", encoding="utf-8") as fh:
    json.dump(payload, fh, indent=2, sort_keys=True)
    fh.write("\n")
PY

  (
    cd "$package_root"
    find . -type f ! -name 'manifest.sha256' -print0 \
      | sort -z \
      | while IFS= read -r -d '' file; do
          file="${file#./}"
          sha256sum "$file"
        done >"$checksum_path"
  )
}

main "$@"
