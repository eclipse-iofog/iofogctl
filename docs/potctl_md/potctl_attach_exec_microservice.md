## potctl attach exec microservice

Attach an Exec Session to a Microservice

### Synopsis

Attach an Exec Session to an existing Microservice.

```
potctl attach exec microservice NAME [flags]
```

### Examples

```
potctl attach exec microservice AppName/MicroserviceName
```

### Options

```
  -h, --help   help for microservice
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl attach exec](potctl_attach_exec.md)	 - Attach an Exec Session to a resource


