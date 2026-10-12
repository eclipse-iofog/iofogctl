# RoleBinding schema

```yaml
apiVersion: datasance.com/v3
kind: RoleBinding
metadata:
  name: <binding-name>
  namespace: my-ecn
roleRef:
  kind: Role
  name: <role-name>
  apiGroup: ""              # optional
subjects:
  - kind: ServiceAccount
    name: <sa-name>
    apiGroup: ""
```

**`describe rolebinding`** prints this same shape (`roleRef` and `subjects` at the document root).

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `roleRef` | **Yes** | Which **Role** to grant |
| `roleRef.kind` | **Yes** | Typically **`Role`** |
| `roleRef.name` | **Yes** | **`metadata.name`** of the Role |
| `roleRef.apiGroup` | No | API group for role reference |
| `subjects` | **Yes** | Who receives the role |
| `subjects[].kind` | **Yes** | e.g. **`ServiceAccount`**, **`User`**, **`Group`** (Controller-supported) |
| `subjects[].name` | **Yes** | Subject name (ServiceAccount **name**, not `app/name`) |
| `subjects[].apiGroup` | No | Subject API group |

ServiceAccounts are **application-scoped** on Controller, but **RoleBinding subjects** use the **ServiceAccount name** only — pairing with **`applicationName`** is enforced when the SA resource is created and when microservices run.

---

## Deploy behavior

| State | Action |
|-------|--------|
| Missing | `CreateRoleBinding` |
| Exists | `UpdateRoleBinding` |

---

## CLI commands

| Command | Args |
|---------|------|
| `get rolebindings` | `-n` |
| `describe rolebinding` | NAME |
| `delete rolebinding` | NAME |

---

## Examples

| File | Shows |
|------|--------|
| [bind-service-account.yaml](examples/bind-service-account.yaml) | Role + Binding + SA |

Full stack: [role/examples/rbac-stack.yaml](../role/examples/rbac-stack.yaml).
