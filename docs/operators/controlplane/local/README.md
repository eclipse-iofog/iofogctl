# Local control plane

Use **`kind: LocalControlPlane`** to run a **single-node** control plane on the **same machine** as the CLI via **Edgelet**. This path is intended primarily for **development and automated tests**, not as a substitute for Kubernetes or remote multi-host production layouts.

## What gets deployed

1. Edgelet is installed on the local host (layered bootstrap).
2. Optional airgap image/binary staging.
3. Optional private registry manifest on Edgelet when `spec.controller.package` includes registry credentials.
4. CLI translates your YAML to **`edgelet.iofog.org/v1`** **`kind: ControlPlane`** and runs **`edgelet deploy -f`**.
5. CLI waits for the Controller API, runs embedded-auth user setup, optional Controller SDK registry for private images.
6. **System agent** is installed with `isSystem: true`, configured, and provisioned. That node becomes the **default router** and **NATS hub** the same way as a remote control plane — [remote networking](../remote/networking.md). Kubernetes hubs are a different path ([kubernetes networking](../kubernetes/networking.md)).
7. Namespace config and optional **`spec.ca`** trust are persisted.

Runtime behavior of the Edgelet manifest (ports, volumes, env mapping, registration) is defined in [Edgelet control plane documentation](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md).

## Prerequisites

- Local host supports Edgelet (native or container engine per agent config).
- **`spec.systemAgent.config.arch`** is **required** (valid values: `amd64`, `arm64`, `arm`, `riscv64`, `auto` where supported).
- Port **54321** must not be in use by a conflicting local agent when using the default system agent layout.

## Deploy

```bash
potctl deploy -f local-controlplane.yaml -n default
```

## Translation

| User YAML | Edgelet manifest |
|-----------|------------------|
| `apiVersion: datasance.com/v3` / `iofog.org/v3` | `apiVersion: edgelet.iofog.org/v1` |
| `kind: LocalControlPlane` | `kind: ControlPlane` |
| `metadata.name` / `-n` | `metadata.name` / `metadata.namespace` on Edgelet CR |

Fields **not** copied to Edgelet include CLI-only **`spec.ca`**, **`spec.iofogUser`**, and global **`routerSiteCA`** / **`natsSiteCA`** blocks (see [securing-cluster.md](securing-cluster.md)).

## Router / NATS CA parity (important)

`routerSiteCA`, `routerLocalCA`, `natsSiteCA`, and `natsLocalCA` are **validated** in local YAML but the CLI **does not** upload them to the Controller API on local deploy today. **Remote** `ControlPlane` deploy does upload them when set.

**Intended behavior:** a future release will give **LocalControlPlane** the same global CA deploy step as remote. Until then, use remote CP documentation for BYO messaging CAs, or import CAs via Controller REST after local deploy (Edgelet rejects inline site/local CA in the manifest).

## Related documents

- [Remote networking](../remote/networking.md) — system-agent hub (same model as local)
- [schema.md](schema.md)
- [securing-cluster.md](securing-cluster.md)
- [examples/minimal.yaml](examples/minimal.yaml)
