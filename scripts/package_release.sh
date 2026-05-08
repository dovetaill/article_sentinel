#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: package_release.sh \
  --package-root <path> \
  --bin-dir <path> \
  --admin-dir <path> \
  --migrations-dir <path> \
  --config-file <path> \
  --deploy-dir <path>
EOF
}

fail() {
  printf '[package-release] error: %s\n' "$*" >&2
  exit 1
}

copy_sql_migrations() {
  local source_dir="$1"
  local dest_dir="$2"
  local copied=0

  shopt -s nullglob
  for migration in "$source_dir"/*.sql; do
    cp "$migration" "$dest_dir"/
    copied=1
  done
  shopt -u nullglob

  [[ "$copied" -eq 1 ]] || fail "no SQL migrations found in $source_dir"
}

main() {
  local package_root=""
  local bin_dir=""
  local admin_dir=""
  local migrations_dir=""
  local config_file=""
  local deploy_dir=""
  local package_parent=""
  local package_name=""
  local stage_root=""
  local stage_package=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --package-root)
        package_root="${2:-}"
        shift 2
        ;;
      --bin-dir)
        bin_dir="${2:-}"
        shift 2
        ;;
      --admin-dir)
        admin_dir="${2:-}"
        shift 2
        ;;
      --migrations-dir)
        migrations_dir="${2:-}"
        shift 2
        ;;
      --config-file)
        config_file="${2:-}"
        shift 2
        ;;
      --deploy-dir)
        deploy_dir="${2:-}"
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

  [[ -n "$package_root" ]] || fail "--package-root is required"
  [[ -d "$bin_dir" ]] || fail "bin dir not found: $bin_dir"
  [[ -d "$admin_dir" ]] || fail "admin dir not found: $admin_dir"
  [[ -d "$migrations_dir" ]] || fail "migrations dir not found: $migrations_dir"
  [[ -f "$config_file" ]] || fail "config file not found: $config_file"
  [[ -d "$deploy_dir" ]] || fail "deploy dir not found: $deploy_dir"

  package_parent="$(dirname "$package_root")"
  package_name="$(basename "$package_root")"
  stage_root="$(mktemp -d "${TMPDIR:-/tmp}/article-sentinel-package.XXXXXX")"
  stage_package="$stage_root/$package_name"

  trap 'rm -rf "$stage_root"' EXIT

  rm -rf "$package_root"
  mkdir -p "$package_parent"
  mkdir -p "$stage_package"/bin "$stage_package"/admin "$stage_package"/migrations "$stage_package"/configs "$stage_package"/deploy

  cp "$bin_dir"/* "$stage_package/bin"/
  cp -R "$admin_dir"/. "$stage_package/admin"/
  copy_sql_migrations "$migrations_dir" "$stage_package/migrations"
  cp "$config_file" "$stage_package/configs"/
  cp -R "$deploy_dir"/. "$stage_package/deploy"/

  mv "$stage_package" "$package_root"
  trap - EXIT
  rm -rf "$stage_root"
}

main "$@"
