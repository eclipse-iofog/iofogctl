# Knowledge schema

Reference for **`kind: Knowledge`** deploy YAML. Field shapes match **Model** (P11-YAML-1).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Knowledge
metadata:
  name: <knowledge-name>       # required; lowercase alphanumeric
  namespace: <optional>
spec:
  # fields below
```

---

## Field reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repo` | string | **Yes** | Repository path for documents / datasets |
| `registry` | string | **One of** | `remote`, `local`, numeric string, or id alias |
| `registryId` | int | registry | Numeric registry id |
| `revision` | string | No | Tag, digest, or content pin |
| `files` | []string | No | File paths or globs within repo |
| `format` | string | No | e.g. **`jsonl`**, **`pdf` bundle** — Controller validates |
| `uuid` | string | Describe only | Omit on deploy |

Registry alias rules and conflict checks are identical to [Model schema](../model/schema.md#registry-resolution).

---

## Validation (CLI)

Same messages as Model: name, repo, registry required; lowercase alphanumeric name.

---

## Initial deploy (create)

Deploy without **`spec.uuid`**. Register HF or OCI registry first if not using **`remote`**.

**Example:** [examples/minimal.yaml](examples/minimal.yaml)

---

## Re-deploy (update / patch)

Same **`metadata.name`** → **UpdateKnowledge**. Use to change revision, files, format, or registry reference.

**Example:** [examples/update-same-name.yaml](examples/update-same-name.yaml)

---

## Describe output

- **`spec`**: `uuid`, `repo`, `revision`, `registryId`, `files`, `format`
- **`status.linkedAgents`**

---

## Attach / detach

| Command | Description |
|---------|-------------|
| `attach knowledge NAME AGENT…` | Link to agents by name |
| `detach knowledge NAME AGENT…` | Unlink listed agents |

---

## Microservice `spec.knowledge`

```yaml
spec:
  knowledge:
    items:
      - name: product-docs
```

Patch-only deploy flags: [Workloads](../workloads/README.md#deploy--f-flags-workload-related).

---

## Agent status fields

After linking, agents report **`knowledgeStatus`**, **`activeKnowledge`**, and **`knowledgeLastUpdate`** (Unix ms on wire; CLI describe may format as RFC3339). **`activeKnowledge`** is a managed count — not `len()` of parsed status rows.

---

## Delete

```bash
potctl delete knowledge <name> -n my-ecn
```
