## potctl delete rolebinding

Delete a RoleBinding

### Synopsis

Delete a RoleBinding from the Controller.

```
potctl delete rolebinding NAME [flags]
```

### Examples

```
potctl delete rolebinding NAME
```

### Options

```
  -h, --help   help for rolebinding
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


