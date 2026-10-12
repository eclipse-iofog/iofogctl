# Controller add-on (`kind: Controller`)

Use a standalone **`Controller`** document to add **one new host** to an existing **remote** control plane without redeploying the full **`ControlPlane`** YAML. The CLI reuses the **stored global spec** (auth, database, images, TLS policy, system microservices) from namespace config and runs the **next-system-agent** path on the new host.

## YAML shape

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Controller
metadata:
  name: ctrl-2                  # must be unique in stored CP
spec:
  host: 10.0.0.12
  ssh:
    user: ubuntu
    keyFile: ~/.ssh/id_rsa
    port: 22
  systemAgent:
    config:
      arch: arm64
  tls:                          # optional — overrides stored global spec.tls for this host
    cert: <base64>
    key: <base64>
  airgap: false
```

Parse alias: documents labeled **`RemoteController`** are accepted as **`Controller`**.

**`spec`** fields match **`RemoteController`**: `host`, `ssh`, `systemAgent`, optional `tls`, optional `airgap`, optional `endpoint` (set by CLI after deploy).

## Prerequisites

The CLI namespace must already contain a deploy-capable **`ControlPlane`** spec:

| Requirement | Reason |
|-------------|--------|
| Prior **`deploy -f`** with full `ControlPlane`, **or** **`connect -f`** with full YAML | Stores `auth`, `controller`, database, etc. |
| **`SupportsControllerAddOn()`** passes | Rejects thin manual connect stubs missing `auth` or `controller` |
| New **`metadata.name`** not already in `controllers[]` | |
| New **`host`** not already used | |
| External **database** if adding second+ controller | SQLite-only CP cannot grow to multi-controller |

## What the add-on runs

**Per new host only** (same as initial CP host pipeline for one row):

1. SSH + Edgelet install (and airgap/registry when applicable).
2. Translate stored global CP + **this** controller → Edgelet ControlPlane manifest.
3. **`edgelet deploy -f`** on the new host.

**Then:**

4. **`DeployNextSystemAgent`** — system agent with **`upstreamRouters: [default-router]`** and **`upstreamNatsServers: [default-nats-hub]`** (defaults merged if you omit upstream lists in config).
5. Append controller to stored **`spec.controllers[]`** and flush config.

## What the add-on skips

These ran during initial **`ControlPlane`** deploy and are **not** repeated:

| Step | Skipped on add-on |
|------|-------------------|
| `WaitForControllerAPI` | Yes |
| `EnsureIofogUserEmbedded` | Yes |
| SDK registry for private `controller.package` | Yes |
| **`DeployGlobalCertificates`** | Yes — CAs must already exist from initial deploy or manual import |

If you need to rotate global messaging CAs, update them on the Controller API or redeploy the full **`ControlPlane`** with new CA blocks before adding hosts that depend on new trust material.

## Deploy order in multi-document files

When a single file contains both **`ControlPlane`** and **`Controller`** documents:

- Control plane executors run first.
- **Controller** add-on executors run after the CP block.
- Validation rejects **duplicate name or host** between `spec.controllers[]` and standalone **`Controller`** entries in the **same file**.

```bash
potctl deploy -f add-controller.yaml -n prod
```

## Multi-document example

```yaml
---
# Optional: only if updating global spec in same apply — usually omit on add-on-only files
apiVersion: datasance.com/v3
kind: Controller
metadata:
  name: edge-west-2
spec:
  host: 203.0.113.20
  ssh:
    user: ec2-user
    keyFile: ~/.ssh/prod.pem
  systemAgent:
    config:
      arch: amd64
      host: 203.0.113.20
```

## HA role of added controllers

Each added host runs its **own Controller container** sharing the **same external database** as existing hosts. The new system agent is always a **next** agent (index ≥ 1): it attaches to **`default-router`** and **`default-nats-hub`** on the **first** controller host from the original deploy order.

Adding a controller **does not** reorder the first host; plan the initial `controllers[]` list with the primary router/NATS host at **index 0**.

See [multi-controller-ha.md](multi-controller-ha.md).

## Troubleshooting

| Error | Cause |
|-------|--------|
| does not support adding controllers | Namespace CP missing full deploy/connect YAML |
| external database is required | Second controller while DB provider empty |
| name/host already exists | Duplicate in stored CP |
| collision in same deploy file | CP `controllers[]` and `Controller` share name or host |

## Related

- [schema.md](schema.md) — global vs per-host fields
- [securing-cluster.md](securing-cluster.md) — per-host `tls` override
