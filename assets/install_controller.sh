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

controller_service() {
    USE_SYSTEMD=`grep -m1 -c systemd /proc/1/comm`
    USE_INITCTL=`which /sbin/initctl | wc -l`
    USE_SERVICE=`which /usr/sbin/service | wc -l`

    if [ $USE_SYSTEMD -eq 1 ]; then
        sudo cp ./iofog-controller-service/iofog-controller.systemd /etc/systemd/system/iofog-controller.service
        sudo chmod 644 /etc/systemd/system/iofog-controller.service
        sudo systemctl daemon-reload
        sudo systemctl enable iofog-controller.service
    elif [ $USE_INITCTL -eq 1 ]; then
        sudo cp ./iofog-controller-service/iofog-controller.initctl /etc/init/iofog-controller.conf
        sudo initctl reload-configuration
    elif [ $USE_SERVICE -eq 1 ]; then
        sudo cp ./iofog-controller-service/iofog-controller.update-rc /etc/init.d/iofog-controller
        sudo chmod +x /etc/init.d/iofog-controller
        update-rc.d iofog-controller defaults
    else
        echo "Unable to setup Controller startup script."
    fi
}

deploy_controller() {
	# Try to stop the instance if is installed
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
		sudo ln -Ffs $(which node) /usr/local/bin/node
	else
		nvm use lts/* || true
	fi
	
	# Set up repo
	if [ ! -z $token ]; then
		install_iofog_controller_snapshot
	fi

	# Install in temporary location
	sudo mkdir -p "$TMP_DIR/controller"
	sudo chmod 0777 "$TMP_DIR/controller"
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

    # Set controller permissions
    sudo chmod 744 -R "$INSTALL_DIR/controller"

    # Startup script
    controller_service

	# Run controller
	sudo iofog-controller start
}

# main
version="$1"
token="$2"
deploy_controller