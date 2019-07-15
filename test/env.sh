#!/bin/sh

######## These variables MUST be updated by the user / automation task

# Space-separated list of user@host
export AGENT_LIST="serge@35.193.16.117 serge@35.222.192.229"

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

######################################################################

echo ""
echo "----- CONFIG -----"
echo ""
echo "${!AGENT_LIST*}: " "$AGENT_LIST"
echo "${!NAMESPACE*}: " "$NAMESPACE"
echo "${!KUBE_CONFIG*}: " "$KUBE_CONFIG"
echo "${!KEY_FILE*}: " "$KEY_FILE"
echo "${!CONTROLLER_IMAGE*}: " "$CONTROLLER_IMAGE"
echo "${!CONNECTOR_IMAGE*}: " "$CONNECTOR_IMAGE"
#echo "${!SCHEDULER_IMAGE*}: " "$SCHEDULER_IMAGE"
echo "${!OPERATOR_IMAGE*}: " "$OPERATOR_IMAGE"
echo "${!KUBELET_IMAGE*}: " "$KUBELET_IMAGE"
echo ""
echo "------------------"
echo ""