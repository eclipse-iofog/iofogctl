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

wait_edgelet_containerd_socket() {
	_timeout="${EDGELET_ATTACH_WAIT_SEC:-120}"
	_elapsed=0
	_sock="/run/edgelet/containerd.sock"
	while [ "$_elapsed" -lt "$_timeout" ]; do
		if [ -S "$_sock" ] || [ -S /var/run/edgelet/containerd.sock ]; then
			[ -S "$_sock" ] || _sock="/var/run/edgelet/containerd.sock"
			_ctr=""
			if [ -x /var/lib/edgelet/data/current/bin/ctr ]; then
				_ctr=/var/lib/edgelet/data/current/bin/ctr
			elif command -v ctr >/dev/null 2>&1; then
				_ctr=ctr
			fi
			if [ -z "$_ctr" ] || "$_ctr" --address "$_sock" version >/dev/null 2>&1; then
				return 0
			fi
		fi
		sleep 2
		_elapsed=$(( _elapsed + 2 ))
	done
	echo "ERROR: edgelet-containerd socket not ready after ${_timeout}s" >&2
	return 1
}

restart_edgelet_containerd_service() {
	case "${INIT_SYSTEM:-unknown}" in
		systemd) maybe_sudo systemctl restart edgelet-containerd 2>/dev/null || true ;;
		openrc) maybe_sudo rc-service edgelet-containerd restart 2>/dev/null || true ;;
		*)
			maybe_sudo systemctl restart edgelet-containerd 2>/dev/null || \
				maybe_sudo service edgelet-containerd restart 2>/dev/null || true
			;;
	esac
	wait_edgelet_containerd_socket || true
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
		info "Restarting edgelet-containerd and edgelet (embedded engine OTA)"
		restart_edgelet_containerd_service
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
