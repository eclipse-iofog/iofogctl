# Role binding (`kind: RoleBinding`)

A **RoleBinding** grants a **Role** to one or more **subjects** (typically **ServiceAccount**) in the Controller RBAC model for an ECN namespace.

## What gets deployed

1. Controller reachable.
2. CLI reads **`roleRef`** and **`subjects`** at the document root.
3. **`GetRoleBinding`** — create or **update** (replace roleRef/subjects).

Identity: **`metadata.name`**.

## Prerequisites

- **Role** named in **`roleRef.name`** should exist (deploy Role earlier in the same multi-doc file).
- **Subjects** often reference **ServiceAccount** names that are created in the same file or already on Controller.

## Deploy

```bash
potctl deploy -f rolebinding.yaml -n my-ecn
potctl get rolebindings -n my-ecn
potctl describe rolebinding app-worker-binding -n my-ecn
potctl delete rolebinding app-worker-binding -n my-ecn
```

## Deploy order

**RoleBinding** runs after **Role** and before **ServiceAccount** in CLI **`kindOrder`**. Subjects may name ServiceAccounts deployed later in the same batch — confirm Controller accepts forward references or deploy Application + SA in a second apply if needed.

## Related documents

- [schema.md](schema.md)
- [examples/bind-service-account.yaml](examples/bind-service-account.yaml)
- [Role](../role/README.md) · [Service account](../service-account/README.md)
