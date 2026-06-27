## potctl nats users

NATS user operations

### Synopsis

NATS-specific user actions such as create/delete and creds retrieval.

### Examples

```
potctl nats users create my-application service-user
potctl nats users creds my-application service-user
potctl nats users creds my-application service-user -o ./service-user.creds
```

### Options

```
  -h, --help   help for users
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl nats](potctl_nats.md)	 - Manage NATS resources
* [potctl nats users create](potctl_nats_users_create.md)	 - Create NATS user under an application account
* [potctl nats users create-mqtt-bearer](potctl_nats_users_create-mqtt-bearer.md)	 - Create MQTT bearer NATS user
* [potctl nats users creds](potctl_nats_users_creds.md)	 - Fetch NATS creds
* [potctl nats users delete](potctl_nats_users_delete.md)	 - Delete NATS user from an application account
* [potctl nats users delete-mqtt-bearer](potctl_nats_users_delete-mqtt-bearer.md)	 - Delete MQTT bearer NATS user


