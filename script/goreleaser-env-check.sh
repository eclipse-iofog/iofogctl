#!/usr/bin/env bash
set -euo pipefail

if [ -z "${OPERATOR_VERSION:-}" ]; then
  echo "Load release env first: eval \"\$(script/goreleaser-env.sh)\" (set GITHUB_REPOSITORY or FLAVOR to pick mirror)" >&2
  exit 1
fi
