## iofogctl legacy

Execute commands using legacy CLI

### Synopsis

Execute commands using legacy CLI

```
iofogctl legacy resource NAME COMMAND ARGS... [flags]
```

### Examples

```
iofogctl legacy controller NAME COMMAND
iofogctl legacy agent      NAME COMMAND
```

### Options

```
  -h, --help   help for legacy
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --detached           Use/Show detached resources
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


