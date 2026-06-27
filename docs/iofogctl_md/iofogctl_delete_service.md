## iofogctl delete service

Delete a Service

### Synopsis

Delete a Service from the Controller.

```
iofogctl delete service NAME [flags]
```

### Examples

```
iofogctl delete service NAME
```

### Options

```
  -h, --help   help for service
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --delete-namespace   Also delete the Kubernetes namespace (never deletes "default")
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl delete](iofogctl_delete.md)	 - Delete an existing ioFog resource


