## potctl delete volume

Delete an Volume

### Synopsis

Delete an Volume.

The Volume will be deleted from the Agents that it is stored on.

```
potctl delete volume NAME [flags]
```

### Examples

```
potctl delete volume NAME
```

### Options

```
  -h, --help   help for volume
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


