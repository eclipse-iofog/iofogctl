# Role schema

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Role
metadata:
  name: <role-name>
  namespace: my-ecn
rules:
  - apiGroups: [""]
    resources: ["microservices"]
    verbs: ["get", "list"]
    resourceNames: []         # optional restrict to named resources
```

**`metadata.name`** is the Controller role name used in **`roleRef.name`**. **`describe role`** prints this same shape (`rules` at the document root). `metadata.namespace` on describe is the namespace you queried.

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `rules` | **Yes** (typical) | List of RBAC rules (Kubernetes-style) |

### Rule object (`RBACRule`)

| Field | Description |
|-------|-------------|
| `apiGroups` | API group strings; `""` for core PoT API groups as defined by Controller |
| `resources` | Resource types (e.g. `microservices`, `agents`, … — Controller catalog) |
| `verbs` | `get`, `list`, `create`, `update`, `delete`, `watch`, … |
| `resourceNames` | Optional: limit rule to named instances |

Exact allowed **resources** and **verbs** are enforced by the **Controller**, not validated in depth by potctl.

---

## Deploy behavior

| State | CLI |
|-------|-----|
| Role missing | `CreateRole` with `name`, `kind`, `rules` |
| Role exists | `UpdateRole` with new `rules` |

---

## Consumption

### RoleBinding

```yaml
roleRef:
  kind: Role
  name: microservice
```

### Microservice (inline ref)

Microservice YAML can set **`spec.serviceAccount.roleRef`** without creating a **ServiceAccount** kind — Controller may still expect a **ServiceAccount** + **RoleBinding** for full RBAC; align with your Controller version.

```yaml
spec:
  serviceAccount:
    roleRef:
      kind: Role
      name: microservice
```

---

## CLI commands

| Command | Args |
|---------|------|
| `get roles` | `-n` |
| `describe role` | NAME |
| `delete role` | NAME |

---

## Examples

| File | Shows |
|------|--------|
| [microservice-api.yaml](examples/microservice-api.yaml) | Minimal rules |
| [rbac-stack.yaml](examples/rbac-stack.yaml) | Role + RoleBinding + ServiceAccount |
