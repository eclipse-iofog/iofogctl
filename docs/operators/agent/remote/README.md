# Remote agent (`kind: Agent`)

Use **`kind: Agent`** to install **edgelet** on a **remote host** over SSH and register the node with the Controller in the target namespace.

## What gets deployed

1. Namespace must already contain a **control plane** with a reachable Controller API.
2. **AgentConfig** is applied first (explicit document or synthesized from `spec.config`) — see [agent-config/README.md](../agent-config/README.md).
3. CLI resolves `package`, airgap, and WASM scope, merges optional **`spec.scripts`**, and runs the [edgelet bootstrap](../edgelet-bootstrap.md) on the host.
4. If `spec.config` is set, AgentConfig may be updated again after bootstrap (facade executor).
5. CLI runs **provision**: `edgelet config --a <controller>` and `edgelet provision <key>`.
6. Agent spec (SSH, UUID, timestamps) is stored under `~/.iofog/v3` for the namespace.

## Prerequisites

- Passwordless **sudo** on the remote host for bootstrap (see edgelet bootstrap doc).
- SSH: **`spec.host`**, **`spec.ssh.user`**, **`spec.ssh.keyFile`** required (port defaults to **22**).
- **`spec.config.arch`** required when using airgap, WASM, or arch-sensitive packages (recommended always).
- Agent **name** must be lowercase alphanumeric; **`default-router`** is reserved.

## Deploy

```bash
potctl deploy -f remote-agent.yaml -n my-ecn
```

## Controller endpoint

| Source | Used for |
|--------|----------|
| `spec.controllerEndpoint` | Provision URL when set |
| Control plane in namespace | Default endpoint from stored CP |

## Same host as Controller

The CLI rejects deploying a **non-system** remote **Agent** on the same hostname as the Controller **public URL**, except the reserved local agent name used for vanilla local scenarios. Do not colocate edge workloads on the controller host without using control plane **systemAgent** semantics. See [system agents](../README.md#system-agents-control-plane-hosts).

## Custom install scripts

Set **`spec.scripts`** — full reference in [edgelet-bootstrap.md](../edgelet-bootstrap.md#yaml-reference-scripts-block).

## Related documents

- [schema.md](schema.md) — field reference
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/with-airgap.yaml](examples/with-airgap.yaml)
- [examples/with-custom-scripts.yaml](examples/with-custom-scripts.yaml)
- [lifecycle.md](../lifecycle.md)
