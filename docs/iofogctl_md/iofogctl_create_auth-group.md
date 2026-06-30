## iofogctl create auth-group

Create a custom embedded auth group

### Synopsis

Manage embedded auth groups (admin/sre only for create, rename, and delete).

System groups (admin, sre, developer, viewer) cannot be created or deleted; their names cannot be changed.
MfaRequired can be toggled on any group. Auth groups are unavailable when the Controller uses external OIDC.

```
iofogctl create auth-group NAME [flags]
```

### Examples

```
iofogctl create auth-group secops --mfa-required -n NAMESPACE
```

### Options

```
  -h, --help           help for auth-group
      --mfa-required   Require TOTP for members of this group at login
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl create](iofogctl_create.md)	 - Create a resource


