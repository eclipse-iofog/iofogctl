## iofogctl attach agent

Attach an Agent to an existing Namespace

### Synopsis

Attach a detached Agent to an existing Namespace.

The Agent will be provisioned with the Controller within the Namespace.

```
iofogctl attach agent NAME [flags]
```

### Examples

```
iofogctl attach agent NAME
```

### Options

```
  -h, --help       help for agent
      --port int   Port number that iofogctl uses to SSH into remote hosts (default 22)
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --detached           Use/Show detached resources
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl attach](iofogctl_attach.md)	 - Attach an existing ioFog resource to Control Plane


