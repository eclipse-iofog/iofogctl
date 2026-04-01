#!/bin/sh
set -x
set -e


CONTROLLER_LOG_DIR="iofog-controller-log"
CONTAINER_NAME="iofog-controller"
EXECUTABLE_FILE=/usr/local/bin/iofog-controller
CONTROLLER_DB=iofog-controller-db


do_uninstall_controller() {
    echo "# Removing ioFog controller..."

    case "$lsb_dist" in
        rhel|fedora|centos|ol|sles|opensuse*) CONTAINER_RUNTIME="podman" ;;
        *) CONTAINER_RUNTIME="docker" ;;
    esac

    case "${INIT_SYSTEM:-systemd}" in
        systemd)
            for f in /etc/systemd/system/iofog-controller.service /etc/containers/systemd/iofog-controller.container; do
                if [ -f "$f" ]; then
                    echo "Disabling and stopping systemd service..."
                    sudo systemctl stop iofog-controller.service 2>/dev/null || true
                    sudo systemctl disable iofog-controller.service 2>/dev/null || true
                    sudo rm -f "$f"
                    sudo systemctl daemon-reload
                    break
                fi
            done
            ;;
        sysvinit|openrc)
            if [ -f /etc/init.d/iofog-controller ]; then
                sudo service iofog-controller stop 2>/dev/null || sudo /etc/init.d/iofog-controller stop 2>/dev/null || true
                [ "$INIT_SYSTEM" = "openrc" ] && sudo rc-update del iofog-controller default 2>/dev/null || true
                sudo update-rc.d -f iofog-controller remove 2>/dev/null || sudo chkconfig --del iofog-controller 2>/dev/null || true
                sudo rm -f /etc/init.d/iofog-controller
            fi
            ;;
        s6)
            sudo s6-svc -d /etc/s6/sv/iofog-controller 2>/dev/null || true
            sudo rm -rf /etc/s6/sv/iofog-controller
            [ -L /etc/s6/adminsv/default/iofog-controller ] && sudo rm -f /etc/s6/adminsv/default/iofog-controller
            ;;
        runit)
            sudo sv stop iofog-controller 2>/dev/null || true
            [ -L /var/service/iofog-controller ] && sudo rm -f /var/service/iofog-controller
            [ -L /etc/runit/runsvdir/default/iofog-controller ] && sudo rm -f /etc/runit/runsvdir/default/iofog-controller
            sudo rm -rf /etc/runit/sv/iofog-controller
            ;;
        upstart)
            sudo initctl stop iofog-controller 2>/dev/null || true
            sudo rm -f /etc/init/iofog-controller.conf
            ;;
        *)
            sudo systemctl stop iofog-controller 2>/dev/null || true
            sudo systemctl disable iofog-controller 2>/dev/null || true
            sudo rm -f /etc/systemd/system/iofog-controller.service /etc/containers/systemd/iofog-controller.container
            sudo systemctl daemon-reload 2>/dev/null || true
            [ -f /etc/init.d/iofog-controller ] && sudo /etc/init.d/iofog-controller stop 2>/dev/null || true
            sudo rm -f /etc/init.d/iofog-controller
            ;;
    esac

    if sudo ${CONTAINER_RUNTIME} ps -a --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
        echo "Stopping and removing the ioFog controller container..."
        sudo ${CONTAINER_RUNTIME} stop ${CONTAINER_NAME} 2>/dev/null || true
        sudo ${CONTAINER_RUNTIME} rm ${CONTAINER_NAME} 2>/dev/null || true
    fi

    # Remove config files
    echo "Checking if the ${CONTAINER_RUNTIME} volume exists..."

    if sudo ${CONTAINER_RUNTIME} volume inspect "${CONTROLLER_DB}" >/dev/null 2>&1; then
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_DB}' found. Removing..."
        sudo ${CONTAINER_RUNTIME} volume rm "${CONTROLLER_DB}"
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_DB}' has been removed."
    else
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_DB}' does not exist. Skipping removal."
    fi

    # Remove log files
    echo "Removing log files..."
    if sudo ${CONTAINER_RUNTIME} volume inspect "${CONTROLLER_LOG_DIR}" >/dev/null 2>&1; then
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_LOG_DIR}' found. Removing..."
        sudo ${CONTAINER_RUNTIME} volume rm "${CONTROLLER_LOG_DIR}"
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_LOG_DIR}' has been removed."
    else
        echo "${CONTAINER_RUNTIME} volume '${CONTROLLER_LOG_DIR}' does not exist. Skipping removal."
    fi


    # Remove the executable script
    if [ -f ${EXECUTABLE_FILE} ]; then
        echo "Removing the iofog-controller executable script..."
        sudo rm -f ${EXECUTABLE_FILE}
    fi

    echo "ioFog controller uninstalled successfully!"
}

. /etc/iofog/controller/init.sh
init

do_uninstall_controller