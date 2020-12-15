#!/bin/sh
set -x

# Check can sudo without password
if ! $(sudo ls /tmp/ > /dev/null); then
	MSG="Unable to successfully use sudo with user $USER on this host.\nUser $USER must be in sudoers group and using sudo without password must be enabled.\nPlease see iofog.org documentation for more details."
	echo $MSG
	exit 1
fi
