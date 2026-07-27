#!/bin/sh
# Init-system service stop/restart helpers (OTA parity with upstream edgelet install.sh).

stop_edgelet_service() {
	_init="${1:-${INIT_SYSTEM:-unknown}}"
	case "${_init}" in
		systemd) maybe_sudo systemctl stop edgelet 2>/dev/null || true ;;
		openrc) maybe_sudo rc-service edgelet stop 2>/dev/null || true ;;
		procd) maybe_sudo /etc/init.d/edgelet stop 2>/dev/null || true ;;
		sysvinit) maybe_sudo /etc/init.d/edgelet stop 2>/dev/null || true ;;
		upstart) maybe_sudo initctl stop edgelet 2>/dev/null || true ;;
		s6) maybe_sudo s6-svc -d /var/run/s6/services/edgelet 2>/dev/null || true ;;
		runit) maybe_sudo sv down edgelet 2>/dev/null || true ;;
		*)
			maybe_sudo /usr/libexec/edgelet/edgelet-shutdown 2>/dev/null || \
				pkill -f '/usr/local/bin/edgelet daemon' 2>/dev/null || true
			;;
	esac
}

edgelet_containerd_socket_ready() {
	_sock="/run/edgelet/containerd.sock"
	if [ ! -S "$_sock" ] && [ ! -S /var/run/edgelet/containerd.sock ]; then
		return 1
	fi
	[ -S "$_sock" ] || _sock="/var/run/edgelet/containerd.sock"
	_ctr=""
	if [ -x /var/lib/edgelet/data/current/bin/ctr ]; then
		_ctr=/var/lib/edgelet/data/current/bin/ctr
	elif command -v ctr >/dev/null 2>&1; then
		_ctr=ctr
	fi
	if [ -z "$_ctr" ]; then
		return 0
	fi
	"$_ctr" --address "$_sock" version >/dev/null 2>&1
}

edgelet_containerd_unit_active() {
	case "${INIT_SYSTEM:-unknown}" in
		systemd)
			maybe_sudo systemctl is-active --quiet edgelet-containerd 2>/dev/null
			;;
		openrc)
			maybe_sudo rc-service edgelet-containerd status 2>/dev/null | grep -qi running
			;;
		*)
			maybe_sudo systemctl is-active --quiet edgelet-containerd 2>/dev/null
			;;
	esac
}

wait_edgelet_containerd_socket() {
	_timeout="${EDGELET_ATTACH_WAIT_SEC:-120}"
	_elapsed=0
	while [ "$_elapsed" -lt "$_timeout" ]; do
		if edgelet_containerd_socket_ready; then
			return 0
		fi
		echo "Waiting for containerd socket... (${_elapsed}s / ${_timeout}s)"
		sleep 2
		_elapsed=$(( _elapsed + 2 ))
	done
	echo "ERROR: edgelet-containerd socket not ready after ${_timeout}s" >&2
	return 1
}

wait_edgelet_containerd_ready() {
	_timeout="${EDGELET_CONTAINERD_READY_SEC:-150}"
	_interval="${EDGELET_CONTAINERD_POLL_SEC:-10}"
	_elapsed=0
	while [ "$_elapsed" -lt "$_timeout" ]; do
		if edgelet_containerd_unit_active && edgelet_containerd_socket_ready; then
			return 0
		fi
		info "waiting for edgelet-containerd (${_elapsed}s / ${_timeout}s)"
		sleep "$_interval"
		_elapsed=$(( _elapsed + _interval ))
	done
	echo "ERROR: edgelet-containerd not ready after ${_timeout}s" >&2
	return 1
}

restart_edgelet_containerd_service() {
	info "Restarting edgelet-containerd (drain may take up to 120s; do not interrupt)"
	case "${INIT_SYSTEM:-unknown}" in
		systemd)
			if ! maybe_sudo systemctl restart --no-block edgelet-containerd 2>/dev/null; then
				maybe_sudo systemctl restart edgelet-containerd 2>/dev/null || true
			fi
			;;
		openrc)
			maybe_sudo rc-service edgelet-containerd restart 2>/dev/null || true
			;;
		*)
			if ! maybe_sudo systemctl restart --no-block edgelet-containerd 2>/dev/null; then
				maybe_sudo systemctl restart edgelet-containerd 2>/dev/null || \
					maybe_sudo service edgelet-containerd restart 2>/dev/null || true
			fi
			;;
	esac
	wait_edgelet_containerd_ready
	info "edgelet-containerd is ready"
}

containerd_restarted_marker() {
	printf '%s' "${EDGELET_SCRIPT_STAGE_DIR:-/tmp/edgelet-scripts}/.containerd-restarted"
}

consume_containerd_restarted_marker() {
	_marker=$(containerd_restarted_marker)
	if [ -f "$_marker" ]; then
		rm -f "$_marker"
		return 0
	fi
	return 1
}

mark_containerd_restarted() {
	_marker=$(containerd_restarted_marker)
	_dir=$(dirname "$_marker")
	maybe_sudo mkdir -p "$_dir" 2>/dev/null || mkdir -p "$_dir"
	echo 1 > "$_marker"
}

restart_edgelet_daemon_service() {
	case "${INIT_SYSTEM:-unknown}" in
		systemd)
			maybe_sudo systemctl reset-failed edgelet 2>/dev/null || true
			maybe_sudo systemctl restart edgelet 2>/dev/null || true
			;;
		openrc)
			maybe_sudo rc-service edgelet restart 2>/dev/null || \
				maybe_sudo rc-service edgelet start 2>/dev/null || true
			;;
		procd)
			maybe_sudo /etc/init.d/edgelet restart 2>/dev/null || \
				maybe_sudo /etc/init.d/edgelet start 2>/dev/null || true
			;;
		sysvinit)
			maybe_sudo /etc/init.d/edgelet restart 2>/dev/null || \
				maybe_sudo /etc/init.d/edgelet start 2>/dev/null || true
			;;
		upstart)
			maybe_sudo initctl restart edgelet 2>/dev/null || \
				maybe_sudo initctl start edgelet 2>/dev/null || true
			;;
		s6)
			if command -v s6-svc >/dev/null 2>&1 && [ -d /var/run/s6/services/edgelet ]; then
				maybe_sudo s6-svc -d /var/run/s6/services/edgelet 2>/dev/null || true
				maybe_sudo s6-svc -u /var/run/s6/services/edgelet 2>/dev/null || true
			fi
			;;
		runit)
			maybe_sudo sv restart edgelet 2>/dev/null || \
				maybe_sudo sv start edgelet 2>/dev/null || true
			;;
		*)
			echo "Error: cannot restart edgelet on init=${INIT_SYSTEM:-unknown}" >&2
			return 1
			;;
	esac
}

# restart_edgelet_services stops then starts edgelet (and edgelet-containerd when embedded).
restart_edgelet_services() {
	_eng="${1:-${CONTAINER_ENGINE:-edgelet}}"
	if [ "$_eng" = "edgelet" ]; then
		if consume_containerd_restarted_marker; then
			info "edgelet-containerd already restarted; restarting edgelet only (OTA)"
		elif consume_restart_data_plane_marker; then
			info "Restarting edgelet-containerd and edgelet (embedded bundle OTA)"
			restart_edgelet_containerd_service
		else
			info "Thin OTA (embed hash unchanged); restarting edgelet only"
			start_edgelet_containerd_unit "${INIT_SYSTEM:-unknown}" false
		fi
	else
		info "Restarting edgelet (containerEngine=${_eng})"
	fi
	restart_edgelet_daemon_service
}

edgelet_version_field() {
	_field="$1"
	edgelet --version 2>/dev/null | awk -v key="${_field}.version" -F': ' '$1 == key {print $2; exit}' | tr -d '[:space:]'
}

edgelet_cli_version() {
	edgelet_version_field cli
}

edgelet_daemon_version() {
	edgelet_version_field daemon
}
