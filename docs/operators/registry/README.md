# Controller registry (`kind: Registry`)

Use **`kind: Registry`** to register an **image or artifact pull endpoint** with the **Controller API**. Registries are referenced by numeric **id** from catalog items, microservice images, and fleet **Model** / **Knowledge** resources.

This is **not** the edgelet host registry (`apiVersion: edgelet.iofog.org/v1`, `kind: Registry`) used during control plane bootstrap for private controller images. That flow lives under [control plane documentation](../controlplane/README.md).

## What gets deployed

1. Namespace must contain a **control plane** with a reachable Controller API (`connect` or successful CP deploy).
2. CLI unmarshals **`spec`** into the registry shape and validates type, privacy, and credentials.
3. If **`spec.id`** is absent or zero → **`CreateRegistry`** (Controller assigns **id**).
4. If **`spec.id` > 0** → **`UpdateRegistry`** (patch existing entry).

Deploy uses **JSON CRUD** on the Controller — there is no separate `apply` command.

## Built-in registry aliases

When other kinds reference a registry by alias (not numeric id):

| Alias | Id | Typical use |
|-------|-----|-------------|
| `remote` | **1** | Default online / Docker Hub–style pulls |
| `local` | **2** | Airgap / pre-loaded images (`from_cache` on edgelet) |

Use `get registries` to list **all** ids (built-in plus user-created).

## Identity and CLI commands

| Operation | Identifier |
|-----------|------------|
| `deploy -f` create | No **`spec.id`** (or `id: 0`) |
| `deploy -f` update | **`spec.id`** from `get registries` |
| `describe registry` | **Numeric id** |
| `delete registry` | **Numeric id** |

**`metadata.name`** is optional and **not** used as the Controller registry key. Identity is **Controller-assigned id** plus **url**.

## Prerequisites

- Connected Controller in the target namespace.
- For **private** registries, credentials must match **`spec.type`** rules (see [schema.md](schema.md)).

## Deploy

```bash
potctl deploy -f registry.yaml -n my-ecn
potctl get registries -n my-ecn
potctl describe registry 3 -n my-ecn
```

Replace `potctl` with `iofogctl` on the Eclipse flavor.

## Re-deploy (patch) workflow

1. Run `get registries` and note the **ID** for your URL.
2. Edit the manifest: keep **`spec.url`** and credentials; set **`spec.id: <ID>`**.
3. Run `deploy -f` again — the CLI calls **update**, not create.

See [examples/patch-update.yaml](examples/patch-update.yaml).

### New registry + Model in one session

**Model** / **Knowledge** YAML needs a **known registry id** or alias (`remote`, `local`). If you **`deploy -f`** a **new** `kind: Registry` in the same file, the Controller assigns an id during that run — you will not know the id until **`get registries`** completes. Typical flow:

1. `deploy -f registry.yaml` (create)
2. `get registries` → note id
3. Set **`spec.registryId`** (or numeric **`spec.registry`**) on Model/Knowledge YAML
4. `deploy -f` fleet artifacts

## Deploy order in multi-document YAML

Registry documents should appear **before** `CatalogItem`, `Model`, `Knowledge`, and `Microservice` kinds that reference their id or alias. The CLI sorts kinds automatically; still list registries first in files you maintain by hand.

## Control plane auto-registration

When you deploy a **LocalControlPlane** or remote **ControlPlane** with a **private** `spec.controller.package` (registry + username + password), the CLI may **also** create a Controller registry entry during CP install. Avoid duplicating the same URL with a separate `kind: Registry` unless you intend to manage it explicitly.

## Related documents

- [schema.md](schema.md) — field reference, validation, retired v3.8 flags
- [examples/](examples/) — OCI/HF public and private, insecure, private CA, patch update

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
| CLI config directory | `~/.iofog/v3` | `~/.iofog/v3` |

Examples use `datasance.com/v3`; substitute `iofog.org/v3` for iofogctl.
