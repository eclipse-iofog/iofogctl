# Remote `Agent` schema

Reference for **`kind: Agent`** deploy YAML (v3.9.x).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Agent
metadata:
  name: <agent-name>           # required; lowercase alphanumeric
  namespace: <optional>        # must match CLI -n if set
  tags: [<optional tags>]
spec:
  # fields below
```

Agent identity in the Controller is **`metadata.name`**. The CLI passes that name to validation and AgentConfig executors.

---

## CLI / host fields (not AgentConfig API only)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `host` | string | **Yes** | SSH target host or IP |
| `ssh.user` | string | **Yes** | SSH user |
| `ssh.keyFile` | string | **Yes** | Path to private key (expanded with `~`) |
| `ssh.port` | int | No | Default **22** |
| `controllerEndpoint` | string | No | Override Controller URL for provision; default from namespace CP |
| `airgap` | bool | No | Stage binaries/images from operator cache; see [airgap](#specairgap) |
| `uuid` | string | No | Read-only after deploy; round-trip / describe |
| `created` | string | No | Read-only CLI timestamp after deploy |

### Validation

- **`metadata.name`:** lowercase alphanumeric; cannot be **`default-router`** (reserved).
- **SSH:** required unless host is special-cased in validation (production paths expect full SSH).
- **Same host as Controller:** non-system agents cannot share the Controller hostname (see [README](README.md)).

---

## `spec.package`

Optional edgelet **binary**, **container image**, and **WASM shims** for bootstrap.

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Edgelet version passed to native `install.sh` when using default embedded install |
| `container.image` | string | Container image when `deploymentType: container` |
| `container.registry` | string | Registry hostname for private pull |
| `container.username` | string | With registry + password for private image |
| `container.password` | string | |
| `wasm.<handler>` | object | WASM shim pack per handler (`spin`, `edgelet-wasm`, …) |
| `wasm.<handler>.url` | string | HTTP(S) tarball URL (CLI extracts locally) |
| `wasm.<handler>.path` | string | Local path for fully offline operator |
| `wasm.<handler>.sha256` | string | Optional integrity check |

**WASM scope:** linux, `deploymentType: native`, `containerEngine` `edgelet` or `docker` — not podman+docker WASM combo. See [edgelet-bootstrap](../edgelet-bootstrap.md#wasm-runtime-scope).

When **`spec.package.container.image`** is set, the CLI may set **`deploymentType: container`** automatically if not specified in config.

---

## `spec.scripts`

Optional override of embedded bootstrap layers.

| Field | Description |
|-------|-------------|
| `dir` | **Required** when block present — directory on **operator machine** |
| `deps.entrypoint` / `deps.args` | Custom dependency layer |
| `install.entrypoint` / `install.args` | Custom edgelet install |
| `uninstall.entrypoint` / `uninstall.args` | Custom uninstall (e.g. `--remove-data`) |

Full merge rules: [edgelet-bootstrap.md](../edgelet-bootstrap.md).

---

## `spec.config`

Controller **AgentConfig** plus edgelet runtime materialization. Agent identity is **`metadata.name`** — there is no `config.name` on **`kind: Agent`**.

| Field | Meaning |
|-------|---------|
| **`spec.host`** | SSH target (see [CLI / host fields](#cli--host-fields-not-agentconfig-api-only)) |
| **`spec.config.host`** | Host registered on the Controller API (provision / platform); may differ from SSH host |

Standalone **`kind: AgentConfig`** uses the same fields at **`spec.*`** instead of **`spec.config.*`**. See [agent-config/schema.md](../agent-config/schema.md).

Example (nested router / NATS):

```yaml
config:
  description: Edge node
  arch: amd64
  host: 10.0.0.50
  deploymentType: native
  containerEngine: edgelet
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

### Identity and geo

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `config.description` | string | No | Free text |
| `config.location` | string | No | Location label |
| `config.latitude` | float | No | Geo latitude |
| `config.longitude` | float | No | Geo longitude |
| `config.arch` | string | **Recommended** | `amd64`, `arm64`, `arm`, `riscv64`, … — synced to Controller **`archId`** at deploy |

### Runtime / edgelet

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.host` | string | Often same as `spec.host` | Controller agent host; required for some router modes |
| `config.deploymentType` | string | `native` | `native` or `container` |
| `config.containerEngine` | string | `edgelet` | `edgelet`, `docker`, or `podman` |
| `config.containerEngineUrl` | string | Engine default | e.g. `unix:///run/edgelet/containerd.sock` |
| `config.networkInterface` | string | No | Network interface to bind |

**`config.isSystem`:** not settable via YAML. System agents use control plane **`systemAgent`**, not **`kind: Agent`**.

### Resource limits and logging

| Field | Type | Description |
|-------|------|-------------|
| `config.diskLimit` | int64 | Disk quota (GiB) |
| `config.diskDirectory` | string | Data directory on host |
| `config.memoryLimit` | int64 | Memory limit (MiB) |
| `config.cpuLimit` | int64 | CPU limit (% — 100 = one core) |
| `config.logLimit` | int64 | Log size limit |
| `config.logDirectory` | string | Log directory |
| `config.logFileCount` | int64 | Rotated log file count |
| `config.statusFrequency` | float | Status report interval |
| `config.changeFrequency` | float | Change detection interval |
| `config.watchdogEnabled` | bool | Watchdog |
| `config.logLevel` | string | e.g. `INFO`, `info` |
| `config.pruningFrequency` | float | Prune interval |
| `config.availableDiskThreshold` | float | Disk pressure threshold (%) |
| `config.timeZone` | string | Host timezone |

### GPS and EdgeGuard

| Field | Type | Description |
|-------|------|-------------|
| `config.gpsMode` | string | e.g. `auto`, `off` |
| `config.gpsScanFrequency` | float | GPS scan interval |
| `config.gpsDevice` | string | Device path or selector |
| `config.edgeGuardFrequency` | float | EdgeGuard interval; **`0`** disables |

HAL / Bluetooth agent fields are **retired** in v3.9.

### Router configuration

#### Top-level references

| Field | Type | Description |
|-------|------|-------------|
| `config.upstreamRouters` | []string | Upstream router **agent names** → UUIDs at deploy |
| `config.networkRouter` | string | Network router agent **name** (only when mode is **`none`**) |

#### `config.routerConfig` (nested)

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.routerConfig.routerMode` | string | `edge` | `edge`, `interior`, or `none` |
| `config.routerConfig.messagingPort` | int | `5671` | Messaging AMQP port |
| `config.routerConfig.edgeRouterPort` | int | — | **Interior** mode only |
| `config.routerConfig.interRouterPort` | int | — | **Interior** mode only |

#### Router validation (edge agents)

- **`none`:** must not set `upstreamRouters` inconsistently; `networkRouter` only valid for **`none`**.
- **`edge`:** do not set `edgeRouterPort` / `interRouterPort`.
- **`interior`:** edge/inter router ports allowed; used for system agents on CP hosts, not typical edge **`Agent`**.

### NATS configuration

#### Top-level references

| Field | Type | Description |
|-------|------|-------------|
| `config.upstreamNatsServers` | []string | NATS hub **agent names** → UUIDs at deploy |

#### `config.natsConfig` (nested)

| Field | Type | Default (edge) | Description |
|-------|------|----------------|-------------|
| `config.natsConfig.natsMode` | string | `leaf` | `leaf`, `server`, or `none` |
| `config.natsConfig.natsServerPort` | int | `4222` | Client port |
| `config.natsConfig.natsLeafPort` | int | `7422` | Leaf port |
| `config.natsConfig.natsClusterPort` | int | — | **Server** mode only |
| `config.natsConfig.natsMqttPort` | int | `8883` | MQTT |
| `config.natsConfig.natsHttpPort` | int | `8222` | Monitoring HTTP |
| `config.natsConfig.jsStorageSize` | string | `10G` | JetStream storage |
| `config.natsConfig.jsMemoryStoreSize` | string | `1G` | JetStream memory |

#### NATS validation (edge agents)

- Do not set `natsClusterPort` unless `natsMode` is **`server`**.

### Defaults and deploy-time processing

| Step | Behavior |
|------|----------|
| `PrepareForControllerAPI` | Edge defaults: router **`edge`**, NATS **`leaf`**, standard ports if omitted |
| `Process` | Resolve `upstreamRouters` / `upstreamNatsServers` names to UUIDs; interior router host rules |
| `Validate` | Router/NATS mode matrix before SDK create/update |
| Bootstrap | Go **`MaterializeEdgeletRuntime`** uses config for host `config.yaml` where applicable |

**`config.archId`:** normally derived from **`config.arch`** — do not set inconsistently.

**Describe / status (not in deploy YAML):** daemon status, usage, platform phase, runtimes — from **`describe agent`** / **`describe agent-config`**.

---

## `spec.airgap`

When **`true`**:

- Native: CLI stages edgelet binary from `~/.iofog/v3/airgap-binaries/` (and related caches) and passes airgap flags to install.
- Images: `edgelet image load` on host when applicable.

Requires coherent **`spec.config.arch`** and package metadata. Remote control plane context may influence transfer options when CP is remote.

---

## Deploy behavior summary

| Step | Component |
|------|-----------|
| 1 | AgentConfig create/update + wait platform ready (batch phase) |
| 2 | `EnsureAgentConfig`, deployment type resolution |
| 3 | Optional airgap / WASM staging |
| 4 | `CustomizeProcedures` + bootstrap |
| 5 | AgentConfig update if `spec.config` non-nil |
| 6 | Provision + namespace persist |

---

## Related

- [agent-config/schema.md](../agent-config/schema.md) — same fields at **`spec.*`** for **`kind: AgentConfig`**
- [lifecycle.md](../lifecycle.md) — delete, detach, configure SSH
