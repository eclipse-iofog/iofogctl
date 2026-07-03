#!/bin/sh
# Apply agent spec to a running edgelet container on desktop hosts, then restart the container.
set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/container_mounts.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

if [ "${EDGELET_INSTALL_MODE:-native}" != "container" ]; then
	exit 0
fi
if ! is_desktop_container_host; then
	exit 0
fi

wait_edgelet_api_running() {
	_iter=0
	_max="${EDGELET_CONFIG_TIMEOUT:-120}"
	while [ "$_iter" -lt "$_max" ]; do
		_out=""
		_out=$(edgelet system status 2>&1) || true
		if echo "$_out" | grep -qi 'daemon is not running'; then
			sleep 1
			_iter=$((_iter + 1))
			continue
		fi
		if echo "$_out" | grep -q 'edgeletDaemon: RUNNING'; then
			return 0
		fi
		if echo "$_out" | grep -q 'runtime.agentPhase: running'; then
			return 0
		fi
		sleep 1
		_iter=$((_iter + 1))
	done
	die "Timed out after ${_max}s waiting for edgelet API before configure"
}

wait_edgelet_api_running

if [ -n "${EDGELET_BOOTSTRAP_CONFIG_CMD:-}" ]; then
	info "Applying edgelet bootstrap configuration"
	sh -c "$EDGELET_BOOTSTRAP_CONFIG_CMD"
else
	info "No EDGELET_BOOTSTRAP_CONFIG_CMD set; skipping edgelet config"
fi

restart_edgelet_container
info "Edgelet container configured and restarted"
