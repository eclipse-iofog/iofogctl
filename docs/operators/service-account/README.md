# Service account (`kind: ServiceAccount`)

A **ServiceAccount** is an **application-scoped** identity on the Controller. It ties an application name to a **Role** via **`roleRef`**. Microservices use **`spec.serviceAccount.roleRef`** to run with that permission model; **RoleBindings** attach the role to the account.

## What gets deployed

1. Controller reachable.
2. **`metadata.applicationName`** is **required** (CLI error if empty).
3. **`GetServiceAccount(applicationName, name)`** — create or **update** (`roleRef`).

Identity: **`(applicationName, metadata.name)`** — not namespace-global.

## Naming and CLI

| Operation | Identifier |
|-----------|------------|
| `describe serviceaccount` | **`APP/NAME`** (e.g. `my-app/worker`) |
| `delete serviceaccount` | **`APP/NAME`** |
| `get serviceaccounts` | Lists APPLICATION + NAME + ROLE |

## Prerequisites

- **`metadata.applicationName`** must match a real **Application** (deploy **`kind: Application`** in the same multi-doc file or earlier apply if Controller requires it).
- **Role** referenced in **`roleRef`** should exist.
- Optional **RoleBinding** linking this SA name as a subject.

## Deploy

```bash
potctl deploy -f sa.yaml -n my-ecn
potctl get serviceaccounts -n my-ecn
potctl describe serviceaccount my-app/worker -n my-ecn
potctl delete serviceaccount my-app/worker -n my-ecn
```

## Microservice usage

```yaml
spec:
  application: my-app
  serviceAccount:
    roleRef:
      kind: Role
      name: app-worker
```

See [microservice schema — RBAC](../workloads/microservice/schema.md#rbac-application-scoped).

Volume mapping **`type: serviceAccount`** (Controller API) mounts SA-related material — rare in user YAML; see Controller docs if enabled.

## Deploy order

**ServiceAccount** runs after **Role** / **RoleBinding** and **before** agents in **`kindOrder`**, but **before** **Application** in the same file. Recommended operator flow:

```text
Application (or same file) → Role → RoleBinding → ServiceAccount → Microservice
```

If create fails with unknown application, deploy **Application** first in a separate `deploy -f`.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/with-microservice.yaml](examples/with-microservice.yaml)
- [Role](../role/README.md) · [Role binding](../role-binding/README.md)
