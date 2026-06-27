## potctl move agent

Move an Agent to another Namespace

### Synopsis

Move an Agent to another Namespace

```
potctl move agent NAME DEST_NAMESPACE [flags]
```

### Examples

```
potctl move agent NAME DEST_NAMESPACE
```

### Options

```
      --force   Move Agent even if it is running Microservices
  -h, --help    help for agent
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl move](potctl_move.md)	 - Move an existing resources inside the current Namespace


