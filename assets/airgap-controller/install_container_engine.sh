#!/bin/sh
# Script to configure Docker/Podman for airgap deployment
# Check-only: verifies container engine is installed (Docker 25+ or Podman 4+), then configures/starts
# Sources init.sh for distribution detection

set -x
set -e

CONTAINER_ENGINE_MSG="This operating system does not support automatic container engine installation. Please install Docker 25+ or Podman 4+ on the target host and re-run, or use an airgap deployment with a pre-installed engine."

check_docker_version() {
    docker_version_num=0
    if command -v docker >/dev/null 2>&1; then
        raw=$(docker -v 2>/dev/null | sed 's/.*version \([^,]*\),.*/\1/' | tr -d '.')
        [ -n "$raw" ] && docker_version_num="$raw"
    fi
    [ "$docker_version_num" -ge 2500 ] 2>/dev/null || return 1
}

check_podman_version() {
    podman_version_num=0
    if command -v podman >/dev/null 2>&1; then
        raw=$(podman --version 2>/dev/null | sed -n 's/.*version \([0-9][0-9]*\).*/\1/p')
        [ -n "$raw" ] && podman_version_num="$raw"
    fi
    [ "$podman_version_num" -ge 4 ] 2>/dev/null || return 1
}

start_docker() {
    set +e
    if $sh_c "docker ps" >/dev/null 2>&1; then
        set -e
        return 0
    fi
    err_code=1
    case "${INIT_SYSTEM:-unknown}" in
        systemd) $sh_c "systemctl start docker" >/dev/null 2>&1; err_code=$? ;;
        sysvinit) $sh_c "service docker start" >/dev/null 2>&1 || $sh_c "/etc/init.d/docker start" >/dev/null 2>&1; err_code=$? ;;
        openrc) $sh_c "rc-service docker start" >/dev/null 2>&1; err_code=$? ;;
        *)
            $sh_c "/etc/init.d/docker start" >/dev/null 2>&1
            err_code=$?
            [ $err_code -ne 0 ] && $sh_c "systemctl start docker" >/dev/null 2>&1 && err_code=0
            [ $err_code -ne 0 ] && $sh_c "service docker start" >/dev/null 2>&1 && err_code=0
            [ $err_code -ne 0 ] && $sh_c "snap start docker" >/dev/null 2>&1 && err_code=0
            ;;
    esac
    set -e
    if [ $err_code -ne 0 ]; then
        echo "Could not start Docker daemon"
        exit 1
    fi
}

start_podman() {
    set +e
    case "${INIT_SYSTEM:-unknown}" in
        systemd)
            $sh_c "systemctl start podman" >/dev/null 2>&1
            $sh_c "systemctl start podman.socket" >/dev/null 2>&1
            ;;
        sysvinit) $sh_c "service podman start" >/dev/null 2>&1 || $sh_c "/etc/init.d/podman start" >/dev/null 2>&1 ;;
        openrc) $sh_c "rc-service podman start" >/dev/null 2>&1 ;;
        *)
            $sh_c "systemctl start podman" >/dev/null 2>&1 || true
            $sh_c "systemctl start podman.socket" >/dev/null 2>&1 || true
            $sh_c "service podman start" >/dev/null 2>&1 || true
            ;;
    esac
    set -e
}

do_modify_daemon() {
    # Skip for Podman installations
    if [ "$USE_PODMAN" = "true" ]; then
        echo "# Configuring Podman for CDI directory support..."

        # Create CDI directories
        $sh_c "mkdir -p /etc/cdi /var/run/cdi"

        # Ensure /etc/containers exists
        $sh_c "mkdir -p /etc/containers"

        # Create containers.conf if it doesn't exist
        if [ ! -f "/etc/containers/containers.conf" ]; then
            $sh_c 'cat > /etc/containers/containers.conf <<EOF
[engine]
runtime = "crun"
cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
EOF'
        else
            # Check if [engine] block exists
            if grep -q "^\[engine\]" /etc/containers/containers.conf; then
                # Ensure runtime is set under [engine]
                if grep -q "^runtime" /etc/containers/containers.conf; then
                    $sh_c "sed -i 's|^runtime *=.*|runtime = \"crun\"|' /etc/containers/containers.conf"
                else
                    $sh_c "sed -i '/^\[engine\]/a runtime = \"crun\"' /etc/containers/containers.conf"
                fi

                # Ensure cdi_spec_dirs is set under [engine]
                if grep -q "^cdi_spec_dirs" /etc/containers/containers.conf; then
                    $sh_c "sed -i 's|^cdi_spec_dirs *=.*|cdi_spec_dirs = [\"/etc/cdi\", \"/var/run/cdi\"]|' /etc/containers/containers.conf"
                else
                    $sh_c "sed -i '/^\[engine\]/a cdi_spec_dirs = [\"/etc/cdi\", \"/var/run/cdi\"]' /etc/containers/containers.conf"
                fi
            else
                # Append full engine block if missing
                $sh_c 'echo -e "\n[engine]\nruntime = \"crun\"\ncdi_spec_dirs = [\"/etc/cdi\", \"/var/run/cdi\"]" >> /etc/containers/containers.conf'
            fi
        fi

        case "${INIT_SYSTEM:-unknown}" in
            systemd)
                $sh_c "systemctl enable podman" 2>/dev/null || true
                $sh_c "systemctl enable podman.socket" 2>/dev/null || true
                ;;
            openrc) $sh_c "rc-update add podman default" 2>/dev/null || true ;;
            sysvinit) $sh_c "update-rc.d podman defaults" 2>/dev/null || $sh_c "chkconfig podman on" 2>/dev/null || true ;;
            *) ;;
        esac
        start_podman
        return
    fi
    
    # Original Docker daemon configuration
    if [ ! -f /etc/docker/daemon.json ]; then
        echo "Creating /etc/docker/daemon.json..."
        $sh_c "mkdir -p /etc/docker"
        $sh_c 'cat > /etc/docker/daemon.json << EOF
{
	"storage-driver": "overlayfs",
    "features": {
        "containerd-snapshotter": true,
        "cdi": true
    },
    "cdi-spec-dirs": ["/etc/cdi/", "/var/run/cdi"]
}
EOF'
    else
        echo "/etc/docker/daemon.json already exists"
    fi
    echo "Restarting Docker daemon..."
    case "${INIT_SYSTEM:-unknown}" in
        systemd)
            $sh_c "systemctl daemon-reload"
            $sh_c "systemctl restart docker"
            ;;
        *)
            $sh_c "systemctl daemon-reload" 2>/dev/null || true
            $sh_c "systemctl restart docker" 2>/dev/null || start_docker
            ;;
    esac
}

# Airgap: determine engine by availability (Docker 25+ or Podman 4+)
determine_container_engine() {
    if check_docker_version; then
        USE_PODMAN="false"
        echo "# Using Docker (25+)"
    elif check_podman_version; then
        USE_PODMAN="true"
        echo "# Using Podman (4+)"
    else
        echo "Error: Docker 25+ or Podman 4+ is required. $CONTAINER_ENGINE_MSG"
        exit 1
    fi
}

. /etc/iofog/controller/init.sh
init

determine_container_engine

if [ "$USE_PODMAN" = "false" ]; then
    start_docker
fi

do_modify_daemon

echo "# Container engine configuration completed successfully"



