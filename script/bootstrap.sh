#!/usr/bin/env bash
#
# bootstrap.sh will check for and install any dependencies we have for building and using iofogctl
#
# Usage: ./bootstrap.sh
#


set -e

# Import our helper functions
. script/utils.sh

prettyTitle "Installing iofogctl Dependencies"
echo

# Check whether Brew is installed
# TODO: Current installation method is macos centric, make it work for linux too.
#if ! checkForInstallation "brew"; then
#    echoInfo " Attempting to install Brew"
#    /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
#fi


#
# All our Go related stuff
#

# Is go installed?
if ! checkForInstallation "go"; then
    echoNotify "\nYou do not have Go installed. Please install and re-run bootstrap."
    exit 1
fi

# Is rice installed?
if [ -z $(command -v rice) ]; then
    echo " Attempting to install 'rice'"
    go install github.com/GeertJohan/go.rice/rice@latest
    if [ -z $(command -v rice) ]; then
        echo ' Could not find command rice after installation - is $GOBIN in $PATH?'
    fi
fi

# Is bats installed?
if ! checkForInstallation "bats"; then
    echoInfo " Attempting to install 'bats'"
    git clone https://github.com/bats-core/bats-core.git && cd bats-core && git checkout tags/v1.1.0 && sudo ./install.sh /usr/local
fi

# Is jq installed?
if ! checkForInstallation "jq"; then
    echoInfo " Attempting to install 'jq'"
    if [ "$(uname -s)" = "Darwin" ]; then
        brew install jq
    else
        sudo apt install jq
    fi
fi

# Is go lint installed?
if ! checkForInstallation "golangci-lint"; then
    if [ "$(uname -s)" = "Darwin" ]; then
        brew install golangci-lint
        brew upgrade golangci-lint
    else
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.33.0
    fi
fi

# CI deps
if [ ! -z "$PIPELINE" ]; then
    ## Is kubernetes-cli installed?
    if ! checkForInstallation "kubectl"; then
        OS=$(uname -s | tr A-Z a-z)
        K8S_VERSION=1.22.7
        echoInfo " Attempting to install kubernetes-cli"
        curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v"$K8S_VERSION"/bin/"$OS"/amd64/kubectl
        chmod +x kubectl
        sudo mv kubectl /usr/local/bin/
    fi
    # Is go-junit-report installed?
    if ! checkForInstallation "go-junit-report"; then
        echoInfo " Attempting to install 'go-junit-report'"
        go install github.com/jstemmer/go-junit-report@latest
    fi
    ## TODO: gcloud
fi
