# Local `LocalAgent` schema

Reference for **`kind: LocalAgent`** deploy YAML (v3.9.x).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: LocalAgent
metadata:
  name: <optional>             # defaults to "local" if omitted
  namespace: <optional>
  tags: [<optional>]
spec:
  # fields below
```

---

## Fields specific to LocalAgent

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `host` | string | No | Legacy/spec host; **`spec.config.host`** takes precedence for Controller host |
| `controllerEndpoint` | string | No | Override Controller URL for provision |
| `airgap` | bool | No | Same as remote agent |
| `uuid` | string | No | Set after deploy |
| `created` | string | No | Set after deploy |

**No `ssh` block** — bootstrap runs locally.

### Port check

For non-system agents, the CLI rejects deploy if TCP **54321** is already in use on the local machine (conflict with default system agent / local CP layout).

### Name default

If `metadata.name` is empty, **`Sanitize`** sets name to **`local`**.

---

## `spec.package`

Identical to remote agent — see [remote/schema.md — package](../remote/schema.md#specpackage).

---

## `spec.scripts`

Identical to remote agent — see [remote/schema.md — scripts](../remote/schema.md#specscripts).

Local bootstrap uses **`LOCAL_INSTALL=1`** (no remote sudo prereq). See [edgelet-bootstrap.md](../edgelet-bootstrap.md).

---

## `spec.config`

Same structure and semantics as **`Agent.spec.config`**. Full field reference: [remote/schema.md — spec.config](remote/schema.md#specconfig) (tables use `config.*` paths under **`spec.config`**).

Agent identity is **`metadata.name`** (defaults to **`local`** if omitted).

**Host resolution for provision and Controller API:**

1. `spec.config.host` if non-empty  
2. `spec.host`  
3. `localhost`

### Identity and geo

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `config.description` | string | No | Free text |
| `config.location` | string | No | Location label |
| `config.latitude` | float | No | Geo latitude |
| `config.longitude` | float | No | Geo longitude |
| `config.arch` | string | **Recommended** | `amd64`, `arm64`, `arm`, `riscv64`, … — synced to **`archId`** |

### Runtime / edgelet

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.host` | string | See precedence above | Controller agent host |
| `config.deploymentType` | string | `native` | `native` or `container` |
| `config.containerEngine` | string | `edgelet` | `edgelet`, `docker`, or `podman` |
| `config.containerEngineUrl` | string | Engine default | Socket URL for engine |
| `config.networkInterface` | string | No | Bind interface |

### Resource limits and logging

| Field | Type | Description |
|-------|------|-------------|
| `config.diskLimit` | int64 | Disk quota (GiB) |
| `config.diskDirectory` | string | Data directory |
| `config.memoryLimit` | int64 | Memory limit (MiB) |
| `config.cpuLimit` | int64 | CPU limit (%) |
| `config.logLimit` | int64 | Log size limit |
| `config.logDirectory` | string | Log directory |
| `config.logFileCount` | int64 | Rotated log files |
| `config.statusFrequency` | float | Status interval |
| `config.changeFrequency` | float | Change interval |
| `config.watchdogEnabled` | bool | Watchdog |
| `config.logLevel` | string | Log level |
| `config.pruningFrequency` | float | Prune interval |
| `config.availableDiskThreshold` | float | Disk threshold (%) |
| `config.timeZone` | string | Timezone |

### GPS and EdgeGuard

| Field | Type | Description |
|-------|------|-------------|
| `config.gpsMode` | string | e.g. `auto`, `off` |
| `config.gpsScanFrequency` | float | GPS scan interval |
| `config.gpsDevice` | string | GPS device |
| `config.edgeGuardFrequency` | float | **`0`** disables |

### Router configuration

| Field | Type | Description |
|-------|------|-------------|
| `config.upstreamRouters` | []string | Upstream router **agent names** → UUIDs at deploy |
| `config.networkRouter` | string | Network router agent **name** (mode **`none`** only) |

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.routerConfig.routerMode` | string | `edge` | `edge`, `interior`, or `none` |
| `config.routerConfig.messagingPort` | int | `5671` | Messaging port |
| `config.routerConfig.edgeRouterPort` | int | — | **Interior** only |
| `config.routerConfig.interRouterPort` | int | — | **Interior** only |

### NATS configuration

| Field | Type | Description |
|-------|------|-------------|
| `config.upstreamNatsServers` | []string | NATS hub **agent names** → UUIDs at deploy |

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.natsConfig.natsMode` | string | `leaf` | `leaf`, `server`, or `none` |
| `config.natsConfig.natsServerPort` | int | `4222` | Client port |
| `config.natsConfig.natsLeafPort` | int | `7422` | Leaf port |
| `config.natsConfig.natsClusterPort` | int | — | **Server** only |
| `config.natsConfig.natsMqttPort` | int | `8883` | MQTT |
| `config.natsConfig.natsHttpPort` | int | `8222` | Monitoring HTTP |
| `config.natsConfig.jsStorageSize` | string | `10G` | JetStream storage |
| `config.natsConfig.jsMemoryStoreSize` | string | `1G` | JetStream memory |

Defaults, validation, and deploy processing: [remote/schema.md — defaults](remote/schema.md#defaults-and-deploy-time-processing). Desktop **container** on darwin/windows: [edgelet-bootstrap.md](../edgelet-bootstrap.md#desktop-container-local-behavior-darwin--windows).

---

## `configure agent`

Not supported for LocalAgent. SSH **`configure`** command applies to remote agents only.

---

## Related

- [remote/schema.md](../remote/schema.md) — shared package/scripts/config detail
- [lifecycle.md](../lifecycle.md)
