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
  -h, --help          help for agent
      --host string   Hostname of remote host
      --key string    Path to private SSH key
      --port int      Port number that iofogctl uses to SSH into remote hosts (default 22)
      --user string   Username of remote host
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


