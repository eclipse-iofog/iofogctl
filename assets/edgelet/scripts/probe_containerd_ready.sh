#!/bin/sh
# One-shot: exit 0 if containerd socket ready for edgelet embedded engine.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/service.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

if edgelet_containerd_socket_ready; then
	exit 0
fi
echo "containerd socket not ready" >&2
exit 1
