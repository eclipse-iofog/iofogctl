#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

eval "$("$ROOT/script/goreleaser-env.sh")"

cfg="$(mktemp)"
trap 'rm -f "$cfg"' EXIT
sed "s/__CLI_BINARY_NAME__/${CLI_BINARY_NAME}/g" .goreleaser.yml >"$cfg"
exec goreleaser -f "$cfg" "$@"
