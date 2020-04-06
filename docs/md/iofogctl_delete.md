## iofogctl delete

Delete an existing ioFog resource

### Synopsis

Delete an existing ioFog resource.

```
iofogctl delete [flags]
```

### Examples

```
delete all
delete controller NAME
delete agent NAME
delete application NAME
```

### Options

```
  -f, --file string   YAML file containing resource definitions for Controllers, Agents, and Microservice to delete
  -h, --help          help for delete
      --soft          Don't delete ioFog stack from remote hosts
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
* [iofogctl delete agent](iofogctl_delete_agent.md)	 - Delete an Agent
* [iofogctl delete all](iofogctl_delete_all.md)	 - Delete all resources within a namespace
* [iofogctl delete application](iofogctl_delete_application.md)	 - Delete an application
* [iofogctl delete catalogitem](iofogctl_delete_catalogitem.md)	 - Delete a Catalog item
* [iofogctl delete controller](iofogctl_delete_controller.md)	 - Delete a Controller
* [iofogctl delete microservice](iofogctl_delete_microservice.md)	 - Delete a Microservice
* [iofogctl delete namespace](iofogctl_delete_namespace.md)	 - Delete a Namespace
* [iofogctl delete registry](iofogctl_delete_registry.md)	 - Delete a registry
* [iofogctl delete volume](iofogctl_delete_volume.md)	 - Delete an Volume


