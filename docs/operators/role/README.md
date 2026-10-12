# Role (`kind: Role`)

A **Role** defines **RBAC rules** (API groups, resources, verbs) stored in the **Controller API** for an ECN namespace. **RoleBindings** attach roles to **ServiceAccounts**; microservices can reference a role inline via **`spec.serviceAccount.roleRef`**.

This is **ECN Controller RBAC**, not the Kubernetes **Role** installed with the [iofog-operator control plane](../controlplane/kubernetes/README.md).

## What gets deployed

1. Controller reachable in namespace.
2. CLI reads **`rules`** at the document root.
3. **`GetRole(name)`** — absent → **create**; present → **update** rules.

Identity: **`metadata.name`**.

## Prerequisites

- Control plane with Controller API.
- Deploy **before** **RoleBinding** documents that reference this role name.

## Deploy

```bash
potctl deploy -f role.yaml -n my-ecn
potctl get roles -n my-ecn
potctl describe role microservice -n my-ecn
potctl delete role microservice -n my-ecn
```

## Typical stack

```text
Role → RoleBinding (subjects) → ServiceAccount (applicationName + roleRef) → Microservice (serviceAccount.roleRef)
```

See [role-binding](../role-binding/README.md), [service-account](../service-account/README.md), and [microservice RBAC](../workloads/microservice/schema.md#rbac-application-scoped).

## Deploy order

CLI **`kindOrder`** runs **Role** before **RoleBinding**, **ServiceAccount**, agents, and **Application**. Rules can be uploaded early; bindings and microservices need matching application names at runtime.

## Related documents

- [schema.md](schema.md)
- [examples/microservice-api.yaml](examples/microservice-api.yaml)
- [examples/rbac-stack.yaml](examples/rbac-stack.yaml) (multi-doc with binding + SA)
