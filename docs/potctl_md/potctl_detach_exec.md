## potctl detach exec

Detach an Exec Session to a resource

### Synopsis

Detach an Exec Session to a Microservice or Agent.

### Examples

```
potctl detach exec microservice AppName/MicroserviceName
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl detach](potctl_detach.md)	 - Detach one ioFog resource from another
* [potctl detach exec agent](potctl_detach_exec_agent.md)	 - Detach an Exec Session from an Agent
* [potctl detach exec microservice](potctl_detach_exec_microservice.md)	 - Detach an Exec Session to a Microservice


