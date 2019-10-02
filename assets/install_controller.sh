#!/bin/sh
set -x
set -e

INSTALL_DIR="/opt/iofog"
TMP_DIR="/tmp/iofog"

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

controller_service() {
    USE_SYSTEMD=`grep -m1 -c systemd /proc/1/comm`
    USE_INITCTL=`which initctl | wc -l`
    USE_SERVICE=`which service | wc -l`

    if [ $USE_SYSTEMD -eq 1 ]; then
        sudo cp /tmp/iofog-controller-service/iofog-controller.systemd /etc/systemd/system/iofog-controller.service
        sudo chmod 644 /etc/systemd/system/iofog-controller.service
        sudo systemctl daemon-reload
        sudo systemctl enable iofog-controller.service
    elif [ $USE_INITCTL -eq 1 ]; then
        sudo cp /tmp/iofog-controller-service/iofog-controller.initctl /etc/init/iofog-controller.conf
        sudo initctl reload-configuration
    elif [ $USE_SERVICE -eq 1 ]; then
        sudo cp /tmp/iofog-controller-service/iofog-controller.update-rc /etc/init.d/iofog-controller
        sudo chmod +x /etc/init.d/iofog-controller
        update-rc.d iofog-controller defaults
    else
        echo "Unable to setup Controller startup script."
    fi
}

install_deps() {
	if [ -z "$(command -v lsof)" ]; then
		sudo apt install lsof
	fi

	if [ -z "$(command -v setcap)" ]; then
		sudo apt install libcap2-bin
	fi
}

deploy_controller() {
	# Nuke any existing instances
	if [ ! -z $(sudo lsof -ti tcp:51121) ]; then
		sudo lsof -ti tcp:51121 | xargs sudo kill
	fi

	# nvm
	load_existing_nvm
	if [ -z "$(command -v nvm)" ]; then
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.34.0/install.sh | bash
		export NVM_DIR="${HOME}/.nvm"
		[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
		nvm install lts/*
		NODE=$(which node)
		sudo ln -Ffs "$NODE" /usr/local/bin/node
	else
		nvm use lts/* || true
	fi

	# npmrc
	if [ -z "$(command -v npmrc)" ]; then
		npm i npmrc -g
	fi

	# If token is provided, set up private registry
	if [ ! -z $token ]; then
		if [ ! -z $(npmrc | grep iofog)]; then
			npmrc -c iofog
			npmrc iofog
		fi
		curl -s https://"$token":@packagecloud.io/install/repositories/iofog/iofog-controller-snapshots/script.node.sh | force_npm=1 bash
		mv ~/.npmrc ~/.npmrcs/npmrc
		ln -s ~/.npmrcs/npmrc ~/.npmrc
	else
		npmrc default
	fi

	# Install in temporary location
	sudo mkdir -p "$TMP_DIR/controller"
	sudo chmod -R 777 "$TMP_DIR/controller"
	if [ -z $version ]; then
		npm install -g -f minipass@2.7.0 iofogcontroller --unsafe-perm --prefix "$TMP_DIR/controller"
	else
		npm install -g -f minipass@2.7.0 "iofogcontroller@$version" --unsafe-perm --prefix "$TMP_DIR/controller"
	fi

	# Move files into $INSTALL_DIR/controller
	sudo mkdir -p "$INSTALL_DIR/"
	sudo rm -rf "$INSTALL_DIR/controller" # Clean possible previous install
	sudo mv "$TMP_DIR/controller/" "$INSTALL_DIR/"

	sudo mkdir -p /var/log/iofog-controller
	sudo chmod -R 777 /var/log/iofog-controller

	# Symbolic links
	if [ ! -f "/usr/local/bin/iofog-controller" ]; then
		sudo ln -fFs "$INSTALL_DIR/controller/bin/iofog-controller" /usr/local/bin/iofog-controller
        sudo chmod 777 /usr/local/bin/iofog-controller
	fi

    # Set controller permissions
    sudo chmod 777 -R "$INSTALL_DIR/controller"

    # Startup script
    controller_service

    # Allow node to listen on port 80
    sudo setcap 'cap_net_bind_service=+ep' $(which node)

	# Run controller
    iofog-controller start
}

# main
version="$1"
token="$2"
# Optional args
export DB_PROVIDER="$3"
export DB_HOST="$4"
export DB_USER="$5"
export DB_PASSWORD="$6"
export DB_PORT="$7"

install_deps
deploy_controller