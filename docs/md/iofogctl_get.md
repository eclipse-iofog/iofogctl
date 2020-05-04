## iofogctl get

Get information of existing resources

### Synopsis

Get information of existing resources.

Resources like Agents will require a working Controller in the namespace to display all information.

```
iofogctl get RESOURCE [flags]
```

### Examples

```
iofogctl get all
iofogctl get namespaces
iofogctl get controllers

Valid resources are: all, namespaces, controllers, agents, applications, microservices, catalog, registries, volumes

```

### Options

```
  -h, --help   help for get
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


