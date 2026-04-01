## iofogctl nats users

NATS user operations

### Synopsis

NATS-specific user actions such as create/delete and creds retrieval.

### Examples

```
iofogctl nats users create my-application service-user
iofogctl nats users creds my-application service-user
iofogctl nats users creds my-application service-user -o ./service-user.creds
```

### Options

```
  -h, --help   help for users
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl nats](iofogctl_nats.md)	 - Manage NATS resources
* [iofogctl nats users create](iofogctl_nats_users_create.md)	 - Create NATS user under an application account
* [iofogctl nats users create-mqtt-bearer](iofogctl_nats_users_create-mqtt-bearer.md)	 - Create MQTT bearer NATS user
* [iofogctl nats users creds](iofogctl_nats_users_creds.md)	 - Fetch NATS creds
* [iofogctl nats users delete](iofogctl_nats_users_delete.md)	 - Delete NATS user from an application account
* [iofogctl nats users delete-mqtt-bearer](iofogctl_nats_users_delete-mqtt-bearer.md)	 - Delete MQTT bearer NATS user


