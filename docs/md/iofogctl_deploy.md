## iofogctl deploy

Deploy ioFog platform or components on existing infrastructure

### Synopsis

Deploy ioFog platform or individual components on existing infrastructure.

The YAML resource specification file should look like this (two Controllers specified for example only):
```
kind: ControlPlane
apiVersion: iofog.org/v2
metadata:
  name: alpaca-1 # ControlPlane name
spec:
  kube: # K8s
	config: ~/.kube/config # Will deploy a controller in a kubernetes cluster
	images:
		controller: ...
  controllers:
  - name: vanilla
    host: 35.239.157.151 # Will deploy a controller as a standalone binary
    ssh:
      user: serge # SSH user
	  keyFile: ~/.ssh/id_rsa # SSH private key
---
apiVersion: iofog.org/v2
kind: Agent
metadata:
  name: agent1 # Agent name
spec:
  host: 35.239.157.151 # SSH host
  ssh:
    user: serge # SSH User
    keyFile: ~/.ssh/id_rsa # SSH private key
---
apiVersion: iofog.org/v2
kind: Agent
metadata:
  name: agent2
spec:
  host: 35.232.114.32
  ssh:
    user: serge
    keyFile: ~/.ssh/id_rsa

```
The complete description of yaml file definition can be found at iofog.org

```
iofogctl deploy [flags]
```

### Examples

```
deploy -f platform.yaml
```

### Options

```
  -f, --file string   YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy
  -h, --help          help for deploy
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


