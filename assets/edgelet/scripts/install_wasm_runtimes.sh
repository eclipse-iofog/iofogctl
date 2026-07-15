#!/bin/sh
# Install pre-staged WASM containerd shims. Go resolves/extracts artifacts; no curl/tar here.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/service.sh"

if [ "${EDGELET_WASM_INSTALL:-}" != "1" ]; then
	exit 0
fi

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

if [ "$EDGELET_OS" != "linux" ]; then
	info "Skipping WASM runtime install on os=$EDGELET_OS"
	exit 0
fi

if [ "${DEPLOYMENT_TYPE:-native}" = "container" ]; then
	info "Skipping WASM runtime install for deploymentType=container"
	exit 0
fi

CONTAINER_ENGINE="${CONTAINER_ENGINE:-edgelet}"

install_one() {
	_handler="$1"
	_src="$2"
	_name="$3"
	[ -n "$_handler" ] || return 0
	[ -n "$_src" ] || return 0
	[ -n "$_name" ] || return 0
	[ -f "$_src" ] || die "WASM shim source missing for ${_handler}: ${_src}"
	_dest="/usr/local/bin/${_name}"
	if [ -f "$_dest" ] && cmp -s "$_src" "$_dest" 2>/dev/null; then
		info "WASM shim ${_name} unchanged"
		return 0
	fi
	maybe_sudo install -m 755 "$_src" "$_dest"
	info "Installed WASM shim ${_name}"
	echo 1 > "${EDGELET_WASM_CHANGED_FLAG:-/tmp/edgelet-wasm-changed.$$}"
}

EDGELET_WASM_CHANGED_FLAG="${EDGELET_SCRIPT_STAGE_DIR:-/tmp/edgelet-scripts}/wasm/.changed"
maybe_sudo mkdir -p "$(dirname "$EDGELET_WASM_CHANGED_FLAG")"
rm -f "$EDGELET_WASM_CHANGED_FLAG"

handler_env_key() {
	echo "$1" | tr '[:lower:]' '[:upper:]' | tr '-' '_'
}

install_from_manifest() {
	_manifest="$1"
	[ -f "$_manifest" ] || die "WASM manifest not found: $_manifest"
	# shellcheck disable=SC2016
	grep -o '{"handler":"[^"]*","src":"[^"]*","name":"[^"]*"}' "$_manifest" | while read -r _obj; do
		_handler=$(printf '%s' "$_obj" | sed -n 's/.*"handler":"\([^"]*\)".*/\1/p')
		_src=$(printf '%s' "$_obj" | sed -n 's/.*"src":"\([^"]*\)".*/\1/p')
		_name=$(printf '%s' "$_obj" | sed -n 's/.*"name":"\([^"]*\)".*/\1/p')
		install_one "$_handler" "$_src" "$_name"
	done
}

install_from_env() {
	_handlers="${EDGELET_WASM_HANDLERS:-}"
	[ -n "$_handlers" ] || return 0
	_old_ifs=$IFS
	IFS=','
	# shellcheck disable=SC2086
	set -- $_handlers
	IFS=$_old_ifs
	for _handler in "$@"; do
		_key=$(handler_env_key "$_handler")
		eval "_src=\${EDGELET_WASM_BIN_${_key}:-}"
		eval "_name=\${EDGELET_WASM_NAME_${_key}:-}"
		install_one "$_handler" "$_src" "$_name"
	done
}

wait_docker_ready() {
	_timeout="${EDGELET_DOCKER_WAIT_SEC:-120}"
	_elapsed=0
	while [ "$_elapsed" -lt "$_timeout" ]; do
		if command -v docker >/dev/null 2>&1 && docker ps >/dev/null 2>&1; then
			return 0
		fi
		sleep 2
		_elapsed=$(( _elapsed + 2 ))
	done
	echo "ERROR: docker not ready after ${_timeout}s" >&2
	return 1
}

restart_engine_if_needed() {
	_changed=0
	if [ -f "$EDGELET_WASM_CHANGED_FLAG" ]; then
		_changed=1
	fi
	rm -f "$EDGELET_WASM_CHANGED_FLAG"

	if [ "$_changed" != "1" ]; then
		info "WASM shims unchanged; skipping engine restart"
		return 0
	fi
	if [ "${EDGELET_WASM_SKIP_RESTART:-}" = "1" ]; then
		info "Skipping engine restart (fresh install or pre-start)"
		return 0
	fi
	if [ "${EDGELET_WASM_DEFER_ENGINE_RESTART:-}" = "1" ]; then
		info "Deferring engine restart to potctl (WASM shims installed)"
		return 0
	fi
	case "$CONTAINER_ENGINE" in
		edgelet)
			info "Restarting edgelet-containerd after WASM shim update"
			restart_edgelet_containerd_service
			mark_containerd_restarted
			;;
		docker)
			info "Restarting docker after WASM shim update"
			case "${INIT_SYSTEM:-unknown}" in
				systemd) maybe_sudo systemctl restart docker ;;
				*) maybe_sudo systemctl restart docker 2>/dev/null || maybe_sudo service docker restart 2>/dev/null || true ;;
			esac
			wait_docker_ready
			;;
		*)
			info "Skipping engine restart for containerEngine=$CONTAINER_ENGINE"
			;;
	esac
}

if [ -n "${EDGELET_WASM_HANDLERS:-}" ]; then
	install_from_env
elif [ -n "${EDGELET_WASM_MANIFEST:-}" ] && [ -f "$EDGELET_WASM_MANIFEST" ]; then
	install_from_manifest "$EDGELET_WASM_MANIFEST"
fi

restart_engine_if_needed
