#!/bin/sh
# uninstall.sh — Edgelet uninstaller
#
# Usage:
# sudo sh uninstall.sh [--remove-data]
#
# --remove-data also removes config, data, runtime, logs, and backup directories

set -e

die() { echo "ERROR: $1" >&2; exit 1; }
info() { echo ">>> $1"; }

detect_os() {
 _u=$(uname -s)
 case "${_u}" in
 Linux) echo "linux" ;;
 Darwin) echo "darwin" ;;
 MINGW*|MSYS*|CYGWIN*|Windows_NT) echo "windows" ;;
 *) die "Unsupported OS: ${_u}" ;;
 esac
}

windows_program_data_edgelet() {
 echo "${ProgramData:-/c/ProgramData}/Edgelet"
}

share_dir_for_os() {
 case "$1" in
 linux) echo "/usr/share/edgelet" ;;
 darwin) echo "/usr/local/share/edgelet" ;;
 windows) echo "$(windows_program_data_edgelet)/scripts" ;;
 *) die "Unsupported OS: $1" ;;
 esac
}

binary_path_for_os() {
 case "$1" in
 linux|darwin) echo "/usr/local/bin/edgelet" ;;
 windows)
 _pf="${ProgramFiles:-/c/Program Files}"
 echo "${_pf}/Edgelet/edgelet.exe"
 ;;
 *) die "Unsupported OS: $1" ;;
 esac
}

# OpenRC ships /sbin/openrc-run on Alpine even when PID 1 is busybox (Lima
# template:alpine). Require a running OpenRC supervisor, not merely openrc-run.
openrc_is_pid1() {
 [ -x /sbin/openrc-run ] || return 1
 if rc-status -s >/dev/null 2>&1; then
 return 0
 fi
 if [ -f /etc/inittab ] && grep -q '/sbin/openrc' /etc/inittab 2>/dev/null; then
 return 0
 fi
 _init="$(readlink -f /sbin/init 2>/dev/null || readlink /sbin/init 2>/dev/null || true)"
 case "${_init}" in
 *openrc*) return 0 ;;
 esac
 return 1
}

procd_is_openwrt() {
 [ -x /sbin/procd ] && [ -f /etc/rc.common ] || return 1
 return 0
}

detect_init() {
 if command -v systemctl >/dev/null 2>&1 && [ -d /etc/systemd/system ]; then
 echo "systemd"
 return 0
 fi
 if procd_is_openwrt; then
 echo "procd"
 return 0
 fi
 if openrc_is_pid1 && {
 command -v openrc >/dev/null 2>&1 \
 || [ -x /sbin/openrc-run ] \
 || [ -f /sbin/openrc ]
 }; then
 echo "openrc"
 return 0
 fi
 if command -v initctl >/dev/null 2>&1 && [ -d /etc/init ]; then
 echo "upstart"
 return 0
 fi
 if [ -d /etc/s6 ] || command -v s6-svc >/dev/null 2>&1; then
 echo "s6"
 return 0
 fi
 if command -v runsvdir >/dev/null 2>&1 || [ -d /etc/runit ]; then
 echo "runit"
 return 0
 fi
 if [ -f /etc/inittab ] || command -v update-rc.d >/dev/null 2>&1 || command -v chkconfig >/dev/null 2>&1; then
 echo "sysvinit"
 return 0
 fi
 echo "unknown"
}

[ "$(id -u)" -eq 0 ] || die "Must be run as root. Try: sudo $0 $*"

OS=$(detect_os)
SHARE_DIR=$(share_dir_for_os "$OS")
BINARY_PATH=$(binary_path_for_os "$OS")

REMOVE_DATA=false
for arg in "$@"; do
 case "${arg}" in
 --remove-data) REMOVE_DATA=true ;;
 --help|-h)
 echo "Usage: $0 [--remove-data]"
 echo ""
 echo " --remove-data also delete config, data, and backup directories"
 exit 0 ;;
 *) die "Unknown option: ${arg}" ;;
 esac
done

stop_systemd() {
 if systemctl is-active --quiet edgelet-containerd 2>/dev/null; then
 systemctl stop edgelet-containerd
 fi
 if systemctl is-active --quiet edgelet 2>/dev/null; then
 systemctl stop edgelet
 fi
 systemctl disable edgelet edgelet-containerd 2>/dev/null || true
 rm -f /etc/systemd/system/edgelet.service
 rm -f /etc/systemd/system/edgelet-containerd.service
 rm -rf /etc/systemd/system/edgelet.service.d
 systemctl daemon-reload 2>/dev/null || true
 info "systemd service removed."
}

stop_procd() {
 /etc/init.d/edgelet stop 2>/dev/null || true
 /etc/init.d/edgelet disable 2>/dev/null || true
 rm -f /etc/init.d/edgelet
 info "procd init script removed."
}

stop_openrc() {
 rc-service edgelet-containerd stop 2>/dev/null || true
 rc-service edgelet stop 2>/dev/null || true
 rc-update del edgelet-containerd default 2>/dev/null || true
 rc-update del edgelet default 2>/dev/null || true
 rm -f /etc/init.d/edgelet /etc/init.d/edgelet-containerd
 info "OpenRC service removed."
}

stop_sysvinit() {
 /etc/init.d/edgelet stop 2>/dev/null || true
 update-rc.d -f edgelet remove 2>/dev/null || true
 chkconfig --del edgelet 2>/dev/null || true
 rm -f /etc/init.d/edgelet
 info "SysV init service removed."
}

stop_upstart() {
 initctl stop edgelet 2>/dev/null || true
 rm -f /etc/init/edgelet.conf
 initctl reload-configuration 2>/dev/null || true
 info "Upstart service removed."
}

stop_s6() {
 s6-svc -d /var/run/s6/services/edgelet 2>/dev/null || true
 rm -rf /var/run/s6/services/edgelet
 rm -rf /etc/s6/edgelet
 info "s6 service removed."
}

stop_runit() {
 sv down edgelet 2>/dev/null || true
 rm -f /etc/service/edgelet /var/service/edgelet /service/edgelet 2>/dev/null || true
 rm -rf /etc/runit/edgelet
 info "runit service removed."
}

stop_fallback() {
 pkill -f "/usr/local/bin/edgelet daemon" 2>/dev/null || true
 pkill -f "edgelet daemon" 2>/dev/null || true
 info "Background edgelet processes stopped (if any)."
}

lazy_umount_edgelet() {
 if ! command -v umount >/dev/null 2>&1; then
 return 0
 fi
 mount 2>/dev/null | grep -E '/run/edgelet|/var/run/edgelet|/var/lib/edgelet' | awk '{print $3}' | \
 sort -r | while read -r mp; do
 [ -n "$mp" ] || continue
 umount -l "${mp}" 2>/dev/null || true
 done
}

remove_init_service() {
 _init=$(detect_init 2>/dev/null) || _init="unknown"
 if [ -f /etc/systemd/system/edgelet.service ]; then
 stop_systemd
 elif [ -f /etc/init.d/edgelet ] && [ -x /sbin/procd ]; then
 stop_procd
 elif [ -f /etc/init.d/edgelet ] && command -v openrc >/dev/null 2>&1; then
 stop_openrc
 elif [ -f /etc/init/edgelet.conf ]; then
 stop_upstart
 elif [ -f /etc/init.d/edgelet ]; then
 stop_sysvinit
 elif [ -d /etc/s6/edgelet ]; then
 stop_s6
 elif [ -d /etc/runit/edgelet ] || [ -L /var/service/edgelet ]; then
 stop_runit
 elif [ "$_init" != "" ] && [ "$_init" != "unknown" ]; then
 case "$_init" in
 systemd) stop_systemd ;;
 openrc) stop_openrc ;;
 procd) stop_procd ;;
 sysvinit) stop_sysvinit ;;
 upstart) stop_upstart ;;
 s6) stop_s6 ;;
 runit) stop_runit ;;
 *) stop_fallback ;;
 esac
 else
 stop_fallback
 fi
}

remove_init_service
lazy_umount_edgelet

# Container deployment instance
EDGELET_CONTAINER_NAME="${EDGELET_CONTAINER_NAME:-edgelet}"
docker rm -f "$EDGELET_CONTAINER_NAME" 2>/dev/null || true
podman rm -f "$EDGELET_CONTAINER_NAME" 2>/dev/null || true

rm -f "$BINARY_PATH"
info "Binary removed."

case "$OS" in
 linux)
 rm -rf /usr/libexec/edgelet
 info "Init helpers removed from /usr/libexec/edgelet/"
 ;;
esac

rm -rf "$SHARE_DIR"
info "Bundled scripts removed from ${SHARE_DIR}/"

if [ "${REMOVE_DATA}" = "true" ]; then
 info "Removing agent data directories..."
 lazy_umount_edgelet
 case "$OS" in
 linux)
 rm -rf /var/lib/edgelet
 rm -rf /var/lib/edgelet-containerd
 rm -rf /run/edgelet
 rm -rf /var/log/edgelet
 rm -rf /var/backups/edgelet
 rm -rf /etc/edgelet
 ;;
 darwin)
 rm -rf /var/lib/edgelet
 rm -rf /var/run/edgelet
 rm -rf /var/log/edgelet
 rm -rf /var/backups/edgelet
 rm -rf /etc/edgelet
 ;;
 windows)
 _pd=$(windows_program_data_edgelet)
 rm -rf "${_pd}/data"
 rm -rf "${_pd}/config"
 rm -rf "${_pd}/run"
 rm -rf "${_pd}/log"
 rm -rf "${_pd}/scripts"
 rm -rf "${_pd}"
 ;;
 esac
 info "Data, backups, and configuration removed."
else
 info "Data directories preserved (use --remove-data to remove):"
 case "$OS" in
 linux)
 info " /var/lib/edgelet"
 info " /var/lib/edgelet-containerd"
 info " /run/edgelet"
 info " /var/log/edgelet"
 info " /var/backups/edgelet"
 info " /etc/edgelet"
 ;;
 darwin)
 info " /var/lib/edgelet"
 info " /var/run/edgelet"
 info " /var/log/edgelet"
 info " /var/backups/edgelet"
 info " /etc/edgelet"
 ;;
 windows)
 _pd=$(windows_program_data_edgelet)
 info " ${_pd}/data"
 info " ${_pd}/config"
 info " ${_pd}/run"
 info " ${_pd}/log"
 ;;
 esac
fi

info ""
info "Edgelet has been uninstalled."
