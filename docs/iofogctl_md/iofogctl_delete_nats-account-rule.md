## iofogctl delete nats-account-rule

Delete a NATS account rule

### Synopsis

Delete a NATS account rule from the Controller.

```
iofogctl delete nats-account-rule NAME [flags]
```

### Examples

```
iofogctl delete nats-account-rule NAME
```

### Options

```
  -h, --help   help for nats-account-rule
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


