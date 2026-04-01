#!/bin/sh
set -x
set -e

AGENT_LOG_FOLDER=/var/log/iofog-agent
AGENT_BACKUP_FOLDER=/var/backups/iofog-agent
AGENT_MESSAGE_FOLDER=/var/lib/iofog-agent
AGENT_SHARE_FOLDER=/usr/share/iofog-agent
SAVED_AGENT_CONFIG_FOLDER=/tmp/agent-config-save
AGENT_CONTAINER_NAME="iofog-agent"
ETC_DIR=/etc/iofog/agent

do_check_install() {
	if command_exists iofog-agent; then
		local VERSION=$(sudo iofog-agent version | head -n1 | sed "s/ioFog//g" | tr -d ' ' | tr -d "\n")
		if [ "$VERSION" = "$agent_version" ]; then
			echo "Agent $VERSION already installed."
			exit 0
		fi
	fi
}

do_stop_iofog() {
	if ! command_exists iofog-agent; then
		return 0
	fi
	case "${INIT_SYSTEM:-systemd}" in
		systemd)
			sudo systemctl stop iofog-agent 2>/dev/null || true
			;;
		sysvinit|openrc)
			sudo service iofog-agent stop 2>/dev/null || sudo /etc/init.d/iofog-agent stop 2>/dev/null || true
			;;
		s6)
			sudo s6-svc -d /etc/s6/sv/iofog-agent 2>/dev/null || true
			;;
		runit)
			sudo sv stop iofog-agent 2>/dev/null || true
			;;
		upstart)
			sudo initctl stop iofog-agent 2>/dev/null || true
			;;
		*)
			sudo systemctl stop iofog-agent 2>/dev/null || sudo service iofog-agent stop 2>/dev/null || true
			;;
	esac
	(docker stop ${AGENT_CONTAINER_NAME} 2>/dev/null || podman stop ${AGENT_CONTAINER_NAME} 2>/dev/null) || true
}

do_create_env() {
ENV_FILE_NAME=iofog-agent.env # Used as an env file in systemd

ENV_FILE="$ETC_DIR/$ENV_FILE_NAME"

# Env file (for systemd)
rm -f "$ENV_FILE"
touch "$ENV_FILE"

echo "IOFOG_AGENT_IMAGE=${agent_image}" >> "$ENV_FILE"
echo "IOFOG_AGENT_TZ=${agent_tz}" >> "$ENV_FILE"

}

do_install_iofog() {
	echo "# Installing ioFog agent (airgap mode)..."
	
    for FOLDER in ${ETC_DIR} ${AGENT_LOG_FOLDER} ${AGENT_BACKUP_FOLDER} ${AGENT_MESSAGE_FOLDER} ${AGENT_SHARE_FOLDER}; do
        if [ ! -d "$FOLDER" ]; then
            echo "Creating folder: $FOLDER"
            sudo mkdir -p "$FOLDER"
            sudo chmod 775 "$FOLDER"
        fi
    done
	do_create_env

    USE_PODMAN="false"
    case "$lsb_dist" in
        rhel|centos|fedora|ol|sles|opensuse*) USE_PODMAN="true" ;;
    esac
    if [ "$USE_PODMAN" = "true" ]; then
        CONTAINER_RUNTIME="podman"
        SOCK_MOUNT="-v /run/podman/podman.sock:/run/podman/podman.sock:rw"
    else
        CONTAINER_RUNTIME="docker"
        SOCK_MOUNT="-v /var/run/docker.sock:/var/run/docker.sock:rw"
    fi

    if [ "${INIT_SYSTEM:-systemd}" = "systemd" ]; then
        if [ "$USE_PODMAN" = "true" ]; then
            echo "Using Podman (Quadlet) for container management..."
            sudo mkdir -p /etc/containers/systemd
            cat <<EOF | sudo tee /etc/containers/systemd/iofog-agent.container > /dev/null
[Unit]
Description=ioFog Agent Service
After=podman.service
Requires=podman.service

[Container]
ContainerName=${AGENT_CONTAINER_NAME}
Image=${agent_image}
PodmanArgs=--privileged --stop-timeout=60
EnvironmentFile=${ETC_DIR}/iofog-agent.env
Network=host
Volume=/run/podman/podman.sock:/run/podman/podman.sock:rw
Volume=iofog-agent-config:/etc/iofog-agent:rw
Volume=/var/log/iofog-agent:/var/log/iofog-agent:rw
Volume=/var/backups/iofog-agent:/var/backups/iofog-agent:rw
Volume=/usr/share/iofog-agent:/usr/share/iofog-agent:rw
Volume=/var/lib/iofog-agent:/var/lib/iofog-agent:rw
LogDriver=journald

[Service]
Restart=always

[Install]
WantedBy=default.target
EOF
            sudo systemctl daemon-reload
            sudo systemctl restart podman 2>/dev/null || true
            sudo systemctl enable iofog-agent.service
            sudo systemctl start iofog-agent.service
        else
            echo "Using Docker (systemd) for container management..."
            cat <<EOF | sudo tee /etc/systemd/system/iofog-agent.service > /dev/null
[Unit]
Description=ioFog Agent Service
After=docker.service
Requires=docker.service

[Service]
Restart=always
ExecStartPre=-/usr/bin/docker rm -f ${AGENT_CONTAINER_NAME}
ExecStart=/usr/bin/docker run --rm --name ${AGENT_CONTAINER_NAME} \\
--env-file ${ETC_DIR}/iofog-agent.env \\
-v /var/run/docker.sock:/var/run/docker.sock:rw \\
-v iofog-agent-config:/etc/iofog-agent:rw \\
-v /var/log/iofog-agent:/var/log/iofog-agent:rw \\
-v /var/backups/iofog-agent:/var/backups/iofog-agent:rw \\
-v /usr/share/iofog-agent:/usr/share/iofog-agent:rw \\
-v /var/lib/iofog-agent:/var/lib/iofog-agent:rw \\
--net=host \\
--privileged \\
--stop-timeout 60 \\
--attach stdout \\
--attach stderr \\
${agent_image}
ExecStop=/usr/bin/docker stop ${AGENT_CONTAINER_NAME}

[Install]
WantedBy=default.target
EOF
            sudo systemctl daemon-reload
            sudo systemctl enable iofog-agent.service
            sudo systemctl start iofog-agent.service
        fi
    else
        echo "Using $CONTAINER_RUNTIME with $INIT_SYSTEM for container management..."
        RUN_CMD="${CONTAINER_RUNTIME} run --rm -d --name ${AGENT_CONTAINER_NAME} --env-file ${ETC_DIR}/iofog-agent.env ${SOCK_MOUNT} -v iofog-agent-config:/etc/iofog-agent:rw -v /var/log/iofog-agent:/var/log/iofog-agent:rw -v /var/backups/iofog-agent:/var/backups/iofog-agent:rw -v /usr/share/iofog-agent:/usr/share/iofog-agent:rw -v /var/lib/iofog-agent:/var/lib/iofog-agent:rw --net=host --privileged --stop-timeout 60 ${agent_image}"
        RUN_CMD_FG="${CONTAINER_RUNTIME} run --rm --name ${AGENT_CONTAINER_NAME} --env-file ${ETC_DIR}/iofog-agent.env ${SOCK_MOUNT} -v iofog-agent-config:/etc/iofog-agent:rw -v /var/log/iofog-agent:/var/log/iofog-agent:rw -v /var/backups/iofog-agent:/var/backups/iofog-agent:rw -v /usr/share/iofog-agent:/usr/share/iofog-agent:rw -v /var/lib/iofog-agent:/var/lib/iofog-agent:rw --net=host --privileged --stop-timeout 60 ${agent_image}"
        STOP_CMD="${CONTAINER_RUNTIME} stop ${AGENT_CONTAINER_NAME}"
        case "$INIT_SYSTEM" in
            sysvinit|openrc)
                sudo tee /etc/init.d/iofog-agent > /dev/null <<INITSCRIPT
#!/bin/sh
### BEGIN INIT INFO
# Provides:          iofog-agent
# Required-Start:    \$network \$local_fs
# Required-Stop:     \$network \$local_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
### END INIT INFO
case "\$1" in
    start)
        n=0; while [ \$n -lt 30 ]; do [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] && break; n=\$((n+1)); sleep 1; done
        [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] || { echo "Container engine socket not available"; exit 1; }
        if ${CONTAINER_RUNTIME} ps --format '{{.Names}}' 2>/dev/null | grep -q "^${AGENT_CONTAINER_NAME}\$"; then exit 0; fi
        $RUN_CMD
        ;;
    stop) $STOP_CMD 2>/dev/null || true ;;
    restart) \$0 stop; \$0 start ;;
    status)
        if ${CONTAINER_RUNTIME} ps --format '{{.Names}}' 2>/dev/null | grep -q "^${AGENT_CONTAINER_NAME}\$"; then echo "running"; exit 0; else echo "stopped"; exit 1; fi
        ;;
    *) echo "Usage: \$0 {start|stop|restart|status}"; exit 1 ;;
esac
exit 0
INITSCRIPT
                sudo chmod +x /etc/init.d/iofog-agent
                if [ "$INIT_SYSTEM" = "openrc" ]; then
                    sudo rc-update add iofog-agent default 2>/dev/null || true
                    sudo rc-service iofog-agent start
                else
                    sudo update-rc.d iofog-agent defaults 2>/dev/null || sudo chkconfig iofog-agent on 2>/dev/null || true
                    sudo service iofog-agent start 2>/dev/null || sudo /etc/init.d/iofog-agent start
                fi
                ;;
            s6)
                sudo mkdir -p /etc/s6/sv/iofog-agent
                printf '#!/bin/sh\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/s6/sv/iofog-agent/run > /dev/null
                sudo chmod +x /etc/s6/sv/iofog-agent/run
                [ -d /etc/s6/adminsv/default ] && sudo ln -sf /etc/s6/sv/iofog-agent /etc/s6/adminsv/default/iofog-agent 2>/dev/null || true
                sudo s6-svc -u /etc/s6/sv/iofog-agent 2>/dev/null || true
                ;;
            runit)
                sudo mkdir -p /etc/runit/sv/iofog-agent
                printf '#!/bin/sh\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/runit/sv/iofog-agent/run > /dev/null
                sudo chmod +x /etc/runit/sv/iofog-agent/run
                [ -d /var/service ] && sudo ln -sf /etc/runit/sv/iofog-agent /var/service/iofog-agent 2>/dev/null || true
                [ -d /etc/runit/runsvdir/default ] && sudo ln -sf /etc/runit/sv/iofog-agent /etc/runit/runsvdir/default/iofog-agent 2>/dev/null || true
                sudo sv start iofog-agent 2>/dev/null || true
                ;;
            upstart)
                printf 'description "IoFog Agent container"\nstart on runlevel [2345]\nstop on runlevel [!2345]\nrespawn\nrespawn limit 10 5\nexec %s\n' "$RUN_CMD_FG" | sudo tee /etc/init/iofog-agent.conf > /dev/null
                sudo initctl reload-configuration 2>/dev/null || true
                sudo initctl start iofog-agent 2>/dev/null || true
                ;;
            *)
                sudo tee /etc/init.d/iofog-agent > /dev/null <<INITSCRIPT
#!/bin/sh
case "\$1" in start) n=0; while [ \$n -lt 30 ]; do [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] && break; n=\$((n+1)); sleep 1; done; [ -S /var/run/docker.sock ] || [ -S /run/podman/podman.sock ] || { echo "Container engine socket not available"; exit 1; }; $RUN_CMD ;; stop) $STOP_CMD ;; restart) $STOP_CMD; $RUN_CMD ;; *) echo "Usage: \$0 {start|stop|restart}"; exit 1 ;; esac
INITSCRIPT
                sudo chmod +x /etc/init.d/iofog-agent
                sudo /etc/init.d/iofog-agent start
                ;;
        esac
    fi

    EXECUTABLE_FILE=/usr/local/bin/iofog-agent
    if [ "$USE_PODMAN" = "true" ]; then
        cat <<'EOF' | sudo tee ${EXECUTABLE_FILE} > /dev/null
#!/bin/sh
CONTAINER_NAME="iofog-agent"
if ! podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: The iofog-agent container is not running."
    exit 1
fi
exec podman exec ${CONTAINER_NAME} iofog-agent "$@"
EOF
    else
        cat <<'EOF' | sudo tee ${EXECUTABLE_FILE} > /dev/null
#!/bin/sh
CONTAINER_NAME="iofog-agent"
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: The iofog-agent container is not running."
    exit 1
fi
exec docker exec ${CONTAINER_NAME} iofog-agent "$@"
EOF
    fi
    sudo chmod +x ${EXECUTABLE_FILE}

    echo "ioFog agent installation completed!"
}

do_start_iofog(){
	case "${INIT_SYSTEM:-systemd}" in
		systemd) sudo systemctl start iofog-agent >/dev/null 2>&1 & ;;
		sysvinit|openrc) sudo service iofog-agent start 2>/dev/null || sudo /etc/init.d/iofog-agent start 2>/dev/null & ;;
		s6) sudo s6-svc -u /etc/s6/sv/iofog-agent 2>/dev/null & ;;
		runit) sudo sv start iofog-agent 2>/dev/null & ;;
		upstart) sudo initctl start iofog-agent 2>/dev/null & ;;
		*) sudo systemctl start iofog-agent 2>/dev/null || sudo /etc/init.d/iofog-agent start 2>/dev/null & ;;
	esac
	local STATUS=""
	local ITER=0
	while [ "$STATUS" != "RUNNING" ]; do
		ITER=$((ITER+1))
		if [ "$ITER" -gt 600 ]; then
			echo "Timed out waiting for Agent to be RUNNING"
			exit 1
		fi
		sleep 1
		STATUS=$(sudo iofog-agent status 2>/dev/null | cut -f2 -d: | head -n 1 | tr -d '[:space:]')
		echo "${STATUS}"
	done
	sudo iofog-agent "config -cf 10 -sf 10"
	if [ "$lsb_dist" = "rhel" ] || [ "$lsb_dist" = "centos" ] || [ "$lsb_dist" = "fedora" ] || [ "$lsb_dist" = "ol" ] || [ "$lsb_dist" = "sles" ] || [ "$lsb_dist" = "opensuse" ]; then
		sudo iofog-agent "config -c unix:///var/run/podman/podman.sock"
	fi
}

agent_image="$1"
agent_tz="$2"
echo "Using variables"
echo "version: $agent_image"
echo "timezone: $agent_tz"
. /etc/iofog/agent/init.sh
init
do_check_install
do_stop_iofog
do_install_iofog
do_start_iofog



