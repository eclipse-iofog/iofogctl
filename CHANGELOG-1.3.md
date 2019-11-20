## v1.3.0
* Add support for HA Controllers and Connectors
* Refactor YAML specifications for HA support
* Add more support for deploying microservices
* Display microservice deployment status
* Allow for users to deploy Connectors to dedicated remote hosts
* Allow for users to optionally deploy Connectors on K8s cluster
* Add Connector as a standalone resource for iofogctl commands
* Add NodePort service support for k8s install
* Replace -q with -v
* Integrate with ioFog K8s operator for deploying K8s Controller, Connector, and Kubelet
* Force resource names to be lowercase alphanumeric with - characters
* Add microservice logs command for microservices on remote agents
* Add --force to delete namespace command
* Refactor and improve parallelism during commands like deploy -f and delete all
* Fix vanilla Controller install bugs and make the installation idempotent
* Stop logging stdout to file during SSH sessions and dump to terminal instead
