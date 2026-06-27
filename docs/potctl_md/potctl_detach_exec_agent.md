## potctl detach exec agent

Remove fog debug exec from an Agent

### Synopsis

Remove the debug microservice provisioned for Agent exec via DELETE /iofog/{uuid}/exec.

```
potctl detach exec agent NAME [flags]
```

### Examples

```
potctl detach exec agent AgentName
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

* [potctl detach exec](potctl_detach_exec.md)	 - Remove fog debug exec from an Agent


