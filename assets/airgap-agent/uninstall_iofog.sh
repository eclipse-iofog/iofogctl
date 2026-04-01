#!/bin/sh
set -x
set -e

AGENT_CONFIG_FOLDER=iofog-agent-config
AGENT_LOG_FOLDER=/var/log/iofog-agent
AGENT_BACKUP_FOLDER=/var/backups/iofog-agent
AGENT_MESSAGE_FOLDER=/var/lib/iofog-agent
EXECUTABLE_FILE=/usr/local/bin/iofog-agent
CONTAINER_NAME="iofog-agent"

do_uninstall_iofog() {
    echo "# Removing ioFog agent..."

    case "$lsb_dist" in
        rhel|fedora|centos|ol|sles|opensuse*) CONTAINER_RUNTIME="podman" ;;
        *) CONTAINER_RUNTIME="docker" ;;
    esac

    case "${INIT_SYSTEM:-systemd}" in
        systemd)
            for f in /etc/systemd/system/iofog-agent.service /etc/containers/systemd/iofog-agent.container; do
                if [ -f "$f" ]; then
                    echo "Disabling and stopping systemd service..."
                    sudo systemctl stop iofog-agent.service 2>/dev/null || true
                    sudo systemctl disable iofog-agent.service 2>/dev/null || true
                    sudo rm -f "$f"
                    sudo systemctl daemon-reload
                    break
                fi
            done
            ;;
        sysvinit|openrc)
            if [ -f /etc/init.d/iofog-agent ]; then
                sudo service iofog-agent stop 2>/dev/null || sudo /etc/init.d/iofog-agent stop 2>/dev/null || true
                [ "$INIT_SYSTEM" = "openrc" ] && sudo rc-update del iofog-agent default 2>/dev/null || true
                sudo update-rc.d -f iofog-agent remove 2>/dev/null || sudo chkconfig --del iofog-agent 2>/dev/null || true
                sudo rm -f /etc/init.d/iofog-agent
            fi
            ;;
        s6)
            sudo s6-svc -d /etc/s6/sv/iofog-agent 2>/dev/null || true
            sudo rm -rf /etc/s6/sv/iofog-agent
            [ -L /etc/s6/adminsv/default/iofog-agent ] && sudo rm -f /etc/s6/adminsv/default/iofog-agent
            ;;
        runit)
            sudo sv stop iofog-agent 2>/dev/null || true
            [ -L /var/service/iofog-agent ] && sudo rm -f /var/service/iofog-agent
            [ -L /etc/runit/runsvdir/default/iofog-agent ] && sudo rm -f /etc/runit/runsvdir/default/iofog-agent
            sudo rm -rf /etc/runit/sv/iofog-agent
            ;;
        upstart)
            sudo initctl stop iofog-agent 2>/dev/null || true
            sudo rm -f /etc/init/iofog-agent.conf
            ;;
        *)
            sudo systemctl stop iofog-agent 2>/dev/null || true
            sudo systemctl disable iofog-agent 2>/dev/null || true
            sudo rm -f /etc/systemd/system/iofog-agent.service /etc/containers/systemd/iofog-agent.container
            sudo systemctl daemon-reload 2>/dev/null || true
            [ -f /etc/init.d/iofog-agent ] && sudo /etc/init.d/iofog-agent stop 2>/dev/null || true
            sudo rm -f /etc/init.d/iofog-agent
            ;;
    esac

    if sudo ${CONTAINER_RUNTIME} ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
        echo "Stopping and removing the ioFog agent container..."
        sudo ${CONTAINER_RUNTIME} stop ${CONTAINER_NAME} 2>/dev/null || true
        sudo ${CONTAINER_RUNTIME} rm ${CONTAINER_NAME} 2>/dev/null || true
    fi

    # Remove config files
    echo "Checking if the ${CONTAINER_RUNTIME} volume exists..."

    if sudo ${CONTAINER_RUNTIME} volume inspect "${AGENT_CONFIG_FOLDER}" >/dev/null 2>&1; then
        echo "${CONTAINER_RUNTIME} volume '${AGENT_CONFIG_FOLDER}' found. Removing..."
        sudo ${CONTAINER_RUNTIME} volume rm "${AGENT_CONFIG_FOLDER}"
        echo "${CONTAINER_RUNTIME} volume '${AGENT_CONFIG_FOLDER}' has been removed."
    else
        echo "${CONTAINER_RUNTIME} volume '${AGENT_CONFIG_FOLDER}' does not exist. Skipping removal."
    fi

    # Remove log files
    echo "Removing log files..."
    sudo rm -rf ${AGENT_LOG_FOLDER}

    # Remove backup files
    echo "Removing backup files..."
    sudo rm -rf ${AGENT_BACKUP_FOLDER}

    # Remove message files
    echo "Removing message files..."
    sudo rm -rf ${AGENT_MESSAGE_FOLDER}

    # Remove the executable script
    if [ -f ${EXECUTABLE_FILE} ]; then
        echo "Removing the iofog-agent executable script..."
        sudo rm -f ${EXECUTABLE_FILE}
    fi

    echo "ioFog agent uninstalled successfully!"
}

. /etc/iofog/agent/init.sh
init

do_uninstall_iofog
