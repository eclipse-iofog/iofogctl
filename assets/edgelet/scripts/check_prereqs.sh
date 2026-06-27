#!/bin/sh
# Remote: passwordless sudo required. Local: set LOCAL_INSTALL=1 to skip sudo probe.
set -e
set -x

if [ "${LOCAL_INSTALL:-0}" = "1" ]; then
	echo "# Local install: skipping passwordless sudo check"
	exit 0
fi

if ! sudo ls /tmp/ > /dev/null 2>&1; then
	MSG="Unable to successfully use sudo with user $USER on this host.\nUser $USER must be in sudoers group and using sudo without password must be enabled."
	echo "$MSG"
	exit 1
fi
