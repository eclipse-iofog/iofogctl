## potctl exec microservice

Connect to an Exec Session of a Microservice

### Synopsis

Connect to an Exec Session of a Microservice to interact with its container.

```
potctl exec microservice AppName/MsvcName [flags]
```

### Examples

```
potctl exec microservice AppName/MicroserviceName
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

* [potctl exec](potctl_exec.md)	 - Connect to an Exec Session of a resource


