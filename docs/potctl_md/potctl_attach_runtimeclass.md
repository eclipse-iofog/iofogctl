## potctl attach runtimeclass

Attach a RuntimeClass to existing Agents

### Synopsis

Attach a RuntimeClass to existing Agents. RuntimeClass can only be linked to edgelet agents.

```
potctl attach runtimeclass NAME AGENT_NAME1 AGENT_NAME2 [flags]
```

### Examples

```
potctl attach runtimeclass NAME AGENT_NAME1 AGENT_NAME2
```

### Options

```
  -h, --help   help for runtimeclass
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl attach](potctl_attach.md)	 - Attach one ioFog resource to another


