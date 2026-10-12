# ApplicationTemplate

## What it is

An **ApplicationTemplate** is a **reusable blueprint** for creating **Applications**. It stores:

- Human-readable **description**
- **Variables** (keys, descriptions, default values)
- A nested **application** skeleton (structure the Controller merges when you instantiate)

Templates are **Controller API objects**. Deploy with **`kind: ApplicationTemplate`** — not a Kubernetes CRD.

Operators use templates when the same app shape (microservice layout, NATS policy, variable slots) must be deployed many times with different parameter values.

---

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: ApplicationTemplate
metadata:
  name: <template-name>    # required; lowercase alphanumeric
  namespace: my-ecn          # optional; must match CLI -n if set
spec:
  # template body (see below)
```

---

## Deploy behavior

| Phase | SDK / CLI |
|-------|-----------|
| Parse | **`yaml.Unmarshal`** (non-strict) on **`spec`** |
| Lookup | `GetApplicationTemplate(name)` |
| Create | `CreateApplicationTemplateFromYAML` |
| Update | `UpdateApplicationTemplateFromYAML` |

Instantiate a template from an **Application** via **`spec.template`** (see [Application schema](../application/schema.md)).

---

## Spec fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional duplicate of metadata name |
| `description` | string | Operator-facing summary |
| `variables` | []object | Template parameters |
| `variables[].key` | string | Variable name referenced in nested YAML |
| `variables[].description` | string | Help text |
| `variables[].defaultValue` | any | JSON-compatible default (scalar or object) |
| `variables[].value` | any | Optional fixed value when not using Application override |
| `application` | object | Nested application spec (microservices, natsConfig, etc.) |

Variable placeholders in nested YAML typically use template syntax consumed by the Controller when **`Application.spec.template`** supplies **`variables`**.

---

## Initial deploy vs update

| | Create | Update |
|---|--------|--------|
| **Trigger** | Template name not on Controller | Same **`metadata.name`** re-deployed |
| **Action** | Create from YAML | Update from YAML |

Changing a template does **not** automatically redeploy Applications already created from it — update those Applications separately.

---

## Commands

```bash
potctl deploy -f app-template.yaml -n my-ecn
potctl describe application-template my-template -n my-ecn
potctl delete application-template my-template -n my-ecn
```

---

## Examples

- [minimal.yaml](examples/minimal.yaml)
