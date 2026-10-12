# MicroserviceTemplate

## What it is

A **MicroserviceTemplate** is a **reusable blueprint for one microservice** — variables plus a **microservice** skeleton (images, container, config, catalog hooks).

Use it when many services share the same container shape but differ by agent name, image tag, fleet model names, or env.

Instantiation happens on **`kind: Microservice`** via **`spec.template`** (name + variables). There is **no** separate CLI subcommand to “instantiate template” — include **`template`** in the microservice YAML and **`deploy -f`**.

---

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: MicroserviceTemplate
metadata:
  name: <template-name>
  namespace: my-ecn
spec:
  description: ...
  variables: ...
  microservice: ...
```

---

## Deploy behavior

| Phase | Behavior |
|-------|----------|
| Parse | non-strict **`yaml.Unmarshal`** on spec |
| Deploy | `DeployMicroserviceTemplate` → create/update YAML on Controller |

---

## Spec fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Summary for operators |
| `variables` | []object | Same shape as ApplicationTemplate variables |
| `variables[].key` | string | Required per entry |
| `variables[].description` | string | Optional |
| `variables[].defaultValue` | any | Default when MS YAML omits override |
| `microservice` | object | Partial or full microservice spec (Controller template engine) |

The **`microservice`** block mirrors [Microservice](../microservice/schema.md) fields but may use placeholders (for example `{{application}}`, `{{agent-name}}`).

---

## Microservice `spec.template` reference

When deploying **`kind: Microservice`**:

```yaml
spec:
  template:
    name: my-ms-template
    variables:
      agent-name: edge-01
      application: myapp
```

**`variables`** may be a **map** or a **list** of `{key, value}` pairs (SDK accepts both).

---

## Initial deploy vs update

Re-deploy same **`metadata.name`** → update template on Controller. Existing microservices already deployed from the template are **not** auto-updated.

---

## Commands

```bash
potctl deploy -f ms-template.yaml -n my-ecn
potctl describe microservice-template my-ms-template -n my-ecn
potctl delete microservice-template my-ms-template -n my-ecn
potctl get microservice-templates -n my-ecn
```

---

## Examples

- [minimal.yaml](examples/minimal.yaml)
