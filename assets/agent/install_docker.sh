#!/bin/sh
set -x
set -e

start_docker() {
	set +e
	# check if docker is running
	if ! $sh_c "docker ps" >/dev/null 2>&1; then
		# Try init.d
		$sh_c "/etc/init.d/docker start"
		local err_code=$?
		# Try systemd
		if [ $err_code -ne 0 ]; then
			$sh_c "service docker start"
			err_code=$?
		fi
		# Try snapd
		if [ $err_code -ne 0 ]; then
			$sh_c "snap docker start"
			err_code=$?
		fi
		if [ $err_code -ne 0 ]; then
			echo "Could not start Docker daemon"
			exit 1
		fi
	fi
	set -e
}

do_configure_overlay() {
	local driver="$DOCKER_STORAGE_DRIVER"
	if [ -z "$driver" ]; then
		driver="overlay"
	fi
	echo "# Configuring /etc/systemd/system/docker.service.d/overlay.conf..."
	if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
		if [ ! -d "/etc/systemd/system/docker.service.d" ]; then
			$sh_c "mkdir -p /etc/systemd/system/docker.service.d"
		fi
		if [ ! -f "/etc/systemd/system/docker.service.d/overlay.conf" ] || ! grep -Fxq "ExecStart=/usr/bin/dockerd --storage-driver $driver -H unix:// -H tcp://127.0.0.1:2375" "/etc/systemd/system/docker.service.d/overlay.conf"; then
			$sh_c 'echo "[Service]" > /etc/systemd/system/docker.service.d/overlay.conf'
			$sh_c 'echo "ExecStart=" >> /etc/systemd/system/docker.service.d/overlay.conf'
			$sh_c "echo \"ExecStart=/usr/bin/dockerd --storage-driver $driver -H unix:// -H tcp://127.0.0.1:2375\" >> /etc/systemd/system/docker.service.d/overlay.conf"
		fi
		$sh_c "systemctl daemon-reload"
		$sh_c "service docker restart"
	fi
}

do_install_docker() {
	# Check that Docker 18.09.2 or greater is installed
	if command_exists docker; then
		docker_version=$(docker -v | sed 's/.*version \(.*\),.*/\1/' | tr -d '.')
		if [ "$docker_version" -ge 18090 ]; then
			echo "# Docker $docker_version already installed"
			start_docker
			do_configure_overlay
			return
		fi
	fi
	echo "# Installing Docker..."
	case "$dist_version" in
		"stretch")
			$sh_c "apt install -y apt-transport-https ca-certificates curl gnupg2 software-properties-common"
			curl -fsSL https://download.docker.com/linux/debian/gpg | $sh_c "apt-key add -"
			$sh_c "sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/debian $(lsb_release -cs) stable\""
			$sh_c "apt-get update -y"
			$sh_c "sudo apt install -y docker-ce"
		;;
    7|8)
      $sh_c "sudo yum install -y yum-utils || echo 'yum-utils already installed'"
      $sh_c "sudo yum-config-manager \
            --add-repo \
            https://download.docker.com/linux/centos/docker-ce.repo"
      $sh_c "sudo yum install docker-ce docker-ce-cli containerd.io -y"
    ;;
		*)
			curl -fsSL https://get.docker.com/ | sh
		;;
	esac
	
	if ! command_exists docker; then
		echo "Failed to install Docker"
		exit 1
	fi
	start_docker
	do_configure_overlay
}

. /etc/iofog/agent/init.sh
init
do_install_docker