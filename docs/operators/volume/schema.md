# Volume schema

```yaml
apiVersion: datasance.com/v3
kind: Volume
metadata:
  name: <volume-name>
  namespace: my-ecn
spec:
  name: <volume-name>
  agents:
    - <agent-name>
  source: <local-directory>
  destination: <absolute-path-on-agent>
  permissions: "0755"
```

All **`spec`** fields are unmarshalled into the volume record stored under namespace config.

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `name` | **Yes** | Volume identifier; microservice **`hostDestination`** when `type: volume` |
| `agents` | **Yes** | Agent **names** in this ECN namespace |
| `source` | **Yes** | Path to a **directory on the potctl host** to copy from |
| `destination` | **Yes** | Absolute directory on each target agent (created if missing) |
| `permissions` | **Yes** | Unix permission string applied during remote copy (e.g. `"0777"`, `"0755"`) |

---

## Remote deploy mechanics

For each remote agent concurrently:

1. SSH connect using agent **`user`**, **`host`**, **`keyFile`**, **`port`**.
2. Create **`destination`** and temp folder **`/tmp/iofogctlssh` + destination**.
3. **`CopyFolderTo`** from local **`source`** → temp on agent.
4. Move contents into **`destination`** with `sudo cp -pR`.

Failures surface per-agent errors.

---

## Local agent semantics

When **`agents`** includes a **LocalAgent**:

- CLI prints that volume deploy is **unnecessary** (host filesystem is used directly).
- If **`source` ≠ `destination`**, CLI warns about possible bind/mount confusion.

Config still records the volume for naming consistency.

---

## Validation (CLI)

| Rule | Error |
|------|-------|
| `source` not a directory | Source must be a directory |
| Remote agent on loopback host | Not supported for local Agents |
| Missing SSH | ValidateSSH error |
| Unknown agent name | Namespace lookup error |

---

## Storage location

Volume definitions live in **`~/.iofog/v3/<namespace>/`** namespace YAML (not Controller). **`describe volume`** reads from this config.

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Single remote agent |
