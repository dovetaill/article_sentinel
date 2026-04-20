#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-/tmp/article-sentinel-go-cache}"
mkdir -p "$GOCACHE"

# Keep verification aligned with the canonical starter suite.
go test ./...
