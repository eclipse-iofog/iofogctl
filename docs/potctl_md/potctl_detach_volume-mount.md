## potctl detach volume-mount

Detach a Volume Mount from existing Agents

### Synopsis

Detach a Volume Mount from existing Agents.

```
potctl detach volume-mount NAME AGENT_NAME1 AGENT_NAME2 [flags]
```

### Examples

```
potctl detach volume-mount NAME AGENT_NAME1 AGENT_NAME2
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

* [potctl detach](potctl_detach.md)	 - Detach one ioFog resource from another


