## potctl configure auth-group

Update an embedded auth group

### Synopsis

Manage embedded auth groups (admin/sre only for create, rename, and delete).

System groups (admin, sre, developer, viewer) cannot be created or deleted; their names cannot be changed.
MfaRequired can be toggled on any group. Auth groups are unavailable when the Controller uses external OIDC.

```
potctl configure auth-group NAME [flags]
```

### Examples

```
potctl configure auth-group secops --name platform-ops -n NAMESPACE
potctl configure auth-group viewer --mfa-required=true -n NAMESPACE
```

### Options

```
  -h, --help           help for auth-group
      --mfa-required   Require TOTP for members of this group at login
      --name string    New group name (custom groups only)
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl configure](potctl_configure.md)	 - Configure potctl or ioFog resources


