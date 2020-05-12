## iofogctl describe

Get detailed information of existing resources

### Synopsis

Get detailed information of existing resources.

Resources such as Agents require a working Controller in the namespace in order to be described.

```
iofogctl describe resource NAME [flags]
```

### Examples

```
iofogctl describe namespace
iofogctl describe controlplane
iofogctl describe controller NAME
iofogctl describe agent NAME
iofogctl describe agent-config NAME
iofogctl describe microservice NAME

Valid resources are: namespace, controlplane, controller, agent, agent-config, microservice, application, volume, route

```

### Options

```
  -h, --help                 help for describe
  -o, --output-file string   YAML output file
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


