# Kubernetes control plane

Use **`kind: KubernetesControlPlane`** to install the ioFog operator in a Kubernetes namespace, apply a **`ControlPlane`** custom resource, wait until the stack is ready, and register the CLI user against the Controller API.

## Two-layer model

| Layer | `kind` | Purpose |
|-------|--------|---------|
| **CLI deploy YAML** | `KubernetesControlPlane` | What you author in `deploy -f` |
| **Cluster custom resource** | `ControlPlane` | What the operator reconciles |

The CLI **translates** most of `spec` into the operator CR. Fields that never appear on the CR are **CLI-only** (kubeconfig, trust CA, CLI user, operator image).

The operator watches the translated **`ControlPlane`** CR, creates Controller, Router, and NATS, and registers the hub. See [How the operator works](operator.md) and [Kubernetes networking](networking.md). Field-level CR mapping stays in [schema.md](schema.md).

## Prerequisites

- `kubectl` access via `spec.config` (kubeconfig path)
- CLI namespace (`-n`) equals the **Kubernetes namespace** where the operator and CR live
- If both `metadata.namespace` and `-n` are set, they **must match**

## Deploy sequence (what the CLI does)

1. Parse and validate [schema](schema.md) rules (auth, database vs replicas, ingress when ClusterIP, NATS replicas, vault).
2. Install the **operator** Deployment, ServiceAccount, Role, and RoleBinding in the target namespace (direct Kubernetes API — not Helm in the default CLI path).
3. Apply or update the **`ControlPlane`** CR with fixed name **`pot`** (Datasance) or **`iofog`** (Eclipse).
4. Wait until the ControlPlane reports **ready** (operator reconciles Router, NATS if enabled, and Controller).
5. Resolve the **Controller API endpoint** (LoadBalancer, Ingress named **`controller`**, or `spec.controller.publicUrl`).
6. If `spec.ca` is set, store it in the CLI namespace **trust store** (`~/.iofog/v3/trust/<namespace>/`).
7. Wait until the Controller API accepts connections.
8. If `auth.mode: embedded`, ensure the **`iofogUser`** exists in the Controller (bootstrap login; create user if missing).
9. Persist control plane and discovered controller pods in namespace config.

## CLI-only vs operator CR

| YAML field | On operator CR? |
|------------|-----------------|
| `spec.config` | No — local kubeconfig path |
| `spec.ca` | No — CLI trust for HTTPS API |
| `spec.iofogUser` | No — CLI login identity |
| `spec.images.operator` | No — see [schema — spec.images](schema.md#specimages) |
| `spec.auth`, `database`, `controller`, `events`, `services`, `ingresses`, `replicas`, `images` (`pullSecret`, controller/router/nats), `nats`, `vault` | Yes (translated) |

After deploy, the CLI may record **`controllerPods`** in stored config (discovered pod names). That block is output/state, not part of your deploy file.

## Related documents

- [operator.md](operator.md) — status machine, Controller reconcile, what the operator does not deploy
- [networking.md](networking.md) — default router, NATS hub, CA import, ConfigMaps Controller patches
- [schema.md](schema.md) — field-by-field reference
- [securing-cluster.md](securing-cluster.md) — Controller, Router, NATS TLS and CLI trust
- [examples/](examples/) — sample manifests

## Example (minimal)

| Example | Pattern |
|---------|---------|
| [examples/ingress.yaml](examples/ingress.yaml) | **B1** — Ingress TLS, Controller HTTP |
| [examples/controller-ingress.yaml](examples/controller-ingress.yaml) | **B2** — shared Ingress + pod Secret, re-encrypt |
| [examples/minimal.yaml](examples/minimal.yaml) | LoadBalancer-oriented minimal spec |

```bash
potctl deploy -f kubernetes-controlplane.yaml -n iofog
```
