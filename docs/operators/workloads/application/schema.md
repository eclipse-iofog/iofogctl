# Application

## What it is

An **Application** is a **logical grouping** of microservices on the Controller — the unit for:

- Starting/stopping a set of services together (`StartApplication` on deploy)
- **NATS application policy** (`natsConfig`)
- Optional **template instantiation** (`spec.template` → ApplicationTemplate)

Every **user microservice** belongs to exactly one application name (`myapp`, `wasm`, etc.). System workloads use **`system-<agentName>`** applications managed by the Controller.

An Application YAML may **embed microservices** under **`spec.microservices[]`** (one-shot app + services file) **or** you deploy microservices as separate **`kind: Microservice`** documents in the same multi-doc file.

---

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: Application
metadata:
  name: <application-name>
  namespace: my-ecn
spec:
  # see below
```

---

## Deploy behavior

| Step | SDK |
|------|-----|
| Parse spec | non-strict into generic object |
| Lookup | `GetApplicationByName` |
| Create / update | `CreateApplicationFromYAML` / `UpdateApplicationFromYAML` |
| Start | **`StartApplication`** after create/update |

Embedded **`microservices`** in Application YAML are deployed as part of the application YAML upload path on the Controller (not separate Microservice kind executors in the same pass unless also listed as separate documents).

---

## Spec fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Optional duplicate of metadata name |
| `microservices` | []Microservice | Inline microservice specs (same fields as [Microservice](../microservice/schema.md)) |
| `natsConfig` | object | Application-level NATS access — must use **`natsConfig`**, not a top-level **`natsAccess`** |
| `natsConfig.natsAccess` | bool | Enable NATS for this application (creates account JWT, resolver bundle). Required **before** any microservice **`natsAccess`** in this app |
| `natsConfig.natsRule` | string | [NatsAccountRule](../../nats-account-rule/README.md) **`metadata.name`**. Omit when `natsAccess: true` → **`default-account`**. Per-microservice users: [Microservice NATS](../microservice/schema.md#nats) |

Controller runtime (account creation, resolver, lifecycle): [NATS access](../nats-access/README.md).
| `template` | object | Instantiate from ApplicationTemplate |
| `template.name` | string | ApplicationTemplate metadata.name |
| `template.variables` | [] or map | Override template variables (key/value) |
| `id` | int | Read-only Controller id (describe) |

### Template instantiation

When **`spec.template`** is set, the Controller merges the ApplicationTemplate **`application`** skeleton with supplied **variables**. You may still add **`microservices`** for additional services not in the template.

---

## Initial deploy vs update

| | Create | Update |
|---|--------|--------|
| Condition | Application name absent | Name exists |
| CLI | deploy `-f` | Same file, same **`metadata.name`** |

Updating an application YAML **does not** delete microservices removed from **`spec.microservices`** unless Controller YAML semantics say so — verify with **`get microservices`** and **`describe`**.

---

## Relationship to Microservice kind

| Approach | When to use |
|----------|-------------|
| **Application only** with `spec.microservices[]` | Small apps, demos, single file |
| **Application + Microservice** docs | Independent service lifecycle, clearer `App/Msvc` CLI names |
| **Microservice only** | App already exists; add service |

Standalone **`kind: Microservice`** **requires** **`spec.application`** (or FQ metadata name).

---

## Commands

```bash
potctl deploy -f application.yaml -n my-ecn
potctl get applications -n my-ecn
potctl describe application myapp -n my-ecn
potctl delete application myapp -n my-ecn
```

System applications are listed via **`get system-applications`** (not user **`delete application`**).

---

## Examples

- [wasm-inline-microservices.yaml](examples/wasm-inline-microservices.yaml)
- [from-template.yaml](examples/from-template.yaml)
