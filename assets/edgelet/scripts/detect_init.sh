#!/bin/sh
# Detect host OS, arch, and init system for edgelet install layers.
set -e

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

detect_os() {
	case "$(uname -s)" in
		Linux) EDGELET_OS=linux ;;
		Darwin) EDGELET_OS=darwin ;;
		CYGWIN*|MINGW*|MSYS*) EDGELET_OS=windows ;;
		*)
			echo "Error: unsupported operating system: $(uname -s)"
			exit 1
			;;
	esac
	export EDGELET_OS
}

detect_arch() {
	if [ -n "${EDGELET_ARCH:-}" ] && [ "$EDGELET_ARCH" != "auto" ]; then
		export EDGELET_ARCH
		return
	fi

	machine="$(uname -m)"
	case "$machine" in
		x86_64|amd64) EDGELET_ARCH=amd64 ;;
		aarch64|arm64) EDGELET_ARCH=arm64 ;;
		armv7l|armv6l|arm) EDGELET_ARCH=arm ;;
		riscv64) EDGELET_ARCH=riscv64 ;;
		*)
			echo "Error: unsupported architecture: $machine"
			exit 1
			;;
	esac
	export EDGELET_ARCH
}

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

detect_init_system() {
	INIT_SYSTEM=unknown
	if [ "$EDGELET_OS" = "linux" ]; then
		if command_exists systemctl && [ -d /etc/systemd/system ]; then
			INIT_SYSTEM=systemd
		elif procd_is_openwrt; then
			INIT_SYSTEM=procd
		elif openrc_is_pid1 && {
			command_exists openrc \
			|| [ -x /sbin/openrc-run ] \
			|| [ -f /sbin/openrc ]
		}; then
			INIT_SYSTEM=openrc
		elif command_exists initctl && [ -d /etc/init ]; then
			INIT_SYSTEM=upstart
		elif [ -d /etc/s6 ] || command_exists s6-svc; then
			INIT_SYSTEM=s6
		elif command_exists runsvdir || [ -d /etc/runit ]; then
			INIT_SYSTEM=runit
		elif [ -f /etc/inittab ] || command_exists update-rc.d || command_exists chkconfig; then
			INIT_SYSTEM=sysvinit
		fi
	elif [ "$EDGELET_OS" = "darwin" ]; then
		INIT_SYSTEM=launchd
	elif [ "$EDGELET_OS" = "windows" ]; then
		INIT_SYSTEM=windows
	fi
	export INIT_SYSTEM
}

init() {
	detect_os
	detect_arch
	detect_init_system
	echo "# Detected edgelet host: os=$EDGELET_OS arch=$EDGELET_ARCH init=$INIT_SYSTEM"
}

if [ "${BASH_SOURCE:-$0}" = "$0" ] && [ -z "${EDGELET_DETECT_SOURCED:-}" ]; then
	init
fi
