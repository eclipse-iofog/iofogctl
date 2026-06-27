## potctl create namespace

Create a Namespace

### Synopsis

Create a Namespace.

A Namespace contains all components of an Edge Compute Network.

A single instance of potctl can be used to manage any number of Edge Compute Networks.

```
potctl create namespace NAME [flags]
```

### Examples

```
potctl create namespace NAME
```

### Options

```
  -h, --help   help for namespace
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl create](potctl_create.md)	 - Create a resource


