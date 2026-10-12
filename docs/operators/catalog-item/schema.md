# CatalogItem schema

Reference for **`kind: CatalogItem`** deploy YAML.

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: CatalogItem
metadata:
  name: <catalog-name>         # required; lowercase alphanumeric
  namespace: <optional>
spec:
  # fields below (SDK CatalogItem shape)
```

---

## Field reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `registry` | int or alias | No | **`RegistryRef`**: integer id, or **`remote`** (1), **`local`** (2). Default **1** if omitted |
| `amd64` | string | One of arch | Container image for amd64 |
| `arm64` | string | images | Container image for arm64 |
| `riscv64` | string | required | Container image for riscv64 |
| `arm` | string | | Container image for arm (32-bit) |
| `description` | string | No | Human-readable text |
| `name` | string | No | Duplicate of metadata name; prefer **`metadata.name`** |
| `id` | int | No | Read-only on Controller; **not** used by CLI deploy for upsert |
| `configExample` | string | No | Optional example config snippet (round-trip / tooling) |

Image strings are typically **`registry/path:tag`** resolved via the **`registry`** id (where the registry hostname is defined on the Controller registry entry).

### Registry alias reference

| YAML value | Id |
|------------|-----|
| `remote` | 1 |
| `local` | 2 |
| `3`, `4`, … | User-created **`kind: Registry`** ids |

---

## Validation (CLI)

| Condition | Error |
|-----------|--------|
| Missing name | Name must be specified |
| Invalid name | Lowercase alphanumeric |
| No arch images | At least one image must be specified |

---

## Initial deploy (create)

1. Choose **`metadata.name`** (unique in the ECN).
2. Set **`spec.registry`** to the id that can pull your images.
3. Set at least one arch image field.
4. Omit **`spec.id`** — Controller assigns id on create.

**Example:** [examples/minimal.yaml](examples/minimal.yaml)

---

## Re-deploy (update / patch)

**Upsert by name** — no **`spec.id`** required in YAML.

1. Keep the same **`metadata.name`**.
2. Change images, description, or **`registry`** as needed.
3. Run `deploy -f` again — CLI loads existing item by name and calls **UpdateCatalogItem**.

Changing **`registry`** or image digests affects microservices that reference this catalog item; Controller may reject deletes while references exist.

**Example:** [examples/multi-arch.yaml](examples/multi-arch.yaml) (create); re-apply the same file with updated tags to patch.

---

## Delete

```bash
potctl delete catalogitem <metadata.name> -n my-ecn
```

Resolves name → id, then deletes. Expect errors if system items or active microservices still reference the entry.

---

## Relationship to Microservice

Microservices reference catalog entries via **`spec.images.catalogId`** or catalog-backed deploy paths. Full microservice YAML is documented under [Microservice](../workloads/microservice/schema.md).
