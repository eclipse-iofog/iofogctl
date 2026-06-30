## potctl delete

Delete an existing ioFog resource

### Synopsis

Delete an existing ioFog resource.

```
potctl delete [flags]
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
  -v, --verbose            Toggle for displaying verbose output of potctl
```

### SEE ALSO

* [potctl](potctl.md)	 - 
* [potctl delete agent](potctl_delete_agent.md)	 - Delete an Agent
* [potctl delete all](potctl_delete_all.md)	 - Delete all resources within a namespace
* [potctl delete application](potctl_delete_application.md)	 - Delete an application
* [potctl delete application-template](potctl_delete_application-template.md)	 - Delete an application-template
* [potctl delete auth-group](potctl_delete_auth-group.md)	 - Delete a custom embedded auth group
* [potctl delete catalogitem](potctl_delete_catalogitem.md)	 - Delete a Catalog item
* [potctl delete certificate](potctl_delete_certificate.md)	 - Delete a Certificate
* [potctl delete configmap](potctl_delete_configmap.md)	 - Delete a ConfigMap
* [potctl delete controller](potctl_delete_controller.md)	 - Delete a Controller
* [potctl delete microservice](potctl_delete_microservice.md)	 - Delete a Microservice
* [potctl delete namespace](potctl_delete_namespace.md)	 - Delete a Namespace
* [potctl delete nats-account-rule](potctl_delete_nats-account-rule.md)	 - Delete a NATS account rule
* [potctl delete nats-user-rule](potctl_delete_nats-user-rule.md)	 - Delete a NATS user rule
* [potctl delete registry](potctl_delete_registry.md)	 - Delete a Registry
* [potctl delete role](potctl_delete_role.md)	 - Delete a Role
* [potctl delete rolebinding](potctl_delete_rolebinding.md)	 - Delete a RoleBinding
* [potctl delete secret](potctl_delete_secret.md)	 - Delete a Secret
* [potctl delete service](potctl_delete_service.md)	 - Delete a Service
* [potctl delete serviceaccount](potctl_delete_serviceaccount.md)	 - Delete a ServiceAccount
* [potctl delete volume](potctl_delete_volume.md)	 - Delete an Volume
* [potctl delete volume-mount](potctl_delete_volume-mount.md)	 - Delete a Volume Mount


