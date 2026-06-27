#!/usr/bin/env bash
# Export version pins (versions.mk) and flavor ldflags for goreleaser.
# Flavor resolution (first match wins):
#   1. FLAVOR env (manual override)
#   2. GITHUB_REPOSITORY (CI / local release smoke)
#   3. default iofog
#
# Usage:
#   eval "$(script/goreleaser-env.sh)"
#   GITHUB_REPOSITORY=Datasance/potctl script/goreleaser-release.sh release --clean
#   FLAVOR=datasance script/goreleaser-release.sh release --snapshot --clean
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

resolve_flavor() {
  if [ -n "${FLAVOR:-}" ]; then
    echo "$FLAVOR"
    return
  fi
  case "${GITHUB_REPOSITORY:-}" in
    Datasance/potctl) echo datasance ;;
    eclipse-iofog/iofogctl) echo iofog ;;
    *) echo iofog ;;
  esac
}

FLAVOR="$(resolve_flavor)"

while IFS= read -r line; do
  case "$line" in
    ''|\#*) continue ;;
    *' ?='*)
      key="${line%% ?=*}"
      val="${line#* ?= }"
      export "$key=$val"
      ;;
  esac
done < versions.mk

case "$FLAVOR" in
  datasance)
    export CLI_BINARY_NAME=potctl
    export CLI_CRD_GROUP=datasance.com
    export CLI_API_VERSION=datasance.com/v3
    export CLI_CP_CR_NAME=pot
    export IMAGE_REGISTRY=ghcr.io/datasance
    export CLI_DOCS_URL=https://docs.datasance.com
    export PACKAGE_REPO_BASE=https://downloads.datasance.com
    export OCI_SOURCE_REPO=https://github.com/Datasance/potctl
    export EDGELET_RELEASE_BASE=https://github.com/Datasance/edgelet/releases/download
    export EDGELET_GITHUB_REPO=Datasance/edgelet
    ;;
  iofog)
    export CLI_BINARY_NAME=iofogctl
    export CLI_CRD_GROUP=iofog.org
    export CLI_API_VERSION=iofog.org/v3
    export CLI_CP_CR_NAME=iofog
    export IMAGE_REGISTRY=ghcr.io/eclipse-iofog
    export CLI_DOCS_URL=https://iofog.org
    export PACKAGE_REPO_BASE=https://iofog.datasance.com
    export OCI_SOURCE_REPO=https://github.com/eclipse-iofog/iofogctl
    export EDGELET_RELEASE_BASE=https://github.com/eclipse-iofog/edgelet/releases/download
    export EDGELET_GITHUB_REPO=eclipse-iofog/edgelet
    ;;
  *)
    echo "FLAVOR must be iofog or datasance (got: $FLAVOR)" >&2
    exit 1
    ;;
esac

export FLAVOR

if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
  return 0 2>/dev/null || exit 0
fi

vars=(
  FLAVOR
  OPERATOR_VERSION CONTROLLER_VERSION ROUTER_VERSION NATS_VERSION
  EDGELET_BINARY_VERSION EDGELET_IMAGE_TAG
  CLI_BINARY_NAME CLI_CRD_GROUP CLI_API_VERSION CLI_CP_CR_NAME
  IMAGE_REGISTRY CLI_DOCS_URL PACKAGE_REPO_BASE OCI_SOURCE_REPO
  EDGELET_RELEASE_BASE EDGELET_GITHUB_REPO
)

for v in "${vars[@]}"; do
  printf 'export %s=%q\n' "$v" "${!v}"
done
