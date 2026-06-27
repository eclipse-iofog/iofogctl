## potctl nats

Manage NATS resources

### Synopsis

Manage NATS-specific operations exposed by Controller APIs. Use get/describe/deploy/delete for CRUD-style NATS resources.

### Examples

```
potctl nats operator describe
potctl nats accounts ensure my-app --nats-rule default-account-rule
potctl nats users create my-app service-user
potctl nats users creds my-app service-user
```

### Options

```
  -h, --help   help for nats
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl](potctl.md)	 - 
* [potctl nats accounts](potctl_nats_accounts.md)	 - NATS account operations
* [potctl nats operator](potctl_nats_operator.md)	 - NATS operator operations
* [potctl nats users](potctl_nats_users.md)	 - NATS user operations


