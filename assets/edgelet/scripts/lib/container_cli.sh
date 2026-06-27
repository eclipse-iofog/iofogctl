#!/bin/sh
# Host CLI wrapper for containerized edgelet (docker/podman exec into running container).

install_container_cli_wrapper() {
	_engine="$1"
	_container="${2:-edgelet}"
	_os="${3:-linux}"

	case "$_engine" in
		docker|podman) ;;
		*) die "container CLI wrapper requires docker or podman engine (got $_engine)" ;;
	esac

	case "$_os" in
		linux|darwin) ;;
		*) die "container CLI wrapper is unsupported on os=$_os" ;;
	esac

	if desktop_container_local; then
		_stage="${EDGELET_SCRIPT_STAGE_DIR:-}"
		[ -n "$_stage" ] || die "EDGELET_SCRIPT_STAGE_DIR is required for desktop local container install"
		_dest="${_stage}/bin/edgelet"
		mkdir -p "${_stage}/bin"
		cat <<EOF > "$_dest"
#!/bin/sh
CONTAINER_NAME="${_container}"
ENGINE="${_engine}"

if [ "\$1" = "daemon" ]; then
	echo "Error: edgelet daemon is managed by the \${ENGINE} container service." >&2
	exit 1
fi

if ! "\$ENGINE" ps --format '{{.Names}}' 2>/dev/null | grep -q "^\${CONTAINER_NAME}\$"; then
	echo "Error: The edgelet container is not running." >&2
	exit 1
fi
exec "\$ENGINE" exec "\$CONTAINER_NAME" edgelet "\$@"
EOF
		chmod 755 "$_dest"
		info "Installed container CLI wrapper at ${_dest} (engine=${_engine} container=${_container})"
		return 0
	fi

	_dest="$(binary_path_for_os "$_os")"
	maybe_sudo mkdir -p "$(dirname "$_dest")"
	cat <<EOF | maybe_sudo tee "$_dest" >/dev/null
#!/bin/sh
CONTAINER_NAME="${_container}"
ENGINE="${_engine}"

if [ "\$1" = "daemon" ]; then
	echo "Error: edgelet daemon is managed by the \${ENGINE} container service." >&2
	exit 1
fi

if ! "\$ENGINE" ps --format '{{.Names}}' 2>/dev/null | grep -q "^\${CONTAINER_NAME}\$"; then
	echo "Error: The edgelet container is not running." >&2
	exit 1
fi
exec "\$ENGINE" exec "\$CONTAINER_NAME" edgelet "\$@"
EOF
	maybe_sudo chmod 755 "$_dest"
	info "Installed container CLI wrapper at ${_dest} (engine=${_engine} container=${_container})"
}
