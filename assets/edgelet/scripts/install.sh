#!/bin/sh
# install.sh — Edgelet installer (potctl chunked fork; upstream parity for upgrade/rollback)
#
# Usage:
#   sudo ./install.sh --version=v1.0.0-rc.4
#   sudo ./install.sh --airgap --bin-path=/path/to/edgelet-linux-amd64 --version=v1.0.0-rc.4
#   sudo ./install.sh --upgrade --version=v1.0.0-rc.4
#   sudo ./install.sh --rollback
#
# potctl deploy uses --skip-config and --skip-start (config/start handled by iofogctl/potctl).

set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/paths.sh"
. "$SCRIPT_DIR/lib/receipt.sh"
. "$SCRIPT_DIR/lib/binary.sh"

EDGELET_VERSION="${EDGELET_VERSION:-latest}"
CONTAINER_ENGINE=""
ACTION="install"
AIRGAP=false
BIN_PATH=""
FORCE_CONFIG=false
WITH_SAMPLE_CA=false
ARCH_OVERRIDE=""
CHECKSUM_FILE=""
EXPECTED_SHA256=""
SKIP_CONFIG=false
SKIP_START=false

for arg in "$@"; do
	case "${arg}" in
		--version=*) EDGELET_VERSION="${arg#*=}" ;;
		--arch=*) ARCH_OVERRIDE="${arg#*=}" ;;
		--container-engine=*) CONTAINER_ENGINE="${arg#*=}" ;;
		--bin-path=*) BIN_PATH="${arg#*=}" ;;
		--checksum-path=*) CHECKSUM_FILE="${arg#*=}" ;;
		--expected-sha256=*) EXPECTED_SHA256="${arg#*=}" ;;
		--airgap) AIRGAP=true ;;
		--upgrade) ACTION="upgrade" ;;
		--rollback) ACTION="rollback" ;;
		--force-config) FORCE_CONFIG=true ;;
		--with-sample-ca) WITH_SAMPLE_CA=true ;;
		--skip-config) SKIP_CONFIG=true ;;
		--skip-start) SKIP_START=true ;;
		--help|-h)
			echo "Usage: $0 [--version=VERSION] [--arch=ARCH] [--container-engine=ENGINE]"
			echo "       [--airgap] [--bin-path=PATH] [--upgrade] [--rollback]"
			echo "       [--skip-config] [--skip-start]  # potctl internal"
			exit 0
			;;
		*) die "Unknown option: ${arg}" ;;
	esac
done

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init

OS="$EDGELET_OS"
ARCH="${ARCH_OVERRIDE:-$EDGELET_ARCH}"
if [ -z "$CONTAINER_ENGINE" ]; then
	CONTAINER_ENGINE="${CONTAINER_ENGINE:-$(default_container_engine_for_os "$OS")}"
fi
case "$CONTAINER_ENGINE" in
	edgelet)
		[ "$OS" = "linux" ] || die "containerEngine=edgelet is linux-only"
		;;
	docker|podman) ;;
	*) die "Invalid --container-engine (use edgelet, docker, or podman)" ;;
esac

if [ "$OS" != "windows" ]; then
	[ "$(id -u)" -eq 0 ] || die "Must be run as root. Try: sudo $0 $*"
fi

init_platform_paths "$OS"
BINARY_PATH="$(binary_path_for_os "$OS")"
INIT="${INIT_SYSTEM:-unknown}"

info "OS: ${OS}  Arch: ${ARCH}  Init: ${INIT}  Engine: ${CONTAINER_ENGINE}  Action: ${ACTION}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "${TMPDIR}"' EXIT

if [ "$AIRGAP" = false ] && [ "$EDGELET_VERSION" = "latest" ] && [ "$ACTION" != "rollback" ] && [ -z "$BIN_PATH" ]; then
	EDGELET_VERSION=$(fetch_latest_version)
fi
info "Version: ${EDGELET_VERSION}"

_run_post_install() {
	if [ "$SKIP_START" = true ]; then
		info "Skipping daemon start (--skip-start); use start_edgelet.sh"
		return 0
	fi
	if [ "$OS" = "linux" ]; then
		EDGELET_INSTALL_MODE=native CONTAINER_ENGINE="$CONTAINER_ENGINE" \
			"$SCRIPT_DIR/install_init_units.sh" || true
		EDGELET_INSTALL_MODE=native CONTAINER_ENGINE="$CONTAINER_ENGINE" \
			"$SCRIPT_DIR/start_edgelet.sh"
	elif [ "$OS" = "darwin" ]; then
		"$SCRIPT_DIR/start_edgelet.sh"
	fi
}

if [ "$ACTION" = "rollback" ]; then
	[ -f "$PREVIOUS_FILE" ] || die "No ${PREVIOUS_FILE} found."
	_pv=$(kv_get "$PREVIOUS_FILE" "previous_version")
	_pos=$(kv_get "$PREVIOUS_FILE" "previous_os")
	_parch=$(kv_get "$PREVIOUS_FILE" "previous_arch")
	_peng=$(kv_get "$PREVIOUS_FILE" "previous_container_engine")
	_purl=$(kv_get "$PREVIOUS_FILE" "previous_download_url")
	_cfgbak=$(kv_get "$PREVIOUS_FILE" "config_backup_path")
	[ -n "$_pos" ] || _pos="$OS"
	[ -n "$_parch" ] || _parch="$ARCH"
	CONTAINER_ENGINE="${_peng:-edgelet}"
	EDGELET_VERSION="$_pv"
	_staged="${TMPDIR}/edgelet-bin"
	_cached=$(cached_binary_path "$_pv" "$_pos" "$_parch")
	if [ -n "$_cached" ]; then
		cp "$_cached" "$_staged"
	elif [ -n "$BIN_PATH" ]; then
		verify_binary_checksum "$BIN_PATH"
		cp "$BIN_PATH" "$_staged"
	elif [ "$AIRGAP" = true ]; then
		die "rollback with --airgap requires --bin-path or a cached binary"
	else
		curl -fsSL -o "$_staged" "$_purl" || die "Failed to download rollback binary"
	fi
	install_binary_file "$_staged" "$BINARY_PATH"
	install_dirs_for_os "$OS"
	if [ "$FORCE_CONFIG" != true ] && [ -f "$_cfgbak" ]; then
		maybe_sudo install -m 640 "$_cfgbak" "$CONFIG_FILE"
	fi
	_sha=$(sha256_file "$BINARY_PATH")
	write_install_receipt "$EDGELET_VERSION" "$_pos" "$_parch" "$CONTAINER_ENGINE" "$_purl" "$_sha" "rollback"
	_run_post_install
	"$SCRIPT_DIR/bundled.sh" || true
	info "Rollback to ${EDGELET_VERSION} complete."
	exit 0
fi

if [ "$ACTION" = "upgrade" ]; then
	[ -f "$BINARY_PATH" ] || die "Edgelet not installed; run install first"
	[ -f "$RECEIPT_FILE" ] || die "Missing ${RECEIPT_FILE}"
	_cur_ver=$(kv_get "$RECEIPT_FILE" "installed_version")
	_cur_os=$(kv_get "$RECEIPT_FILE" "os")
	_cur_arch=$(kv_get "$RECEIPT_FILE" "arch")
	_cur_eng=$(kv_get "$RECEIPT_FILE" "container_engine")
	_cur_src=$(kv_get "$RECEIPT_FILE" "source_url")
	_cur_sha=$(kv_get "$RECEIPT_FILE" "binary_sha256")
	[ -n "$_cur_os" ] || _cur_os="$OS"
	[ -n "$_cur_arch" ] || _cur_arch="$ARCH"
	[ -n "$_cur_eng" ] || _cur_eng="$CONTAINER_ENGINE"
	if [ "$EDGELET_VERSION" = "latest" ] && [ -z "$BIN_PATH" ]; then
		EDGELET_VERSION=$(fetch_latest_version)
	fi
	_cfg_backup="${BACKUP_DIR}/config.yaml.$(date +%Y%m%d%H%M%S 2>/dev/null || date +%s)"
	cp "$CONFIG_FILE" "$_cfg_backup" 2>/dev/null || true
	cache_binary "$_cur_ver" "$_cur_os" "$_cur_arch" "$BINARY_PATH"
	write_previous_release "$_cur_ver" "$_cur_os" "$_cur_arch" "$_cur_eng" "$_cur_src" "$_cur_sha" "$_cfg_backup"
	stop_edgelet_daemon_desktop "$OS"
	_staged="${TMPDIR}/edgelet-bin"
	download_or_stage_binary "$_staged"
	verify_binary_checksum "$_staged"
	install_binary_file "$_staged" "$BINARY_PATH"
	install_dirs_for_os "$OS"
	_sha=$(sha256_file "$BINARY_PATH")
	_method="upgrade"
	[ "$AIRGAP" = true ] && _method="upgrade-airgap"
	write_install_receipt "$EDGELET_VERSION" "$OS" "$ARCH" "$CONTAINER_ENGINE" "$(compute_source_url)" "$_sha" "$_method"
	"$SCRIPT_DIR/bundled.sh" || true
	_run_post_install
	info "Upgrade to ${EDGELET_VERSION} complete."
	exit 0
fi

# fresh install
if command -v edgelet >/dev/null 2>&1; then
	installed=$(edgelet --version 2>/dev/null | head -n1 | tr -d '[:space:]')
	if [ -n "$EDGELET_VERSION" ] && [ "$installed" = "$EDGELET_VERSION" ]; then
		info "Edgelet $EDGELET_VERSION already installed."
		"$SCRIPT_DIR/bundled.sh" || true
		exit 0
	fi
fi

_staged="${TMPDIR}/edgelet-bin"
download_or_stage_binary "$_staged"
verify_binary_checksum "$_staged"
install_dirs_for_os "$OS"
install_binary_file "$_staged" "$BINARY_PATH"
_sha=$(sha256_file "$BINARY_PATH")
_method="install"
[ "$AIRGAP" = true ] && _method="install-airgap"
write_install_receipt "$EDGELET_VERSION" "$OS" "$ARCH" "$CONTAINER_ENGINE" "$(compute_source_url)" "$_sha" "$_method"
"$SCRIPT_DIR/bundled.sh" || true

if [ "$SKIP_START" = true ]; then
	info "Skipping daemon start (--skip-start); potctl will start after config materialization."
	exit 0
fi

_run_post_install
info "edgelet ${EDGELET_VERSION} installed (os=${OS} engine=${CONTAINER_ENGINE})."
