#!/bin/sh
set -x
set -e

INSTALL_DIR="/opt/iofog"
TMP_DIR="/tmp/iofog"

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
		sudo iofog-controller stop || true
	fi
	# nvm
	load_existing_nvm
	if [ -z "$(command -v nvm)" ]; then
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash
		export NVM_DIR="${HOME}/.nvm"
		[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
		nvm install lts/*
		sudo ln -s $(which node) /usr/local/bin/node
	else
		nvm use lts/* || true
	fi
	
	# Set up repo
	if [ ! -z $token ]; then
		install_iofog_controller_snapshot
	fi

	# Install in temporary location
	mkdir -p "$TMP_DIR/controller"
	if [ -z $version ]; then
		npm install -g -f iofogcontroller --unsafe-perm --prefix "$TMP_DIR/controller"
	else
		npm install -g -f "iofogcontroller@$version" --unsafe-perm --prefix "$TMP_DIR/controller"
	fi

	# Move files into $INSTALL_DIR/controller
	sudo mkdir -p "$INSTALL_DIR/"
	sudo rm -rf "$INSTALL_DIR/controller" # Clean possible previous install
	sudo mv "$TMP_DIR/controller/" "$INSTALL_DIR/"

	# Symbolic links
	if [ ! -f "/usr/local/bin/iofog-controller" ]; then
		sudo ln -fFs "$INSTALL_DIR/controller/bin/iofog-controller" /usr/local/bin/iofog-controller
	fi

	# Run controller
	sudo iofog-controller start
}

config_connector() {
	# Move binaries into $INSTALL_DIR/connector
	CONNECTOR_DIR="$INSTALL_DIR/connector"
	sudo mkdir -p "$CONNECTOR_DIR"
	if [ -f "/usr/bin/iofog-connector" ]; then # Package installed properly
		sudo rm -rf "$CONNECTOR_DIR/*" # Clean possible previous install
		sudo mv /usr/bin/iofog-connector* "$CONNECTOR_DIR/"
		sudo chmod 0775 "$CONNECTOR_DIR/iofog-connector"
	fi

	# Symbolic links
	if [ ! -f "/usr/local/bin/iofog-connector" ]; then
		sudo ln -fFs "$CONNECTOR_DIR/iofog-connector" /usr/local/bin/iofog-connector
		# Connector is hard coded to look into /usr/bin for .jar and .jard
		sudo ln -fFs "$CONNECTOR_DIR/iofog-connectord.jar" /usr/bin/iofog-connectord.jar
		sudo ln -fFs "$CONNECTOR_DIR/iofog-connector.jar" /usr/bin/iofog-connector.jar
	fi

	echo '{
		"ports": [
			"6000-6050"
		],
		"exclude": [
		],
		"broker":12345,
		"address":"0.0.0.0",
		"dev": true
	}' | sudo tee /etc/iofog-connector/iofog-connector.conf

	sudo chmod 0775 /etc/iofog-connector
}

deploy_connector() {
	# Install if does not exist 
	if [ ! -z $(command -v apt-get) ]; then
		# Debian/Ubuntu
		if [ ! -z $(command -v iofog-connector) ]; then
			# Stop existing deployments
			sudo service iofog-connector stop || true
		fi
		sudo apt-get -y update
		sudo apt-get -y install openjdk-8-jre
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.deb.sh | sudo bash
		sudo apt-get -y install iofog-connector
		config_connector
		sudo service iofog-connector start
	elif [ ! -z $(command -v yum) ]; then
		# Redhat/Fedora/CentOS
		if [ ! -z $(command -v iofog-connector) ]; then
			# Stop existing deployments
			sudo systemctl stop iofog-connector || true
		fi
		su -c "yum update -y"
		su -c "yum install java-1.8.0-openjdk -y"
		curl -s https://packagecloud.io/install/repositories/iofog/iofog-connector/script.rpm.sh | sudo bash
		sudo yum install iofog-connector -y
		config_connector
		sudo systemctl start iofog-connector
	else
		echo "Could not detect system package manager"
		exit 1
	fi
}

# main
version="$1"
token="$2"
deploy_controller
deploy_connector