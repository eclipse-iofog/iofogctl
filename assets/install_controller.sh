#!/bin/sh
set -x
set -e

install_iofog_controller_snapshot() {
	echo "# Installing ioFog Controller snapshot (dev) repo"
	echo
	token="?master_token="$token
	npmrc_file=${HOME}/.npmrc
	epoch_time=`date +%s`
	save_npmrc=${HOME}/${epoch_time}.npmrc.bak
	if [ -f npmrc_file ]; then
		# Save any prexisting .npmrc
		mv ${HOME}/.npmrc ${save_npmrc}
	fi
	curl -s https://packagecloud.io/install/repositories/iofog/iofog-controller-snapshots/script.node.sh$token | bash
	if [ -f save_npmrc ]; then
		# Append previous npmrc configuration
		echo "" >> npmrc_file
		cat save_npmrc >> npmrc_file
	fi
}

# TODO: Handle specifying a version for Connector
deploy_controller() {
	# Install if does not exist
	if [ ! -z $(command -v iofog-controller) ]; then
		iofog-controller stop
	else
		# nvm
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash
		export NVM_DIR="$HOME/.nvm"
		[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
		[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
		nvm install lts/*
		nvm alias default lts/*
		if [ ! -z $token ]; then
			install_iofog_controller_snapshot
		fi
		if [ -z $version ]; then
			npm install -g iofogcontroller --unsafe-perm
		else
			npm install -g "iofogcontroller@$version" --unsafe-perm
		fi
	fi
	# TODO: This env var is used to change default port. Replace dev env and specify default port outright
	NODE_ENV=development iofog-controller start
}

deploy_connector() {
	# Install if does not exist 
	if [ ! -z $(command -v iofog-connector) ]; then
		# Stop existing deployments
		sudo service iofog-connector stop
	else
		# Debian/Ubuntu
		sudo apt-get -y update
		sudo apt-get -y install openjdk-8-jre
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.deb.sh | sudo bash
		sudo apt-get -y install iofog-connector

		# Redhat/Fedora/CentOS
		#su -c "yum install java-1.8.0-openjdk"
		#curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.rpm.sh | sudo bash
		#sudo yum install iofog-connector

		echo '{
  "ports": [
    "6000-6001",
  ],
  "address": "0.0.0.0"
}' | sudo tee /etc/iofog-connector/iofog-connector.conf
	fi
	sudo service iofog-connector start
}

# main
version="$1"
token="$2"
deploy_controller
deploy_connector