# Offline image (`kind: OfflineImage`)

**OfflineImage** pulls container images on the **operator workstation**, transfers them to **remote agents** (airgap load path), and **registers or updates** a **CatalogItem** on the Controller with **registry id 2** (`local` / offline registry).

Use it for air-gapped edges or preloading images before microservices reference **`catalogItem.name`**.

## What gets deployed

1. Validate **`spec.name`**, at least one arch image, and at least one **remote** agent.
2. For each agent: resolve platform from agent **arch**, pick matching image ref, build airgap transfer options from agent config.
3. Pull/cache images locally under **`~/.iofog/v3`** offline cache (unless **`--no-cache`**).
4. Transfer to agents concurrently (**`--transfer-pool`**, default **2**).
5. **`CreateCatalogItem`** or **`UpdateCatalogItem`** (may prompt before overwrite).

OfflineImage is **not** a long-lived Controller CRUD object like Secret; re-running deploy repeats pull/transfer/catalog upsert.

## Constraints

| Rule | Detail |
|------|--------|
| Agents | **Remote agents only** — LocalAgent targets error |
| SSH | Required per agent (`ValidateSSH`) |
| Controller | Must exist (catalog registration step) |
| Registry id | Catalog entries use **`registryID: 2`** (built-in **local** / offline) |

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: OfflineImage
metadata:
  name: my-api-bundle
  namespace: my-ecn
spec:
  name: my-api-bundle
  amd64: ghcr.io/datasance/example/api:1.0.0
  arm64: ghcr.io/datasance/example/api:1.0.0-arm64
  agent:
    - edge-node-1
  auth:
    username: pull-user
    password: pull-token
```

**`metadata.name`** and **`spec.name`** should align; CLI enforces lowercase alphanumeric on **`spec.name`**.

## Deploy flags

| Flag | Effect |
|------|--------|
| **`--no-cache`** | Do not reuse downloaded image cache for this run |
| **`--transfer-pool N`** | Max concurrent agent transfers (default **2**) |

Also documented in [workloads README](../workloads/README.md).

## Deploy

```bash
potctl deploy -f offline.yaml -n my-ecn --transfer-pool 4
```

There is no `get offline-image` — use **`get catalog`** for the registered item name.

Microservices can reference:

```yaml
container:
  catalogItem:
    name: my-api-bundle
```

Or legacy image fields with registry alias **`local`**.

## Prerequisites

- Remote agents with valid arch (`amd64`, `arm64`, `arm`, `riscv64`).
- Network from **potctl host** to registries for initial pull (unless cache populated).
- Optional **`auth`** when pulls need credentials (both username and password required if block present).

## Deploy order

**OfflineImage** runs after **Volume** and before **VolumeMount** / **Registry** in CLI ordering. Place after **agents** and before **Microservice** documents that use the catalog name.

Pairs with agent **`airgap: true`** — [remote agent schema](../agent/remote/schema.md).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/multi-agent.yaml](examples/multi-agent.yaml)
- [Catalog item](../catalog-item/README.md)
