## iofogctl delete certificate

Delete a Certificate

### Synopsis

Delete a Certificate from the Controller.

```
iofogctl delete certificate NAME [flags]
```

### Examples

```
iofogctl delete certificate NAME
```

### Options

```
  -h, --help   help for certificate
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


