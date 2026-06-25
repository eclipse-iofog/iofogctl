## potctl detach exec microservice

Detach an Exec Session to a Microservice

### Synopsis

Detach an Exec Session to an existing Microservice.

```
potctl detach exec microservice NAME [flags]
```

### Examples

```
potctl detach exec microservice AppName/MicroserviceName
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

* [potctl detach exec](potctl_detach_exec.md)	 - Detach an Exec Session to a resource


