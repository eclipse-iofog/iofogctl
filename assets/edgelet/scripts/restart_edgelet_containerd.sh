#!/bin/sh
# Restart edgelet-containerd without blocking wait — Go polls via probe_containerd_ready.sh.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/service.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

export EDGELET_START_NO_WAIT=1
restart_edgelet_containerd_service
