## potctl move microservice

Move a Microservice to another Agent in the same Namespace

### Synopsis

Move a Microservice to another Agent in the same Namespace

```
potctl move microservice NAME AGENT_NAME [flags]
```

### Examples

```
potctl move microservice NAME AGENT_NAME
```

### Options

```
  -h, --help   help for microservice
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl move](potctl_move.md)	 - Move an existing resources inside the current Namespace


