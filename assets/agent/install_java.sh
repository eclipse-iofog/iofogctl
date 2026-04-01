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
	if [ "$java_major_version" -ge "17" ]  && [ "$java_minor_version" -ge "0" ]; then
		echo "Java $java_major_version.$java_minor_version  already installed."
		exit 0
	fi
}

do_install_java() {
	echo "# Installing java 17..."
	echo ""
	os_arch=$(getconf LONG_BIT)
	is_arm=""
	if [ "$lsb_dist" = "raspbian" ] || [ "$(uname -m)" = "armv7l" ] || [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "armv8" ]; then
		is_arm="-arm"
	fi
	case "$lsb_dist" in
		ubuntu|debian|raspbian|mendel)
			$sh_c "apt-get update -y"
			$sh_c "apt install -y openjdk-17-jdk"
		;;
		fedora|centos|rhel|ol)
			$sh_c "yum install -y java-17-openjdk"
		;;
		sles|opensuse*)
			$sh_c "zypper refresh"
			$sh_c "zypper install -y java-17-openjdk"
		;;
		*)
			echo "Unsupported distribution: $lsb_dist"
			exit 1
		;;
	esac
}

do_install_deps() {
	local installer=""
	case "$lsb_dist" in
		ubuntu|debian|raspbian|mendel)
			installer="apt"
		;;
		fedora|centos|rhel|ol)
			installer="yum"
		;;
		sles|opensuse*)
			installer="zypper"
		;;
		*)
			echo "Unsupported distribution: $lsb_dist"
			exit 1
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