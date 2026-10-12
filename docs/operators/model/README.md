# Fleet model (`kind: Model`)

Use **`kind: Model`** to declare a **large artifact** (weights, GGUF bundle, etc.) stored in a **registry-backed repo** on the Controller. Models are **fleet resources**: you **attach** them to agents so edge nodes can pull and run inference workloads.

Models use **YAML parse → JSON CRUD** (`CreateModel` / `UpdateModel`) — not multipart YAML upload.

## What gets deployed

1. Controller API available in the namespace.
2. A **registry** entry must exist for the **`spec.registry`** / **`spec.registryId`** you reference.
3. CLI upserts by **`metadata.name`**: create if missing, else update.

## Fleet linking

After deploy, link models to agents by **name**:

```bash
potctl attach model smollm2-135m edge-a edge-b -n my-ecn
potctl detach model smollm2-135m edge-a -n my-ecn
potctl describe model smollm2-135m -n my-ecn
```

`describe` includes **`status.linkedAgents`** when links exist.

## Microservice catalog binding

Microservices can reference models in **`spec.models`** (catalog bind). That flow is part of application deploy; see [Microservice](../workloads/microservice/schema.md) for full **`spec.models`** shape and deploy examples.

**Catalog-only updates** on an existing microservice use **`deploy -f --patch-model`** — see [Workloads](../workloads/README.md#deploy--f-flags-workload-related) (Controller owns rebuild rules).

## Prerequisites

- **`metadata.name`**: lowercase alphanumeric.
- **`spec.repo`**: required repository path (semantics depend on **`spec.format`** and registry type — Controller validates).
- **`spec.registry`** or **`spec.registryId`**: required (aliases **`remote`**, **`local`**, or numeric id).

## Deploy order

In multi-doc YAML:

```text
Registry → Model → … → Application → Microservice
```

Place **Registry** documents before **Model**.

## Deploy

```bash
potctl deploy -f model.yaml -n my-ecn
potctl get models -n my-ecn
potctl delete model smollm2-135m -n my-ecn
```

## Name collision with Knowledge

A **Model** and a **Knowledge** resource **may share the same name** — they are separate Controller namespaces.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/with-registry-alias.yaml](examples/with-registry-alias.yaml)
- [examples/fleet-stack.yaml](examples/fleet-stack.yaml) — registry + model + knowledge + runtimeclass
- [Registry](../registry/README.md)
- [Knowledge](../knowledge/README.md)
- [RuntimeClass](../runtime-class/README.md)

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
