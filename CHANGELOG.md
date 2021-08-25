# Changelog

## [unreleased]

* Remove Microservice output from `get all` command
* Fix Debian Stretch Docker installation

## [v3.0.0-alpha] - 11 March 2021

* Add Application Template support
* Add Agent `upgrade` and `rollback` commands
* Add custom installation plugin support for Agent deploy
* Add Docker pull percentage to `iofogctl get microservices` output
* Parallelize `iofogctl get all` command to hide latency
* Add `config` field to `Agent` kind to allow custom configuration on Agent deploy
* Remove Agent configuration field from Microservice kind
* Fix bug showing `--detached` as flag for all commands
* Fix bug preventing deployment of Apps/Microservices in same YAML file as Control Plane
* Improve error output of SSH operations during deploys
* Update K8s deploy to ignore errors from irrelevant Namespaces given ioFog can be deployed w/ cluster-scope
* Update `pkg/util/ssh.go` to read keys before dialling connection

## [v2.0.1] - 2020-09-10

* Update default versions of Agent, Router, and Proxy to 2.0.1

## [v2.0.0] - 2020-08-06

### Features

* Add `iofogctl move AGENT NAMESPACE` command
* Show current Namespace with asterisk in `iofogctl get namespaces` output
* Rename `default-namespace` to `current-namespace` for `iofogctl configure` command
* Improve error handling when deploying K8s Control Plane. Operator failures are detected and reported
* Print Controller version on `iofogctl get controllers`

### Bugs

* Fix K8s Control Plane deployment to be idempotent
* Fix volume deploy not working when src dir is empty
* Fix parallel processes running iofogctl trampling over `~/.iofog/v2/config.yaml`
* Fix SSH error output
* Fix success output when detaching Agents
* Fix renaming current Namespace
* No longer reinstall Agent and its deps on remote host during Agent deployment


## [v2.0.0-rc1] - 2020-04-30

### Features

* Check agent name clash during attach command
* Update delete using file to ignore not found errors
* Update catalog item update request to update the registry

### Bugs

* Fix base64 password logic in connect
* Fix Volume to local Windows

## [v2.0.0-beta5] - 2020-04-23

### Features

* Update pipeline image defaults and add proxy and router variables
* Add force option to agent delete
* Update Application Status
* Increase timeout waiting for local deploy containers
* Check Router address on k8s deploy
* Add more retry conditions for k8s deploy
* Update K8s deploy to wait for Default Router
* Update K8s control plane to return real pod name and status
* Use separate dir for v2 config and namespaces and remove conversions
* Encode passwords and add --generate flag to connect command
* Re-order volume deployment

### Bugs

* Update operator module with Application CR fix
* Fix attach command for external Agents
* Regenerate Agent cache every time we run Agent commands
* Fix disconnect command when namespace does not exist

## [v2.0.0-beta4] - 2020-04-09

### Features

* Return error if metadata namespace and flag namespace dont match
* Remove Remote from ControlPlane and Controller kinds
* Error if nothing to execute from YAML file
* Improve url parsing when connecting to an ECN
* Increase timeout when installing Agent

### Bugs

* Fix local agent config

## [v2.0.0-beta3] - 2020-04-08

### Features

* Add ecn flag to version command
* Update header to version output
* Groom command help output
* Update flag checking in connect command
* Remove --name from kube connect command
* Remove config from msvc get
* Update Agent deploy for system Agent deploy to be idempotent
* Remove unnecessary volumes field from agent config type
* Update configure command arguments
* Allow deployment of local agent with remote Controller

### Bugs

* Fix getAddressAndPort for get controllers
* Fix failing to delete unprovisioned agents bug
* Modify local deploy output and fix disconnect on default namespace
* Fix configure k8s controlPlane
* Fix empty image names for operator
* Fix get all output
* Flush on namespace conversion and dont return _detached on GetNamespaces
* Change iofog client timeout config
* Fix get all output and add namespace to msvc/app get output

## [v2.0.0-beta2] - 2020-03-17

### Features

* Make local agent container network host, and force config to be standalone interior router
* Add delete, describe, and get functionality for volumes
* Update CRD versioning support and update logic
* Stop deleting CRDs and update CRD on deploy
* Force update CRDs on deploy
* Add agentVersion and controllerVersion link-time variables for vanilla deploys
* Set router image in deploy k8s
* Allow RouterImage to be configured for K8s deploy
* Allow update agent host using configure

### Bugs

* Fix iofogctl view

## [v2.0.0-beta] - 2020-03-12

### Features

* Add KubernetesControlPlane, RemoteControlPlane, and LocalControlPlane kinds
* Add RemoteController and LocalController kinds
* Add support for new Public Ports and Routers
* Remove Connector kind and associated procedures
* Add attach and detach Agent commands
* Add prune command
* Add configure default namespace feature
* Add Volume kind and deployment procedures
* Add move Microservices command
* Add rename Microservices command
* Update delete agent command to deprovision before deleting agent from controller

## [v1.3.0]

* Add client package to the repo
* Re-organize the repo to maintain multiple packages
  
[Unreleased]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta3..HEAD
[v2.0.0-rc1]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta4..v2.0.0-beta5
[v2.0.0-beta5]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta4..v2.0.0-beta5
[v2.0.0-beta4]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta3..v2.0.0-beta4
[v2.0.0-beta3]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta2..v2.0.0-beta3
[v2.0.0-beta2]: https://github.com/eclipse-iofog/iofogctl/compare/v2.0.0-beta..v2.0.0-beta2
[v2.0.0-beta]: https://github.com/eclipse-iofog/iofogctl/tree/v2.0.0-beta
