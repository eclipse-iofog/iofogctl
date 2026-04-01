#!/bin/sh
set -x
set -e

AGENT_CONFIG_FOLDER=/etc/iofog-agent/
AGENT_LOG_FOLDER=/var/log/iofog-agent/

do_uninstall_iofog() {
	echo "# Removing ioFog agent..."

	case "$lsb_dist" in
		ubuntu|debian|raspbian)
			$sh_c "apt-get -y --purge autoremove iofog-agent"
			;;
		fedora|centos|rhel|ol)
			$sh_c "yum remove -y iofog-agent"
			;;
		sles|opensuse)
			$sh_c "zypper remove -y iofog-agent"
			;;
		*)
			echo "Error: Unsupported Linux distribution: $lsb_dist"
			exit 1
			;;
	esac

	# Remove config files
	$sh_c "rm -rf ${AGENT_CONFIG_FOLDER}"

	# Remove log files
	$sh_c "rm -rf ${AGENT_LOG_FOLDER}"
}

. /etc/iofog/agent/init.sh
init

do_uninstall_iofog