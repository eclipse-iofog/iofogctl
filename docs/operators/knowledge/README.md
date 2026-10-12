# Fleet knowledge (`kind: Knowledge`)

Use **`kind: Knowledge`** for **document, dataset, or vector-index artifacts** stored in a registry-backed repo — same operational pattern as **Model**, different artifact role on the Controller.

Knowledge uses **YAML parse → JSON CRUD** (`CreateKnowledge` / `UpdateKnowledge`).

## What gets deployed

1. Controller API in the namespace.
2. Registry referenced by **`spec.registry`** or **`spec.registryId`** must exist.
3. Upsert by **`metadata.name`** (create or update).

## Fleet linking

```bash
potctl attach knowledge product-docs edge-a -n my-ecn
potctl detach knowledge product-docs edge-a -n my-ecn
potctl describe knowledge product-docs -n my-ecn
```

## Microservice catalog binding

Microservices reference knowledge in **`spec.knowledge`**:

```yaml
spec:
  knowledge:
    items:
      - name: product-docs
```

See [Microservice](../workloads/microservice/schema.md) for deploy and catalog binding details.

**`deploy -f --patch-knowledge`** is documented under [Workloads](../workloads/README.md#deploy--f-flags-workload-related). It may be combined with **`--patch-model`** on the same Microservice-only file.

## Prerequisites

Same naming and registry rules as [Model](../model/schema.md).

## Deploy order

```text
Registry → Model → Knowledge → RuntimeClass → … → Microservice
```

## Deploy

```bash
potctl deploy -f knowledge.yaml -n my-ecn
potctl get knowledge -n my-ecn
potctl delete knowledge product-docs -n my-ecn
```

## Name collision with Model

Knowledge and Model names are **independent** — both may be named `foo`.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/update-same-name.yaml](examples/update-same-name.yaml)
- [Model](../model/README.md)
- [Registry](../registry/README.md)

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
