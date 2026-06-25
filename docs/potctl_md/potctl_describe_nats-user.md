## potctl describe nats-user

Get detailed information about a NATS user

### Synopsis

Get detailed information about a NATS user.

```
potctl describe nats-user APP_NAME USER_NAME [flags]
```

### Options

```
  -h, --help            help for nats-user
      --jwt             Output decoded JWT payload only
      --output string   Output format: yaml|json|wide (default "yaml")
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl describe](potctl_describe.md)	 - Get detailed information of an existing resources


