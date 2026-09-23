#!/bin/sh
# Embed-hash helpers for fat OTA drain before binary replace.
# Requires lib/common.sh (maybe_sudo, info) when sourced.

restart_data_plane_marker_path() {
	printf '%s' "${EDGELET_SCRIPT_STAGE_DIR:-${SCRIPT_DIR:-/tmp/edgelet-scripts}}/.restart-data-plane"
}

write_restart_data_plane_marker() {
	_val="$1"
	_path=$(restart_data_plane_marker_path)
	_dir=$(dirname "$_path")
	maybe_sudo mkdir -p "$_dir" 2>/dev/null || mkdir -p "$_dir"
	if [ "$_val" = true ]; then
		echo true > "$_path"
	else
		echo false > "$_path"
	fi
}

clear_restart_data_plane_marker() {
	rm -f "$(restart_data_plane_marker_path)"
}

restart_data_plane_marker_present() {
	[ -f "$(restart_data_plane_marker_path)" ]
}

restart_data_plane_requested() {
	_path=$(restart_data_plane_marker_path)
	[ -f "$_path" ] || return 1
	_val=$(tr -d '[:space:]' < "$_path" 2>/dev/null || true)
	[ "$_val" = true ]
}

consume_restart_data_plane_marker() {
	_requested=false
	if restart_data_plane_requested; then
		_requested=true
	fi
	clear_restart_data_plane_marker
	[ "$_requested" = true ]
}

edgelet_data_root() {
	if [ -n "${EDGELET_DATA_DIR:-}" ]; then
		echo "${EDGELET_DATA_DIR}"
		return 0
	fi
	echo "/var/lib/edgelet"
}

# embed_bundle_ready is true when dir has an executable fat runtime and the
# same companion files the extract path requires before it will run.
embed_bundle_ready() {
	_dir="$1"
	[ -n "$_dir" ] || return 1
	_fat="${_dir}/bin/edgelet"
	[ -f "$_fat" ] && [ -x "$_fat" ] || return 1
	_magic=$(od -An -tx1 -N 4 "$_fat" 2>/dev/null | tr -d ' \n' || true)
	[ "$_magic" = "7f454c46" ] || return 1
	[ -f "${_dir}/bin/containerd-shim-runc-v2" ] || return 1
	[ -f "${_dir}/bin/aux/xtables-legacy-multi" ] || return 1
	[ -f "${_dir}/bin/ip" ] || return 1
	[ -f "${_dir}/bin/busybox" ] || return 1
	[ -L "${_dir}/bin/aux/iptables" ] || return 1
	[ "$(readlink "${_dir}/bin/aux/iptables" 2>/dev/null || true)" = "xtables-legacy-multi" ]
}

# installed_embed_hash prints the ready data/current bundle hash, or nothing
# when the symlink is missing or the fat runtime is not ready.
installed_embed_hash() {
	_link="$(edgelet_data_root)/data/current"
	[ -L "$_link" ] || return 0
	_target=$(readlink "$_link" 2>/dev/null || true)
	[ -n "$_target" ] || return 0
	case "$_target" in
		/*) ;;
		*) _target="$(CDPATH= cd -- "$(dirname "$_link")" && pwd)/${_target}" ;;
	esac
	embed_bundle_ready "$_target" || return 0
	basename "$_target"
}

# binary_embed_hash reads embed hash from a linux thin binary (edgelet version --verbose).
binary_embed_hash() {
	_bin="$1"
	[ -x "$_bin" ] || return 0
	"$_bin" version --verbose 2>/dev/null \
		| sed -n 's/^  embed hash: //p' \
		| head -1
}

# should_restart_data_plane is true when containerEngine=edgelet, the new embed
# hash is set, and the ready current bundle is missing or different.
# Returns false for docker/podman or a thin binary with no embed hash.
should_restart_data_plane() {
	_eng="$1"
	_old="$2"
	_new="$3"
	[ "$_eng" = "edgelet" ] || return 1
	[ -n "$_new" ] || return 1
	[ -z "$_old" ] && return 0
	[ "$_old" != "$_new" ]
}

# quiesce_data_plane_for_replace drains labeled workloads through CRI using the
# staged binary. Control stays up. It does not use edgelet.sock and does not
# fall back to the installed binary. Non-zero leaves that binary in place.
quiesce_data_plane_for_replace() {
	_bin="$1"
	if [ -z "$_bin" ] || [ ! -x "$_bin" ]; then
		return 1
	fi
	info "Draining data plane before embed replace"
	"$_bin" --quiet runtime drain --direct
}

# replace_staged_edgelet_binary installs a staged thin binary.
# Fat OTA drains and verifies while control is still up, stops the data plane,
# then replaces the binary. A matching ready hash, or docker/podman, stops
# control only and does not stop edgelet-containerd.
# Verify failure exits non-zero and does not replace the binary.
# Sets OTA_RESTART_DATA_PLANE=true or false and writes the start-layer marker.
replace_staged_edgelet_binary() {
	_staged="$1"
	OTA_RESTART_DATA_PLANE=false
	[ -f "$_staged" ] || die "Staged binary missing: ${_staged}"
	chmod 755 "$_staged" || die "Cannot prepare staged binary: ${_staged}"
	if [ "$OS" = "linux" ]; then
		_old_embed=$(installed_embed_hash)
		_new_embed=$(binary_embed_hash "$_staged")
		if should_restart_data_plane "$CONTAINER_ENGINE" "$_old_embed" "$_new_embed"; then
			OTA_RESTART_DATA_PLANE=true
			info "Embed hash: ${_old_embed:-<none>} -> ${_new_embed}; data-plane restart required"
			if quiesce_data_plane_for_replace "$_staged"; then
				stop_edgelet_containerd_unit "$INIT"
			elif ! edgelet_containerd_unit_active \
				&& [ "$(binary_embed_hash "$BINARY_PATH")" = "$_new_embed" ]; then
				info "Data plane already stopped after a previous verified drain; binary replace continues"
			else
				die "Data-plane drain did not verify; binary was not replaced"
			fi
		else
			stop_edgelet_service "$INIT"
		fi
	else
		stop_edgelet_daemon_desktop "$OS"
	fi
	install_binary_file "$_staged" "$BINARY_PATH"
	write_restart_data_plane_marker "$OTA_RESTART_DATA_PLANE"
}

start_edgelet_containerd_unit() {
	_init="$1"
	_restart="$2"
	case "${_init}" in
		systemd)
			maybe_sudo systemctl enable edgelet-containerd 2>/dev/null || true
			if [ "$_restart" = true ]; then
				info "Embedded bundle hash changed; starting edgelet-containerd (data plane)"
				maybe_sudo systemctl reset-failed edgelet-containerd 2>/dev/null || true
				maybe_sudo systemctl start edgelet-containerd
			else
				maybe_sudo systemctl start edgelet-containerd 2>/dev/null || true
			fi
			;;
		openrc)
			maybe_sudo rc-update add edgelet-containerd default 2>/dev/null || true
			if [ "$_restart" = true ]; then
				info "Embedded bundle hash changed; starting edgelet-containerd (data plane)"
				maybe_sudo rc-service edgelet-containerd start 2>/dev/null || true
			else
				maybe_sudo rc-service edgelet-containerd start 2>/dev/null || true
			fi
			;;
	esac
}

# record_restart_data_plane_decision compares a ready embed hash to a binary.
# Pass the staged binary as $2. The default is the installed path, which is
# only valid before that path has been replaced.
record_restart_data_plane_decision() {
	_old_embed="$1"
	_bin="${2:-$BINARY_PATH}"
	_restart_dp=false
	if [ "$OS" = "linux" ]; then
		_new_embed=$(binary_embed_hash "$_bin")
		if should_restart_data_plane "$CONTAINER_ENGINE" "$_old_embed" "$_new_embed"; then
			_restart_dp=true
			info "Embed hash: ${_old_embed:-<none>} -> ${_new_embed}; data-plane restart required"
		fi
	fi
	write_restart_data_plane_marker "$_restart_dp"
}
