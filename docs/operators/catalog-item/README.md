# Catalog item (`kind: CatalogItem`)

Use **`kind: CatalogItem`** to define **reusable container images per architecture** in the Controller **catalog**. Microservices can reference catalog items by id or name instead of repeating image strings.

Catalog items are **Controller API** resources (not Kubernetes CRDs). Apply them with **`deploy -f`** after registries exist and before applications that depend on custom catalog entries.

## What gets deployed

1. Namespace must have a reachable Controller API.
2. CLI unmarshals **`spec`** into `apps.CatalogItem`.
3. **`GetCatalogItemByName`** — if missing → **create**; if present → **update** (upsert by **name**).

Identity is **`metadata.name`** (recommended). The same name in **`spec.name`** is allowed but **`metadata.name`** is the CLI convention used across other kinds.

## Prerequisites

- At least one **registry id** for `spec.registry` (often **`remote`** / **1** or a custom registry you created with `kind: Registry`).
- **At least one** of `amd64`, `arm64`, `riscv64`, or `arm` image references.

## Deploy

```bash
potctl deploy -f catalog-item.yaml -n my-ecn
potctl get catalog -n my-ecn
potctl delete catalogitem my-image -n my-ecn
```

There is **no** `describe catalogitem` command; use **`get catalog`** or deploy/describe microservices that reference the item.

## System catalog items

The Controller ships **system** catalog entries (router, NATS, debugger) with low numeric ids. Catalog items tied to **system microservices** (historically ids **1–3**) must not be deleted in production. Custom items receive higher ids.

## Deploy order

In multi-document YAML, **`CatalogItem`** runs after **Registry**, **Model**, **Knowledge**, **RuntimeClass**, and **MicroserviceTemplate**, and **before** **Application** and **Microservice**. Define registries and any custom catalog images before microservices that point at them.

## Offline images

`kind: OfflineImage` can register or update catalog entries automatically (registry id **2** / local). That path is separate from standalone **`CatalogItem`** YAML — see [offline-image](../offline-image/README.md).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/multi-arch.yaml](examples/multi-arch.yaml)
- [Registry](../registry/README.md) — custom pull endpoints

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
