## potctl nats accounts

NATS account operations

### Synopsis

NATS-specific account actions for applications.

### Examples

```
potctl nats accounts ensure my-application --nats-rule default-account-rule
```

### Options

```
  -h, --help   help for accounts
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl nats](potctl_nats.md)	 - Manage NATS resources
* [potctl nats accounts ensure](potctl_nats_accounts_ensure.md)	 - Ensure NATS account for application


