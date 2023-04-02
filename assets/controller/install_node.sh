#!/bin/sh
set -x
set -e

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

install_node() {
	load_existing_nvm
	if [ -z "$(command -v nvm)" ]; then
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash
		export NVM_DIR="${HOME}/.nvm"
		[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
	fi
	nvm install  v18.15.0
	nvm use  v18.15.0
	ln -Ffs $(which node) /usr/local/bin/node
	ln -Ffs $(which npm) /usr/local/bin/npm

	# npmrc
	if [ -z "$(command -v npmrc)" ]; then
		npm i npmrc -g
	fi
	ln -Ffs $(which npmrc) /usr/local/bin/npmrc
}

install_node