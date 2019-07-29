#!/bin/sh

######## These variables MUST be updated by the user / automation task

# Space-separated list of user@host
export AGENT_LIST="user@host user2@host2"

# Single user@host
export VANILLA_CONTROLLER="user@host"

# Token to access develop versions of Controller
export PACKAGE_CLOUD_TOKEN=""

######################################################################


######## These variables can be left with their defaults if necessary

# Specify a non-existent ephemeral namespace for testing purposes
export NAMESPACE="testing"

# Kubernetes configuration file required to use kubectl
export KUBE_CONFIG="~/.kube/config"

# SSH private key that can be used to log into agents specified by AGENTS variable
export KEY_FILE="~/.ssh/id_rsa"

# Images of ioFog services deployed on Kubernetes cluster
export CONTROLLER_IMAGE="gcr.io/focal-freedom-236620/controller:latest"
export CONNECTOR_IMAGE="gcr.io/focal-freedom-236620/connector:latest"
#export SCHEDULER_IMAGE="gcr.io/focal-freedom-236620/scheduler:develop"
export OPERATOR_IMAGE="gcr.io/focal-freedom-236620/operator:develop"
export KUBELET_IMAGE="gcr.io/focal-freedom-236620/kubelet:develop"

# Controller version for vanilla deploys
export VANILLA_VERSION="1.2.2-b2538"

######################################################################

echo ""
echo "----- CONFIG -----"
echo ""
echo "${!AGENT_LIST*}: " "$AGENT_LIST"
echo "${!PACKAGE_CLOUD_TOKEN*}: " "$PACKAGE_CLOUD_TOKEN"
echo "${!NAMESPACE*}: " "$NAMESPACE"
echo "${!KUBE_CONFIG*}: " "$KUBE_CONFIG"
echo "${!KEY_FILE*}: " "$KEY_FILE"
echo "${!CONTROLLER_IMAGE*}: " "$CONTROLLER_IMAGE"
echo "${!CONNECTOR_IMAGE*}: " "$CONNECTOR_IMAGE"
#echo "${!SCHEDULER_IMAGE*}: " "$SCHEDULER_IMAGE"
echo "${!OPERATOR_IMAGE*}: " "$OPERATOR_IMAGE"
echo "${!KUBELET_IMAGE*}: " "$KUBELET_IMAGE"
echo "${!VANILLA_VERSION*}: " "$VANILLA_VERSION"
echo ""
echo "------------------"
echo ""