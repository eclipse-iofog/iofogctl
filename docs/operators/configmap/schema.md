# ConfigMap schema

```yaml
apiVersion: datasance.com/v3
kind: ConfigMap
metadata:
  name: <name>
  namespace: my-ecn
spec:
  immutable: <bool>    # optional pointer semantics on update
  useVault: <bool>     # optional; Controller vault integration
data:
  <key>: <string>
```

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `metadata.name` | **Yes** | Unique ConfigMap name (lowercase alphanumeric) |
| `spec.immutable` | No | When set on **create**, stored on Controller. On **update**, explicit value overrides; omitted → keep existing immutable flag |
| `spec.useVault` | No | When set on create/update, toggles vault-backed storage when changed from existing |
| `data` | **Yes** (deploy) | String map of configuration entries |

---

## Deploy behavior

| State | CLI |
|-------|-----|
| Absent | `CreateConfigMap` with `data`, `immutable`, `useVault` |
| Present | `UpdateConfigMap` — always sends **`data`**; **`immutable`** / **`useVault`** follow merge rules in [factory.go](../../../internal/deploy/configmap/factory.go) |

If **`useVault`** in YAML equals the existing Controller value, CLI omits the field from the update request (no-op).

---

## Consumption

### Microservice environment

```yaml
- key: FEATURE_FLAGS
  valueFromConfigMap: platform-config/features
```

Reference format: **`configmap-name/dataKey`**.

### VolumeMount

```yaml
spec:
  configMapName: platform-config
```

---

## CLI commands

| Command | Args |
|---------|------|
| `get configmaps` | `-n` |
| `describe configmap` | NAME |
| `delete configmap` | NAME |

---

## Validation

| Condition | Result |
|-----------|--------|
| Invalid name | CLI error |
| No Controller | CLI error |
| Controller rejects immutable change | HTTP error |

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Basic `data` |
| [immutable.yaml](examples/immutable.yaml) | `spec.immutable: true` on create |
