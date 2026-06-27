#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

eval "$("$ROOT/script/goreleaser-env.sh")"
exec goreleaser "$@"
