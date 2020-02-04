## iofogctl connect

Connect to an existing ioFog cluster

### Synopsis

Connect to an existing ioFog cluster.

This command must be executed within an empty or non-existent namespace.
All resources provisioned with the corresponding Controller will become visible under the namespace.
All ssh access will be configured as provided in the config file.
See iofog.org for the YAML format.

```
iofogctl connect [flags]
```

### Examples

```
iofogctl connect -f platform.yaml
iofogctl connect --kube FILE --name NAME --email EMAIL --pass PASSWORD
iofogctl connect --ecn-addr ENDPOINT --name NAME --email EMAIL --pass PASSWORD
```

### Options

```
      --ecn-addr string   URL of Edge Compute Network to connect to
      --email string      ioFog user email address
  -f, --file string       YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy
      --force             Overwrite existing namespace
  -h, --help              help for connect
      --kube string       Kubernetes config file. Typically ~/.kube/config
      --name string       Name you would like to assign to Controller
      --pass string       ioFog user password
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


