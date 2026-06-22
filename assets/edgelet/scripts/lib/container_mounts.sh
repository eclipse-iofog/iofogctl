#!/bin/sh
# Container volume/bind helpers for edgelet container deployment.
# Requires lib/common.sh (is_desktop_container_host, maybe_sudo).

edgelet_container_volume_mounts() {
	echo "-v /etc/edgelet:/etc/edgelet:rw \
-v /var/lib/edgelet:/var/lib/edgelet:rw \
-v /var/lib/edgelet-containerd:/var/lib/edgelet-containerd:rw \
-v /var/log/edgelet:/var/log/edgelet:rw \
-v /tmp/edgelet:/tmp/edgelet:rw"
}

prepare_edgelet_container_host_dirs() {
	if is_desktop_container_host; then
		echo "# skipping host FHS prep on ${EDGELET_OS} container deploy (VM-local bind paths)"
		return 0
	fi
	maybe_sudo mkdir -p /etc/edgelet /var/lib/edgelet /var/lib/edgelet-containerd \
		/var/log/edgelet /var/run/edgelet /run/edgelet /tmp/edgelet
}

container_runtime_bin() {
	case "${CONTAINER_ENGINE:-docker}" in
		docker) echo "docker" ;;
		podman) echo "podman" ;;
		*) die "container deployment requires docker or podman engine (got ${CONTAINER_ENGINE:-})" ;;
	esac
}

restart_edgelet_container() {
	_run=$(container_runtime_bin)
	_name="${EDGELET_CONTAINER_NAME:-edgelet}"
	info "Restarting edgelet container ${_name}"
	maybe_sudo "$_run" restart "$_name"
}
