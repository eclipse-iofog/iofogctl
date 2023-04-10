#!/bin/sh
set -x
set -e

INSTALL_DIR="/opt/iofog"
TMP_DIR="/tmp/iofog"
ETC_DIR="/etc/iofog/controller"

controller_service() {
    USE_SYSTEMD=`grep -m1 -c systemd /proc/1/comm`
    USE_INITCTL=`which initctl | wc -l`
    USE_SERVICE=`which service | wc -l`

    if [ $USE_SYSTEMD -eq 1 ]; then
        cp "$ETC_DIR/service/iofog-controller.systemd" /etc/systemd/system/iofog-controller.service
        chmod 644 /etc/systemd/system/iofog-controller.service
        systemctl daemon-reload
        systemctl enable iofog-controller.service
    elif [ $USE_INITCTL -eq 1 ]; then
        cp "$ETC_DIR/service/iofog-controller.initctl" /etc/init/iofog-controller.conf
        initctl reload-configuration
    elif [ $USE_SERVICE -eq 1 ]; then
        cp "$ETC_DIR/service/iofog-controller.update-rc" /etc/init.d/iofog-controller
        chmod +x /etc/init.d/iofog-controller
        update-rc.d iofog-controller defaults
    else
        echo "Unable to setup Controller startup script."
    fi
}

install_package() {
		if [ -z "$(command -v apt)" ]; then
			echo "Unsupported distro"
			exit 1
		fi
		apt update -qq
		apt install -y $1
}

install_deps() {
	if [ -z "$(command -v curl)" ]; then
        install_package "curl"
	fi

	if [ -z "$(command -v lsof)" ]; then
        install_package "lsof"
	fi

	if [ -z "$(command -v make)" ]; then
        install_package "build-essential"
	fi

	if [ -z "$(command -v python2)" ]; then
        install_package "python"
	fi
}

deploy_controller() {
	# Nuke any existing instances
	if [ ! -z "$(lsof -ti tcp:51121)" ]; then
		lsof -ti tcp:51121 | xargs kill
	fi

#	 If token is provided, set up private repo
	if [ ! -z $token ]; then
		if [ ! -z $(npmrc | grep iofog) ]; then
			npmrc -c iofog
			npmrc iofog
		fi
		curl -s https://"$token":@packagecloud.io/install/repositories/"$repo"/script.node.sh?package_id=7463817 | force_npm=1 bash
		mv ~/.npmrc ~/.npmrcs/npmrc
		ln -s ~/.npmrcs/npmrc ~/.npmrc
	else
		npmrc default
	fi
	# Save DB
	if [ -f "$INSTALL_DIR/controller/lib/node_modules/@iofog/iofogcontroller/package.json" ]; then
		# If iofog-controller is not running, it will fail to stop - ignore that failure.
		node $INSTALL_DIR/controller/lib/node_modules/@iofog/iofogcontroller/scripts/scripts-api.js preuninstall > /dev/null 2>&1 || true
	fi

	# Install in temporary location
	mkdir -p "$TMP_DIR/controller"
	chmod 0777 "$TMP_DIR/controller"
	if [ -z $version ]; then
		npm install -g -f @iofog/iofogcontroller --unsafe-perm --prefix "$TMP_DIR/controller"
	else
		npm install -g -f "@iofog/iofogcontroller@$version" --unsafe-perm --prefix "$TMP_DIR/controller"
	fi
	# Move files into $INSTALL_DIR/controller
	mkdir -p "$INSTALL_DIR/"
	rm -rf "$INSTALL_DIR/controller" # Clean possible previous install
	mv "$TMP_DIR/controller/" "$INSTALL_DIR/"

	# Restore DB
	if [ -f "$INSTALL_DIR/controller/lib/node_modules/@iofog/iofogcontroller/package.json" ]; then
		node $INSTALL_DIR/controller/lib/node_modules/@iofog/iofogcontroller/scripts/scripts-api.js postinstall > /dev/null 2>&1 || true
	fi

	# Symbolic links
	if [ ! -f "/usr/local/bin/iofog-controller" ]; then
		ln -fFs "$INSTALL_DIR/controller/bin/iofog-controller" /usr/local/bin/iofog-controller
	fi

	# Set controller permissions
	chmod 744 -R "$INSTALL_DIR/controller"

	# Startup script
	controller_service

	# Run controller
	. /opt/iofog/config/controller/env.sh
	iofog-controller start
}

# main
version="$1"
repo=$([ -z "$2" ] && echo "iofog/iofog-controller-snapshots" || echo "$2")
token="$3"

install_deps
deploy_controller