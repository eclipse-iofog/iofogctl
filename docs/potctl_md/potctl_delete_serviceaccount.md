## potctl delete serviceaccount

Delete a ServiceAccount

### Synopsis

Delete a ServiceAccount from the Controller. ServiceAccounts are application-scoped; use APPLICATION_NAME/SERVICE_ACCOUNT_NAME (e.g. myapp/my-sa).

```
potctl delete serviceaccount APPLICATION_NAME/SERVICE_ACCOUNT_NAME [flags]
```

### Examples

```
potctl delete serviceaccount myapp/my-sa
```

### Options

```
  -h, --help   help for serviceaccount
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


