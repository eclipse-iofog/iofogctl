## v1.2.3
* Add deploy Application feature
* Update K8s load balancers to avoid unnecessary KubeProxy hops
* Fix Agent details not updating during get Agent command
* Add automatic Doc generation
* Add BASH autocompletion
* Spin up VMs dynamically in build pipeline
* Implement multi-stage build pipeline to improve speed
* Move from bats to bats-core
* Move from TAP to JUnit for test reporting in build pipeline
* Split functional tests into K8s and Vanilla

## v1.2.2
* Implement vanilla Controller deploy command (non-k8s)
* Changes to legacy and logs commands for vanilla Controller deploys
* Update install procedures to use static assets for install scripts
* Stabilize functional tests and the build pipeline
* Refactor iofog package into install and client packages
* Refactor client package

## v1.2.1
* Add quiet mode (-q)
* Remove various shorthand flags from deploy commands
* Add functional test suite and integrate to build pipeline
* Disable scheduler deploy
* Update default images for ioFog services on Kubernetes deployment

## v1.2.0
* Update get agents command to report server-side information only
* Add spinner indicator while commands process
* Print namespace more judiciously
* Move config.yaml to ~/.iofog/
* Print notifications to stderr
* Improve SSH error reporting
* Fix missing IP and port bug
* Fix uptime and age outputs
* Fix relative path input bug
* Check for existing user before generating a new one on deploy
* Add unit tests
* Use branch name to reference install scripts
* Ignore empty image names and use defaults instead for Kubernetes deployment
