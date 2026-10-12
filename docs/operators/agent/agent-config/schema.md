# AgentConfig schema

Reference for **`kind: AgentConfig`** (v3.9.x). Fields below live at **`spec.*`**. For **`Agent`** / **`LocalAgent`**, use the same fields under **`spec.config.*`** — full tables in [remote/schema.md](../remote/schema.md#specconfig) and [local/schema.md](../local/schema.md#specconfig).

```yaml
apiVersion: datasance.com/v3
kind: AgentConfig
metadata:
  name: my-edge
spec:
  description: Edge node
  arch: amd64
  host: 10.0.0.5
  routerConfig:
    routerMode: edge
    messagingPort: 5671
  natsConfig:
    natsMode: leaf
    natsServerPort: 4222
  upstreamRouters:
    - default-router
  upstreamNatsServers:
    - default-nats-hub
```

---

## CLI wrapper fields (top of spec)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Display name; defaults to **`metadata.name`** if empty |
| `description` | string | No | Free text |
| `location` | string | No | Location label |
| `latitude` | float | No | Geo |
| `longitude` | float | No | Geo |
| `arch` | string | **Recommended** | `amd64`, `arm64`, `arm`, `riscv64`, … — mapped to Controller **`archId`** at deploy |

---

## Runtime / edgelet host

| Field | Type | Default (edge agent) | Description |
|-------|------|----------------------|-------------|
| `host` | string | From agent spec | Controller agent host; interior router mode may require host |
| `deploymentType` | string | `native` | `native` or `container` |
| `containerEngine` | string | `edgelet` | `edgelet`, `docker`, or `podman` |
| `containerEngineUrl` | string | Engine default | e.g. `unix:///run/edgelet/containerd.sock` |
| `networkInterface` | string | No | Bind interface |

**`isSystem`:** not settable via YAML (`yaml:"-"` in SDK). System agents are created by control plane deploy with internal flags.

---

## Resource limits and logging

| Field | Type | Description |
|-------|------|-------------|
| `diskLimit` | int64 | Disk quota (GiB, agent semantics) |
| `diskDirectory` | string | Data directory on host |
| `memoryLimit` | int64 | Memory limit (MiB) |
| `cpuLimit` | int64 | CPU limit (% — 100 = one core) |
| `logLimit` | int64 | Log size limit |
| `logDirectory` | string | Log directory |
| `logFileCount` | int64 | Rotated log files |
| `statusFrequency` | float | Status report interval |
| `changeFrequency` | float | Change detection interval |
| `watchdogEnabled` | bool | Watchdog |
| `logLevel` | string | e.g. `INFO`, `info` |
| `pruningFrequency` | float | Prune interval |
| `availableDiskThreshold` | float | Disk pressure threshold (%) |
| `timeZone` | string | Host timezone |

---

## GPS and EdgeGuard

| Field | Type | Description |
|-------|------|-------------|
| `gpsMode` | string | e.g. `auto`, `off` |
| `gpsScanFrequency` | float | Scan interval |
| `gpsDevice` | string | Device path or selector |
| `edgeGuardFrequency` | float | EdgeGuard interval; **`0`** disables |

HAL / Bluetooth fields from older releases are **retired** in v3.9.

---

## Router configuration

Fabric behavior (listeners, TLS CAs, connectors, Service **`bridges`**): [router-fabric.md](../networking/router-fabric.md).

### Top-level references

| Field | Type | Description |
|-------|------|-------------|
| `upstreamRouters` | []string | **Agent names** at YAML time → UUIDs at deploy |
| `networkRouter` | string | Network router agent **name** (mode **`none`** only) |

### `routerConfig` block (nested — preferred)

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `routerMode` | string | `edge` | `edge`, `interior`, or `none` |
| `messagingPort` | int | `5671` | Messaging AMQP port |
| `edgeRouterPort` | int | — | **Interior** mode only |
| `interRouterPort` | int | — | **Interior** mode only |

**System agent defaults:** `routerMode` **interior**, ports filled by `ApplySystemAgentDefaults`.

### Validation highlights

- **`none`:** cannot set `upstreamRouters` or non-none `networkRouter` combinations incorrectly.
- **`edge` / `leaf`:** cannot set `edgeRouterPort` / `interRouterPort`.
- **`networkRouter`:** only when `routerMode` is **`none`**.

---

## NATS configuration

Fabric behavior (server vs leaf, JetStream, resolver bundles): [nats-fabric.md](../networking/nats-fabric.md). Workload creds: [NATS access](../../workloads/nats-access/README.md).

### Top-level references

| Field | Type | Description |
|-------|------|-------------|
| `upstreamNatsServers` | []string | **Agent names** (NATS hub agents) → UUIDs at deploy |

### `natsConfig` block (nested — preferred)

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `natsMode` | string | `leaf` | `leaf`, `server`, or `none` |
| `natsServerPort` | int | `4222` | Client port |
| `natsLeafPort` | int | `7422` | Leaf port |
| `natsClusterPort` | int | — | **Server** mode only |
| `natsMqttPort` | int | `8883` | MQTT |
| `natsHttpPort` | int | `8222` | Monitoring HTTP |
| `jsStorageSize` | string | `10G` | JetStream storage |
| `jsMemoryStoreSize` | string | `1G` | JetStream memory |

**System agent defaults:** `natsMode` **server**, cluster port and JetStream sizes applied automatically.

---

## Deploy-time processing

| Step | Behavior |
|------|----------|
| `PrepareForControllerAPI` | Edge vs system defaults; arch → `archId` |
| `Process` | Name → UUID for upstream lists; host rules for interior router |
| `Validate` | Router/NATS mode matrix |
| SDK create/update | Controller persists config; platform reconciles async |

**Read-only on describe (not in deploy spec):** daemon status, CPU/memory/disk usage, platform phase, runtime classes, model/knowledge status — see `describe agent-config` output.

---

## `archId`

Normally derived from **`arch`**. Do not set both inconsistently. Invalid **`arch`** strings may map to **`0`** / auto behavior depending on Controller validation.

---

## Related

- [README.md](README.md)
- [remote/schema.md](../remote/schema.md) — SSH, package, scripts
- [control plane systemAgent](../../controlplane/remote/schema.md) — per-controller host config
