## potctl exec agent

Connect to an Exec Session of an Agent

### Synopsis

Connect to an Exec Session of an Agent to interact with its container.

```
potctl exec agent AgentName [flags]
```

### Examples

```
potctl exec agent AgentName
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

* [potctl exec](potctl_exec.md)	 - Connect to an Exec Session of a resource


