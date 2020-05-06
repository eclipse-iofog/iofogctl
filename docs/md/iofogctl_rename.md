## iofogctl rename

Rename the iofog resources that are currently deployed

### Synopsis

Rename the iofog resources that are currently deployed

### Examples

```
iofogctl rename namespace NAME NEW_NAME
iofogctl rename controller NAME NEW_NAME
iofogctl rename agent NAME NEW_NAME
iofogctl rename microservice NAME NEW_NAME
iofogctl rename application NAME NEW_NAME
```

### Options

```
  -h, --help   help for rename
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
* [iofogctl rename agent](iofogctl_rename_agent.md)	 - Rename an Agent
* [iofogctl rename application](iofogctl_rename_application.md)	 - Rename a Application
* [iofogctl rename controller](iofogctl_rename_controller.md)	 - Rename a Controller
* [iofogctl rename microservice](iofogctl_rename_microservice.md)	 - Rename a Microservice
* [iofogctl rename namespace](iofogctl_rename_namespace.md)	 - Rename the Namespace of your ECN


