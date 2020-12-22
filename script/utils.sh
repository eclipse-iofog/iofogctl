#!/usr/bin/env bash
#
# utils.sh - a simple set of functions for use throughout our shell scripts
#
# Usage : source utils.sh
#

# Export the name of the script for later use
THIS_SCRIPT="$(basename "${0}")"
export THIS_SCRIPT

#
# Check for DEBUG - looks for the DEBUG environment variable. If it
# is found, it enables more debugging output.
checkForDebug() {

	# Allow anyone to set a DEBUG environment variable for extra output
	if env | grep -q "DEBUG"; then
		echoNotify " :: DEBUG is set                                  ::"
		set -x
	fi
}

#
# Display a nice title line for any output. You can optionally populate it with a string
#
# Usage: prettyTitle "Bootstrapping ioFog"
#
prettyTitle() {
	echoInfo "## $1 ####################################################"
}

#
# Display a nice header for any command line script. You can optionally populate it with a string
#
# Usage: prettyHeader "Bootstrapping ioFog"
#
prettyHeader() {
	echoInfo "## $1 ####################################################"
	echoInfo "## Copyright (C) 2020, Edgeworx, Inc."
	echo
}

#
# Extract and display our script options in a pretty manner
#
# Usage: prettyUsage ${THIS_SCRIPT}
#
prettyUsage() {

	grep -e '--.*")' <"${1}" | grep -v 'grep -e' | tr -d '"' | tr ')' '\t'
}

#
# Check for Installation will check to see whether a particular command exists in
# the $PATH of the current shell. Optionally, you can check for a specific version.
#
# Usage: checkForInstallation protoc "libprotoc 3.6.1"
#
checkForInstallation() {

	# Does the command exist?
	if [[ ! "$(command -v "$1")" ]]; then
		echoError " [!] $1 not found"
		return 1
	else
		# Are we looking for a specific version?
		if [[ ! -z "$2" ]]; then
			if [[ "$2" != "$($1 --version)" ]]; then
				echoError " !! $1 is the wrong version. Found $($1 --version) but expected $2"
				return 1
			fi
		fi
		echoSuccess " [x] $1 $2 found at $(command -v "$1")"
		return 0
	fi
}

#
# Check OS Platform attempts to determine which platform we're running on currently.
# It will export the results into two env variables that can be used in your scripts:
#   $OS_ID - the name of the host os. E.g. ubuntu
#   $D_NUM - the version number of the os E.g. 16
#
checkOSPlatform() {
	if [[ "$(uname -s)" = "Darwin" ]]; then
		ID=macos
		D_NUM="$(sysctl kern.osproductversion | awk '{print $2}' | awk -F '.' '{print $1}')"
	else
		. /etc/os-release
		D_NUM="$(echo ${VERSION_ID} | awk -F '.' '{print $1}')"

		# Push us into privileged if we aren't already
		if [[ ! "$(id -u)" = "0" ]] && [[ ! "$1" = "--help" ]]; then
			echoNotify " If you are a new developer please run this script with --help"
			exec sudo -E "PATH=$PATH" sh "${THIS_SCRIPT}" "$@"
		fi
	fi

	export ID D_NUM
}

#
# The following are a bunch or pretty printing echo methods
#

# This is the list of colors used in our messages
NO_FORMAT="\\033[0m"
C_SKYBLUE1="\\033[38;5;117m"
C_DEEPSKYBLUE4="\\033[48;5;25m"
RED="\\033[38;5;1m"
GREEN="\\033[38;5;28m"

# Need this as bash and sh require different args for the echo command
if [ "${BASH_VERSION}" ]; then
    PRINTARGS="-e"
fi

# Basic subtle output
echoInfo() {
	echo ${PRINTARGS} "${C_SKYBLUE1}$1 ${NO_FORMAT}"
}

# Highlighted output with a background
echoNotify() {
	echo ${PRINTARGS} "${C_DEEPSKYBLUE4}${1} ${NO_FORMAT}"
}

# Hurrah!
echoSuccess() {
	echo ${PRINTARGS} "${GREEN}$1 ${NO_FORMAT}"
}

# Houston, we have a problem!
echoError() {
	echo ${PRINTARGS} "${RED}$1 ${NO_FORMAT}"
}

# Are we in debug mode?
checkForDebug
