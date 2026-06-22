#!/bin/sh
# Configure an existing docker/podman engine. Does not install packages.
set -e
set -x

if [ "${EDGELET_OS:-}" = "darwin" ] || [ "${EDGELET_OS:-}" = "windows" ]; then
	echo "# Skipping configure_container_engine on ${EDGELET_OS} (desktop container runtime is user-managed)"
	exit 0
fi

CONTAINER_ENGINE="${CONTAINER_ENGINE:-docker}"

check_docker_version() {
	docker_version_num=0
	if command_exists docker; then
		raw=$(docker -v 2>/dev/null | sed 's/.*version \([^,]*\),.*/\1/' | tr -d '.')
		[ -n "$raw" ] && docker_version_num="$raw"
	fi
	[ "$docker_version_num" -ge 2610 ] 2>/dev/null
}

check_podman_version() {
	podman_version_num=0
	if command_exists podman; then
		raw=$(podman --version 2>/dev/null | sed -n 's/.*version \([0-9][0-9]*\).*/\1/p')
		[ -n "$raw" ] && podman_version_num="$raw"
	fi
	[ "$podman_version_num" -ge 3 ] 2>/dev/null
}

command_exists() {
	command -v "$@" > /dev/null 2>&1
}

start_docker() {
	if command_exists docker && docker ps >/dev/null 2>&1; then
		return 0
	fi
	case "${INIT_SYSTEM:-unknown}" in
		systemd) sudo systemctl start docker ;;
		*) sudo service docker start 2>/dev/null || sudo systemctl start docker ;;
	esac
}

start_podman() {
	case "${INIT_SYSTEM:-unknown}" in
		systemd)
			sudo systemctl start podman 2>/dev/null || true
			sudo systemctl start podman.socket 2>/dev/null || true
			;;
		*) sudo service podman start 2>/dev/null || true ;;
	esac
}

configure_docker() {
	if ! command_exists docker; then
		echo "Error: docker is not installed. Install Docker 26.10+ manually or provide scripts.deps."
		exit 1
	fi
	if ! check_docker_version; then
		echo "Error: Docker 26.10+ is required."
		exit 1
	fi
	start_docker
	if [ ! -f /etc/docker/daemon.json ]; then
		sudo mkdir -p /etc/docker
		sudo tee /etc/docker/daemon.json >/dev/null <<'EOF'
{
  "storage-driver": "overlayfs",
  "features": {
    "containerd-snapshotter": true,
    "cdi": true
  },
  "cdi-spec-dirs": ["/etc/cdi/", "/var/run/cdi"]
}
EOF
	else
		echo "# /etc/docker/daemon.json already exists"
	fi
}

configure_podman() {
	if ! command_exists podman; then
		echo "Error: podman is not installed. Install Podman 3.0+ manually or provide scripts.deps."
		exit 1
	fi
	if ! check_podman_version; then
		echo "Error: Podman 3.0+ is required."
		exit 1
	fi
	sudo mkdir -p /etc/cdi /var/run/cdi /etc/containers
	if [ ! -f /etc/containers/containers.conf ]; then
		sudo tee /etc/containers/containers.conf >/dev/null <<'EOF'
[engine]
runtime = "crun"
cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
EOF
	fi
	start_podman
}

case "$CONTAINER_ENGINE" in
	docker) configure_docker ;;
	podman) configure_podman ;;
	*)
		echo "Error: unsupported container engine for configure step: $CONTAINER_ENGINE"
		exit 1
		;;
esac

echo "# Container engine $CONTAINER_ENGINE configured"
