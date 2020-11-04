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
                  controlplane
                  controller     NAME
                  agent          NAME
                  agent-config   NAME
                  application    NAME
                  microservice   NAME
                  volume         NAME
                  route          NAME
                  edge-resource  NAME/VERSION
```

### Options

```
      --detached             Specify command is to run against detached resources
  -h, --help                 help for describe
  -o, --output-file string   YAML output file
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


