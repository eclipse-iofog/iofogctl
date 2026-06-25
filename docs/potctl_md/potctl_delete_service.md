## potctl delete service

Delete a Service

### Synopsis

Delete a Service from the Controller.

```
potctl delete service NAME [flags]
```

### Examples

```
potctl delete service NAME
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
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl delete](potctl_delete.md)	 - Delete an existing ioFog resource


