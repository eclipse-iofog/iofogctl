# Service (`kind: Service`)

**Service** publishes a **TCP port through the router**. It is **not** a [Microservice](../workloads/microservice/README.md) and it is **not** the Kubernetes `Service` block inside [control plane](../controlplane/README.md) YAML — though on a **Kubernetes control plane**, Controller may **create** a Kubernetes `Service` in the controller namespace for some types.

Controller turns one Service row into:

1. A **`tcpConnector`** on the router that can reach the backend.
2. A **`tcpListener`** on the hub router, and on every Edgelet agent whose tags **select** this service.
3. When the control plane is **Kubernetes** and `spec.type` is **`microservice`**, **`agent`**, or **`external`**, a Kubernetes `Service` that forwards **`servicePort`** to the assigned **`bridgePort`** on router pods.

Create returns immediately with **`provisioningStatus: pending`**. **potctl deploy** waits until hub provisioning is **`ready`** or **`failed`** (and may call **`reconcile service`** once on failure). **`ready`** means the hub connector, hub listener, and (when required) the Kubernetes Service succeeded. **Edge listeners** on tagged agents reconcile separately on the fog platform.

## Service types

| `spec.type` | `spec.resource` (you write) | Connector lives on | K8s `Service` (K8s CP only) |
|-------------|----------------------------|--------------------|-------------------------------|
| **`microservice`** | `appName/microserviceName` or microservice UUID | Agent that runs the microservice | Yes — needs `servicePort` + `k8sType` |
| **`agent`** | Agent name or UUID | That agent | Yes |
| **`k8s`** | Existing cluster DNS name or IP | Default router | **No** (type allowed only on K8s CP) |
| **`external`** | DNS or IP outside the cluster | Default router | Yes |

**`microservice`** and **`agent`** require a **router** on the backend agent. Router mode **`none`** is rejected.

**`k8s`** and **`external`** require **`spec.defaultBridge: default-router`**. For **`microservice`** / **`agent`**, **`defaultBridge`** may instead be the UUID of an upstream router this agent’s router is already connected to.

## Control plane: remote vs Kubernetes

| Control plane | `servicePort` / `k8sType` in YAML | `serviceEndpoint` |
|---------------|-----------------------------------|-------------------|
| **Remote** (not `kubernetes`) | Omit — Controller sets `servicePort` to assigned **`bridgePort`** | Hub router host |
| **Kubernetes** | **Required** for `microservice`, `agent`, `external` | **`LoadBalancer`**: ingress IP or hostname; ClusterIP / NodePort do not populate from K8s |

See [schema.md](schema.md) and [examples/](examples/).

## What potctl deploys

1. Controller reachable in the namespace.
2. Unmarshal **`spec`**; send **`metadata.tags`** on create/update (**tags are not under `spec`**).
3. **`CreateService`** or **`UpdateService`**, then wait for hub **`ready`**.

Identity: **`metadata.name`** — DNS label pattern and reserved names are enforced by Controller (see [schema.md](schema.md)).

**Immutable after create:** **`spec.type`**, **`spec.defaultBridge`**. Change them by delete + recreate. **`bridgePort`** is assigned once from `BRIDGE_PORTS_RANGE` (default `10024–65535`) and is **not** accepted in YAML.

## Prerequisites

- **`type: microservice`:** Application and microservice exist; backend agent has a router (not `none`). Router fabric: [Agent router-fabric](../agent/networking/router-fabric.md).
- **`type: agent`:** Target agent exists with a router.
- **`type: k8s` / `external`:** Kubernetes control plane for `k8s`; default router available.
- **Edge access:** Agents that should expose the bridge locally need [distribution tags](distribution.md) (`service.iofog.org/tag: …`), not bare service tokens on the agent.

## Deploy

```bash
potctl deploy -f service.yaml -n my-ecn
potctl get services -n my-ecn
potctl describe service checkout -n my-ecn
potctl delete service checkout -n my-ecn

# Hub provisioning failed — retry reconcile (deploy may also retry once)
potctl reconcile service checkout -n my-ecn
```

**`get services`** shows TYPE, RESOURCE, TARGET PORT, BRIDGE PORT, STATUS (`provisioningStatus`). **`describe service -o yaml`** round-trips deploy fields plus assigned **`bridgePort`** / **`serviceEndpoint`**; use **`get services`** for live **STATUS**.

After successful deploy, the CLI prints that the **service hub** is provisioned and **edge bridge listeners on tagged fogs may still be converging**.

## Tags (`metadata.tags`)

Two uses of the same list:

1. **Edgelet distribution** — bare tokens (e.g. `checkout`) matched against agent tags prefixed with **`service.iofog.org/tag:`** (default; see [distribution.md](distribution.md)).
2. **Kubernetes annotations** — tags containing **`:`** (e.g. `service.beta.kubernetes.io/aws-load-balancer-type: nlb`) are copied to the Controller-created K8s Service and **skipped** for agent matching.

## Deploy order

**Service** is **last** in CLI **`kindOrder`** (after **Microservice**). Typical flow: agents (with distribution tags) → applications → microservices → **Service**.

## Related documents

- [schema.md](schema.md) — fields, immutables, connector hosts
- [distribution.md](distribution.md) — agent tags, `all`, annotation vs distribution
- [examples/microservice-remote.yaml](examples/microservice-remote.yaml)
- [examples/microservice-kubernetes.yaml](examples/microservice-kubernetes.yaml)
- [examples/backend-k8s.yaml](examples/backend-k8s.yaml)
- [examples/stack-with-agent-tags.yaml](examples/stack-with-agent-tags.yaml)
- [Agent tags](../agent/README.md#service-distribution-tags)
- Hub storage: Kubernetes ConfigMap **`iofog-router`** ([kubernetes networking](../controlplane/kubernetes/networking.md)); remote/local hub is the system-agent **`router`** microservice config ([remote networking](../controlplane/remote/networking.md))
- Controller reference (internal): [service-handoff.md](controller-service-notes.md)
