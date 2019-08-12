## iofogctl deploy agent

Bootstrap and provision an edge host

### Synopsis

Bootstrap an edge host with the ioFog Agent stack and provision it with a Controller.

A Controller must first be deployed within the corresponding namespace in order to provision the Agent.

```
iofogctl deploy agent NAME [flags]
```

### Examples

```
iofogctl deploy agent NAME --local
iofogctl deploy agent NAME --user root --host 32.23.134.3 --key-file ~/.ssh/id_rsa
```

### Options

```
  -h, --help              help for agent
      --host string       IP or hostname of host the Agent is being deployed on
      --key-file string   Filename of SSH private key used to access host. Corresponding *.pub must be in same dir. Must be RSA key.
  -l, --local             Configure for local deployment. Cannot be used with other flags
      --port int          SSH port to use when deploying agent to host (default 22)
      --user string       Username of host the Agent is being deployed on
```

### Options inherited from parent commands

```
      --config string      CLI configuration file (default is ~/.iofog/config.yaml)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -q, --quiet              Toggle for displaying verbose output
  -v, --verbose            Toggle for displaying verbose output of API client
```

### SEE ALSO

* [iofogctl deploy](iofogctl_deploy.md)	 - Deploy ioFog platform or components on existing infrastructure

###### Auto generated by spf13/cobra on 9-Aug-2019