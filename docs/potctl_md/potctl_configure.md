## potctl configure

Configure potctl or ioFog resources

### Synopsis

Configure potctl or ioFog resources

If you would like to replace the host value of Remote Controllers or Agents, you should delete and redeploy those resources.

```
potctl configure RESOURCE NAME [flags]
```

### Examples

```
potctl configure current-namespace NAME

potctl configure controller  NAME --user USER --key KEYFILE --port PORTNUM
                   controllers
                   agent
                   agents
				   controlplane

potctl configure controlplane --kube FILE
```

### Options

```
      --ca string       Path to PEM CA certificate for controller TLS (persisted to namespace config)
      --ca-b64 string   Base64-encoded PEM CA certificate for controller TLS (persisted to namespace config)
      --detached        Specify command is to run against detached resources
  -h, --help            help for configure
      --key string      Path to private SSH key
      --kube string     Path to Kubernetes configuration file
      --port int        Port number that potctl uses to SSH into remote hosts
      --user string     Username of remote host
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl](potctl.md)	 - 
* [potctl configure auth-group](potctl_configure_auth-group.md)	 - Update an embedded auth group


