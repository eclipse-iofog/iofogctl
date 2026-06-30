## iofogctl delete auth-group

Delete a custom embedded auth group

### Synopsis

Manage embedded auth groups (admin/sre only for create, rename, and delete).

System groups (admin, sre, developer, viewer) cannot be created or deleted; their names cannot be changed.
MfaRequired can be toggled on any group. Auth groups are unavailable when the Controller uses external OIDC.

```
iofogctl delete auth-group NAME [flags]
```

### Examples

```
iofogctl delete auth-group secops -n NAMESPACE
```

### Options

```
  -h, --help   help for auth-group
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


