#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <module_name>"
  echo "Example: $0 article"
  exit 1
fi

module_name="$1"
if [[ ! "$module_name" =~ ^[a-z][a-z0-9_]*$ ]]; then
  echo "module_name must match ^[a-z][a-z0-9_]*$"
  exit 1
fi

src_dir="internal/modules/post"
dst_dir="internal/modules/${module_name}"
if [[ ! -d "$src_dir" ]]; then
  echo "source module not found: $src_dir"
  exit 1
fi
if [[ -e "$dst_dir" ]]; then
  echo "target already exists: $dst_dir"
  exit 1
fi

module_type="$(awk -F'_' '{for(i=1;i<=NF;i++){printf toupper(substr($i,1,1)) substr($i,2)}}' <<<"$module_name")"
module_plural="${module_name}s"
module_type_plural="${module_type}s"

cp -R "$src_dir" "$dst_dir"

for file in "$dst_dir"/*.go; do
  perl -0pi -e "s/\\bPosts\\b/${module_type_plural}/g; s/\\bposts\\b/${module_plural}/g; s/\\bPost\\b/${module_type}/g; s/\\bpost\\b/${module_name}/g" "$file"
done

if [[ -f "$dst_dir/post_test.go" ]]; then
  mv "$dst_dir/post_test.go" "$dst_dir/${module_name}_test.go"
fi

cat <<MSG
Created module skeleton: $dst_dir

This script only does obvious token replacement. Complete these manual steps next:
1) Review model naming, table name, and field semantics in model.go.
2) Review repository queries and indexes for your domain.
3) Review service validation rules and domain errors.
4) Review handler routes and response messages.
5) Register routes in internal/api/register/router.go.
6) Register model in internal/app/bootstrap/schema.go.
7) Extend module tests beyond the copied baseline.
MSG
