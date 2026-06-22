#!/bin/sh
# Shared helpers for potctl edgelet install scripts.

die() { echo "ERROR: $1" >&2; exit 1; }
info() { echo ">>> $1"; }

edgelet_host_os() {
	case "${EDGELET_OS:-}" in
		darwin|windows|linux)
			echo "$EDGELET_OS"
			;;
		*)
			case "$(uname -s)" in
				Darwin) echo darwin ;;
				MINGW*|MSYS*|CYGWIN*) echo windows ;;
				*) echo linux ;;
			esac
			;;
	esac
}

is_desktop_container_host() {
	case "$(edgelet_host_os)" in
		darwin|windows) return 0 ;;
		*) return 1 ;;
	esac
}

# desktop_container_local is true for potctl local install of container edgelet on darwin/windows.
desktop_container_local() {
	[ "${LOCAL_INSTALL:-0}" = "1" ] \
		&& [ "${EDGELET_INSTALL_MODE:-native}" = "container" ] \
		&& is_desktop_container_host
}

maybe_sudo() {
	if desktop_container_local; then
		"$@"
		return
	fi
	if [ "$(id -u)" -eq 0 ]; then
		"$@"
	else
		sudo "$@"
	fi
}
