## iofogctl detach exec agent

Remove fog debug exec from an Agent

### Synopsis

Remove the debug microservice provisioned for Agent exec via DELETE /iofog/{uuid}/exec.

```
iofogctl detach exec agent NAME [flags]
```

### Examples

```
iofogctl detach exec agent AgentName
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

* [iofogctl detach exec](iofogctl_detach_exec.md)	 - Remove fog debug exec from an Agent


