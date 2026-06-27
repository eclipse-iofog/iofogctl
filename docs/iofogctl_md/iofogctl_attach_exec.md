## iofogctl attach exec

Provision fog debug exec on an Agent

### Synopsis

Provision fog debug exec resources. Use exec agent to open an interactive shell after provisioning.

### Examples

```
iofogctl attach exec agent AgentName
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl attach](iofogctl_attach.md)	 - Attach one ioFog resource to another
* [iofogctl attach exec agent](iofogctl_attach_exec_agent.md)	 - Provision a fog debug exec microservice on an Agent


