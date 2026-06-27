## potctl prune agent

Remove all dangling images from Agent

### Synopsis

Remove all the images which are not used by existing containers on the specified Agent

```
potctl prune agent NAME [flags]
```

### Examples

```
potctl prune agent NAME
```

### Options

```
      --detached   Specify command is to run against detached resources
  -h, --help       help for agent
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl prune](potctl_prune.md)	 - prune ioFog resources


