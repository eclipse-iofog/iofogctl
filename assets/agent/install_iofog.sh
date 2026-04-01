#!/bin/sh
set -x
set -e

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
	if command_exists iofog-agent; then
		sudo systemctl stop iofog-agent
	fi
}



do_install_iofog() {
	AGENT_CONFIG_FOLDER=/etc/iofog-agent
	SAVED_AGENT_CONFIG_FOLDER=/tmp/agent-config-save
	echo "# Installing ioFog agent..."

	# Save iofog-agent config
	if [ -d ${AGENT_CONFIG_FOLDER} ]; then
		sudo rm -rf ${SAVED_AGENT_CONFIG_FOLDER}
		sudo mkdir -p ${SAVED_AGENT_CONFIG_FOLDER}
		sudo cp -r ${AGENT_CONFIG_FOLDER}/* ${SAVED_AGENT_CONFIG_FOLDER}/
	fi

	echo $lsb_dist
	case "$lsb_dist" in
		fedora|rhel|ol|centos)
			$sh_c "yum update -y"
			$sh_c "yum install -y iofog-agent-$agent_version-1.noarch"
			;;
		sles|opensuse)
			$sh_c "zypper refresh"
			$sh_c "zypper install -y iofog-agent=$agent_version"
			;;
		*)
			$sh_c "apt update -qy"
			$sh_c "apt install --allow-downgrades iofog-agent=$agent_version -qy"
			;;
	esac

	# Restore iofog-agent config
	if [ -d ${SAVED_AGENT_CONFIG_FOLDER} ]; then
		sudo mv ${SAVED_AGENT_CONFIG_FOLDER}/* ${AGENT_CONFIG_FOLDER}/
		sudo rmdir ${SAVED_AGENT_CONFIG_FOLDER}
	fi
	sudo chmod 775 ${AGENT_CONFIG_FOLDER}
}

do_start_iofog(){

	sudo systemctl start iofog-agent > /dev/null 2&>1 &
	local STATUS=""
	local ITER=0
	while [ "$STATUS" != "RUNNING" ] ; do
    ITER=$((ITER+1))
    if [ "$ITER" -gt 600 ]; then
      echo 'Timed out waiting for Agent to be RUNNING'
      exit 1;
    fi
    sleep 1
    STATUS=$(sudo iofog-agent status | cut -f2 -d: | head -n 1 | tr -d '[:space:]')
    echo "${STATUS}"
	done
	sudo iofog-agent "config -cf 10 -sf 10"
	if [ "$lsb_dist" = "rhel" ] || [ "$lsb_dist" = "centos" ] || [ "$lsb_dist" = "fedora" ] || [ "$lsb_dist" = "ol" ] || [ "$lsb_dist" = "sles" ] || [ "$lsb_dist" = "opensuse" ]; then
        sudo iofog-agent "config -c unix:///var/run/podman/podman.sock"
    fi   
}

agent_version="$1"
echo "Using variables"
echo "version: $agent_version"

. /etc/iofog/agent/init.sh
init

# Native agent is supported only on package-managed OSes (deb/rpm) with systemd
if [ "$PACKAGE_TYPE" != "deb" ] && [ "$PACKAGE_TYPE" != "rpm" ]; then
	echo "Error: This operating system is not supported for native agent installation."
	echo "Please deploy the agent as a container (container agent) on this host."
	exit 1
fi
if [ "$INIT_SYSTEM" != "systemd" ]; then
	echo "Error: Native agent is supported only on systemd. This system uses $INIT_SYSTEM."
	echo "Please deploy the agent as a container (container agent) on this host."
	exit 1
fi

do_check_install
do_stop_iofog
do_install_iofog
do_start_iofog