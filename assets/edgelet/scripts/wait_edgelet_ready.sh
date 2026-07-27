#!/bin/sh
# Wait until edgelet init services and daemon API report edgeletDaemon: RUNNING.
# Remote bootstrap uses Go reconnecting SSH + probe_edgelet_ready.sh (Phase 9).
# Keep this script for local operator use and debugging.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/receipt.sh"
. "$SCRIPT_DIR/lib/service.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

CONTAINER_ENGINE="${CONTAINER_ENGINE:-edgelet}"

wait_init_services() {
	case "$INIT_SYSTEM" in
		systemd)
			maybe_sudo systemctl is-enabled edgelet >/dev/null 2>&1 || return 0
			_iter=0
			while [ "$_iter" -lt 120 ]; do
				if maybe_sudo systemctl is-active edgelet >/dev/null 2>&1; then
					return 0
				fi
				sleep 1
				_iter=$((_iter + 1))
			 done
			;;
		openrc)
			maybe_sudo rc-service edgelet status >/dev/null 2>&1 && return 0
			;;
		procd)
			[ -x /etc/init.d/edgelet ] && return 0
			;;
		launchd|none)
			return 0
			;;
	esac
	return 0
}

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

edgelet_runtime_engine_ready() {
	_out="$1"
	echo "$_out" | grep -qi 'runtime.engineReady:[[:space:]]*true'
}

edgelet_status_running() {
	_out="$1"
	if edgelet_api_starting "$_out"; then
		return 1
	fi
	if edgelet_daemon_status_running "$(edgelet_daemon_status_value "$_out")"; then
		if [ "$CONTAINER_ENGINE" = "edgelet" ] && ! edgelet_runtime_engine_ready "$_out"; then
			return 1
		fi
		return 0
	fi
	if edgelet_runtime_agent_running "$_out"; then
		return 0
	fi
	return 1
}

edgelet_api_socket_ready() {
	[ -S /run/edgelet/edgelet.sock ] || [ -S /var/run/edgelet/edgelet.sock ]
}

wait_edgelet_api_socket() {
	_iter=0
	_max="${EDGELET_API_SOCKET_TIMEOUT:-180}"
	while [ "$_iter" -lt "$_max" ]; do
		if edgelet_api_socket_ready; then
			return 0
		fi
		sleep 1
		_iter=$((_iter + 1))
	done
	return 1
}

wait_edgelet_api_desktop_container() {
	_iter=0
	_max="${EDGELET_READY_TIMEOUT:-600}"
	while [ "$_iter" -lt "$_max" ]; do
		_out=""
		_out=$(edgelet system status 2>&1) || true
		if edgelet_status_running "$_out"; then
			info "edgelet daemon is RUNNING"
			wait_for_expected_daemon_version
			return 0
		fi
		echo "# waiting for edgelet RUNNING (${_iter}s)..."
		sleep 1
		_iter=$((_iter + 1))
	done
	die "Timed out after ${_max}s waiting for edgelet RUNNING"
}

edgelet_daemon_running() {
	pgrep -f '[e]dgelet daemon' >/dev/null 2>&1
}

edgelet_version_matches_receipt() {
	_expected="${1:-}"
	[ -n "$_expected" ] || return 0
	_daemon_ver=$(edgelet_daemon_version)
	[ -n "$_daemon_ver" ] || return 1
	[ "$_daemon_ver" = "$_expected" ]
}

wait_for_expected_daemon_version() {
	_expected=""
	if [ -f "$RECEIPT_FILE" ]; then
		_expected=$(kv_get "$RECEIPT_FILE" "installed_version")
	fi
	[ -n "$_expected" ] || return 0

	_iter=0
	_max="${EDGELET_VERSION_WAIT_TIMEOUT:-120}"
	while [ "$_iter" -lt "$_max" ]; do
		if edgelet_version_matches_receipt "$_expected"; then
			info "edgelet daemon.version matches receipt (${_expected})"
			return 0
		fi
		_got=$(edgelet_daemon_version)
		echo "# waiting for daemon.version=${_expected} (got=${_got:-unknown}, ${_iter}s)..."
		sleep 1
		_iter=$((_iter + 1))
	done
	die "Timed out after ${_max}s waiting for daemon.version=${_expected}"
}

wait_edgelet_api() {
	if desktop_container_local; then
		wait_edgelet_api_desktop_container
		return 0
	fi

	if ! wait_edgelet_api_socket; then
		echo "# edgelet API socket not ready; continuing with process/API poll..."
	fi

	_iter=0
	_max="${EDGELET_READY_TIMEOUT:-600}"
	while [ "$_iter" -lt "$_max" ]; do
		if ! edgelet_daemon_running; then
			echo "# waiting for edgelet daemon process (${_iter}s)..."
			sleep 1
			_iter=$((_iter + 1))
			continue
		fi
		_out=""
		_out=$(maybe_sudo edgelet system status 2>&1) || true
		if edgelet_status_running "$_out"; then
			info "edgelet daemon is RUNNING"
			wait_for_expected_daemon_version
			return 0
		fi
		_status=$(edgelet_daemon_status_value "$_out")
		if edgelet_api_starting "$_out"; then
			_status="starting"
		fi
		echo "# waiting for edgelet RUNNING (${_iter}s) status=${_status:-unknown}"
		sleep 1
		_iter=$((_iter + 1))
	done
	die "Timed out after ${_max}s waiting for edgelet RUNNING"
}

wait_init_services
wait_edgelet_api
