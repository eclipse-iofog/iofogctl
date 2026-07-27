#!/bin/sh
# One-shot: exit 0 if edgelet init unit active and daemon reports RUNNING.
# Remote bootstrap: potctl polls via reconnecting SSH (Phase 9); not wait_edgelet_ready.sh.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/service.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

edgelet_api_starting() {
	_out="$1"
	echo "$_out" | grep -qi 'daemon is not running' && return 0
	echo "$_out" | grep -q 'Edgelet API is still initializing' && return 0
	echo "$_out" | grep -qi 'LOCAL_API_STARTING' && return 0
	echo "$_out" | grep -qi 'Local API is starting' && return 0
	echo "$_out" | grep -qi 'DAEMON_UNAVAILABLE' && return 0
	return 1
}

edgelet_daemon_status_value() {
	_out="$1"
	echo "$_out" | awk -F': ' '/^edgeletDaemon:/ {print $2; exit}' | tr -d '[:space:]'
}

edgelet_daemon_status_running() {
	_status=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')
	[ "$_status" = "running" ]
}

edgelet_runtime_agent_running() {
	_out="$1"
	echo "$_out" | grep -qi 'runtime.agentPhase:[[:space:]]*running'
}

probe_edgelet_daemon_running() {
	_out=$(maybe_sudo edgelet system status 2>&1) || true
	if edgelet_api_starting "$_out"; then
		return 1
	fi
	if edgelet_daemon_status_running "$(edgelet_daemon_status_value "$_out")"; then
		return 0
	fi
	if edgelet_runtime_agent_running "$_out"; then
		return 0
	fi
	return 1
}

if ! edgelet_unit_active; then
	echo "edgelet init unit not active" >&2
	exit 1
fi

if probe_edgelet_daemon_running; then
	exit 0
fi

echo "edgelet daemon not RUNNING" >&2
exit 1
