#!/bin/sh
# Install edgelet init units, engine drop-ins, and container deployment host units.
set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

CONTAINER_ENGINE="${CONTAINER_ENGINE:-edgelet}"
DEPLOYMENT_TYPE="${DEPLOYMENT_TYPE:-native}"
EDGELET_INSTALL_MODE="${EDGELET_INSTALL_MODE:-native}"
EDGELET_CONTAINER_IMAGE="${EDGELET_CONTAINER_IMAGE:-}"
EDGELET_TZ="${EDGELET_TZ:-UTC}"
EDGELET_LIBEXEC="/usr/libexec/edgelet"
EDGELET_CONTAINER_NAME="${EDGELET_CONTAINER_NAME:-edgelet}"

. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/container_engine.sh"
. "$SCRIPT_DIR/lib/container_mounts.sh"

install_shutdown_helper() {
 mkdir -p "${EDGELET_LIBEXEC}"
 cat > "/usr/libexec/edgelet/edgelet-shutdown" << 'EDGELET_SHUTDOWN_EOF'
#!/bin/sh
# edgelet-shutdown — shared control-plane stop entry for all init systems (Plan 10).
# Plan 11: v1 default skips MS drain on control stop (shutdownPolicy=leave-running for docker/podman;
# embedded split uses attach-only). shutdownGracePeriodSeconds applies to optional maintenance drain
# and data-plane (edgelet-containerd) stop — not control-plane MS drain by default.
set -e
EDGELET="${EDGELET_BIN:-/usr/local/bin/edgelet}"
exec "${EDGELET}" shutdown "$@"
EDGELET_SHUTDOWN_EOF
 chmod 755 "/usr/libexec/edgelet/edgelet-shutdown"
}
write_systemd_edgelet_service() {
 cat > "/etc/systemd/system/edgelet.service" << 'EDGELET_SYSTEMD_EDGELET_SERVICE_EOF'
[Unit]
Description=Edgelet daemon (control plane)
Documentation=https://github.com/eclipse-iofog/edgelet
Wants=network-online.target
After=network-online.target
StartLimitIntervalSec=300
StartLimitBurst=20

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mountpoint -q /sys/fs/bpf || mount -t bpf bpf /sys/fs/bpf 2>/dev/null || true'
ExecStart=/usr/local/bin/edgelet daemon
ExecStop=/usr/libexec/edgelet/edgelet-shutdown
Restart=always
RestartSec=2s
# Default 120s = shutdownGracePeriodSeconds (90) + 30s buffer; edgelet-engine drop-in may override.
TimeoutStopSec=120s
KillMode=process
Delegate=yes
DelegateSubgroup=supervisor
KillSignal=SIGTERM
SendSIGKILL=yes
User=root
StandardOutput=journal
StandardError=journal
SyslogIdentifier=edgelet

# Embedded engine needs host paths under /etc/cni, /run, /var, /opt (monolithic unit).
# Tighten on edgelet-containerd.service after Plan 11 data-plane split.
NoNewPrivileges=no

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EDGELET_SYSTEMD_EDGELET_SERVICE_EOF
 chmod 644 "/etc/systemd/system/edgelet.service"
}
write_systemd_edgelet_containerd_service() {
 cat > "/etc/systemd/system/edgelet-containerd.service" << 'EDGELET_SYSTEMD_EDGELET_CONTAINERD_SERVICE_EOF'
# Edgelet embedded containerd (data plane) — Plan 11 workload continuity
[Unit]
Description=Edgelet embedded containerd (data plane)
Documentation=https://github.com/eclipse-iofog/edgelet
Before=edgelet.service
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=300
StartLimitBurst=5
# Intentionally NOT PartOf=edgelet.service: control restart/stop must not
# stop the data plane (Plan 11 attach-only). Full teardown: stop both units.

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mountpoint -q /sys/fs/bpf || mount -t bpf bpf /sys/fs/bpf 2>/dev/null || true'
ExecStart=/usr/local/bin/edgelet runtime-bootstrap
ExecStopPost=-/usr/local/bin/edgelet runtime reap-orphans
Restart=always
RestartSec=5s
# Data-plane stop: drain MS via runtime-bootstrap SIGTERM handler + shutdownGracePeriodSeconds
TimeoutStopSec=120s
KillMode=process
KillSignal=SIGTERM
SendSIGKILL=yes
User=root
StandardOutput=journal
StandardError=journal
SyslogIdentifier=edgelet-containerd
Delegate=yes
DelegateSubgroup=containerd
NoNewPrivileges=no
LimitNOFILE=65536
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EDGELET_SYSTEMD_EDGELET_CONTAINERD_SERVICE_EOF
 chmod 644 "/etc/systemd/system/edgelet-containerd.service"
}
write_systemd_dropin_docker() {
 cat > "/etc/systemd/system/edgelet.service.d/docker.conf" << 'EDGELET_SYSTEMD_DROPIN_DOCKER_EOF'
# Engine ordering drop-in — docker (Plan 10). Installed when containerEngine=docker.
[Unit]
After=network-online.target docker.service
Wants=docker.service
EDGELET_SYSTEMD_DROPIN_DOCKER_EOF
 chmod 644 "/etc/systemd/system/edgelet.service.d/docker.conf"
}
write_systemd_dropin_podman() {
 cat > "/etc/systemd/system/edgelet.service.d/podman.conf" << 'EDGELET_SYSTEMD_DROPIN_PODMAN_EOF'
# Engine ordering drop-in — podman (Plan 10). Installed when containerEngine=podman.
[Unit]
After=network-online.target podman.socket
Wants=podman.socket
EDGELET_SYSTEMD_DROPIN_PODMAN_EOF
 chmod 644 "/etc/systemd/system/edgelet.service.d/podman.conf"
}
write_systemd_dropin_edgelet() {
 cat > "/etc/systemd/system/edgelet.service.d/edgelet.conf" << 'EDGELET_SYSTEMD_DROPIN_EDGELET_EOF'
# Engine ordering drop-in — embedded edgelet (Plan 11). Installed when containerEngine=edgelet.
[Unit]
Wants=edgelet-containerd.service
After=network-online.target edgelet-containerd.service

[Service]
Environment=EDGELET_RUNTIME_SPLIT=1
# Control-plane stop: default leave-running (skip MS drain). Grace + buffer matches shutdownGracePeriodSeconds default 90.
TimeoutStopSec=120s
EDGELET_SYSTEMD_DROPIN_EDGELET_EOF
 chmod 644 "/etc/systemd/system/edgelet.service.d/edgelet.conf"
}
write_openrc_edgelet_init() {
 cat > "/etc/init.d/edgelet" << 'EDGELET_OPENRC_EDGELET_INIT_EOF'
#!/sbin/openrc-run

name="edgelet"
description="Edgelet daemon (control plane)"
command="/usr/local/bin/edgelet"
command_args="daemon"
supervisor="supervise-daemon"
respawn_delay=2
respawn_max=0
respawn_period=3600
command_user="root"
pidfile="/run/${RC_SVCNAME}.pid"
output_log="/var/log/edgelet/daemon.log"
error_log="/var/log/edgelet/daemon.log"
shutdown="/usr/libexec/edgelet/edgelet-shutdown"

depend() {
 need net
 after firewall
%%EDGELET_ENGINE_NEED%%
}

wait_containerd_attach_socket() {
 _timeout="${EDGELET_ATTACH_WAIT_SEC:-120}"
 _elapsed=0
 _sock="/run/edgelet/containerd.sock"
 while [ "${_elapsed}" -lt "${_timeout}" ]; do
 if [ -S "${_sock}" ] || [ -S /var/run/edgelet/containerd.sock ]; then
 [ -S "${_sock}" ] || _sock="/var/run/edgelet/containerd.sock"
 _ctr=""
 if [ -x /var/lib/edgelet/data/current/bin/ctr ]; then
 _ctr=/var/lib/edgelet/data/current/bin/ctr
 elif command -v ctr >/dev/null 2>&1; then
 _ctr=ctr
 fi
 if [ -z "${_ctr}" ] || "${_ctr}" --address "${_sock}" version >/dev/null 2>&1; then
 return 0
 fi
 fi
 sleep 2
 _elapsed=$(( _elapsed + 2 ))
 done
 ewarn "data-plane containerd socket not ready after ${_timeout}s"
 return 1
}

start_pre() {
 /usr/local/bin/edgelet cgroup-preflight || return $?
 if [ -f /etc/init.d/edgelet-containerd ]; then
 export EDGELET_RUNTIME_SPLIT=1
 wait_containerd_attach_socket || return $?
 fi
}

stop_pre() {
 if [ -f /etc/init.d/edgelet-containerd ]; then
 export EDGELET_RUNTIME_SPLIT=1
 fi
}

stop() {
 ebegin "Stopping ${RC_SVCNAME}"
 if ! "${shutdown}"; then
 # Fallback when EdgeletAPI is down: single TERM/KILL, no SSD --retry (hangs after graceful stop).
 if [ -f "${pidfile}" ]; then
 _pid=$(cat "${pidfile}" 2>/dev/null || true)
 if [ -n "${_pid}" ] && kill -0 "${_pid}" 2>/dev/null; then
 kill -TERM "${_pid}" 2>/dev/null || true
 sleep 2
 kill -KILL "${_pid}" 2>/dev/null || true
 fi
 fi
 rm -f "${pidfile}" /run/edgelet/edgelet.pid 2>/dev/null || true
 eend 1
 return 1
 fi
 rm -f "${pidfile}" /run/edgelet/edgelet.pid 2>/dev/null || true
 # Plan 11 split: data-plane containerd is edgelet-containerd service — do not reap child here.
 if [ ! -f /etc/init.d/edgelet-containerd ]; then
 _pids=$(pgrep -f edgelet-containerd-child 2>/dev/null || true)
 for _p in ${_pids}; do kill -TERM "${_p}" 2>/dev/null || true; done
 sleep 2
 _pids=$(pgrep -f edgelet-containerd-child 2>/dev/null || true)
 for _p in ${_pids}; do kill -KILL "${_p}" 2>/dev/null || true; done
 fi
 eend 0
}
EDGELET_OPENRC_EDGELET_INIT_EOF
 chmod 755 "/etc/init.d/edgelet"
}
write_openrc_edgelet_containerd_init() {
 cat > "/etc/init.d/edgelet-containerd" << 'EDGELET_OPENRC_EDGELET_CONTAINERD_INIT_EOF'
#!/sbin/openrc-run

# Edgelet embedded containerd (data plane) — Plan 11 workload continuity
name="edgelet-containerd"
description="Edgelet embedded containerd (data plane)"
command="/usr/local/bin/edgelet"
command_args="runtime-bootstrap"
supervisor="supervise-daemon"
respawn_delay=5
respawn_max=5
respawn_period=300
command_user="root"
pidfile="/run/${RC_SVCNAME}.pid"
output_log="/var/log/edgelet/containerd.log"
error_log="/var/log/edgelet/containerd.log"

depend() {
 need net
 need edgelet-cgroup-prep
 before edgelet
}

start_pre() {
 # C3: light preflight only; primary cgroup bootstrap is in runtime-bootstrap (C1).
 mountpoint -q /sys/fs/bpf 2>/dev/null || mount -t bpf bpf /sys/fs/bpf 2>/dev/null || true
 /usr/local/bin/edgelet cgroup-preflight || return $?
}

start_post() {
 # Block dependency satisfaction until CRI socket answers (Plan 11 split).
 _timeout="${EDGELET_CONTAINERD_READY_SEC:-120}"
 _elapsed=0
 _sock="/run/edgelet/containerd.sock"
 while [ "${_elapsed}" -lt "${_timeout}" ]; do
 if [ -S "${_sock}" ] || [ -S /var/run/edgelet/containerd.sock ]; then
 [ -S "${_sock}" ] || _sock="/var/run/edgelet/containerd.sock"
 _ctr=""
 if [ -x /var/lib/edgelet/data/current/bin/ctr ]; then
 _ctr=/var/lib/edgelet/data/current/bin/ctr
 elif command -v ctr >/dev/null 2>&1; then
 _ctr=ctr
 fi
 if [ -z "${_ctr}" ] || "${_ctr}" --address "${_sock}" version >/dev/null 2>&1; then
 return 0
 fi
 fi
 sleep 2
 _elapsed=$(( _elapsed + 2 ))
 done
 ewarn "containerd socket not ready after ${_timeout}s"
 return 1
}

stop() {
 ebegin "Stopping ${RC_SVCNAME}"
 if [ -f "${pidfile}" ]; then
 _pid=$(cat "${pidfile}" 2>/dev/null || true)
 if [ -n "${_pid}" ] && kill -0 "${_pid}" 2>/dev/null; then
 kill -TERM "${_pid}" 2>/dev/null || true
 _grace="${EDGELET_CONTAINERD_STOP_SEC:-30}"
 _elapsed=0
 while kill -0 "${_pid}" 2>/dev/null; do
 if [ "${_elapsed}" -ge "${_grace}" ]; then
 kill -KILL "${_pid}" 2>/dev/null || true
 break
 fi
 sleep 2
 _elapsed=$(( _elapsed + 2 ))
 done
 fi
 fi
 rm -f "${pidfile}" 2>/dev/null || true
 /usr/local/bin/edgelet runtime reap-orphans || ewarn "orphan reap exited non-zero"
 eend 0
}
EDGELET_OPENRC_EDGELET_CONTAINERD_INIT_EOF
 chmod 755 "/etc/init.d/edgelet-containerd"
}
write_openrc_edgelet_cgroup_prep_init() {
 cat > "/etc/init.d/edgelet-cgroup-prep" << 'EDGELET_OPENRC_EDGELET_CGROUP_PREP_INIT_EOF'
#!/sbin/openrc-run

# Early cgroup v2 delegation for LXC/VM machine roots (OrbStack Alpine, etc.).
# Moby/dind-style reparent before enabling subtree_control on root and /.lxc.
name="edgelet-cgroup-prep"
description="Edgelet cgroup delegation prep (machine root)"

depend() {
 before edgelet-containerd edgelet
}

cgroup_delegate_controllers() {
 _dir="$1"
 _ctrl="${_dir}/cgroup.controllers"
 _sub="${_dir}/cgroup.subtree_control"
 [ -f "${_ctrl}" ] || return 0
 for _c in cpu memory pids; do
 grep -qw "${_c}" "${_ctrl}" 2>/dev/null || continue
 grep -qw "${_c}" "${_sub}" 2>/dev/null && continue
 echo "+${_c}" >> "${_sub}" 2>/dev/null || true
 done
}

cgroup_reparent_procs() {
 _from="$1"
 _to="$2"
 [ -f "${_from}/cgroup.procs" ] || return 0
 [ -s "${_from}/cgroup.procs" ] || return 0
 mkdir -p "${_to}"
 xargs -rn1 < "${_from}/cgroup.procs" > "${_to}/cgroup.procs" 2>/dev/null || true
}

start() {
 # No-op on bare-metal / systemd VM layouts without LXC machine cgroup.
 [ -d /sys/fs/cgroup/.lxc ] || return 0

 ebegin "Preparing cgroup delegation for machine root"
 _cg="/sys/fs/cgroup"

 cgroup_reparent_procs "${_cg}" "${_cg}/init"
 cgroup_delegate_controllers "${_cg}"

 cgroup_reparent_procs "${_cg}/.lxc" "${_cg}/.lxc/init"
 cgroup_delegate_controllers "${_cg}/.lxc"

 eend 0
}

stop() {
 return 0
}
EDGELET_OPENRC_EDGELET_CGROUP_PREP_INIT_EOF
 chmod 755 "/etc/init.d/edgelet-cgroup-prep"
}
write_procd_edgelet() {
 cat > "/etc/init.d/edgelet" << 'EDGELET_PROCD_EDGELET_EOF'
#!/bin/sh /etc/rc.common
# Edgelet control plane (OpenWrt procd)

START=99
STOP=10
USE_PROCD=1

start_service() {
 /usr/local/bin/edgelet cgroup-preflight || return $?

 procd_open_instance
 procd_set_param command /usr/local/bin/edgelet
 procd_append_param command daemon
 procd_set_param respawn 3600 5 5
 procd_set_param stdout 1
 procd_set_param stderr 1
 procd_close_instance
}

stop_service() {
 /usr/libexec/edgelet/edgelet-shutdown
}
EDGELET_PROCD_EDGELET_EOF
 chmod 755 "/etc/init.d/edgelet"
}
write_sysvinit_edgelet() {
 cat > "/etc/init.d/edgelet" << 'EDGELET_SYSVINIT_EDGELET_EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides: edgelet
# Required-Start: $network $remote_fs
# Required-Stop: $network $remote_fs
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: Edgelet daemon
### END INIT INFO

DAEMON=/usr/local/bin/edgelet
SHUTDOWN=/usr/libexec/edgelet/edgelet-shutdown
PIDFILE=/var/run/edgelet.pid
LOGFILE=/var/log/edgelet/daemon.log
KILLTIMEOUT=30

. /lib/lsb/init-functions

preflight() {
 "${DAEMON}" cgroup-preflight
}

case "$1" in
 start)
 log_daemon_msg "Starting edgelet"
 preflight || exit $?
 start-stop-daemon --start --background --make-pidfile --pidfile "$PIDFILE" \
 --exec "$DAEMON" -- daemon >>"$LOGFILE" 2>&1
 log_end_msg $?
 ;;
 stop)
 log_daemon_msg "Stopping edgelet"
 if [ -x "$SHUTDOWN" ]; then
 "$SHUTDOWN" && log_end_msg 0 && exit 0
 fi
 start-stop-daemon --stop --pidfile "$PIDFILE" --retry TERM/${KILLTIMEOUT}/KILL/5
 log_end_msg $?
 ;;
 restart|force-reload)
 $0 stop
 $0 start
 ;;
 status)
 status_of_proc -p "$PIDFILE" "$DAEMON" edgelet
 ;;
 *)
 echo "Usage: $0 {start|stop|restart|status}"
 exit 1
 ;;
esac

exit 0
EDGELET_SYSVINIT_EDGELET_EOF
 chmod 755 "/etc/init.d/edgelet"
}
write_upstart_edgelet() {
 cat > "/etc/init/edgelet.conf" << 'EDGELET_UPSTART_EDGELET_EOF'
description "Edgelet daemon (control plane)"
author "Datasance"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
respawn limit 20 300

pre-start script
 /usr/local/bin/edgelet cgroup-preflight || exit $?
end script

exec /usr/local/bin/edgelet daemon >> /var/log/edgelet/daemon.log 2>&1

post-stop script
 /usr/libexec/edgelet/edgelet-shutdown || true
end script
EDGELET_UPSTART_EDGELET_EOF
 chmod 644 "/etc/init/edgelet.conf"
}
write_s6_run() {
 cat > "/etc/s6/edgelet/run" << 'EDGELET_S6_RUN_EOF'
#!/bin/sh
exec 2>&1
/usr/local/bin/edgelet cgroup-preflight || exit $?
exec /usr/local/bin/edgelet daemon
EDGELET_S6_RUN_EOF
 chmod 755 "/etc/s6/edgelet/run"
}
write_s6_finish() {
 cat > "/etc/s6/edgelet/finish" << 'EDGELET_S6_FINISH_EOF'
#!/bin/sh
/usr/libexec/edgelet/edgelet-shutdown 2>/dev/null || true
exit "${1:-0}"
EDGELET_S6_FINISH_EOF
 chmod 755 "/etc/s6/edgelet/finish"
}
write_runit_run() {
 cat > "/etc/runit/edgelet/run" << 'EDGELET_RUNIT_RUN_EOF'
#!/bin/sh
LOG=/var/log/edgelet/daemon.log
exec 2>&1
/usr/local/bin/edgelet cgroup-preflight >>"$LOG" 2>&1 || exit $?
exec /usr/local/bin/edgelet daemon >>"$LOG" 2>&1
EDGELET_RUNIT_RUN_EOF
 chmod 755 "/etc/runit/edgelet/run"
}
write_runit_finish() {
 cat > "/etc/runit/edgelet/finish" << 'EDGELET_RUNIT_FINISH_EOF'
#!/bin/sh
/usr/libexec/edgelet/edgelet-shutdown 2>/dev/null || true
exit "${1:-0}"
EDGELET_RUNIT_FINISH_EOF
 chmod 755 "/etc/runit/edgelet/finish"
}

openrc_engine_need_line() {
 case "$1" in
 docker) printf '%s\n' ' need docker' ;;
 podman) printf '%s\n' ' need podman' ;;
 edgelet) printf '%s\n' ' need edgelet-containerd' ;;
 *) printf '%s\n' '' ;;
 esac
}

apply_openrc_engine_deps() {
 _eng="$1"
 _dest="$2"
 _need="$(openrc_engine_need_line "${_eng}")"
 if [ -f "${_dest}" ]; then
 # shellcheck disable=SC2016
 awk -v need="${_need}" '
 /%%EDGELET_ENGINE_NEED%%/ {
 if (need != "") print need
 next
 }
 { print }
 ' "${_dest}" > "${_dest}.tmp" && mv "${_dest}.tmp" "${_dest}"
 fi
}

install_init_helpers() {
 install_shutdown_helper
}

install_systemd_dropin() {
 _eng="$1"
 _root="$2"
 _dropdir="/etc/systemd/system/edgelet.service.d"
 mkdir -p "${_dropdir}"
 rm -f "${_dropdir}/docker.conf" "${_dropdir}/podman.conf" "${_dropdir}/edgelet.conf"
 if [ -n "${_root}" ]; then
 case "${_eng}" in
 docker)
 install -m 644 "${_root}/systemd/edgelet.service.d/docker.conf" "${_dropdir}/docker.conf"
 ;;
 podman)
 install -m 644 "${_root}/systemd/edgelet.service.d/podman.conf" "${_dropdir}/podman.conf"
 ;;
 edgelet)
 install -m 644 "${_root}/systemd/edgelet.service.d/edgelet.conf" "${_dropdir}/edgelet.conf"
 systemctl enable edgelet-containerd 2>/dev/null || true
 ;;
 esac
 else
 case "${_eng}" in
 docker) write_systemd_dropin_docker ;;
 podman) write_systemd_dropin_podman ;;
 edgelet)
 write_systemd_dropin_edgelet
 systemctl enable edgelet-containerd 2>/dev/null || true
 ;;
 esac
 fi
}

write_container_systemd_unit() {
	_engine="$1"
	_image="$2"
	_tz="$3"
	_run_bin=""
	case "$_engine" in
		docker) _run_bin="/usr/bin/docker" ;;
		podman) _run_bin="/usr/bin/podman" ;;
		*) echo "Error: container deployment requires docker or podman engine"; exit 1 ;;
	esac
	_sock_mount=$(container_engine_sock_mount)
	_vol_mounts=$(edgelet_container_volume_mounts)
	prepare_edgelet_container_host_dirs
	maybe_sudo tee /etc/systemd/system/edgelet.service > /dev/null <<EOF
[Unit]
Description=Edgelet agent container
After=network-online.target ${_engine}.service
Wants=network-online.target ${_engine}.service

[Service]
Restart=always
RestartSec=5
ExecStartPre=-${_run_bin} rm -f ${EDGELET_CONTAINER_NAME}
ExecStart=${_run_bin} run --rm --name ${EDGELET_CONTAINER_NAME} \
  -e TZ=${_tz} \
  -e EDGELET_DAEMON="container" \
  ${_sock_mount} \
  ${_vol_mounts} \
  --net=host --privileged --stop-timeout 60 \
  ${_image}
ExecStop=${_run_bin} stop ${EDGELET_CONTAINER_NAME}

[Install]
WantedBy=multi-user.target
EOF
	maybe_sudo systemctl daemon-reload
	maybe_sudo systemctl enable edgelet.service
	echo "# systemd container unit edgelet.service installed (engine=${_engine})"
}

write_container_openrc_init() {
	_engine="$1"
	_image="$2"
	_tz="$3"
	_run_bin=""
	case "$_engine" in
		docker) _run_bin="docker" ;;
		podman) _run_bin="podman" ;;
		*) echo "Error: container deployment requires docker or podman engine"; exit 1 ;;
	esac
	_sock_mount=$(container_engine_sock_mount)
	_vol_mounts=$(edgelet_container_volume_mounts)
	prepare_edgelet_container_host_dirs
	maybe_sudo tee /etc/init.d/edgelet > /dev/null <<EOF
#!/bin/sh
case "\$1" in
start)
	${_run_bin} run -d --rm --name ${EDGELET_CONTAINER_NAME} -e TZ=${_tz} -e EDGELET_DAEMON=container ${_sock_mount} \
	${_vol_mounts} \
	--net=host --privileged --stop-timeout 60 ${_image}
	;;
stop)
	${_run_bin} stop ${EDGELET_CONTAINER_NAME} 2>/dev/null || true
	;;
restart)
	\$0 stop; \$0 start
	;;
*) echo "Usage: \$0 {start|stop|restart}"; exit 1 ;;
esac
EOF
	maybe_sudo chmod +x /etc/init.d/edgelet
	maybe_sudo rc-update add edgelet default 2>/dev/null || true
	echo "# OpenRC container init edgelet installed (engine=${_engine})"
}

install_container_init_units() {
	_engine="${CONTAINER_ENGINE:-docker}"
	_image="${EDGELET_CONTAINER_IMAGE:-}"
	_tz="${EDGELET_TZ:-UTC}"
	if [ -z "$_image" ]; then
		echo "Error: EDGELET_CONTAINER_IMAGE is required for container deployment"
		exit 1
	fi
	case "$INIT_SYSTEM" in
		systemd) write_container_systemd_unit "$_engine" "$_image" "$_tz" ;;
		openrc) write_container_openrc_init "$_engine" "$_image" "$_tz" ;;
		launchd)
			echo "# launchd container unit deferred to start_edgelet.sh on darwin"
			;;
		*)
			echo "Error: container deployment init unit not supported for init=$INIT_SYSTEM"
			exit 1
			;;
	esac
}

install_native_init_units() {
	_init="$INIT_SYSTEM"
	_eng="$CONTAINER_ENGINE"

	if [ "$EDGELET_OS" = "darwin" ] || [ "$_init" = "launchd" ]; then
		echo "# launchd native install deferred to start_edgelet.sh on darwin"
		return 0
	fi

	mkdir -p /var/log/edgelet
	install_init_helpers
	case "${_init}" in
		systemd)
			mkdir -p /etc/cni/net.d /run/edgelet /run/containerd
			chmod 755 /run/edgelet /run/containerd 2>/dev/null || true
			write_systemd_edgelet_service
			write_systemd_edgelet_containerd_service
			install_systemd_dropin "${_eng}" ""
			maybe_sudo systemctl daemon-reload
			if [ "${_eng}" = "edgelet" ]; then
				maybe_sudo systemctl enable edgelet-containerd 2>/dev/null || true
			fi
			maybe_sudo systemctl enable edgelet
			echo "# systemd unit edgelet.service installed (engine=${_eng})"
			;;
		openrc)
			write_openrc_edgelet_init
			write_openrc_edgelet_cgroup_prep_init
			write_openrc_edgelet_containerd_init
			maybe_sudo rc-update add edgelet-cgroup-prep sysinit 2>/dev/null || true
			apply_openrc_engine_deps "${_eng}" /etc/init.d/edgelet
			chmod 755 /etc/init.d/edgelet
			if [ "${_eng}" = "edgelet" ]; then
				maybe_sudo rc-update add edgelet-containerd default 2>/dev/null || true
			fi
			maybe_sudo rc-update add edgelet default 2>/dev/null || true
			echo "# OpenRC service edgelet installed (engine=${_eng})"
			;;
		procd)
			write_procd_edgelet
			maybe_sudo /etc/init.d/edgelet enable 2>/dev/null || true
			echo "# procd init script edgelet installed (engine=${_eng})"
			;;
		sysvinit)
			write_sysvinit_edgelet
			if command -v update-rc.d >/dev/null 2>&1; then
				maybe_sudo update-rc.d edgelet defaults 2>/dev/null || true
			elif command -v chkconfig >/dev/null 2>&1; then
				maybe_sudo chkconfig --add edgelet 2>/dev/null || true
			fi
			echo "# SysV init script edgelet installed"
			;;
		upstart)
			write_upstart_edgelet
			echo "# Upstart job edgelet installed"
			;;
		s6)
			mkdir -p /etc/s6/edgelet
			write_s6_run
			write_s6_finish
			echo "# s6 service installed under /etc/s6/edgelet"
			;;
		runit)
			mkdir -p /etc/runit/edgelet
			write_runit_run
			write_runit_finish
			if [ -d /etc/runit ]; then
				ln -sf /etc/runit/edgelet /etc/service/edgelet 2>/dev/null || 				ln -sf /etc/runit/edgelet /var/service/edgelet 2>/dev/null || true
			fi
			echo "# runit service installed under /etc/runit/edgelet"
			;;
		launchd)
			echo "# launchd native install deferred to start_edgelet.sh on darwin"
			;;
		windows)
			echo "# windows init unit stubs: use edgelet.exe service tooling when available"
			;;
		*)
			echo "Error: no supported init system detected (${_init})"
			exit 1
			;;
	esac
}

case "$EDGELET_INSTALL_MODE" in
	container) install_container_init_units ;;
	native|"") install_native_init_units ;;
	*) echo "Error: unknown EDGELET_INSTALL_MODE=$EDGELET_INSTALL_MODE"; exit 1 ;;
esac
