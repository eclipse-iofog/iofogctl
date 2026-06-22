#!/bin/sh
# Container engine socket URL/path helpers (aligned with pkg/iofog/install/edgelet_config.go).

if ! type die >/dev/null 2>&1; then
	die() { echo "ERROR: $1" >&2; exit 1; }
fi

default_container_engine_url() {
	_engine=$(echo "$1" | tr '[:upper:]' '[:lower:]')
	case "$_engine" in
		docker) echo "unix:///var/run/docker.sock" ;;
		podman) echo "unix:///run/podman/podman.sock" ;;
		*) echo "unix:///run/edgelet/containerd.sock" ;;
	esac
}

unix_url_to_path() {
	_url="$1"
	case "$_url" in
		unix://*) echo "${_url#unix://}" ;;
		*) echo "$_url" ;;
	esac
}

read_edgelet_config_profile_value() {
	_key="$1"
	_file="${EDGELET_CONFIG_FILE:-/etc/edgelet/config.yaml}"
	[ -f "$_file" ] || return 1
	awk -v key="$_key" '
		/^  production:/ { p=1; next }
		p && /^  [a-zA-Z]/ && !/^  production:/ { exit }
		p && index($0, "    " key ":") == 1 {
			sub(/^    [^:]*:[[:space:]]*/, "")
			gsub(/^"/, ""); gsub(/"$/, "")
			gsub(/^'\''/, ""); gsub(/'\''$/, "")
			print
			exit
		}
	' "$_file"
}

resolve_container_engine_sock_url() {
	if [ -n "${EDGELET_CONTAINER_ENGINE_URL:-}" ]; then
		echo "$EDGELET_CONTAINER_ENGINE_URL"
		return 0
	fi
	_url=$(read_edgelet_config_profile_value "containerEngineUrl" 2>/dev/null) || true
	if [ -n "$_url" ]; then
		echo "$_url"
		return 0
	fi
	_engine="${CONTAINER_ENGINE:-docker}"
	_cfg_engine=$(read_edgelet_config_profile_value "containerEngine" 2>/dev/null) || true
	if [ -n "$_cfg_engine" ]; then
		_engine="$_cfg_engine"
	fi
	default_container_engine_url "$_engine"
}

resolve_container_engine_sock_path() {
	unix_url_to_path "$(resolve_container_engine_sock_url)"
}

container_engine_sock_mount() {
	_path=$(resolve_container_engine_sock_path)
	[ -n "$_path" ] || die "container engine socket path is empty"
	echo "-v ${_path}:${_path}:rw"
}
