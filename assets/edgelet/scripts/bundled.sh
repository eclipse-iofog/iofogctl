#!/bin/sh
# Publish potctl edgelet script bundle into the canonical OS share directory.
set -e

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/paths.sh"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init
init_platform_paths "$EDGELET_OS"

if desktop_container_local; then
	info "Skipping bundled publish on desktop local container deploy"
	exit 0
fi

_share="$(share_dir_for_os "$EDGELET_OS")"
maybe_sudo mkdir -p "$_share" "$_share/lib"
maybe_sudo chmod 755 "$_share" "$_share/lib" 2>/dev/null || true

_publish() {
	_src="$1"
	_name="$2"
	[ -f "$_src" ] || return 0
	maybe_sudo install -m 755 "$_src" "$_share/$_name"
}

for _script in \
	check_prereqs.sh \
	detect_init.sh \
	install_deps.sh \
	configure_container_engine.sh \
	install.sh \
	install_wasm_runtimes.sh \
	uninstall.sh \
	install_init_units.sh \
	install_container.sh \
	start_edgelet.sh \
	probe_containerd_ready.sh \
	probe_edgelet_ready.sh \
	wait_edgelet_ready.sh \
	bundled.sh
do
	_publish "$SCRIPT_DIR/$_script" "$_script"
done

for _lib in "$SCRIPT_DIR"/lib/*.sh; do
	[ -f "$_lib" ] || continue
	_publish "$_lib" "lib/$(basename "$_lib")"
done

# Persist the installing CLI's release repo. A later edgelet --upgrade runs
# these scripts with no EDGELET_GITHUB_REPO in the environment.
if [ -n "${EDGELET_GITHUB_REPO:-}" ]; then
	case "$EDGELET_GITHUB_REPO" in
		*[!A-Za-z0-9._/-]*)
			die "Invalid EDGELET_GITHUB_REPO"
			;;
	esac
	_flavor_tmp=$(mktemp)
	printf 'EDGELET_GITHUB_REPO=%s\n' "$EDGELET_GITHUB_REPO" > "$_flavor_tmp"
	maybe_sudo install -m 644 "$_flavor_tmp" "$_share/flavor.env"
	rm -f "$_flavor_tmp"
	info "Pinned EDGELET_GITHUB_REPO=${EDGELET_GITHUB_REPO}"
fi

info "Bundled edgelet scripts at ${_share}/"
