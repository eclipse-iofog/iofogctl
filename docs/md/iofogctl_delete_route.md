## iofogctl delete route

Delete a Route

### Synopsis

Delete a Route.

The corresponding Microservices will no longer be able to reach each other using ioMessages.

```
iofogctl delete route NAME [flags]
```

### Examples

```
iofogctl delete route NAME
```

### Options

```
  -h, --help   help for route
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl delete](iofogctl_delete.md)	 - Delete an existing ioFog resource


