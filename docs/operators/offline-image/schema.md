# OfflineImage schema

```yaml
apiVersion: datasance.com/v3
kind: OfflineImage
metadata:
  name: <bundle-name>
  namespace: my-ecn
spec:
  name: <bundle-name>
  amd64: <image-ref>
  arm64: <image-ref>
  arm: <image-ref>
  riscv64: <image-ref>
  agent:
    - <remote-agent-name>
  auth:
    username: <string>
    password: <string>
```

At least **one** of **`amd64`**, **`arm64`**, **`arm`**, **`riscv64`** is required. Each agent receives the image matching its platform.

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `name` | **Yes** | Catalog item name and bundle id (lowercase alphanumeric) |
| `amd64` | One of arch fields | Image for `linux/amd64` agents |
| `arm64` | One of arch fields | Image for `linux/arm64` agents |
| `arm` | One of arch fields | Image for `linux/arm` agents |
| `riscv64` | One of arch fields | Image for `linux/riscv64` agents |
| `agent` | **Yes** | List of **remote** agent names |
| `auth.username` | If `auth` set | Registry username |
| `auth.password` | If `auth` set | Registry password or token |

---

## Execution pipeline

```text
validate → buildAgentPlans → prepareArtifacts (pull/cache)
  → distribute (airgap transfer per agent) → registerCatalogItem
```

### Catalog registration

- Lookup **`GetCatalogItemByName(spec.name)`**.
- Missing → **create** with **`RegistryID: 2`**, description `Offline image bundle <name>`.
- Exists → interactive confirm → **update** images if approved.

Built-in registry alias **`local`** maps to id **2** — see [registry README](../registry/README.md).

---

## Platform mapping

Agent **`arch`** from Controller config resolves to platform; missing arch-specific image errors with a clear message (e.g. arm64 agent without `arm64:` field).

---

## Cache paths

CLI stores artifacts under namespace offline cache dirs (`GetOfflineImageCacheDir`). **`--no-cache`** forces fresh pulls where applicable.

---

## Validation (CLI)

| Condition | Error |
|-----------|--------|
| Empty `name` | spec must include a name |
| No agents | must include at least one agent |
| No image arch | must include at least one architecture image |
| Local agent target | only supports remote Agents |
| Partial `auth` | requires both username and password |

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Single agent, amd64 + arm64 |
| [multi-agent.yaml](examples/multi-agent.yaml) | Multiple agents + auth |
