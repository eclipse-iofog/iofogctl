#!/bin/bash

set -e

OS=$(uname -s | tr A-Z a-z)
HELM_VERSION=2.13.1
K8S_VERSION=1.13.4

# Go Dep
if [[ -z $(command -v dep) ]]; then
    go get -u github.com/golang/dep/cmd/dep
fi