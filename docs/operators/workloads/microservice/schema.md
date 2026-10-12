# Microservice

## What it is

A **Microservice** is the **smallest deployable workload unit** on an edge agent: one container (or WASM module) with image reference, resource limits, ports, environment, optional fleet **model/knowledge** mounts, and optional **RuntimeClass** handler.

For **container** workloads (not WASM), runtime shape depends on the **target agent’s** [`containerEngine`](../../agent/agent-config/schema.md):

| `containerEngine` | On the agent |
|-------------------|--------------|
| **`edgelet`** (default) | Edgelet runs the microservice as a **single-container pod** on its built-in engine (edgelet pod abstraction, not a Kubernetes Pod). |
| **`docker`** or **`podman`** | Edgelet orchestrates the same logical microservice as **one container** on that external engine. |

It always runs on **exactly one agent** (`spec.agent.name`) and belongs to **one application** (`spec.application` or FQ metadata name).

The Controller schedules pull, create, and health-check on that agent’s container engine.

---

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: Microservice
metadata:
  name: <msvc-name>              # or application/msvc
  namespace: my-ecn
spec:
  application: <app-name>        # required unless metadata.name contains /
  # …
```

---

## Deploy behavior

| Mode | CLI | Controller path |
|------|-----|-----------------|
| **Full deploy** | default `deploy -f` | **`yaml.UnmarshalStrict`** → `DeployMicroservice` → create/update **FromYAML** |
| **Patch models only** | `--patch-model` | `PatchMicroserviceModels` — file must be Microservice-only, **`spec.models` required** |
| **Patch knowledge only** | `--patch-knowledge` | `PatchMicroserviceKnowledge` — **`spec.knowledge` required** |
| **Both patches** | both flags | Both PATCH calls (models first) |

| Create vs update | SDK sets UUID on existing name in application → **update**; else **create** |
| System MS | **Cannot create** via user deploy; updates use system YAML API when UUID known |

Set **`spec.rebuild: true`** when you need Controller to force container rebuild on full deploy (Controller still owns final rebuild heuristics for catalog PATCH).

---

## Identity and naming

| Source | Result |
|--------|--------|
| `metadata.name: api` + `spec.application: myapp` | Deploy as **`myapp/api`** |
| `metadata.name: myapp/api` | Application **myapp**, microservice **api** |

CLI commands (`describe`, `logs`, `exec`, `delete`) use **`application/microservice`**.

---

## Spec — grouped reference

### Core

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `application` | string | Yes* | Application name (*unless FQ metadata.name) |
| `name` | string | No | Duplicate of metadata microservice name |
| `agent` | object | **Yes** | Placement |
| `agent.name` | string | **Yes** | Controller agent name |
| `schedule` | int | No | Scheduling hint (Controller-defined) |
| `config` | object | No | Opaque JSON config for the microservice |
| `rebuild` | bool | No | Request rebuild on deploy |
| `uuid` | string | No | **Describe only** — do not set on create |

### Images

| Field | Type | Description |
|-------|------|-------------|
| `images` | object | Container images per architecture |
| `images.registry` | int / alias | **`remote`** (1), **`local`** (2), or registry id — [RegistryRef](../../registry/schema.md) |
| `images.amd64` | string | Image for amd64 |
| `images.arm64` | string | Image for arm64 |
| `images.riscv64` | string | Image for riscv64 |
| `images.arm` | string | Image for arm |
| `images.catalogId` | int | Use [CatalogItem](../../catalog-item/schema.md) instead of inline image strings |

Provide **catalogId** **or** per-arch images — catalog wins in describe when set.

### Container (summary)

See [Container appendix](#appendix-container-fields) for every field.

| Group | Fields | Purpose |
|-------|--------|---------|
| Security | `isPrivileged`, `runAsUser`, `runAsGroup`, `readOnlyRootFilesystem`, `capAdd`, `capDrop` | Linux security |
| Network | `hostNetworkMode`, `ports`, `extraHosts` | Exposure, DNS — [extraHosts](#container-extrahosts), [sysctls vs host network](#container-sysctls) |
| WASM / runtime | `platform`, `runtime` | e.g. `wasi/wasm` + [RuntimeClass](../../runtime-class/README.md) handler name |
| Resources | `cpus`, `memoryLimit`, `memoryReservation`, `memorySwap`, `shmSize`, `cpuSetCpus`, `ulimits` | [CPU/memory units](#container-resources-cpu-and-memory) |
| Storage | `volumes`, `tmpfs` | [Volume mappings](#container-volumes) and tmpfs |
| Process | `workingDir`, `entrypoint`, `commands`, `env` | Container process — [environment](#container-environment-env) |
| Health | `healthCheck` | Docker-style healthcheck |
| Devices | `devices`, `cdiDevices`, `annotations`, `sysctls` | Hardware / CDI — [sysctls](#container-sysctls) |

#### Container resources (CPU and memory)

| Field | Unit | Semantics |
|-------|------|-----------|
| `cpus` | **cores** (float) | **`1`** = one full CPU, **`1.5`** = one and a half cores. This is **not** the 0–100 utilization scale used in **status.cpuUsage** on describe. |
| `memoryLimit` | **MiB** | Hard memory cap for the container |
| `memoryReservation` | **MiB** | Soft reservation / minimum |
| `memorySwap` | **MiB** | Total memory + swap limit; **`-1`** means **unlimited** swap |
| `shmSize` | **MiB** | Size of **`/dev/shm`** |
| `cpuSetCpus` | cpuset string | CPUs the container may use (e.g. `0-3`, `0,1`) |

#### Container volumes

Each entry in **`container.volumes`** is a **volume mapping** from host-side storage into the container filesystem.

| Field | Required | Description |
|-------|----------|-------------|
| `hostDestination` | **Yes** | Host-side identifier: **absolute host path** (`bind`), [Volume](../../volume/README.md) name (`volume`), or [VolumeMount](../../volume-mount/README.md) name (`volumeMount`) — depends on **`type`** |
| `containerDestination` | **Yes** | Mount path inside the container |
| `accessMode` | **Yes** | **`ro`** (read-only) or **`rw`** (read-write) |
| `type` | No | Mapping kind — see table below |
| `scope` | No | **`private`** or **`shared`** — **only when `type: volume`** (named Volume resource). If omitted → Controller stores **`private`** |

**`type` values:**

| `type` | Meaning |
|--------|---------|
| **`bind`** | Bind-mount a **host path** — **`hostDestination`** is the path on the agent (e.g. `/tmp/data`, `/var/log/myapp`). No `scope`. |
| **`volumeMount`** | Mount a **VolumeMount** resource — **`hostDestination`** is the VolumeMount **`metadata.name`**. On microservice deploy, Controller **auto-attaches** that VolumeMount to **`spec.agent`** ([volume-mount](../../volume-mount/README.md)) |
| **`volume`** | Mount a **`kind: Volume`** on the agent — **`hostDestination`** is the Volume **name**; supports **`scope`** |
| *(omit `type`)* | Treated like a **host path bind** in many deployments (same as **`bind`**). Prefer **`type: bind`** for clarity. |

Other Controller mapping kinds (e.g. **`serviceAccount`**) may appear in API payloads; user deploy YAML most often uses **`bind`**, **`volumeMount`**, and **`volume`**.

**`scope` (named volumes only):**

| `scope` | Behavior |
|---------|----------|
| **`private`** | Default. Volume data is isolated to this microservice instance |
| **`shared`** | Shared named volume — multiple microservices on the same agent can use the same Volume name |

`scope` is **ignored** for **`bind`**, **`volumeMount`**, and omitted-type path binds — only applies when **`type: volume`**.

Example:

```yaml
container:
  volumes:
    - hostDestination: /var/lib/myapp/cache
      containerDestination: /cache
      accessMode: rw
      type: bind
    - hostDestination: nats-creds-operations-center-alert-dashboard
      containerDestination: /etc/nats/creds
      accessMode: ro
      type: volumeMount
    - hostDestination: config
      containerDestination: /app/config
      accessMode: rw
      type: volume
    - hostDestination: nodered-config
      containerDestination: /data
      accessMode: rw
      type: volume
      scope: shared
```

Deploy [**Volume**](../../volume/README.md) / [**VolumeMount**](../../volume-mount/README.md) before the microservice when using named volumes. For **`type: volumeMount`**, Controller **links** the mount to **`spec.agent`** when the microservice is deployed — manual **`attach volume-mount`** is only needed for agents without a referencing microservice ([volume-mount](../../volume-mount/README.md)).

Full example file: [examples/volumes.yaml](examples/volumes.yaml).

#### Container ulimits

**`container.ulimits`** is a map from **limit name** → `{ soft, hard }` (integer values). **`-1`** on soft/hard means **unlimited** (same convention as Docker).

**Allowed keys:**

`core`, `cpu`, `data`, `fsize`, `locks`, `memlock`, `msgqueue`, `nice`, `nofile`, `nproc`, `rss`, `rtprio`, `rttime`, `sigpending`, `stack`

Example:

```yaml
ulimits:
  nofile:
    soft: 65536
    hard: 65536
  nproc:
    soft: 4096
    hard: 4096
```

Controller rejects unknown limit names.

#### Container sysctls

**`container.sysctls`** is a map of **sysctl name → value** (string). Only a **safe allowlist** is permitted on the Controller (not arbitrary kernel tuning).

**Allowed sysctl names:**

| Key |
|-----|
| `kernel.shm_rmid_forced` |
| `net.ipv4.ip_local_port_range` |
| `net.ipv4.tcp_syncookies` |
| `net.ipv4.ping_group_range` |
| `net.ipv4.ip_unprivileged_port_start` |
| `net.ipv4.ip_local_reserved_ports` |
| `net.ipv4.tcp_keepalive_time` |
| `net.ipv4.tcp_fin_timeout` |
| `net.ipv4.tcp_keepalive_intvl` |
| `net.ipv4.tcp_keepalive_probes` |
| `net.ipv4.tcp_rmem` |
| `net.ipv4.tcp_wmem` |
| `net.ipv4.tcp_slow_start_after_idle` |
| `net.ipv4.tcp_notsent_lowat` |

**Validation rules:**

- Any sysctl whose name starts with **`net.`** is **rejected** when **`hostNetworkMode: true`** (container uses the host network stack).
- **IPC-namespaced** sysctl names are **rejected** when **`ipcMode: host`**.

#### Container extraHosts

**`container.extraHosts`** adds **`/etc/hosts`** entries inside the container (Docker **`--add-host`** style).

| Field | Who sets it | Description |
|-------|-------------|-------------|
| `name` | You | Hostname the container resolves (maps to **`/etc/hosts`**) |
| `address` | You | Literal IP/hostname **or** a Controller **template expression** (see below) |
| `value` | Controller | **Resolved** IP written into the microservice spec for the agent that runs the workload |

In deploy YAML you normally set only **`name`** and **`address`**. The Controller fills **`value`**:

- **Literal `address`** (plain IPv4/IPv6 or hostname) — **`value`** equals **`address`**.
- **Template `address`** (`${…}`) — Controller evaluates the expression per deployment context and sets **`value`** to the resolved IP. The template string stays in **`address`**; edgelet uses **`value`** when applying **`--add-host`**.

**Template expressions** (non-exhaustive; Controller defines the grammar):

| Pattern | Meaning |
|---------|---------|
| `${Agents.<agentName>}` | IP of the Agent named **`<agentName>`** |
| `${Apps.<applicationName>.<microserviceName>.local}` | IP of microservice **`<microserviceName>`** in application **`<applicationName>`**, on the **same Agent** as this microservice instance |

Example:

```yaml
extraHosts:
  - name: literalHost
    address: "192.168.0.120" # static IP; value will match address
  - name: agentHost
    address: "${Agents.foo}" # IP of Agent named foo
  - name: appHostLocal
    address: "${Apps.foo.bar.local}" # IP of microservice bar in app foo, co-located on this Agent
```

After deploy, **`describe microservice`** may show both **`address`** (your template or literal) and **`value`** (resolved IP). Do not rely on hand-authoring **`value`** for templates — set **`address`** only.

Full example file: [examples/extra-hosts.yaml](examples/extra-hosts.yaml).

#### Container environment (`env`)

**`container.env`** is a list of environment variables injected into the container at start. Each item sets one **`key`**. The value comes from **exactly one** of three sources (Controller **oneOf**; the CLI YAML parser enforces the same rule).

| Field | Required | Meaning |
|-------|----------|---------|
| `key` | Yes | Container environment variable name |
| `value` | One of three | Literal string (stored as-is) |
| `valueFromSecret` | One of three | Reference: **`secret-name/dataKey`** |
| `valueFromConfigMap` | One of three | Reference: **`configmap-name/dataKey`** |

**Rule:** per entry, set **exactly one** of **`value`**, **`valueFromSecret`**, or **`valueFromConfigMap`**.

Reference format is **`resourceName/dataKey`**: the **`Secret`** or **`ConfigMap`** must already exist in the same ECN namespace as the microservice (`metadata.namespace`), and the named **data key** must be present in that resource’s top-level **`data`** map (deploy YAML shape for those kinds).

**Literal:**

```yaml
- key: BASE_URL
  value: "http://localhost:8080/data"
```

**From Secret:**

```yaml
- key: API_TOKEN
  valueFromSecret: my-secret/token
```

**From ConfigMap:**

```yaml
- key: CONFIG_JSON
  valueFromConfigMap: app-config/settings
```

**Combined example:**

```yaml
container:
  env:
    - key: LOG_LEVEL
      value: debug
    - key: DB_PASSWORD
      valueFromSecret: postgres-creds/password
    - key: FEATURE_FLAGS
      valueFromConfigMap: platform-config/features
```

Deploy [**Secret**](../../secret/README.md) / [**ConfigMap**](../../configmap/README.md) in the same namespace **before** (or in the same multi-doc file as) the microservice that references them.

**Typical Controller errors:**

| Situation | HTTP |
|-----------|------|
| No `value`, `valueFromSecret`, or `valueFromConfigMap` | **400** |
| More than one of those fields set | **400** |
| Malformed `secret-name/dataKey` or `configmap-name/dataKey` | **400** |
| Unknown Secret/ConfigMap or missing data key | **404** / **400** |

Full example file: [examples/env.yaml](examples/env.yaml).

### NATS

Requires [Application](../application/schema.md) **`spec.natsConfig.natsAccess: true`** for the same **`spec.application`**.

| Field | Type | Description |
|-------|------|-------------|
| `natsConfig` | object | Per-microservice NATS |
| `natsConfig.natsAccess` | bool | Enable NATS — Controller creates user JWT, creds secret, **volumeMount** at `/etc/nats/creds`, and **`NATS_CREDS_PATH`** / **`NATS_SERVER_URL`** (do not hand-author creds mounts when using this flag) |
| `natsConfig.natsRule` | string | [NatsUserRule](../../nats-user-rule/README.md) **`metadata.name`**. Omit when `natsAccess: true` → **`default-user`**. |
| `natsEnabled` | bool | Alias of **`natsConfig.natsAccess`** (microservice YAML only) |

Account JWT policy is [NatsAccountRule](../../nats-account-rule/README.md) on the application only. Changing **`natsRule`** rotates the user key and overwrites the creds file at the same path.

Runtime and lifecycle: [NATS access](../nats-access/README.md) · [runtime](../nats-access/runtime.md) · [lifecycle](../nats-access/lifecycle.md).

### Fleet catalogs

| Field | Type | Description |
|-------|------|-------------|
| `models` | object | Bind [Model](../../model/README.md) artifacts |
| `models.bindPath` | string | Mount path in container |
| `models.permissions` | string | **`ro`** or **`rw`** |
| `models.items` | [] | `{ name: <Model metadata.name> }` |
| `knowledge` | object | Bind [Knowledge](../../knowledge/README.md) |
| `knowledge.bindPath` | string | Mount path |
| `knowledge.permissions` | string | **`ro`** or **`rw`** |
| `knowledge.items` | [] | `{ name: <Knowledge metadata.name> }` |

Deploy fleet resources and **attach** models/knowledge to agents before expecting Ready mounts.

### Template

| Field | Type | Description |
|-------|------|-------------|
| `template.name` | string | [MicroserviceTemplate](../microservice-template/schema.md) name |
| `template.variables` | map or list | Instance variables |

### RBAC (application-scoped)

| Field | Type | Description |
|-------|------|-------------|
| `serviceAccount.roleRef.kind` | string | **`Role`** |
| `serviceAccount.roleRef.name` | string | Role name in same application |

Deploy [Role](../../role/README.md), [RoleBinding](../../role-binding/README.md), and [ServiceAccount](../../service-account/README.md) when using full app-scoped RBAC. **`spec.application`** must match **`ServiceAccount` `metadata.applicationName`**. Example stack: [service-account/examples/with-microservice.yaml](../../service-account/examples/with-microservice.yaml).

Inline **`serviceAccount.roleRef`** alone references a **Role** by name; combine with RoleBinding + ServiceAccount kind when Controller requires explicit SA objects.

### Status (describe only)

`describe microservice` adds **`status`** and **`execStatus`** (not for deploy YAML):

| Block | Examples |
|-------|----------|
| `status` | `status`, `containerId`, `cpuUsage`, `memoryUsage`, `healthStatus`, `restartCount`, … |
| `execStatus` | Exec session state |

---

## Initial deploy (create)

1. Ensure **Application** exists (or deploy Application in same file **before** Microservice in manual ordering).
2. Ensure **agent** exists and arch matches images.
3. Provide **`images`** or **`catalogId`**, **`agent`**, **`container.ports`** as required by your stack.
4. **`deploy -f`** without patch flags.

**Example:** [examples/minimal.yaml](examples/minimal.yaml)

---

## Re-deploy (update / patch)

| Method | When |
|--------|------|
| Full **`deploy -f`** | Change image, container, agent, template, etc. |
| **`--patch-model`** | Change only **`spec.models`** (bind path, permissions, item list) |
| **`--patch-knowledge`** | Change only **`spec.knowledge`** |

Catalog PATCH may or may not rebuild the container — **Controller decides** (empty↔non-empty catalog, bind path, permissions vs item-only changes).

**Examples:** [examples/patch-models.yaml](examples/patch-models.yaml)

---

## Commands

```bash
potctl deploy -f microservice.yaml -n my-ecn
potctl deploy -f microservice.yaml --patch-model -n my-ecn
potctl describe microservice myapp/api -n my-ecn
potctl delete microservice myapp/api -n my-ecn
potctl rebuild microservice myapp/api -n my-ecn
```

Debugging: [debugging.md](../debugging.md).

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | App + agent + images + ports |
| [wasm.yaml](examples/wasm.yaml) | WASM platform + runtime |
| [template-and-models.yaml](examples/template-and-models.yaml) | Template ref + fleet models |
| [patch-models.yaml](examples/patch-models.yaml) | `--patch-model` file |
| [volumes.yaml](examples/volumes.yaml) | volumeMount + volume + shared scope |
| [extra-hosts.yaml](examples/extra-hosts.yaml) | literal + `${Agents…}` + `${Apps….local}` |
| [env.yaml](examples/env.yaml) | literal + Secret + ConfigMap refs |

---

## Appendix: container fields

| Field | Type | Description |
|-------|------|-------------|
| `hostNetworkMode` | bool | Use host network stack |
| `isPrivileged` | bool | Privileged container |
| `runAsUser` | string | UID |
| `runAsGroup` | string | GID |
| `readOnlyRootFilesystem` | bool | Read-only root FS |
| `ipcMode` | string | IPC mode (engine-specific) |
| `pidMode` | string | PID namespace mode |
| `platform` | string | OCI platform (e.g. `wasi/wasm`) |
| `runtime` | string | Runtime handler ([RuntimeClass](../../runtime-class/README.md)) |
| `capAdd` | []string | Added capabilities |
| `capDrop` | []string | Dropped capabilities |
| `annotations` | object | Arbitrary annotations map |
| `sysctls` | map[string]string | [Allowlisted sysctls](#container-sysctls) |
| `ulimits` | map[string]object | [Allowed keys](#container-ulimits): `{ soft, hard }`; `-1` = unlimited |
| `cpuSetCpus` | string | cpuset CPUs |
| `cpus` | float | CPU limit in **cores** (`1` = one CPU) |
| `memoryLimit` | int64 | **MiB** |
| `memoryReservation` | int64 | **MiB** |
| `memorySwap` | int64 | **MiB**; **`-1`** = unlimited |
| `shmSize` | int64 | **MiB** (`/dev/shm`) |
| `cdiDevices` | []string | CDI device IDs |
| `devices` | []object | `{ hostPath, containerPath, permissions }` |
| `volumes` | []object | [Volume mappings](#container-volumes) |
| `tmpfs` | []object | `{ containerPath, size (MiB), mode }` |
| `extraHosts` | []object | [extraHosts](#container-extrahosts): `name`, `address` (you); `value` (Controller-resolved IP) |
| `env` | []object | [Environment](#container-environment-env): `key` + one of `value`, `valueFromSecret`, `valueFromConfigMap` |
| `ports` | []object | `{ internal, external, protocol, … }` |
| `workingDir` | string | Working directory |
| `entrypoint` | []string | Override entrypoint |
| `commands` | []string | Command args |
| `healthCheck.test` | []string | Healthcheck command |
| `healthCheck.interval` | int64 | Interval ms |
| `healthCheck.timeout` | int64 | Timeout ms |
| `healthCheck.retries` | int | Retries |
| `healthCheck.startPeriod` | int64 | Start period ms |
| `healthCheck.startInterval` | int64 | Start interval ms |

Use **`potctl describe microservice App/Msvc -o file.yaml`** for authoritative round-trip of fields your Controller version supports.
