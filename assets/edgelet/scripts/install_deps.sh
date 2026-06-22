#!/bin/sh
# Install deps layer. Skips when containerEngine=edgelet.
set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
CONTAINER_ENGINE="${1:-edgelet}"
DEPLOYMENT_TYPE="${2:-native}"

if [ "$CONTAINER_ENGINE" = "edgelet" ]; then
	echo "# Skipping install_deps: containerEngine=edgelet (deploymentType=$DEPLOYMENT_TYPE)"
	exit 0
fi

export CONTAINER_ENGINE
EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

if [ "$EDGELET_OS" = "darwin" ] || [ "$EDGELET_OS" = "windows" ]; then
	echo "# Skipping configure_container_engine on ${EDGELET_OS} (desktop container runtime is user-managed)"
	exit 0
fi

. "$SCRIPT_DIR/configure_container_engine.sh"
