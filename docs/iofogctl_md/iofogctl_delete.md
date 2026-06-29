## iofogctl delete

Delete an existing ioFog resource

### Synopsis

Delete an existing ioFog resource.

```
iofogctl delete [flags]
```

### Options

```
      --delete-namespace   Also delete the Kubernetes namespace (never deletes "default")
  -f, --file string        YAML file containing specifications for ioFog resources to deploy
  -h, --help               help for delete
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 
* [iofogctl delete agent](iofogctl_delete_agent.md)	 - Delete an Agent
* [iofogctl delete all](iofogctl_delete_all.md)	 - Delete all resources within a namespace
* [iofogctl delete application](iofogctl_delete_application.md)	 - Delete an application
* [iofogctl delete application-template](iofogctl_delete_application-template.md)	 - Delete an application-template
* [iofogctl delete auth-group](iofogctl_delete_auth-group.md)	 - Delete a custom embedded auth group
* [iofogctl delete catalogitem](iofogctl_delete_catalogitem.md)	 - Delete a Catalog item
* [iofogctl delete certificate](iofogctl_delete_certificate.md)	 - Delete a Certificate
* [iofogctl delete configmap](iofogctl_delete_configmap.md)	 - Delete a ConfigMap
* [iofogctl delete controller](iofogctl_delete_controller.md)	 - Delete a Controller
* [iofogctl delete microservice](iofogctl_delete_microservice.md)	 - Delete a Microservice
* [iofogctl delete namespace](iofogctl_delete_namespace.md)	 - Delete a Namespace
* [iofogctl delete nats-account-rule](iofogctl_delete_nats-account-rule.md)	 - Delete a NATS account rule
* [iofogctl delete nats-user-rule](iofogctl_delete_nats-user-rule.md)	 - Delete a NATS user rule
* [iofogctl delete registry](iofogctl_delete_registry.md)	 - Delete a Registry
* [iofogctl delete role](iofogctl_delete_role.md)	 - Delete a Role
* [iofogctl delete rolebinding](iofogctl_delete_rolebinding.md)	 - Delete a RoleBinding
* [iofogctl delete secret](iofogctl_delete_secret.md)	 - Delete a Secret
* [iofogctl delete service](iofogctl_delete_service.md)	 - Delete a Service
* [iofogctl delete serviceaccount](iofogctl_delete_serviceaccount.md)	 - Delete a ServiceAccount
* [iofogctl delete volume](iofogctl_delete_volume.md)	 - Delete an Volume
* [iofogctl delete volume-mount](iofogctl_delete_volume-mount.md)	 - Delete a Volume Mount


