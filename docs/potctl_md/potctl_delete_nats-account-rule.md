## potctl delete nats-account-rule

Delete a NATS account rule

### Synopsis

Delete a NATS account rule from the Controller.

```
potctl delete nats-account-rule NAME [flags]
```

### Examples

```
potctl delete nats-account-rule NAME
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
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl delete](potctl_delete.md)	 - Delete an existing ioFog resource


