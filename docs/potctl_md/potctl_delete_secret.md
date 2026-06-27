## potctl delete secret

Delete a Secret

### Synopsis

Delete a Secret from the Controller.

```
potctl delete secret NAME [flags]
```

### Examples

```
potctl delete secret NAME
```

### Options

```
  -h, --help   help for secret
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


