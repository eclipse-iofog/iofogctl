#!/bin/sh
# Embed-hash helpers for conditional edgelet-containerd restart (upstream install.sh 9a9ded8 parity).
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

# installed_embed_hash returns the basename of data/current (embed SHA256 prefix), or empty.
installed_embed_hash() {
	_link="/var/lib/edgelet/data/current"
	[ -L "$_link" ] || return 0
	basename "$(readlink -f "$_link" 2>/dev/null || readlink "$_link")"
}

# binary_embed_hash reads embed hash from a linux thin binary (edgelet version --verbose).
binary_embed_hash() {
	_bin="$1"
	[ -x "$_bin" ] || return 0
	"$_bin" version --verbose 2>/dev/null \
		| sed -n 's/^  embed hash: //p' \
		| head -1
}

# should_restart_data_plane is true when containerEngine=edgelet and the embed hash changed
# (or data/current is not yet set). Returns false for docker/podman or lite builds without embed.
should_restart_data_plane() {
	_eng="$1"
	_old="$2"
	_new="$3"
	[ "$_eng" = "edgelet" ] || return 1
	[ -n "$_new" ] || return 1
	[ -z "$_old" ] && return 0
	[ "$_old" != "$_new" ]
}

start_edgelet_containerd_unit() {
	_init="$1"
	_restart="$2"
	case "${_init}" in
		systemd)
			maybe_sudo systemctl enable edgelet-containerd 2>/dev/null || true
			if [ "$_restart" = true ]; then
				info "Embedded bundle hash changed; restarting edgelet-containerd (data plane)"
				maybe_sudo systemctl stop edgelet-containerd 2>/dev/null || true
				maybe_sudo systemctl reset-failed edgelet-containerd 2>/dev/null || true
				maybe_sudo systemctl start edgelet-containerd
			else
				maybe_sudo systemctl start edgelet-containerd 2>/dev/null || true
			fi
			;;
		openrc)
			maybe_sudo rc-update add edgelet-containerd default 2>/dev/null || true
			if [ "$_restart" = true ]; then
				info "Embedded bundle hash changed; restarting edgelet-containerd (data plane)"
				maybe_sudo rc-service edgelet-containerd stop 2>/dev/null || true
				maybe_sudo rc-service edgelet-containerd start 2>/dev/null || true
			else
				maybe_sudo rc-service edgelet-containerd start 2>/dev/null || true
			fi
			;;
	esac
}

record_restart_data_plane_decision() {
	_old_embed="$1"
	_restart_dp=false
	if [ "$OS" = "linux" ]; then
		_new_embed=$(binary_embed_hash "$BINARY_PATH")
		if should_restart_data_plane "$CONTAINER_ENGINE" "$_old_embed" "$_new_embed"; then
			_restart_dp=true
			info "Embed hash: ${_old_embed:-<none>} -> ${_new_embed}; data-plane restart required"
		fi
	fi
	write_restart_data_plane_marker "$_restart_dp"
}
