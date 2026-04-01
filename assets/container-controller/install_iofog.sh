#!/bin/sh
set -x
set -e

# INSTALL_DIR="/opt/iofog"
TMP_DIR="/tmp/iofog"
ETC_DIR="/etc/iofog/controller"
CONTROLLER_LOG_FOLDER=/var/log/iofog-controller
CONTROLLER_CONTAINER_NAME="iofog-controller"

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

do_stop_iofog_controller() {
    if ! command_exists iofog-controller; then
        return 0
    fi
    case "${INIT_SYSTEM:-systemd}" in
        systemd)
            sudo systemctl stop iofog-controller 2>/dev/null || true
            ;;
        sysvinit|openrc)
            sudo service iofog-controller stop 2>/dev/null || sudo /etc/init.d/iofog-controller stop 2>/dev/null || true
            ;;
        s6)
            sudo s6-svc -d /etc/s6/sv/iofog-controller 2>/dev/null || true
            ;;
        runit)
            sudo sv stop iofog-controller 2>/dev/null || true
            ;;
        upstart)
            sudo initctl stop iofog-controller 2>/dev/null || true
            ;;
        *)
            sudo systemctl stop iofog-controller 2>/dev/null || sudo service iofog-controller stop 2>/dev/null || true
            ;;
    esac
    (docker stop ${CONTROLLER_CONTAINER_NAME} 2>/dev/null || podman stop ${CONTROLLER_CONTAINER_NAME} 2>/dev/null) || true
}

do_install_iofog_controller() {
    echo "# Installing ioFog controller..."

    for FOLDER in ${ETC_DIR} ${CONTROLLER_LOG_FOLDER}; do
        if [ ! -d "$FOLDER" ]; then
            echo "Creating folder: $FOLDER"
            sudo mkdir -p "$FOLDER"
            sudo chmod 775 "$FOLDER"
        fi
    done

    USE_PODMAN="false"
    case "$lsb_dist" in
        rhel|centos|fedora|ol|sles|opensuse*) USE_PODMAN="true" ;;
    esac

    CONTROLLER_RUN_ARGS="-e IOFOG_CONTROLLER_IMAGE=${controller_image} --env-file ${ETC_DIR}/iofog-controller.env -v iofog-controller-db:/home/runner/.npm-global/lib/node_modules/@eclipse-iofog/iofogcontroller/src/data/sqlite_files/:rw -v iofog-controller-log:/var/log/iofog-controller:rw -p 51121:51121 -p 80:8008 --stop-timeout 60 ${controller_image}"

    if [ "${INIT_SYSTEM:-systemd}" = "systemd" ]; then
        if [ "$USE_PODMAN" = "true" ]; then
            echo "Creating Quadlet container file for ioFog controller..."
            sudo mkdir -p /etc/containers/systemd
            cat <<EOF | sudo tee /etc/containers/systemd/iofog-controller.container > /dev/null
[Unit]
Description=ioFog Controller Service
After=podman.service
Requires=podman.service

[Container]
ContainerName=${CONTROLLER_CONTAINER_NAME}
Image=${controller_image}
PodmanArgs=--stop-timeout=60
Environment=IOFOG_CONTROLLER_IMAGE=${controller_image}
EnvironmentFile=${ETC_DIR}/iofog-controller.env
Volume=iofog-controller-db:/home/runner/.npm-global/lib/node_modules/@eclipse-iofog/iofogcontroller/src/data/sqlite_files/:rw
Volume=iofog-controller-log:/var/log/iofog-controller:rw
PublishPort=51121:51121
PublishPort=80:8008
LogDriver=journald

[Service]
Restart=always

[Install]
WantedBy=default.target
EOF
            sudo systemctl daemon-reload
            sudo systemctl restart podman 2>/dev/null || true
            sudo systemctl enable iofog-controller.service
            sudo systemctl start iofog-controller.service
        else
            echo "Creating systemd service for ioFog controller..."
            cat <<EOF | sudo tee /etc/systemd/system/iofog-controller.service > /dev/null
[Unit]
Description=ioFog Controller Service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker rm -f ${CONTROLLER_CONTAINER_NAME}
ExecStart=/usr/bin/docker run --rm --name ${CONTROLLER_CONTAINER_NAME} \\
${CONTROLLER_RUN_ARGS}
ExecStop=/usr/bin/docker stop ${CONTROLLER_CONTAINER_NAME}

[Install]
WantedBy=default.target
EOF
            sudo systemctl daemon-reload
            sudo systemctl enable iofog-controller.service
            sudo systemctl start iofog-controller.service
        fi
    else
        if [ "$USE_PODMAN" = "true" ]; then
            RUN_CMD="podman run --rm -d --name ${CONTROLLER_CONTAINER_NAME} ${CONTROLLER_RUN_ARGS}"
            RUN_CMD_FG="podman run --rm --name ${CONTROLLER_CONTAINER_NAME} ${CONTROLLER_RUN_ARGS}"
        else
            RUN_CMD="docker run --rm -d --name ${CONTROLLER_CONTAINER_NAME} ${CONTROLLER_RUN_ARGS}"
            RUN_CMD_FG="docker run --rm --name ${CONTROLLER_CONTAINER_NAME} ${CONTROLLER_RUN_ARGS}"
        fi
        if [ "$USE_PODMAN" = "true" ]; then
            STOP_CMD="podman stop ${CONTROLLER_CONTAINER_NAME}"
        else
            STOP_CMD="docker stop ${CONTROLLER_CONTAINER_NAME}"
        fi

        case "$INIT_SYSTEM" in
            sysvinit|openrc)
                sudo tee /etc/init.d/iofog-controller > /dev/null <<INITSCRIPT
#!/bin/sh
### BEGIN INIT INFO
# Provides:          iofog-controller
# Required-Start:    \$network \$local_fs
# Required-Stop:     \$network \$local_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
### END INIT INFO
case "\$1" in
    start)
        n=0; while [ \$n -lt 30 ]; do [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] && break; n=\$((n+1)); sleep 1; done
        [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] || { echo "Container engine socket not available"; exit 1; }
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTROLLER_CONTAINER_NAME}\$"; then exit 0; fi
        if podman ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTROLLER_CONTAINER_NAME}\$"; then exit 0; fi
        $RUN_CMD
        ;;
    stop) $STOP_CMD 2>/dev/null || true ;;
    restart) \$0 stop; \$0 start ;;
    status)
        if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTROLLER_CONTAINER_NAME}\$"; then echo "running"; exit 0; fi
        if podman ps --format '{{.Names}}' 2>/dev/null | grep -q "^${CONTROLLER_CONTAINER_NAME}\$"; then echo "running"; exit 0; fi
        echo "stopped"; exit 1
        ;;
    *) echo "Usage: \$0 {start|stop|restart|status}"; exit 1 ;;
esac
exit 0
INITSCRIPT
                sudo chmod +x /etc/init.d/iofog-controller
                if [ "$INIT_SYSTEM" = "openrc" ]; then
                    sudo rc-update add iofog-controller default 2>/dev/null || true
                    sudo rc-service iofog-controller start
                else
                    sudo update-rc.d iofog-controller defaults 2>/dev/null || sudo chkconfig iofog-controller on 2>/dev/null || true
                    sudo service iofog-controller start 2>/dev/null || sudo /etc/init.d/iofog-controller start
                fi
                ;;
            s6)
                sudo mkdir -p /etc/s6/sv/iofog-controller
                printf '#!/bin/sh\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/s6/sv/iofog-controller/run > /dev/null
                sudo chmod +x /etc/s6/sv/iofog-controller/run
                [ -d /etc/s6/adminsv/default ] && sudo ln -sf /etc/s6/sv/iofog-controller /etc/s6/adminsv/default/iofog-controller 2>/dev/null || true
                sudo s6-svc -u /etc/s6/sv/iofog-controller 2>/dev/null || true
                ;;
            runit)
                sudo mkdir -p /etc/runit/sv/iofog-controller
                printf '#!/bin/sh\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/runit/sv/iofog-controller/run > /dev/null
                sudo chmod +x /etc/runit/sv/iofog-controller/run
                [ -d /var/service ] && sudo ln -sf /etc/runit/sv/iofog-controller /var/service/iofog-controller 2>/dev/null || true
                [ -d /etc/runit/runsvdir/default ] && sudo ln -sf /etc/runit/sv/iofog-controller /etc/runit/runsvdir/default/iofog-controller 2>/dev/null || true
                sudo sv start iofog-controller 2>/dev/null || true
                ;;
            upstart)
                printf 'description "IoFog Controller container"\nstart on runlevel [2345]\nstop on runlevel [!2345]\nrespawn\nrespawn limit 10 5\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/init/iofog-controller.conf > /dev/null
                sudo initctl reload-configuration 2>/dev/null || true
                sudo initctl start iofog-controller 2>/dev/null || true
                ;;
            *)
                sudo tee /etc/init.d/iofog-controller > /dev/null <<INITSCRIPT
#!/bin/sh
case "\$1" in start) n=0; while [ \$n -lt 30 ]; do [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] && break; n=\$((n+1)); sleep 1; done; [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] || { echo "Container engine socket not available"; exit 1; }; $RUN_CMD ;; stop) $STOP_CMD ;; restart) $STOP_CMD; $RUN_CMD ;; *) echo "Usage: \$0 {start|stop|restart}"; exit 1 ;; esac
INITSCRIPT
                sudo chmod +x /etc/init.d/iofog-controller
                sudo /etc/init.d/iofog-controller start
                ;;
        esac
    fi

    EXECUTABLE_FILE=/usr/local/bin/iofog-controller
    if [ "$USE_PODMAN" = "true" ]; then
        cat <<'EOF' | sudo tee ${EXECUTABLE_FILE} > /dev/null
#!/bin/sh
CONTAINER_NAME="iofog-controller"
if ! podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: The iofog-controller container is not running."
    exit 1
fi
exec podman exec ${CONTAINER_NAME} iofog-controller "$@"
EOF
    else
        cat <<'EOF' | sudo tee ${EXECUTABLE_FILE} > /dev/null
#!/bin/sh
CONTAINER_NAME="iofog-controller"
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: The iofog-controller container is not running."
    exit 1
fi
exec docker exec ${CONTAINER_NAME} iofog-controller "$@"
EOF
    fi
    sudo chmod +x ${EXECUTABLE_FILE}

    echo "ioFog controller installation completed!"
}

# main
controller_image="$1"

. /etc/iofog/controller/init.sh
init
do_stop_iofog_controller
do_install_iofog_controller
