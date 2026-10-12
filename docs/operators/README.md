# Operator documentation

Human-written guides for deploying and operating PoT / ioFog resources with **potctl** and **iofogctl**.

| Document | Description |
|----------|-------------|
| [Control plane (Kubernetes, local, remote)](controlplane/README.md) | Deploy YAML kinds, operator reconcile, hub router/NATS, TLS, and multi-host HA for v3.9.x |
| [Agents (remote, local, AgentConfig)](agent/README.md) | Edge agent deploy YAML, Controller agent configuration, lifecycle, debugging (`logs` / `exec`), and edgelet bootstrap |
| [Agent networking (router + NATS fabric)](agent/networking/README.md) | System `router` / `nats` microservices from declarative agent config |
| [Edgelet bootstrap (relocated)](agent/edgelet-bootstrap.md) | Layered bootstrap pipeline, platform matrix, `scripts` YAML, and merge logic |
| [Secret](secret/README.md) | Controller secrets, microservice `valueFromSecret`, CA import material |
| [ConfigMap](configmap/README.md) | Controller config maps, `valueFromConfigMap`, VolumeMount backing store |
| [Certificate authority](certificate-authority/README.md) | ECN CA catalog (PKI): self-signed, direct, k8s-secret |
| [Certificate](certificate/README.md) | Leaf TLS certificates signed by a CA |
| [Volume](volume/README.md) | Stage host directories onto agents (CLI namespace config + SSH) |
| [Volume mount](volume-mount/README.md) | Mount Secret/ConfigMap on agents (`attach volume-mount`) |
| [Offline image](offline-image/README.md) | Airgap image pull/transfer + offline catalog registration |
| [Registry (Controller)](registry/README.md) | Image and artifact pull registries (`oci` / `hf`), create vs update by `spec.id` |
| [Catalog item](catalog-item/README.md) | Multi-arch catalog images for microservices |
| [Model (fleet)](model/README.md) | Registry-backed model artifacts, attach to agents, microservice catalog bind |
| [Knowledge (fleet)](knowledge/README.md) | Document and dataset artifacts, attach, microservice catalog bind |
| [RuntimeClass (Controller)](runtime-class/README.md) | WASM/runtime handlers, link to edgelet agents |
| [Role](role/README.md) | ECN RBAC rules |
| [Role binding](role-binding/README.md) | Grant roles to ServiceAccount subjects |
| [Service account](service-account/README.md) | Application-scoped identity + `roleRef` |
| [NATS account rule](nats-account-rule/README.md) | Account JWT templates → bind on **Application** `natsConfig.natsRule` |
| [NATS user rule](nats-user-rule/README.md) | User JWT templates → bind on **Microservice** `natsConfig.natsRule` |
| [NATS access (workloads)](workloads/nats-access/README.md) | Creds mount, env, resolver, app/ms lifecycle after `natsAccess: true` |
| [Service (TCP interconnection)](service/README.md) | Router `tcpConnector` / `tcpListener`; optional K8s Service — not a Microservice |
| [Workloads (apps and microservices)](workloads/README.md) | ApplicationTemplate, MicroserviceTemplate, Application, Microservice, deploy flags, logs/exec |

Generated CLI reference lives under [`../potctl_md/`](../potctl_md/) and [`../iofogctl_md/`](../iofogctl_md/).

## ECN resource deploy order

When one YAML file defines several kinds, the CLI sorts documents using **`kindOrder`** in the deploy executor (not file order). Abbreviated flow for operator planning:

```text
Secret → CertificateAuthority → Certificate → ConfigMap
→ Role → RoleBinding → ServiceAccount → NatsAccountRule → NatsUserRule
→ Agent | LocalAgent → ApplicationTemplate
→ Volume → OfflineImage → VolumeMount
→ Registry → Model → Knowledge → RuntimeClass → MicroserviceTemplate → CatalogItem
→ Application → Microservice → Service
```

Practical rules:

- **Secret / ConfigMap** before **VolumeMount** and microservice env references.
- **CertificateAuthority** before **Certificate**.
- **Agents** before **Volume**, **OfflineImage**, and **`attach volume-mount`**.
- **Registry** before **Model**, **Knowledge**, and standalone **CatalogItem** (OfflineImage registers catalog id **2** itself).
- **Catalog / offline bundle** before **Microservice** that references **`catalogItem.name`**.
- **Role → RoleBinding → ServiceAccount** before microservices using RBAC; **`applicationName`** must match **Application** (same or prior apply).
- **NatsAccountRule** before Application **`natsConfig.natsAccess: true`**; **NatsUserRule** before microservice **`natsConfig.natsAccess: true`** (account rule on Application, user rule on Microservice).
- **Service** after **Microservice** (and agents tagged for [distribution](service/distribution.md)) when publishing a TCP bridge to a backend.

Global Router/NATS CAs on control plane deploy are separate from ECN **`CertificateAuthority`** — see [control plane securing cluster](controlplane/remote/securing-cluster.md).

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
| CLI config directory | `~/.iofog/v3` | `~/.iofog/v3` |

Examples in this tree often use Datasance paths; substitute your flavor’s `apiVersion` and image registry.
