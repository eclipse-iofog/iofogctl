#!/bin/sh
set -x
set -e

java_major_version=0
java_minor_version=0
do_check_install() {
	if command_exists java; then
        java_major_version="$(java --version | head -n1 | awk '{print $2}' | cut -d. -f1)"
        java_minor_version="$(java --version | head -n1 | awk '{print $2}' | cut -d. -f2)"
	fi
	if [ "$java_major_version" -ge "11" ]  && [ "$java_minor_version" -ge "0" ]; then
		echo "Java $java_major_version.$java_minor_version  already installed."
		exit 0
	fi
}

do_install_java() {
	echo "# Installing java 11..."
	echo ""
	os_arch=$(getconf LONG_BIT)
	is_arm=""
	if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
		is_arm="-arm"
	fi
	case "$lsb_dist" in
		ubuntu)
			$sh_c "add-apt-repository ppa:openjdk-r/ppa -y"
			$sh_c "apt-get update"
			$sh_c "apt install -y openjdk-11-jdk"
		;;
		debian|raspbian|mendel)
			$sh_c "apt-get update"
			$sh_c "apt install -y openjdk-11-jdk"
		;;
		fedora|centos)
			$sh_c "yum install -y openjdk-11-jdk"
		;;
	esac
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
	while ! $sh_c "$installer update" && [ "$iter" -lt 6 ]; do
		sleep 5
		iter=$((iter+1))
	done
}

. /etc/iofog/agent/init.sh
init
do_check_install
do_install_deps
do_install_java