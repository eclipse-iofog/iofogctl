## potctl delete configmap

Delete a ConfigMap

### Synopsis

Delete a ConfigMap from the Controller.

```
potctl delete configmap NAME [flags]
```

### Examples

```
potctl delete configmap NAME
```

### Options

```
  -h, --help   help for configmap
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


