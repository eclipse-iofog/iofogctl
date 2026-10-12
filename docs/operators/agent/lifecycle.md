# Agent lifecycle and related deploy kinds

Operator reference for **potctl** / **iofogctl** commands that change or clean up agents after initial **`deploy -f`**. Deploy and AgentConfig field reference live under [remote/](remote/README.md), [local/](local/README.md), and [agent-config/](agent-config/README.md).

**Applies to v3.9.x.**

---

## Overview

| Command | Controller API | Edgelet on host | Namespace config (`~/.iofog/v3`) |
|---------|----------------|-----------------|----------------------------------|
| `deploy -f` (Agent / LocalAgent) | Create/update via AgentConfig path | Bootstrap + provision | Agent record stored |
| `deploy -f` (AgentConfig only) | Create/update agent | No | No agent host record unless paired |
| `delete agent` | Delete agent (if reachable) | Deprovision + uninstall (`--remove-data`) | Remove agent; trim volumes |
| `detach agent` | Delete agent | Deprovision only (no uninstall) | Move to **detached** store |
| `attach agent` | Recreate + provision | Uses stored SSH/local spec | Attach to namespace |
| `prune agent` | Prune API | Host cleanup per agent type | Unchanged |
| `upgrade agent` | Schedule node upgrade | Async on agent | Unchanged |
| `reconcile agent` | Platform reconcile | Async | Unchanged |
| `configure agent` | No | No | Update **SSH cache only** (remote) |
| `move agent` | Via detach + attach | Same as detach/attach | Namespace change |
| `deploy -f` (Volume) | No | Mount paths on listed agents | Volume record |
| `deploy -f` (OfflineImage) | No | Load images on listed agents | Offline image record |

**System agents** on control plane hosts are torn down via **control plane / controller delete**, not via ordinary `delete agent` on reserved system names. See [README — system agents](README.md#system-agents-control-plane-hosts).

---

## delete agent

```bash
potctl delete agent NAME [-n NAMESPACE] [--force] [--detached]
```

### Attached agent (default)

1. Sync agent list from Controller into namespace config.
2. Unless **`--force`**: refuse if any microservice is still scheduled on the agent UUID.
3. **Deprovision** edgelet on the host (best-effort; errors logged).
4. **Delete agent** on Controller API (when client and UUID available).
5. **Uninstall** edgelet with data removal (`uninstall.sh --remove-data` path).
6. Remove agent from namespace config.
7. **Volumes:** remove volume entries that listed only this agent; remove agent name from multi-agent volumes.

### Remote agent without SSH in config

If `host`, `user`, or `keyFile` are missing, deprovision/uninstall are **skipped** with a notify message. Use **`configure agent`** to restore SSH details, then retry delete.

### Local agent

Same flow on the local host (no SSH). Volume deploy to local agents is limited (see [Volumes](#volumes-kind-volume)).

### `--detached`

Deletes a **detached** agent record from global detached storage only (no Controller or host operations).

### `--force`

Skips the microservice check. Use when the Controller still lists microservices but the agent must be removed.

---

## detach agent

```bash
potctl detach agent NAME [-n NAMESPACE] [--force]
```

Detach **removes the agent from the ECN** while preserving enough state to reattach later.

1. Refuse if an agent with the same name is already in the detached store.
2. Unless **`--force`**: refuse if microservices run on the agent.
3. **Deprovision** edgelet on the host (local or remote).
4. **Delete agent** on the Controller API.
5. Copy agent spec from namespace config to **detached** storage, including **AgentConfig** snapshot from Controller.
6. Remove agent from namespace config.

**Does not** run edgelet **uninstall** on the host—binaries and data may remain until delete or manual cleanup.

---

## attach agent

```bash
potctl attach agent NAME [-n NAMESPACE]
```

1. Load agent from **detached** storage (must exist).
2. Push **AgentConfig** to Controller (create agent record).
3. Run agent deploy executor path: provision edgelet with Controller (SSH/local from stored spec).
4. Add agent to target namespace config; remove from detached store.

Attach assumes edgelet is still installed from before detach. If the host was wiped, run **`deploy -f`** with a fresh `Agent` / `LocalAgent` manifest instead.

---

## move agent

```bash
potctl move agent NAME DEST_NAMESPACE [-n SOURCE_NAMESPACE] [--force]
```

Equivalent to **`detach agent`** in the source namespace (honors **`--force`**) then **`attach agent`** in the destination namespace. Cache is invalidated when source and destination namespaces are the same.

---

## prune agent

```bash
potctl prune agent NAME [-n NAMESPACE] [--detached]
```

Controller-driven **prune** operation plus host-side cleanup:

- **Remote attached:** prune via Controller API and remote host helpers.
- **Local:** local prune path.
- **`--detached`:** prune using detached agent record (remote detached path).

Use when the agent daemon should reclaim resources without full delete. Does not remove the agent from namespace config by itself (see CLI help for your build).

---

## upgrade agent

```bash
potctl upgrade agent NAME [-n NAMESPACE] [--semver VERSION]
```

1. Sync namespace agent cache from Controller.
2. Call Controller **`UpgradeNode`** for the agent UUID.
3. Optional **`--semver`** pins target version (mapped via CLI node version helpers).

Upgrade is **asynchronous** on the agent. Poll status with **`describe agent`** (`isReadyToUpgrade`, version fields).

This upgrades the **agent node software** via Controller orchestration—not the same as redeploying edgelet bootstrap with a new `package.version` in YAML.

---

## reconcile agent

```bash
potctl reconcile agent NAME [-n NAMESPACE]
```

Enqueues a manual **platform reconcile** on the Controller and waits until provisioning completes (`ReconcileAgentByNameAndWait`).

Use after Controller-side config changes or platform drift when the agent should re-apply router/NATS/platform state.

---

## configure agent

```bash
potctl configure agent NAME [-n NAMESPACE] \
  [--user USER] [--key KEYFILE] [--port PORT] [--detached]
```

**Remote agents only.** Updates **SSH fields** in namespace or detached config:

| Flag | Field |
|------|--------|
| `--user` | `ssh.user` |
| `--key` | `ssh.keyFile` (path expanded) |
| `--port` | `ssh.port` |

Does **not** change **AgentConfig** on the Controller. Use **`deploy -f`** with `kind: AgentConfig` or agent `spec.config` for runtime limits, router, or NATS settings.

**LocalAgent** cannot be configured via this command.

---

## describe agent / describe agent-config

```bash
potctl describe agent NAME [-n NAMESPACE]
potctl describe agent-config NAME [-n NAMESPACE]
```

- **`describe agent`:** merged namespace spec (SSH, package, config) plus live **status** from Controller when connected.
- **`describe agent-config`:** Controller configuration and status oriented toward AgentConfig shape (nested `routerConfig` / `natsConfig` in output).

Useful for verifying round-trip YAML and platform readiness after deploy.

---

## Volumes and offline images

| Kind | Full reference |
|------|----------------|
| **`Volume`** | [volume/README.md](../volume/README.md) — SSH staging + `~/.iofog/v3` config (not Controller API) |
| **`OfflineImage`** | [offline-image/README.md](../offline-image/README.md) — remote agents only, catalog id **2** |
| **`VolumeMount`** | [volume-mount/README.md](../volume-mount/README.md) — Controller object + **`attach volume-mount`** |

**Delete agent** still trims **Volume** records in namespace config (drop agent from lists or remove single-agent volumes). See [volume README](../volume/README.md).

---

## exec, attach/detach exec

Agent **debug exec** and **log streaming** (`logs agent`, `exec agent`, `attach exec agent`, `detach exec agent`) are documented in [debugging.md](debugging.md), including auto-provision behavior, local vs remote `logs agent`, and RBAC notes.

Microservice **`logs microservice`** / **`exec microservice`** are documented in [workloads/debugging.md](../workloads/debugging.md).

---

## Related reading

- [Deploy sequence](README.md#deploy-sequence)
- [Edgelet bootstrap](edgelet-bootstrap.md)
- [Control plane delete](../controlplane/remote/README.md) (system agent teardown)
