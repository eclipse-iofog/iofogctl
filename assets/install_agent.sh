#!/bin/sh
set -x
set -e

SUPPORT_MAP="
x86_64-centos-7
x86_64-fedora-26
x86_64-fedora-27
x86_64-fedora-28
x86_64-debian-wheezy
x86_64-debian-jessie
x86_64-debian-stretch
x86_64-debian-buster
x86_64-ubuntu-trusty
x86_64-ubuntu-xenial
x86_64-ubuntu-bionic
x86_64-ubuntu-artful
s390x-ubuntu-xenial
s390x-ubuntu-bionic
s390x-ubuntu-artful
ppc64le-ubuntu-xenial
ppc64le-ubuntu-bionic
ppc64le-ubuntu-artful
aarch64-ubuntu-xenial
aarch64-ubuntu-bionic
aarch64-debian-jessie
aarch64-debian-stretch
aarch64-debian-buster
aarch64-fedora-26
aarch64-fedora-27
aarch64-fedora-28
aarch64-centos-7
armv6l-raspbian-jessie
armv7l-raspbian-jessie
armv6l-raspbian-stretch
armv7l-raspbian-stretch
armv6l-raspbian-buster
armv7l-raspbian-buster
armv7l-debian-jessie
armv7l-debian-stretch
armv7l-debian-buster
armv7l-ubuntu-trusty
armv7l-ubuntu-xenial
armv7l-ubuntu-bionic
armv7l-ubuntu-artful
"


get_distribution() {
	lsb_dist=""
	# Every system that we officially support has /etc/os-release
	if [ -r /etc/os-release ]; then
		lsb_dist="$(. /etc/os-release && echo "$ID")"
		lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
	else
		echo "Unsupported Linux distribution!"
		exit 1
	fi
	echo "# Our distro is '$lsb_dist'"
	echo $lsb_dist
}

# Check if this is a forked Linux distro
check_forked() {
	# Check for lsb_release command existence, it usually exists in forked distros
	if command_exists lsb_release; then
		# Check if the `-u` option is supported
		set +e
		lsb_release -a
		lsb_release_exit_code=$?
		set -e

		# Check if the command has exited successfully, it means we're in a forked distro
		if [ "$lsb_release_exit_code" = "0" ]; then
			# Print info about current distro
			cat <<-EOF
			You're using '$lsb_dist' version '$dist_version'.
			EOF

			# Get the upstream release info
			lsb_dist=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'id' | cut -d ':' -f 2 | tr -d '[:space:]')
			dist_version=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'codename' | cut -d ':' -f 2 | tr -d '[:space:]')

			# Print info about upstream distro
			cat <<-EOF
			Upstream release is '$lsb_dist' version '$dist_version'.
			EOF
		else
			if [ -r /etc/debian_version ] && [ "$lsb_dist" != "ubuntu" ] && [ "$lsb_dist" != "raspbian" ]; then
				if [ "$lsb_dist" = "osmc" ]; then
					# OSMC runs Raspbian
					lsb_dist=raspbian
				else
					# We're Debian and don't even know it!
					lsb_dist=debian
				fi
				dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
				case "$dist_version" in
					10)
						dist_version="buster"
					;;
					9)
						dist_version="stretch"
					;;
					8|'Kali Linux 2')
						dist_version="jessie"
					;;
					7)
						dist_version="wheezy"
					;;
				esac
			elif [ -r /etc/redhat-release ] && [ "$lsb_dist" = "" ]; then
				lsb_dist=redhat
			fi
		fi
	fi
}

command_exists() {
	command -v "$@"
}

do_install_java() {
	echo "# Installing java 8..."
	echo
	java8_version=0
	if command_exists java; then
        java8_version="$(java -version 2>&1 | awk -F '"' '/version/ {print $2}' | grep 1.8 | cut -d'_' -f 2)"
	fi
	if [ "$java8_version" -lt "181" ]; then
		os_arch=$(getconf LONG_BIT)
		is_arm=""
		if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
			is_arm="-arm"
		fi
		cd /opt/
		$sh_c 'wget -q --no-check-certificate '"http://www.edgeworx.io/downloads/jdk/jdk-8u211$is_arm-$os_arch.tar.gz"''
		$sh_c "tar xzf jdk-8u211$is_arm-$os_arch.tar.gz"
		cd /opt/jdk1.8.0_211/	
		case "$lsb_dist" in
			debian|raspbian|ubuntu)
				$sh_c "update-alternatives --install /usr/bin/java java /opt/jdk1.8.0_211/bin/java 1100"
				;;		
			fedora|centos)
				$sh_c "alternatives --install /usr/bin/java java /opt/jdk1.8.0_211/bin/java 4"
				;;
		esac
	fi
}

start_docker() {
	set +e
	# check if docker is running
	if [ ! -f /var/run/docker.pid ]; then
		$sh_c "/etc/init.d/docker start"
		local err_code=$?
		if [ $err_code -ne 0 ]; then
			$sh_c "service docker start"
			err_code=$?
		fi
		if [ $err_code -ne 0 ]; then
			echo "Could not start Docker daemon"
			exit 1
		fi
	fi
	set -e
}

do_install_docker() {
	# Check that Docker 18.09.2 or greater is installed
	if command_exists docker; then
		docker_version=$(docker -v | sed 's/.*version \(.*\),.*/\1/' | tr -d '.')
		if [ "$docker_version" -ge 18090 ]; then
			echo "# Docker $docker_version already installed"
			start_docker
			return
		fi
	fi
	echo "# Installing Docker..."
	curl -fsSL https://get.docker.com/ | sh
	
	if ! command_exists docker; then
		echo "Failed to install Docker"
		exit 1
	fi
	start_docker

	if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
		if [ ! -d "/etc/systemd/system/docker.service.d" ]; then
			$sh_c "mkdir -p /etc/systemd/system/docker.service.d"
		fi
		$sh_c 'echo "[Service]" > /etc/systemd/system/docker.service.d/overlay.conf'
		$sh_c 'echo "ExecStart=" >> /etc/systemd/system/docker.service.d/overlay.conf'
		$sh_c 'echo "ExecStart=/usr/bin/dockerd --storage-driver overlay -H unix:// -H tcp://127.0.0.1:2375" >> /etc/systemd/system/docker.service.d/overlay.conf'
		$sh_c "systemctl daemon-reload"
		$sh_c "service docker restart"
	fi
}

do_stop_iofog() {
	if command_exists iofog-agent; then
		sudo service iofog-agent stop
	fi
}


do_check_iofog_on_arm() {
    if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then

        echo "# We re on ARM ($(uname -m)) : Updating config.xml to use correct docker_url"
		$sh_c 'sed -i -e "s|<docker_url>.*</docker_url>|<docker_url>tcp://127.0.0.1:2375/</docker_url>|g" /etc/iofog-agent/config.xml'

		echo "# Restarting iofog-agent service"
		$sh_c "service iofog-agent stop"
		sleep 3
		$sh_c "service iofog-agent start"
	fi
}

do_install_iofog() {
	echo "# Installing ioFog agent..."

	prefix=$([ -z "$token" ] && echo "" || echo "$token:@")

	case "$lsb_dist" in
		ubuntu)
			curl -s "https://${prefix}packagecloud.io/install/repositories/$repo/script.deb.sh" | $sh_c "bash"
			$sh_c "apt-get install -y --allow-downgrades iofog-agent=$agent_version"
			;;
		fedora|centos)
			curl -s "https://${prefix}packagecloud.io/install/repositories/$repo/script.rpm.sh" | $sh_c "bash"
			$sh_c "yum install -y iofog-agent-"$agent_version"-1.noarch"
			;;
		debian|raspbian)
			curl -s "https://${prefix}packagecloud.io/install/repositories/$repo/script.deb.sh" | $sh_c "bash"
			$sh_c "apt-get install -y --allow-downgrades iofog-agent=$agent_version"
			;;
	esac

	do_check_iofog_on_arm
}

do_install_deps() {
	local installer=""
	case "$lsb_dist" in
		ubuntu|debian|raspbian)
			installer="apt"
			;;
		fedora|centos)
			installer="yum"
			;;
	esac

	local iter=0
	while [ ! $($sh_c "$installer update") ] && [ "$iter" -lt 6 ]; do
		sleep 5
		iter=$((iter+1))
	done

	if [ -z $(command -v wget) ]; then
		$sh_c "$installer install -y wget"
	fi
}

do_install() {
	echo "# Executing iofog install script"
	
	sh_c='sh -c'
	if [ "$user" != 'root' ]; then
		if command_exists sudo; then
			sh_c='sudo -E sh -c'
		elif command_exists su; then
			sh_c='su -c'
		else
			cat >&2 <<-'EOF'
			Error: this installer needs the ability to run commands as root.
			We are unable to find either "sudo" or "su" available to make this happen.
			EOF
			exit 1
		fi
	fi

	get_distribution

	case "$lsb_dist" in

		ubuntu)
			if command_exists lsb_release; then
				dist_version="$(lsb_release --codename | cut -f2)"
			fi
			if [ -z "$dist_version" ] && [ -r /etc/lsb-release ]; then
				dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
			fi
		;;

		debian|raspbian)
			dist_version="$(sed 's/\/.*//' /etc/debian_version | sed 's/\..*//')"
			case "$dist_version" in
				10)
					dist_version="buster"
				;;
				9)
					dist_version="stretch"
				;;
				8)
					dist_version="jessie"
				;;
				7)
					dist_version="wheezy"
				;;
			esac
		;;

		centos)
			if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
				dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
			fi
		;;

		rhel|ol|sles)
			ee_notice "$lsb_dist"
			exit 1
			;;

		*)
			if command_exists lsb_release; then
				dist_version="$(lsb_release --release | cut -f2)"
			fi
			if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
				dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
			fi
		;;

	esac

	# Check if this is a forked Linux distro
	check_forked

	# Check if we actually support this configuration
	if [ "$lsb_dist" = "redhat" ]; then
		cat >&2 <<-'EOF'

		Since Docker Community Edition is not supported for RedHat you have to procceed with installation manually.
		Please visit the following URL for more detailed installation instructions:

		https://iofog.org/install/RHEL

		EOF
		exit 1
	elif ! echo "$SUPPORT_MAP" | grep "$(uname -m)-$lsb_dist-$dist_version"; then
		cat >&2 <<-'EOF'

		Either your platform is not easily detectable or is not supported by this
		installer script.
		Please visit the following URL for more detailed installation instructions:

		https://iofog.org/developer

		EOF
		exit 1
	fi

	do_install_deps

	do_install_java
	
	do_install_docker
	
	do_stop_iofog

	do_install_iofog
}

agent_version="$1"
repo=$([ -z "$2" ] && echo "iofog/iofog-agent" || echo "$2")
token="$3"
echo "Using variables"
echo "version: $agent_version"
echo "repo: $repo"
echo "token: $token"

do_install