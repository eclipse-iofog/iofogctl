#!/bin/sh
# Binary download, verification, and install helpers.

GITHUB_REPO="${EDGELET_GITHUB_REPO:-eclipse-iofog/edgelet}"

release_download_url() {
	_ver="$1"
	_os="$2"
	_arch="$3"
	echo "https://github.com/${GITHUB_REPO}/releases/download/${_ver}/$(binary_basename "$_os" "$_arch")"
}

verify_binary_checksum() {
	_bin="$1"
	[ -f "$_bin" ] || die "Not a file: $_bin"
	if [ -n "$EXPECTED_SHA256" ]; then
		_sum=$(sha256_file "$_bin")
		[ "$_sum" = "$EXPECTED_SHA256" ] || die "SHA256 mismatch (expected $EXPECTED_SHA256 got $_sum)"
		info "SHA256 verified."
	elif [ -n "$CHECKSUM_FILE" ] && [ -f "$CHECKSUM_FILE" ]; then
		_bn=$(basename "$_bin")
		( cd "$(dirname "$_bin")" && grep " ${_bn}\$" "$CHECKSUM_FILE" >/dev/null ) || \
			( cd "$(dirname "$CHECKSUM_FILE")" && sha256sum -c "$CHECKSUM_FILE" ) || \
			die "Checksum file verification failed"
	fi
}

download_or_stage_binary() {
	_dest="$1"
	if [ -n "$BIN_PATH" ]; then
		[ -f "$BIN_PATH" ] || die "Local binary not found: $BIN_PATH"
		verify_binary_checksum "$BIN_PATH"
		cp "$BIN_PATH" "$_dest"
		info "Using local binary: ${BIN_PATH}"
		return 0
	fi
	if [ "$AIRGAP" = true ]; then
		die "--airgap requires --bin-path"
	fi
	_url=$(release_download_url "$EDGELET_VERSION" "$OS" "$ARCH")
	info "Downloading ${_url} ..."
	curl -fsSL -o "$_dest" "$_url" || die "Failed to download release binary"
}

compute_source_url() {
	if [ -n "$BIN_PATH" ]; then
		_real=$(cd "$(dirname "$BIN_PATH")" && pwd)/$(basename "$BIN_PATH")
		echo "file://${_real}"
	else
		release_download_url "$EDGELET_VERSION" "$OS" "$ARCH"
	fi
}

install_binary_file() {
	_src="$1"
	_dest="$2"
	_dir=$(dirname "$_dest")
	maybe_sudo mkdir -p "$_dir"
	maybe_sudo install -m 755 "$_src" "$_dest"
	info "Installed ${_dest}"
}

fetch_latest_version() {
	_ver=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
		| grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/') || true
	[ -n "$_ver" ] || die "Failed to determine latest version"
	echo "$_ver"
}

stop_edgelet_daemon_desktop() {
	_os="$1"
	if pgrep -f '[e]dgelet daemon' >/dev/null 2>&1; then
		pkill -f '[e]dgelet daemon' 2>/dev/null || true
		sleep 1
		info "edgelet daemon stopped."
	fi
	case "${_os}" in
		darwin) rm -f /var/run/edgelet/edgelet.pid ;;
		windows) rm -f "$(windows_program_data_edgelet)/run/edgelet.pid" ;;
	esac
}
