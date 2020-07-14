## iofogctl configure

Configure iofogctl or ioFog resources

### Synopsis

Configure iofogctl or ioFog resources

If you would like to replace the host value of Remote Controllers or Agents, you should delete and redeploy those resources.

```
iofogctl configure resource NAME [flags]
```

### Examples

```
iofogctl configure default-namespace NAME

iofogctl configure controller  NAME --user USER --key KEYFILE --port PORTNUM
                   controllers
                   agent
                   agents

iofogctl configure controlplane --kube FILE
```

### Options

```
  -h, --help          help for configure
      --key string    Path to private SSH key
      --kube string   Path to Kubernetes configuration file
      --port int      Port number that iofogctl uses to SSH into remote hosts
      --user string   Username of remote host
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --detached           Use/Show detached resources
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


