## potctl detach agent

Detaches an Agent

### Synopsis

Detaches an Agent.

The Agent will be deprovisioned from the Controller within the namespace.
The Agent will be removed from Controller.

You cannot detach unprovisioned Agents.

The Agent stack will not be uninstalled from the host.

```
potctl detach agent NAME [flags]
```

### Examples

```
potctl detach agent NAME
```

### Options

```
      --force   Detach Agent even if it is running Microservices
  -h, --help    help for agent
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl detach](potctl_detach.md)	 - Detach one ioFog resource from another


