#!/bin/sh
# Enable/start edgelet daemon or container deployment unit.
set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/container_engine.sh"
. "$SCRIPT_DIR/lib/container_mounts.sh"
. "$SCRIPT_DIR/lib/embed.sh"
. "$SCRIPT_DIR/lib/service.sh"
EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

NO_WAIT=false
while [ $# -gt 0 ]; do
	case "$1" in
		--no-wait)
			NO_WAIT=true
			;;
		*)
			echo "Error: unknown argument: $1" >&2
			exit 1
			;;
	esac
	shift
done
if [ "$NO_WAIT" = true ]; then
	export EDGELET_START_NO_WAIT=1
fi

CONTAINER_ENGINE="${CONTAINER_ENGINE:-edgelet}"
EDGELET_INSTALL_MODE="${EDGELET_INSTALL_MODE:-native}"
EDGELET_CONTAINER_IMAGE="${EDGELET_CONTAINER_IMAGE:-}"
EDGELET_TZ="${EDGELET_TZ:-UTC}"
EDGELET_CONTAINER_NAME="${EDGELET_CONTAINER_NAME:-edgelet}"
EDGELET_SERVICE_ACTION="${EDGELET_SERVICE_ACTION:-start}"

restart_native_linux() {
	restart_edgelet_services "$CONTAINER_ENGINE"
}

redeploy_native_linux() {
	case "$INIT_SYSTEM" in
		systemd)
			if [ "$CONTAINER_ENGINE" = "edgelet" ]; then
				if consume_containerd_restarted_marker; then
					info "edgelet-containerd already restarted; restarting edgelet only (redeploy)"
				elif consume_restart_data_plane_marker; then
					restart_edgelet_containerd_service
				else
					start_edgelet_containerd_unit "$INIT_SYSTEM" false
				fi
				restart_edgelet_daemon_service
			else
				restart_edgelet_daemon_service
			fi
			;;
		openrc)
			if [ "$CONTAINER_ENGINE" = "edgelet" ]; then
				maybe_sudo rc-update add edgelet-containerd default 2>/dev/null || true
				if consume_containerd_restarted_marker; then
					:
				elif consume_restart_data_plane_marker; then
					restart_edgelet_containerd_service
				else
					start_edgelet_containerd_unit "$INIT_SYSTEM" false
				fi
			fi
			maybe_sudo rc-service edgelet restart 2>/dev/null || maybe_sudo rc-service edgelet start
			;;
		procd)
			maybe_sudo /etc/init.d/edgelet restart 2>/dev/null || maybe_sudo /etc/init.d/edgelet start
			;;
		sysvinit)
			maybe_sudo /etc/init.d/edgelet restart 2>/dev/null || maybe_sudo /etc/init.d/edgelet start
			;;
		upstart)
			maybe_sudo initctl restart edgelet 2>/dev/null || maybe_sudo initctl start edgelet
			;;
		s6)
			if command -v s6-svc >/dev/null 2>&1 && [ -d /var/run/s6/services/edgelet ]; then
				maybe_sudo s6-svc -d /var/run/s6/services/edgelet 2>/dev/null || true
				maybe_sudo s6-svc -u /var/run/s6/services/edgelet 2>/dev/null || true
			fi
			;;
		runit)
			maybe_sudo sv restart edgelet 2>/dev/null || maybe_sudo sv start edgelet 2>/dev/null || true
			;;
		*)
			echo "Error: cannot redeploy edgelet on init=$INIT_SYSTEM"
			exit 1
			;;
	esac
}

start_embedded_systemd() {
	maybe_sudo systemctl enable edgelet-containerd 2>/dev/null || true
	maybe_sudo systemctl start edgelet-containerd 2>/dev/null || true
	if [ "${EDGELET_START_NO_WAIT:-0}" != "1" ]; then
		wait_edgelet_containerd_socket || true
	fi
	maybe_sudo systemctl enable edgelet 2>/dev/null || true
	maybe_sudo systemctl stop edgelet 2>/dev/null || true
	maybe_sudo systemctl reset-failed edgelet 2>/dev/null || true
	maybe_sudo systemctl start edgelet
}

start_native_linux() {
	case "$INIT_SYSTEM" in
		systemd)
			if [ "$CONTAINER_ENGINE" = "edgelet" ]; then
				start_embedded_systemd
			else
				maybe_sudo systemctl reset-failed edgelet 2>/dev/null || true
				maybe_sudo systemctl start edgelet
			fi
			;;
		openrc)
			if [ "$CONTAINER_ENGINE" = "edgelet" ]; then
				maybe_sudo rc-update add edgelet-containerd default 2>/dev/null || true
				maybe_sudo rc-service edgelet-containerd start 2>/dev/null || true
			fi
			maybe_sudo rc-service edgelet restart 2>/dev/null || maybe_sudo rc-service edgelet start
			;;
		procd)
			maybe_sudo /etc/init.d/edgelet start
			;;
		sysvinit)
			maybe_sudo /etc/init.d/edgelet restart 2>/dev/null || maybe_sudo /etc/init.d/edgelet start
			;;
		upstart)
			maybe_sudo initctl restart edgelet 2>/dev/null || maybe_sudo initctl start edgelet
			;;
		s6)
			if command -v s6-svc >/dev/null 2>&1 && [ -d /var/run/s6/services/edgelet ]; then
				maybe_sudo s6-svc -u /var/run/s6/services/edgelet 2>/dev/null || true
			fi
			;;
		runit)
			maybe_sudo sv restart edgelet 2>/dev/null || maybe_sudo sv start edgelet 2>/dev/null || true
			;;
		*)
			echo "Error: cannot start edgelet on init=$INIT_SYSTEM"
			exit 1
			;;
	esac
}

edgelet_daemon_running() {
	pgrep -f '[e]dgelet daemon' >/dev/null 2>&1
}

start_edgelet_daemon_desktop() {
	_log_dir="/var/log/edgelet"
	_log_file="${_log_dir}/daemon.log"
	_pid_file="/var/run/edgelet/edgelet.pid"
	case "$EDGELET_OS" in
		darwin)
			maybe_sudo mkdir -p "$_log_dir" /var/run/edgelet
			;;
		windows)
			die "windows native start must be handled by platform-specific tooling"
			;;
		*) return 0 ;;
	esac

	if edgelet_daemon_running; then
		echo "# edgelet daemon already running"
		return 0
	fi

	# Detach under one root shell so the daemon survives bootstrap script exit.
	maybe_sudo sh -c "nohup edgelet daemon >> '${_log_file}' 2>&1 </dev/null & echo \$! > '${_pid_file}'"

	_iter=0
	_max="${EDGELET_START_TIMEOUT:-60}"
	while [ "$_iter" -lt "$_max" ]; do
		if edgelet_daemon_running; then
			break
		fi
		sleep 1
		_iter=$((_iter + 1))
	done

	if ! edgelet_daemon_running; then
		echo "ERROR: edgelet daemon failed to start within ${_max}s; see ${_log_file}" >&2
		if [ -f "$_log_file" ]; then
			tail -20 "$_log_file" >&2 || true
		fi
		exit 1
	fi

	_pid=$(cat "$_pid_file" 2>/dev/null || pgrep -f '[e]dgelet daemon' | head -1)
	echo "# edgelet daemon started in background (pid=${_pid:-unknown}, log=${_log_file})"
}

start_container_linux() {
	case "$INIT_SYSTEM" in
		systemd)
			maybe_sudo systemctl start edgelet
			;;
		openrc)
			maybe_sudo rc-service edgelet start
			;;
		*)
			echo "Error: cannot start container deployment on init=$INIT_SYSTEM"
			exit 1
			;;
	esac
}

start_container_desktop() {
	_image="${EDGELET_CONTAINER_IMAGE:-}"
	if [ -z "$_image" ]; then
		echo "Error: EDGELET_CONTAINER_IMAGE is required"
		exit 1
	fi
	_run=$(container_runtime_bin)
	prepare_edgelet_container_host_dirs
	if $_run ps --format '{{.Names}}' 2>/dev/null | grep -q "^${EDGELET_CONTAINER_NAME}$"; then
		echo "# edgelet container already running"
		return 0
	fi
	_sock_mount=$(container_engine_sock_mount)
	_vol_mounts=$(edgelet_container_volume_mounts)
	# shellcheck disable=SC2086
	maybe_sudo $_run run -d --name "${EDGELET_CONTAINER_NAME}" \
		-e "TZ=${EDGELET_TZ}" \
		-e "EDGELET_DAEMON=container" \
		${_sock_mount} \
		${_vol_mounts} \
		--net=host --privileged --stop-timeout 60 \
		--restart=always \
		"$_image"
	echo "# edgelet container started (${EDGELET_CONTAINER_NAME})"
}

start_container_darwin() {
	start_container_desktop
}

start_container_windows() {
	start_container_desktop
}

case "$EDGELET_INSTALL_MODE" in
	container)
		case "$EDGELET_OS" in
			linux) start_container_linux ;;
			darwin) start_container_darwin ;;
			windows) start_container_windows ;;
			*) echo "Error: container deployment start not supported on os=$EDGELET_OS"; exit 1 ;;
		esac
		;;
	native|"")
		case "$EDGELET_OS" in
			linux)
				case "$EDGELET_SERVICE_ACTION" in
					upgrade) restart_native_linux ;;
					restart) redeploy_native_linux ;;
					*) start_native_linux ;;
				esac
				;;
			darwin) start_edgelet_daemon_desktop ;;
			windows)
				echo "Error: windows native start must be handled by platform-specific tooling"
				exit 1
				;;
			*)
				echo "Error: unsupported os=$EDGELET_OS"
				exit 1
				;;
		esac
		;;
	*)
		echo "Error: unknown EDGELET_INSTALL_MODE=$EDGELET_INSTALL_MODE"
		exit 1
		;;
esac
