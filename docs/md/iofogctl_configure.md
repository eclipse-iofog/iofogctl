## iofogctl configure

Configure iofogctl or SSH details an existing resource

### Synopsis

Configure iofogctl or SSH details for an existing resource

Note that you cannot (and shouldn't need to) configure the host value of Agents.

```
iofogctl configure resource NAME [flags]
```

### Examples

```
iofogctl configure default-namespace NAME
iofogctl configure controller NAME --host HOST --user USER --key KEYFILE --port PORTNUM
iofogctl configure controller NAME --kube KUBECONFIG
iofogctl configure agent NAME --user USER --key KEYFILE --port PORTNUM

iofogctl configure all --user USER --key KEYFILE --port PORTNUM
iofogctl configure controllers --host HOST NAME --user USER --key KEYFILE --port PORTNUM
iofogctl configure agents --user USER --key KEYFILE --port PORTNUM

Valid resources are: controller, agent, all, agents, controllers, default-namespace

```

### Options

```
  -h, --help          help for configure
      --host string   Hostname of remote host
      --key string    Path to private SSH key
      --kube string   Path to Kubernetes configuration file
      --port int      Port number that iofogctl uses to SSH into remote hosts
      --user string   Username of remote host
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "6928137")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


