#!/bin/sh
# OS-specific edgelet paths (aligned with upstream edgelet install.sh).

windows_program_data_edgelet() {
	echo "${ProgramData:-/c/ProgramData}/Edgelet"
}

share_dir_for_os() {
	case "$1" in
		linux) echo "/usr/share/edgelet" ;;
		darwin) echo "/usr/local/share/edgelet" ;;
		windows) echo "$(windows_program_data_edgelet)/scripts" ;;
		*) die "Unsupported OS for share dir: $1" ;;
	esac
}

binary_path_for_os() {
	case "$1" in
		linux|darwin) echo "/usr/local/bin/edgelet" ;;
		windows)
			_pf="${ProgramFiles:-/c/Program Files}"
			echo "${_pf}/Edgelet/edgelet.exe"
			;;
		*) die "Unsupported OS for binary path: $1" ;;
	esac
}

init_platform_paths() {
	_os="$1"
	case "${_os}" in
		linux)
			SHARE_DIR="/usr/share/edgelet"
			CONFIG_DIR="/etc/edgelet"
			RUNTIME_DIR="/run/edgelet"
			;;
		darwin)
			SHARE_DIR="/usr/local/share/edgelet"
			CONFIG_DIR="/etc/edgelet"
			RUNTIME_DIR="/var/run/edgelet"
			;;
		windows)
			_pd=$(windows_program_data_edgelet)
			SHARE_DIR="${_pd}/scripts"
			CONFIG_DIR="${_pd}/config"
			RUNTIME_DIR="${_pd}/run"
			;;
		*) die "Unsupported OS for platform paths: ${_os}" ;;
	esac
	CONFIG_FILE="${CONFIG_DIR}/config.yaml"
	CERT_FILE="${CONFIG_DIR}/cert.crt"
	export SHARE_DIR CONFIG_DIR CONFIG_FILE CERT_FILE RUNTIME_DIR
}

install_dirs_for_os() {
	_os="$1"
	_share="$(share_dir_for_os "$_os")"
	case "${_os}" in
		linux)
			maybe_sudo mkdir -p /etc/edgelet /var/log/edgelet /var/lib/edgelet /run/edgelet \
				/var/lib/edgelet-containerd /var/backups/edgelet /var/backups/edgelet/cache "$_share"
			maybe_sudo chmod 750 /etc/edgelet /var/log/edgelet /var/lib/edgelet 2>/dev/null || true
			;;
		darwin)
			maybe_sudo mkdir -p /etc/edgelet /var/log/edgelet /var/lib/edgelet /var/lib/edgelet-containerd /var/run/edgelet \
				/var/backups/edgelet /var/backups/edgelet/cache "$_share"
			maybe_sudo chmod 750 /etc/edgelet /var/log/edgelet /var/lib/edgelet 2>/dev/null || true
			;;
		windows)
			_pd=$(windows_program_data_edgelet)
			maybe_sudo mkdir -p "${_pd}/data" "${_pd}/config" "${_pd}/run" "${_pd}/log" "${_pd}/scripts" 2>/dev/null || true
			;;
	esac
}

binary_basename() {
	_os="$1"
	_arch="$2"
	case "${_os}" in
		windows) echo "edgelet-${_os}-${_arch}.exe" ;;
		*) echo "edgelet-${_os}-${_arch}" ;;
	esac
}

default_container_engine_for_os() {
	case "$1" in
		linux) echo "edgelet" ;;
		*) echo "docker" ;;
	esac
}
