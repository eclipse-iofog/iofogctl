#!/bin/sh

######## These variables MUST be updated by the user / automation task

# Space-separated list of user@host
export AGENT_LIST="user@host user2@host2"

# Single user@host
export VANILLA_CONTROLLER="user@host"

######################################################################
export VANILLA_VERSION="latest"

######## These variables can be left with their defaults if necessary

# Specify a non-existent ephemeral namespace for testing purposes
export NAMESPACE="testing"

# Kubernetes configuration file required to use kubectl
export KUBE_CONFIG="~/.kube/config"
# Kubernetes configuration file used by tests
export TEST_KUBE_CONFIG="~/.kube/config"

# SSH private key that can be used to log into agents specified by AGENTS variable
export KEY_FILE="~/.ssh/id_rsa"

# Images of ioFog services deployed on Kubernetes cluster or local deploy
export CONTROLLER_IMAGE="gcr.io/focal-freedom-236620/controller:2.0.0-beta2"
export AGENT_IMAGE="gcr.io/focal-freedom-236620/agent:2.0.0-beta2"
export PORT_MANAGER_IMAGE="gcr.io/focal-freedom-236620/port-manager:2.0.0-beta2"
export ROUTER_IMAGE="iofog/router:2.0.0-beta"
export PROXY_IMAGE="iofog/proxy:2.0.0-beta"
export OPERATOR_IMAGE="gcr.io/focal-freedom-236620/operator:2.0.0-beta2"
export KUBELET_IMAGE="gcr.io/focal-freedom-236620/kubelet:2.0.0-beta"

# Controller version for vanilla deploys
export CONTROLLER_VANILLA_VERSION="2.0.0-beta"
export CONTROLLER_REPO="iofog/iofog-controller-snapshots"
# Token to access develop versions of Controller
export CONTROLLER_PACKAGE_CLOUD_TOKEN=""

# Agent version for vanilla deploys
export AGENT_VANILLA_VERSION="2.0.0-beta2"
export AGENT_REPO="iofog/iofog-agent-snapshots"
# Token to access develop versions of Agent
export AGENT_PACKAGE_CLOUD_TOKEN=""
######################################################################

######## These are necessary for HA tests

# Database
export DB_PROVIDER=""
export DB_USER=""
export DB_HOST=""
export DB_PORT=""
export DB_PW=""
export DB_NAME=""

######################################################################

echo ""
echo "----- CONFIG -----"
echo ""
echo "${!AGENT_LIST*}: " "$AGENT_LIST"
echo "${!VANILLA_CONTROLLER*}: " "$VANILLA_CONTROLLER"
echo "${!CONTROLLER_PACKAGE_CLOUD_TOKEN*}: " "$CONTROLLER_PACKAGE_CLOUD_TOKEN"
echo "${!CONTROLLER_VANILLA_VERSION*}: " "$CONTROLLER_VANILLA_VERSION"
echo "${!CONTROLLER_REPO*}: " "$CONTROLLER_REPO"
echo "${!AGENT_VANILLA_VERSION*}: " "$AGENT_VANILLA_VERSION"
echo "${!AGENT_REPO*}: " "$AGENT_REPO"
echo "${!AGENT_PACKAGE_CLOUD_TOKEN*}: " "$AGENT_PACKAGE_CLOUD_TOKEN"
echo "${!NAMESPACE*}: " "$NAMESPACE"
echo "${!KUBE_CONFIG*}: " "$KUBE_CONFIG"
echo "${!TEST_KUBE_CONFIG*}: " "$TEST_KUBE_CONFIG"
echo "${!KEY_FILE*}: " "$KEY_FILE"
echo "${!CONTROLLER_IMAGE*}: " "$CONTROLLER_IMAGE"
echo "${!PORT_MANAGER_IMAGE*}: " "$PORT_MANAGER_IMAGE"
echo "${!ROUTER_IMAGE*}: " "$ROUTER_IMAGE"
echo "${!PROXY_IMAGE*}: " "$PROXY_IMAGE"
echo "${!AGENT_IMAGE*}: " "$AGENT_IMAGE"
echo "${!OPERATOR_IMAGE*}: " "$OPERATOR_IMAGE"
echo "${!KUBELET_IMAGE*}: " "$KUBELET_IMAGE"
echo "${!VANILLA_VERSION*}: " "$VANILLA_VERSION"
echo "${!DB_PROVIDER*}: " "$DB_PROVIDER"
echo "${!DB_USER*}: " "$DB_USER"
echo "${!DB_HOST*}: " "$DB_HOST"
echo "${!DB_PORT*}: " "$DB_PORT"
echo "${!DB_PW*}: " "$DB_PW"
echo "${!DB_NAME*}: " "$DB_NAME"
echo ""
echo "------------------"
echo ""
