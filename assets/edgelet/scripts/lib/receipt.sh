#!/bin/sh
# Install receipt and rollback metadata (/var/backups/edgelet).

BACKUP_DIR="${BACKUP_DIR:-/var/backups/edgelet}"
CACHE_DIR="${CACHE_DIR:-${BACKUP_DIR}/cache}"
RECEIPT_FILE="${RECEIPT_FILE:-${BACKUP_DIR}/install-receipt}"
PREVIOUS_FILE="${PREVIOUS_FILE:-${BACKUP_DIR}/previous-release}"

sha256_file() {
	sha256sum "$1" | awk '{print $1}'
}

kv_get() {
	_file="$1"
	_key="$2"
	[ -f "$_file" ] || { echo ""; return 0; }
	_line=$(grep "^${_key}=" "$_file" | head -1) || true
	[ -n "$_line" ] || { echo ""; return 0; }
	echo "$_line" | sed "s/^${_key}=//"
}

write_install_receipt() {
	_ver="$1"
	_os="$2"
	_arch="$3"
	_eng="$4"
	_url="$5"
	_sha="$6"
	_method="$7"
	maybe_sudo mkdir -p "$BACKUP_DIR"
	_ts=$(date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date -u)
	{
		printf 'installed_version=%s\n' "$_ver"
		printf 'os=%s\n' "$_os"
		printf 'arch=%s\n' "$_arch"
		printf 'container_engine=%s\n' "$_eng"
		printf 'source_url=%s\n' "$_url"
		printf 'installed_at=%s\n' "$_ts"
		printf 'install_method=%s\n' "$_method"
		printf 'binary_sha256=%s\n' "$_sha"
	} | maybe_sudo tee "$RECEIPT_FILE" >/dev/null
	maybe_sudo chmod 600 "$RECEIPT_FILE" 2>/dev/null || true
}

write_previous_release() {
	_pv="$1"
	_pos="$2"
	_parch="$3"
	_peng="$4"
	_purl="$5"
	_psha="$6"
	_cfg="$7"
	maybe_sudo mkdir -p "$BACKUP_DIR"
	{
		printf 'previous_version=%s\n' "$_pv"
		printf 'previous_os=%s\n' "$_pos"
		printf 'previous_arch=%s\n' "$_parch"
		printf 'previous_container_engine=%s\n' "$_peng"
		printf 'previous_download_url=%s\n' "$_purl"
		printf 'previous_binary_sha256=%s\n' "$_psha"
		printf 'config_backup_path=%s\n' "$_cfg"
	} | maybe_sudo tee "$PREVIOUS_FILE" >/dev/null
	maybe_sudo chmod 600 "$PREVIOUS_FILE" 2>/dev/null || true
}

cache_binary() {
	_ver="$1"
	_os="$2"
	_arch="$3"
	_src="$4"
	maybe_sudo mkdir -p "$CACHE_DIR"
	_dest="${CACHE_DIR}/edgelet-${_ver}-${_os}-${_arch}"
	case "${_os}" in
		windows) _dest="${_dest}.exe" ;;
	esac
	maybe_sudo cp "$_src" "$_dest"
	maybe_sudo chmod 755 "$_dest" 2>/dev/null || true
	info "Cached binary at ${_dest}"
}

cached_binary_path() {
	_ver="$1"
	_os="$2"
	_arch="$3"
	_p="${CACHE_DIR}/edgelet-${_ver}-${_os}-${_arch}"
	case "${_os}" in
		windows) _p="${_p}.exe" ;;
	esac
	if [ -f "$_p" ]; then
		echo "$_p"
		return 0
	fi
	echo ""
}
