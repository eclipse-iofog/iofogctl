#!/bin/sh
set -x
set -e

# TODO: Handle specifying a version for Controller/Connector
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
		npm install -g iofogcontroller --unsafe-perm
	fi
	iofog-controller start
}

deploy_connector() {
	# Install if does not exist 
	if [ ! -z $(command -v iofog-connector) ]; then
		# Stop existing deployments
		sudo service iofog-connector stop
	else
		# Debian/Ubuntu
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
deploy_controller
deploy_connector