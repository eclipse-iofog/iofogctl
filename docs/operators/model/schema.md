# Model schema

Reference for **`kind: Model`** deploy YAML.

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Model
metadata:
  name: <model-name>           # required; lowercase alphanumeric
  namespace: <optional>
spec:
  # fields below
```

---

## Field reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repo` | string | **Yes** | Artifact repository path (e.g. Hugging Face repo id, OCI repo path) |
| `registry` | string | **One of** | Alias or numeric string resolved via **`remote`**, **`local`**, or id |
| `registryId` | int | registry | Numeric Controller registry id |
| `revision` | string | No | Pin: tag, digest, or `sha256:…` |
| `files` | []string | No | Subpaths or file list within repo (format-specific) |
| `format` | string | No | Artifact format hint (e.g. **`gguf`**) — Controller validates |
| `uuid` | string | Describe only | Controller-assigned; **do not** set on create/update YAML |

### Registry resolution

| Input | Behavior |
|-------|----------|
| `registry: remote` | Id **1** |
| `registry: local` | Id **2** |
| `registry: "5"` | Id **5** |
| `registryId: 5` | Id **5** |
| Both `registry` and `registryId` | Must refer to the **same** id or CLI errors |

---

## Validation (CLI)

| Condition | Error |
|-----------|--------|
| Missing `metadata.name` | Name must be specified |
| Invalid name | Lowercase alphanumeric |
| Missing `repo` | repo must be specified |
| Missing registry | spec.registry or spec.registryId must be specified |
| Conflicting registry fields | must refer to the same registry |

---

## Initial deploy (create)

1. Register or identify registry id ([Registry](../registry/README.md)).
2. Deploy model YAML **without** `spec.uuid`.
3. Optionally **`attach model`** to agents before or after microservice deploy.

**Examples:** [examples/minimal.yaml](examples/minimal.yaml), [examples/with-registry-alias.yaml](examples/with-registry-alias.yaml)

---

## Re-deploy (update / patch)

**Upsert by `metadata.name`** — same name triggers **UpdateModel** with the spec you provide.

1. Change `repo`, `revision`, `files`, `format`, or registry reference as needed.
2. Re-run `deploy -f` with the **same** `metadata.name`.
3. **`spec.uuid`** in describe output is informational; omit from deploy files.

Controller decides whether linked agents or bound microservices need refresh.

---

## Describe output

`describe model NAME` returns:

- **`spec`**: `uuid`, `repo`, `revision`, `registryId`, `files`, `format`
- **`status.linkedAgents`**: agent names when linked

Use describe to round-trip registry id after changing registry entries.

---

## Attach / detach

| Command | Args |
|---------|------|
| `attach model NAME AGENT…` | Resolves agent **names** to fog UUIDs |
| `detach model NAME AGENT…` | Removes link for listed agents |

Errors **409/400** from Controller (e.g. blocking microservices) print the HTTP body.

---

## Microservice `spec.models`

Bind models into a microservice catalog (names only in items):

```yaml
spec:
  models:
    items:
      - name: smollm2-135m
```

Full microservice deploy, rebuild behavior, and **`--patch-model`** are covered in [Microservice](../workloads/microservice/schema.md).

---

## Delete

```bash
potctl delete model <name> -n my-ecn
```

May fail if microservices or links still reference the model.
