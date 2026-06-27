#!/bin/sh
# Prepare containerized edgelet deployment (deploymentType=container).
set -e
set -x

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib/common.sh"
. "$SCRIPT_DIR/lib/paths.sh"
. "$SCRIPT_DIR/lib/receipt.sh"
. "$SCRIPT_DIR/lib/container_cli.sh"
. "$SCRIPT_DIR/lib/container_mounts.sh"

IMAGE=""
ENGINE=""
TZ="${EDGELET_TZ:-UTC}"

for arg in "$@"; do
	case "${arg}" in
		--image=*) IMAGE="${arg#*=}" ;;
		--engine=*) ENGINE="${arg#*=}" ;;
		--tz=*) TZ="${arg#*=}" ;;
		--help|-h)
			echo "Usage: $0 --image=IMAGE [--engine=docker|podman] [--tz=TZ]"
			exit 0
			;;
		*) die "Unknown option: ${arg}" ;;
	esac
done

[ -n "$IMAGE" ] || die "--image is required"

EDGELET_DETECT_SOURCED=1
. "$SCRIPT_DIR/detect_init.sh"
init
init_platform_paths "$EDGELET_OS"

EDGELET_CONTAINER_NAME="${EDGELET_CONTAINER_NAME:-edgelet}"

export EDGELET_CONTAINER_IMAGE="$IMAGE"
export EDGELET_TZ="$TZ"
export CONTAINER_ENGINE="$ENGINE"
export EDGELET_CONTAINER_NAME
export EDGELET_INSTALL_MODE=container

if [ -z "$ENGINE" ]; then
	ENGINE="docker"
	export CONTAINER_ENGINE="$ENGINE"
fi

if ! is_desktop_container_host; then
	install_dirs_for_os "$EDGELET_OS"
fi

_runtime=""
case "$ENGINE" in
	docker) _runtime="docker" ;;
	podman) _runtime="podman" ;;
	*) die "container deployment requires docker or podman engine (got $ENGINE)" ;;
esac

if ! command -v "$_runtime" >/dev/null 2>&1; then
	die "$_runtime is not installed"
fi

info "Pulling container image $IMAGE"
maybe_sudo "$_runtime" pull "$IMAGE"

install_container_cli_wrapper "$ENGINE" "$EDGELET_CONTAINER_NAME" "$EDGELET_OS"

if desktop_container_local; then
	info "Skipping bundled publish on desktop local container deploy"
else
	"$SCRIPT_DIR/bundled.sh" || true
fi
info "Container edgelet prepared (image=$IMAGE engine=$ENGINE tz=$TZ)"
