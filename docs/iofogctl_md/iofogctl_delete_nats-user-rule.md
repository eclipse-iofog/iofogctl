## iofogctl delete nats-user-rule

Delete a NATS user rule

### Synopsis

Delete a NATS user rule from the Controller.

```
iofogctl delete nats-user-rule NAME [flags]
```

### Examples

```
iofogctl delete nats-user-rule NAME
```

### Options

```
  -h, --help   help for nats-user-rule
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


