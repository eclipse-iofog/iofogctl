#!/bin/sh
# Wait until edgelet init services and daemon API report iofogDaemon: RUNNING.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

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

edgelet_status_running() {
	_out="$1"
	if echo "$_out" | grep -qi 'daemon is not running'; then
		return 1
	fi
	if echo "$_out" | grep -q 'Edgelet API is still initializing'; then
		return 1
	fi
	if echo "$_out" | grep -q 'iofogDaemon: RUNNING'; then
		return 0
	fi
	if echo "$_out" | grep -q 'runtime.agentPhase: running'; then
		return 0
	fi
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

wait_edgelet_api() {
	if desktop_container_local; then
		wait_edgelet_api_desktop_container
		return 0
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
			return 0
		fi
		_status=$(echo "$_out" | awk -F': ' '/^iofogDaemon:/ {print $2; exit}' | tr -d '[:space:]')
		echo "# waiting for edgelet RUNNING (${_iter}s) status=${_status:-unknown}"
		sleep 1
		_iter=$((_iter + 1))
	done
	die "Timed out after ${_max}s waiting for edgelet RUNNING"
}

wait_init_services
wait_edgelet_api
