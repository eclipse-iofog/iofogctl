#!/bin/sh
set -x
set -e

INSTALL_DIR="/opt/iofog"
TMP_DIR="/tmp/iofog"

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

get_distro() {
	distro=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'id' | cut -d ':' -f 2 | tr -d '[:space:]')
	distro_version=$(lsb_release -a 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'codename' | cut -d ':' -f 2 | tr -d '[:space:]')
	echo "$distro"
	echo "$distro_version"
}

deploy_connector() {
	# Install if does not exist 
	if [ ! -z $(command -v apt-get) ]; then
		# Debian/Ubuntu
		if [ ! -z $(command -v iofog-connector) ]; then
			# Stop existing deployments
			sudo service iofog-connector stop || true
		fi
		if [ "$distro" = "ubuntu" ] && [ "$distro_version" = "xenial" ]; then
			sudo add-apt-repository ppa:openjdk-r/ppa
		fi
		sudo apt-get -y update
		sudo apt-get -y install openjdk-11-jre
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
get_distro
deploy_connector