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
	add_npm_repository=1
	if [ -f $npmrc_file ]; then
		if [ ! -z "$(cat $npmrc_file | grep '//packagecloud.io/iofog/iofog-controller-snapshots/npm/:_authToken=')" ]; then
			# assume we don't need to install npm repo
			add_npm_repository=0
		fi
		# Save any prexisting .npmrc
		echo "Moving npmrc"
		mv ${npmrc_file} ${save_npmrc}
	fi
	if [ $add_npm_repository -eq 1 ]; then
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-controller-snapshots/script.node.sh$token | bash
	fi
	if [ -f $save_npmrc ]; then
		echo "Restoring npmrc"
		if [ $add_npm_repository -eq 1 ]; then
			# Append previous npmrc configuration
			echo "" >> $npmrc_file
			cat $save_npmrc >> $npmrc_file
			rm $save_npmrc
		else
			# Save any prexisting .npmrc
			mv ${save_npmrc} ${npmrc_file}
		fi
	fi
}

load_existing_nvm() {
	set +e
	if [ -z "$(command -v nvm)" ]; then
		export NVM_DIR="${HOME}/.nvm"
		mkdir -p $NVM_DIR
		if [ -f "$NVM_DIR/nvm.sh" ]; then
			[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" # This loads nvm
		fi
	fi
	set -e
}

# TODO: Handle specifying a version for Connector
deploy_controller() {
	# Install if does not exist
	if [ ! -z $(command -v iofog-controller) ]; then
		iofog-controller stop || true
	fi
	# nvm
	load_existing_nvm
	if [ -z "$(command -v nvm)" ]; then
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash
		export NVM_DIR="${HOME}/.nvm"
		[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
		nvm install lts/*
	else
		nvm use lts/* || true
	fi
	if [ ! -z $token ]; then
		install_iofog_controller_snapshot
	fi
	if [ -z $version ]; then
		npm install -g -f iofogcontroller --unsafe-perm
	else
		npm install -g -f "iofogcontroller@$version" --unsafe-perm
	fi
	iofog-controller start
}

deploy_connector() {
	# Install if does not exist 
	if [ ! -z $(command -v iofog-connector) ]; then
		# Stop existing deployments
		sudo service iofog-connector stop || true
	fi
	if [ ! -z $(command -v apt-get) ]; then
		# Debian/Ubuntu
		sudo apt-get -y update
		sudo apt-get -y install openjdk-8-jre
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.deb.sh | sudo bash
		sudo apt-get -y install iofog-connector
	elif [ ! -z $(command -v yum) ]; then
		# Redhat/Fedora/CentOS
		su -c "yum update -y"
		su -c "yum install java-1.8.0-openjdk -y"
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.rpm.sh | sudo bash
		sudo yum install iofog-connector -y
	else
		echo "Could not detect system package manager"
		exit 1
	fi

	echo '{
  "ports": [
    "6000-6001"
  ],
  "exclude": [
  ],
  "broker":12345,
  "address":"0.0.0.0",
  "dev": true
}' | sudo tee /etc/iofog-connector/iofog-connector.conf
	sudo service iofog-connector start
}

# main
version="$1"
token="$2"
deploy_controller
deploy_connector