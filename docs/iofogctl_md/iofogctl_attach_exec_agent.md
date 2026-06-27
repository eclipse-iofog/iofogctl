## iofogctl attach exec agent

Provision a fog debug exec microservice on an Agent

### Synopsis

Provision a debug microservice on an Agent for interactive exec via POST /iofog/{uuid}/exec.

```
iofogctl attach exec agent NAME [DEBUG_IMAGE] [flags]
```

### Examples

```
iofogctl attach exec agent AgentName DebugImage
```

### Options

```
  -h, --help   help for agent
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl attach exec](iofogctl_attach_exec.md)	 - Provision fog debug exec on an Agent


