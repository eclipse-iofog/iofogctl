## iofogctl exec microservice

Open an interactive exec session to a Microservice

### Synopsis

Open a WebSocket exec session to a running Microservice. No attach step is required.

```
iofogctl exec microservice AppName/MsvcName [flags]
```

### Examples

```
iofogctl exec microservice AppName/MicroserviceName
```

### Options

```
  -h, --help   help for microservice
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl exec](iofogctl_exec.md)	 - Connect to an Exec Session of a resource


