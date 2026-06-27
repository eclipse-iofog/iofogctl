## potctl exec agent

Open an interactive exec session on an Agent debug shell

### Synopsis

Open a WebSocket exec session to the Agent debug microservice. Provisions fog debug exec automatically when it is not already enabled.

```
potctl exec agent AgentName [DEBUG_IMAGE] [flags]
```

### Examples

```
potctl exec agent AgentName
potctl exec agent AgentName ghcr.io/org/debug:latest
```

### Options

```
  -h, --help   help for agent
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl exec](potctl_exec.md)	 - Connect to an Exec Session of a resource


