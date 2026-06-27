## potctl attach volume-mount

Attach a Volume Mount to existing Agents

### Synopsis

Attach a Volume Mount to existing Agents.

```
potctl attach volume-mount NAME AGENT_NAME1 AGENT_NAME2 [flags]
```

### Examples

```
potctl attach volume-mount NAME AGENT_NAME1 AGENT_NAME2
```

### Options

```
  -h, --help   help for volume-mount
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl attach](potctl_attach.md)	 - Attach one ioFog resource to another


