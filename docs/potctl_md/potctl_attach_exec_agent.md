## potctl attach exec agent

Attach an Exec Session to an Agent

### Synopsis

Attach an Exec Session to an existing Agent.

```
potctl attach exec agent NAME [DEBUG_IMAGE] [flags]
```

### Examples

```
potctl attach exec agent AgentName DebugImage
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

* [potctl attach exec](potctl_attach_exec.md)	 - Attach an Exec Session to a resource


