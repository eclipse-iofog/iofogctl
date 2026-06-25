#!/usr/bin/env bash
set -euo pipefail

if [ -z "${OPERATOR_VERSION:-}" ]; then
  echo "Load release env first: eval \"\$(FLAVOR=${FLAVOR:-iofog} script/goreleaser-env.sh)\"" >&2
  exit 1
fi
