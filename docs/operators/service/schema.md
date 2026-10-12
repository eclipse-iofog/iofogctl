# Service schema

```yaml
apiVersion: datasance.com/v3
kind: Service
metadata:
  name: <service-name>
  namespace: my-ecn          # potctl ECN context
  tags:
    - checkout              # distribution token; see distribution.md
spec:
  type: microservice       # microservice | agent | k8s | external — immutable
  resource: orders/checkout
  targetPort: 8080
  defaultBridge: default-router   # immutable; required default-router for k8s/external
  # Kubernetes control plane only (required for microservice, agent, external):
  servicePort: 80
  k8sType: LoadBalancer    # LoadBalancer | ClusterIP | NodePort
```

**`metadata`** and **`spec`** are required. **`apiVersion`** is not validated on the Controller; use your flavor’s group in potctl YAML. **Tags live under `metadata.tags`, not `spec`.**

---

## Writable fields

| Field | Required | Meaning |
|-------|----------|---------|
| `metadata.name` | yes | DNS label: `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`. Also router **`address`** and Kubernetes Service name when created. **Rejected:** `controller`, `router`, `router-internal`, `docker`, `podman`, `kubernetes`, `nats`, `nats-headless`, `nats-server` |
| `metadata.tags` | no | Distribution tokens and/or K8s annotation tags (see [distribution.md](distribution.md)) |
| `spec.type` | yes | `microservice`, `agent`, `k8s`, or `external`. **Immutable** |
| `spec.resource` | yes | Backend identifier — meaning depends on `type` (below) |
| `spec.targetPort` | yes | TCP port on the backend → `tcpConnector.port` |
| `spec.defaultBridge` | no | Hub router owning the listener. Default **`default-router`**. **Immutable**. `k8s` / `external` must use **`default-router`** |
| `spec.servicePort` | K8s CP, non-`k8s` types | Port clients use on the created Kubernetes Service. Remote CP: Controller overwrites with **`bridgePort`** |
| `spec.k8sType` | K8s CP, non-`k8s` types | `LoadBalancer`, `ClusterIP`, or `NodePort` |

**Not accepted in YAML:** **`bridgePort`** (assigned from `BRIDGE_PORTS_RANGE`, default `10024–65535`, first free port).

---

## `spec.resource` by type

| `type` | You write | Stored as | Connector dials |
|--------|-----------|-----------|-----------------|
| `microservice` | `appName/microserviceName` or microservice UUID | microservice UUID | See connector host table |
| `agent` | agent name or agent UUID | agent UUID | See connector host table |
| `k8s` | DNS or IP of an existing Service | that string | `resource:targetPort` |
| `external` | DNS or IP outside the cluster | that string | `resource:targetPort` |

**`type: k8s`** is rejected unless the control plane is **Kubernetes**. Controller does **not** create a Kubernetes Service for **`k8s`**.

---

## Connector host (`tcpConnector.host`)

Chosen from the backend, not from `resource` alone.

| Backend | Router on agent | Host dialed |
|---------|-----------------|-------------|
| Microservice, edge router, not host network | edge | `{applicationName}.{microserviceName}` |
| Microservice, edge router, host network | edge | `edgelet.default.svc.bridge.local` |
| Microservice, interior router | interior | `127.0.0.1` |
| Agent, edge router | edge | `edgelet.default.svc.bridge.local` |
| Agent, interior router | interior | `127.0.0.1` |
| `k8s` or `external` | connector on default router | the `resource` string |

| `type` | Router that receives `tcpConnector` |
|--------|-------------------------------------|
| `microservice` | agent running the microservice |
| `agent` | that agent |
| `k8s`, `external` | default router |

**`tcpConnector.processId`:** microservice → microservice UUID; agent → `{agentUuid}-local-{targetPort}`; `k8s` → `{resource}-k8s-{targetPort}`; `external` → `{resource}-external-{targetPort}`.

Connector and listener share **`address`** = service **`metadata.name`**. Listener binds **`bridgePort`**.

---

## Assigned / status fields (describe, API)

| Field | Meaning |
|-------|---------|
| `bridgePort` | Router listener port (assigned) |
| `serviceEndpoint` | Where clients reach the hub (remote: router host; K8s LoadBalancer: ingress IP/hostname) |
| `provisioningStatus` | `pending`, `ready`, or `failed` |
| `provisioningError` | Last hub reconcile error when `failed` — retry with **`potctl reconcile service NAME`** |

**`ready`** = hub connector + hub listener + (if applicable) Kubernetes Service. Edge agent listeners are updated by fog reconcile separately.

---

## Kubernetes control plane

When `CONTROL_PLANE=kubernetes`:

| `spec.type` | `k8sType` + `servicePort` | Controller creates K8s Service |
|-------------|----------------------------|--------------------------------|
| `microservice`, `agent`, `external` | required | yes, in controller namespace, name = `metadata.name` |
| `k8s` | not used | no |

Router pod selector uses the Controller distribution label (e.g. `datasance.com/component: router` or `iofog.org/component: router`). Client **`servicePort`** → pod **`bridgePort`**.

**LoadBalancer:** Controller polls load balancer ingress; missing ingress after retries → **`failed`**.

Delete removes the Kubernetes Service for every type **except** **`k8s`**.

---

## Update and delete

- **`type`** and **`defaultBridge`** cannot change — delete and recreate.
- Update rewrites connector and listener. If **`resource`** changed, connector is removed from the previous site first.
- Tag changes union old and new tags so agents that lost a tag drop `-listener` entries.
- Delete removes hub `{name}-connector` / `{name}-listener`, deletes K8s Service when present, and reconciles agents that had matched the old tags.

---

## CLI validation (potctl)

| Rule | Notes |
|------|-------|
| Empty name | Error |
| Name chars | `IsLowerAlphanumeric` in CLI; Controller also enforces DNS label + reserved names |
| No Controller | Error |

**`type` / `resource`** combinations and immutables → Controller **400**.

---

## CLI commands

| Command | Args |
|---------|------|
| `get services` | `-n` |
| `describe service` | NAME; `-o yaml` for deploy-shaped YAML |
| `delete service` | NAME |
| `reconcile service` | NAME — retry hub provisioning |

---

## Examples

| File | Shows |
|------|--------|
| [microservice-remote.yaml](examples/microservice-remote.yaml) | Remote CP — no `servicePort` / `k8sType` |
| [microservice-kubernetes.yaml](examples/microservice-kubernetes.yaml) | LoadBalancer + annotation tag |
| [backend-k8s.yaml](examples/backend-k8s.yaml) | `type: k8s` backend |
| [agent-ssh.yaml](examples/agent-ssh.yaml) | `type: agent` |
| [external-vendor.yaml](examples/external-vendor.yaml) | `type: external` |
| [stack-with-agent-tags.yaml](examples/stack-with-agent-tags.yaml) | Service + Agent distribution tags |

Full router JSON and reconcile details: [service-handoff.md](../service-handoff.md).
