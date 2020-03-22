## iofogctl deploy

Deploy ioFog platform or components on existing infrastructure

### Synopsis

Deploy ioFog platform or individual components on existing infrastructure.
The complete description of yaml file definition can be found at iofog.org

```
iofogctl deploy [flags]
```

### Examples

```
deploy -f platform.yaml
```

### Options

```
  -f, --file string   YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy
  -h, --help          help for deploy
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


