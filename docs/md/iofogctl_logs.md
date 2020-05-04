## iofogctl logs

Get log contents of deployed resource

### Synopsis

Get log contents of deployed resource

```
iofogctl logs RESOURCE NAME [flags]
```

### Examples

```
iofogctl logs controller NAME
iofogctl logs agent NAME
iofogctl logs microservice NAME
```

### Options

```
  -h, --help   help for logs
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


