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
	AGENT_CONFIG_FOLDER=/etc/iofog-agent
	SAVED_AGENT_CONFIG_FOLDER=/tmp/agent-config-save
	PACKAGE_CLOUD_SCRIPT=package_cloud.sh
	echo "# Installing ioFog agent..."

	# Save iofog-agent config
	if [ -d ${AGENT_CONFIG_FOLDER} ]; then
		sudo rm -rf ${SAVED_AGENT_CONFIG_FOLDER}
		sudo mkdir -p ${SAVED_AGENT_CONFIG_FOLDER}
		sudo cp -r ${AGENT_CONFIG_FOLDER}/* ${SAVED_AGENT_CONFIG_FOLDER}/
	fi

	prefix=$([ -z "$token" ] && echo "" || echo "$token:@")
	echo $lsb_dist
	if [ "$lsb_dist" = "fedora" ] || [ "$lsb_dist" = "centos" ]; then
#		$sh_c "yum install yum-utils -y"
		repo_any="$(echo $repo | tr "/" "_")"
		echo "$repo_any"
		repo_file="yum.repos.d/$repo_any.repo"
echo "[$repo_any]
name=$repo_any
baseurl=https://packagecloud.io/$repo/rpm_any/rpm_any/\$basearch
repo_gpgcheck=1
gpgcheck=0
enabled=1
gpgkey=https://packagecloud.io/$repo/gpgkey
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
metadata_expire=300" > "/etc/$repo_file"
		$sh_c "yum -q makecache -y --disablerepo='*' --enablerepo=$repo_any"
		$sh_c "yum --disablerepo='*' --enablerepo=$repo_any install -y iofog-agent-$agent_version-1.noarch"
	else
    repo_any=$(echo $repo | tr "/" "_")
    echo $repo_any
    gpg_key_url="https://packagecloud.io/$repo/gpgkey"
    repo_list_file="sources.list.d/${repo_any}_any.list"
    apt_trusted_keyring_path="/etc/apt/trusted.gpg.d/$repo_any.gpg"
    apt install -qy debian-archive-keyring
    apt install -qy apt-transport-https
    # Import the gpg key
    echo "${gpg_key_url}"
    curl -fsSL "${gpg_key_url}" | gpg --dearmor > "${apt_trusted_keyring_path}"
    $sh_c "apt update -qy"
    # Repo definition
    echo "deb https://packagecloud.io/$repo/any/ any main
    deb-src https://packagecloud.io/$repo/any/ any main" > "/etc/apt/$repo_list_file"
    $sh_c "apt-get update -qy \
    -o Dir::Etc::sourcelist="$repo_list_file" \
    -o Dir::Etc::sourceparts="-" \
    -o APT::Get::List-Cleanup='0'"
    $sh_c "apt install --allow-downgrades iofog-agent=$agent_version -qy"
	fi
	do_check_iofog_on_arm

	# Restore iofog-agent config
	if [ -d ${SAVED_AGENT_CONFIG_FOLDER} ]; then
		sudo mv ${SAVED_AGENT_CONFIG_FOLDER}/* ${AGENT_CONFIG_FOLDER}/
		sudo rmdir ${SAVED_AGENT_CONFIG_FOLDER}
	fi
	sudo chmod 775 ${AGENT_CONFIG_FOLDER}
}

do_start_iofog(){
	# shellcheck disable=SC2261
	sudo service iofog-agent start > /dev/null 2&>1 &
	local STATUS=""
	local ITER=0
	while [ "$STATUS" != "RUNNING" ] ; do
    ITER=$((ITER+1))
    if [ "$ITER" -gt 60 ]; then
      echo 'Timed out waiting for Agent to be RUNNING'
      exit 1;
    fi
    sleep 1
    STATUS=$(sudo iofog-agent status | cut -f2 -d: | head -n 1 | tr -d '[:space:]')
    echo "${STATUS}"
	done
	sudo iofog-agent "config -cf 10 -sf 10"
}

agent_version="$1"
repo=$([ -z "$2" ] && echo "iofog/iofog-agent" || echo "$2")
token="$3"
echo "Using variables"
echo "version: $agent_version"
echo "repo: $repo"
echo "token: $token"

. /etc/iofog/agent/init.sh
init
do_check_install
do_stop_iofog
do_install_iofog
do_start_iofog