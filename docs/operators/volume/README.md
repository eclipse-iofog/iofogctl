# Volume (`kind: Volume`)

A **Volume** stages a **directory tree** from the operator workstation onto one or more **agents**, then records the mapping in **CLI namespace config** (`~/.iofog/v3`). Microservices mount the volume by **name** using **`container.volumes`** with **`type: volume`**.

Volumes are **not** stored in the Controller API. They are edge data paths plus local potctl state.

## What gets deployed

1. **`spec.agents`** must list agent **names** already in namespace config.
2. **`spec.source`** must be an **existing directory on the machine running potctl** (validated with `os.Stat`).
3. For each **remote** agent: SSH copy from `source` → `destination` on the host (with temp staging under `/tmp/iofogctlssh…`).
4. For **local** agents: no file copy — CLI notifies that host filesystem bind is used directly.
5. Namespace config is updated via **`UpdateVolume`** and flushed to disk.

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: Volume
metadata:
  name: model-weights
  namespace: my-ecn
spec:
  name: model-weights
  agents:
    - edge-node-1
  source: /data/models-src
  destination: /var/lib/iofog/volumes/model-weights
  permissions: "0755"
```

**`metadata.name`** and **`spec.name`** should match (CLI sets `spec.name` from metadata when deploying from file).

## Agent constraints

| Agent type | Behavior |
|------------|----------|
| **Remote** | SSH required (`ValidateSSH`). Host must **not** be loopback — “Volume deployment is not supported for local Agents” on remote path |
| **Local** | Records volume in config; warns that deploy is unnecessary — microservices bind host paths directly |

## Prerequisites

- Agents deployed and listed in namespace.
- Remote: SSH keys/user/host configured in agent spec.
- Source directory readable by the user running potctl.

## Deploy

```bash
potctl deploy -f volume.yaml -n my-ecn
potctl get volumes -n my-ecn
potctl describe volume model-weights -n my-ecn
potctl delete volume model-weights -n my-ecn
```

**Delete volume** removes staged data on agents (best-effort per agent) and drops the config entry.

## Microservice usage

```yaml
container:
  volumes:
    - type: volume
      hostDestination: model-weights
      containerDestination: /models
      accessMode: rw
      scope: private   # or shared
```

See [microservice schema — volumes](../workloads/microservice/schema.md#container-volumes).

## Lifecycle interactions

- **`delete agent`** removes the agent from volume agent lists or deletes volumes that only referenced that agent — [agent lifecycle](../agent/lifecycle.md).

## Deploy order

**Volume** runs after **ApplicationTemplate** and before **OfflineImage** / **VolumeMount** in multi-doc deploy. Agents must exist before volume deploy.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
