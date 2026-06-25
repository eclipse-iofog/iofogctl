## potctl delete namespace

Delete a Namespace

### Synopsis

Delete a Namespace.

The Namespace must be empty.

If you would like to delete all resources in the Namespace, use the --force flag.

```
potctl delete namespace NAME [flags]
```

### Examples

```
potctl delete namespace NAME
```

### Options

```
      --force   Force deletion of all resources within the Namespace
  -h, --help    help for namespace
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


