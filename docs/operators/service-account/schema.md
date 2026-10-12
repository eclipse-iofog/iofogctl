# ServiceAccount schema

```yaml
apiVersion: datasance.com/v3
kind: ServiceAccount
metadata:
  name: <sa-name>
  namespace: my-ecn
  applicationName: <app-name>  # required
roleRef:
  kind: Role
  name: <role-name>
  apiGroup: ""
```

**`describe serviceaccount`** prints this same shape (`applicationName` on metadata, `roleRef` at the document root).

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `metadata.name` | **Yes** | ServiceAccount name within the application |
| `metadata.applicationName` | **Yes** | Application that owns this SA — must match **`kind: Application` `metadata.name`** |
| `roleRef.kind` | **Yes** | Typically **`Role`** |
| `roleRef.name` | **Yes** | Role **`metadata.name`** |
| `roleRef.apiGroup` | No | Passed through to Controller |

---

## Deploy behavior

| State | CLI |
|-------|-----|
| SA missing for `(app, name)` | `CreateServiceAccount` |
| Exists | `UpdateServiceAccount` (updates `roleRef`) |

API path is scoped by **application** + **service account name**.

---

## Microservice `serviceAccount` block

Deploy YAML for **Microservice** may include only **`roleRef`** (no separate ServiceAccount kind):

```yaml
serviceAccount:
  roleRef:
    kind: Role
    name: microservice
```

Whether Controller requires a **ServiceAccount** CRUD object in addition to inline **`roleRef`** depends on policy — using **`kind: ServiceAccount`** + **RoleBinding** is the explicit pattern.

---

## CLI commands

| Command | Args |
|---------|------|
| `get serviceaccounts` | `-n` |
| `describe serviceaccount` | **`application/sa-name`** |
| `delete serviceaccount` | **`application/sa-name`** |

---

## Validation (CLI)

| Condition | Result |
|-----------|--------|
| Missing `applicationName` | ServiceAccount must have applicationName |
| No Controller | CLI error |

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | SA + roleRef |
| [with-microservice.yaml](examples/with-microservice.yaml) | Full RBAC + microservice |
