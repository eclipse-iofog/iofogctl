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

check_command_status() {
	if [ $1 -eq 0 ]; then
		echo
		echo "$2"
		echo
	elif [ $1 -eq 776 ]; then
		echo
		echo "$5"
		echo
	elif [ $1 -eq 777 ]; then
		echo
		echo "$4"
		echo 
	else
		echo
		echo "$3"
		echo
		exit $1
	fi
}

disable_package_preconfiguration() {
	if [ "$lsb_dist" = "debian" ]; then
		if [ -f /etc/apt/apt.conf.d/70debconf ]; then
			$sh_c 'ex +"%s@DPkg@//DPkg" -cwq /etc/apt/apt.conf.d/70debconf'
			$sh_c 'dpkg-reconfigure debconf -f noninteractive -p critical'
		fi
	fi
}

add_repo_if_not_exists() {
	repo="$1"
	if ! grep -Fxq "$repo" /etc/apt/sources.list; then
		($sh_c "echo \"$repo\" >> /etc/apt/sources.list")
	fi
}

add_initial_apt_repos_if_not_exist() {
	case "$lsb_dist" in
		debian)
			if [ "$dist_version" = "stretch" ]; then
				add_repo_if_not_exists "deb http://deb.debian.org/debian stretch main"
				add_repo_if_not_exists "deb-src http://deb.debian.org/debian stretch main"
				add_repo_if_not_exists "deb http://deb.debian.org/debian-security/ stretch/updates main"
				add_repo_if_not_exists "deb-src http://deb.debian.org/debian-security/ stretch/updates main"
				add_repo_if_not_exists "deb http://deb.debian.org/debian stretch-updates main"
				add_repo_if_not_exists "deb-src http://deb.debian.org/debian stretch-updates main"
			elif [ "$dist_version" = "jessie" ]; then
				add_repo_if_not_exists "deb http://ftp.de.debian.org/debian jessie main"
			elif [ "$dist_version" = "buster" ]; then
				add_repo_if_not_exists "deb http://ftp.de.debian.org/debian buster main"
			fi
			$sh_c 'apt-get update -qq'
			;;
	esac
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
				command_status=$?
				;;		
			fedora|centos)
				$sh_c "alternatives --install /usr/bin/java java /opt/jdk1.8.0_211/bin/java 4"
				command_status=$?
				;;
		esac
		# Proceeding with existing java if java update failed
		if [ "$command_status" -ne "0" ] && [ ! -z "$java8_version" ]; then
			command_status=776
		fi
	else
		command_status=777
	fi
}

handle_docker_unsuccessful_installation() {
	if ! command_exists docker; then
		# for fedora 28
		if [ "$lsb_dist" == "fedora" ] && [ "$dist_version" == "28" ]; then
			$sh_c "dnf -y -q install https://download.docker.com/linux/fedora/27/x86_64/stable/Packages/docker-ce-18.03.1.ce-1.fc27.x86_64.rpm"
		fi	
	fi
}

start_docker() {
	# check if docker is running
	if [ ! -f /var/run/docker.pid ]; then
		$sh_c "/etc/init.d/docker start"
		command_status=$?
		if [ $command_status -ne 0 ]; then
			$sh_c "service docker start"
			command_status=$?
		fi
	else
		command_status=0	
	fi
}

install_docker_apt() {
	sudo $1 update -qy
	sudo $1 upgrade -qy
	sudo $1 install \
			apt-transport-https \
			ca-certificates \
			curl \
			gnupg-agent \
			software-properties-common -qy
	DISTRO=$(lsb_release -a 2> /dev/null | grep 'Distributor ID' | awk '{print $3}')
	if [ "$DISTRO" == "Ubuntu" ]; then
		curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
	else
		curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -
	fi
	sudo $1 update -qy
	sudo $1 install docker-ce docker-ce-cli containerd.io -qy
}

install_docker_linux() {
    if [ -x "$(command -v apt-get)" ]; then
			install_docker_apt "apt-get"
	elif [ -x "$(command -v apt)" ]; then
			install_docker_apt "apt"
    elif [ -x "$(command -v dnf)" ]; then
        sudo dnf -y install dnf-plugins-core
        sudo dnf config-manager \
            --add-repo \
            https://download.docker.com/linux/fedora/docker-ce.repo
        sudo dnf install docker-ce docker-ce-cli containerd.io -y
    elif [ -x "$(command -v yum)" ]; then
        sudo yum install -y yum-utils \
            device-mapper-persistent-data \
            lvm2 -qy
        sudo yum-config-manager \
            --add-repo \
            https://download.docker.com/linux/centos/docker-ce.repo
        sudo yum install docker-ce docker-ce-cli containerd.io -qy
    else
        handle_docker_unsuccessful_installation
    fi
}

do_install_docker() {
	# Check that Docker 18.09.2 or greater is installed
	if command_exists docker; then
		docker_version=$(docker -v | sed 's/.*version \(.*\),.*/\1/' | tr -d '.')
		if [ "$docker_version" -ge 18090 ]; then
			echo "# Docker $docker_version already installed"
			return
		fi
	fi
	echo "# Installing Docker..."
	# install_docker_linux
	sleep 3
	curl -fsSL https://get.docker.com/ | sh
	
	handle_docker_unsuccessful_installation
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
		command_status=$?
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
	echo
	case "$lsb_dist" in
		ubuntu)
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent/script.deb.sh | $sh_c "bash"
			$sh_c "apt-get install -y iofog-agent"
			command_status=$?
			;;
		fedora|centos)
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent/script.rpm.sh | $sh_c "bash"
			$sh_c "yum install -y iofog-agent"
			command_status=$?
			;;
		debian|raspbian)
			if [ "$lsb_dist" = "debian" ]; then
				$sh_c "apt-get install -y -qq net-tools"
			fi
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent/script.deb.sh | $sh_c "bash"
			$sh_c "apt-get install -y iofog-agent"
			command_status=$?
			;;
	esac

	do_check_iofog_on_arm
}

do_install_iofog_dev() {
	echo "# Installing ioFog agent dev version: "$version 
	echo
	token="?master_token="$token
	case "$lsb_dist" in
		ubuntu)
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent-snapshots/script.deb.sh$token | $sh_c "bash"
			$sh_c "apt-get install -y --allow-downgrades iofog-agent="$version""
			command_status=$?
			;;
		fedora|centos)
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent-snapshots/script.rpm.sh$token  | $sh_c "bash"
			$sh_c "yum install -y iofog-agent-"$version"-1.noarch"
			command_status=$?
			;;
		debian|raspbian)
			if [ "$lsb_dist" = "debian" ]; then
				$sh_c "apt-get install -y --allow-downgrades -qq net-tools"
			fi
			curl -s https://packagecloud.io/install/repositories/iofog/iofog-agent-snapshots/script.deb.sh$token  | $sh_c "bash"
			$sh_c "apt-get install -y --allow-downgrades iofog-agent="$version""
			command_status=$?
			;;
	esac

	do_check_iofog_on_arm
}

do_install() {
	echo "# Executing iofog install script"
	
	command_status=0
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

	disable_package_preconfiguration

	# Run setup for each distro accordingly
	add_initial_apt_repos_if_not_exist
	
	do_install_java
	
	do_install_docker
	
	do_stop_iofog

	if [ "$env" = "dev" ]
	then
		do_install_iofog_dev 
	else
		do_install_iofog
	fi
}

env="$1"
version="$2"
token="$3"
echo "Using variables"
echo "Env: ${env}"
echo "version: ${version}"
echo "token: ${token}"
if [ "$env" = "dev" ]; then
	echo "----> Dev environment"
fi
if ! [ -z "$version" ]; then
	echo "----> Dev environment, version: $version"
fi
if ! [ -z "$token" ]; then
	echo "----> Dev environment, token: $token"
fi

if [ "$env" = "dev" ] && ! [ -z "$version" ] && ! [ -z "$token" ]; then
	echo "Will be installing iofog-agent version $version from snapshot repo"
else 
	env=""
	echo "Will be installing iofog-agent from public repo"
	echo "To install from snapshot repo, run script with additional param 'dev <VERSION> <PACKAGE_CLOUD_TOKEN>'"
fi
do_install