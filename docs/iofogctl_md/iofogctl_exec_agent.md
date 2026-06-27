## iofogctl exec agent

Open an interactive exec session on an Agent debug shell

### Synopsis

Open a WebSocket exec session to the Agent debug microservice. Provisions fog debug exec automatically when it is not already enabled.

```
iofogctl exec agent AgentName [DEBUG_IMAGE] [flags]
```

### Examples

```
iofogctl exec agent AgentName
iofogctl exec agent AgentName ghcr.io/org/debug:latest
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

* [iofogctl exec](iofogctl_exec.md)	 - Connect to an Exec Session of a resource


