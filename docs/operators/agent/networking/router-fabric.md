# Router fabric (service interconnection)

The **router fabric** is the Skupper-style AMQP mesh between edge agents. It is separate from user **`kind: Service`** TCP publishing — Services add **`tcpConnector`** / **`tcpListener`** entries to the router config’s **`bridges`** section ([Service](../../service/README.md)).

Configure the fabric on **`Agent` / `LocalAgent` `spec.config`** or **`AgentConfig` `spec`** ([schema](../agent-config/schema.md#router-configuration)).

---

## What you declare

An agent needs a **name**, **architecture**, and **host** (unless `routerMode: none`). Default **`routerMode`** is **`edge`**.

| `routerMode` | Role | Ports Controller publishes |
|--------------|------|----------------------------|
| **`edge`** | Workload node; connects inward; does not accept other routers | AMQPS **`messagingPort`** (default **5671**) |
| **`interior`** | Fabric node; accepts edge and other interior routers | AMQPS **5671**, edge **45671**, inter-router **55671** (defaults) |
| **`none`** | No router process on this agent | None — use **`networkRouter`** for AMQP clients |

**`host`** is required unless the role is **`none`**.

Creating **`interior`** yourself requires **`edgeRouterPort`** and **`interRouterPort`** on create. On a **remote** control plane, the **first** agent in an empty cluster is promoted to system agent **after** validation and receives interior defaults even if the role was omitted.

**`upstreamRouters`:** agent name, agent uuid, or **`default-router`**. An **edge** router cannot be an upstream. If omitted, Controller attaches to **default-router** plus every router on a **system agent** (deduplicated).

**`networkRouter`:** only when **`routerMode: none`** — names which existing router this agent uses (default **default-router**).

**`routerMode: none`** rejects **`kind: Service`** backends of type **`microservice`** or **`agent`** (no local router). See [Service prerequisites](../../service/README.md).

---

## What Controller builds (`edge` or `interior`)

Platform reconcile, without further YAML:

1. Creates fabric CAs if needed; issues this agent’s router certificates.
2. Creates system application **`system-{agentName}`**.
3. Creates catalog microservice **`router`** on that agent.
4. Writes router config (listeners, TLS profiles, connectors to upstreams).
5. Publishes listener ports on the microservice.
6. Mounts certificate secrets read-only on the router container.
7. Assigns service account and **`NET_RAW`**.

The process reads **`/tmp/skrouterd.json`** (`QDROUTERD_CONF`). **`SKUPPER_SITE_ID`** is the agent uuid. Certs under **`/etc/skupper-router-certs`**.

**Interior** routers use **host networking** so edge and inter-router ports bind on the host. **Edge** routers do not.

---

## TLS by default

Two **self-signed** CAs (60 months), created on first use — not [ECN CertificateAuthority](../../certificate-authority/README.md) resources.

| CA | Signs | Used for |
|----|-------|----------|
| **`router-site-ca`** | **`router-site-server-{agentName}`** | Edge and inter-router links between routers |
| **`default-router-local-ca`** | **`router-local-server-{agentName}`**, **`router-local-agent-{agentName}`** | AMQPS on this node; local clients with client certs |

Site cert SANs include agent **`host`** (or **`localhost`**). Local certs add **`localhost`**, **`127.0.0.1`**, **`router.default.svc.bridge.local`**, docker internal hosts, and on the default router **`router.{namespace}.svc.cluster.local`**.

Secrets mount at **`/etc/skupper-router-certs/{profileName}/`**. SASL **`EXTERNAL`** — certificate is identity; no router password.

Host change or crossing **`none`** → reissue certs and flag **`volumeMounts`**.

**`routerMode: none`** still creates **`default-router-local-ca`** and **`router-local-agent-{agentName}`** for clients of **`networkRouter`**.

---

## Listeners

Every **edge** and **interior** router:

| Listener | Port | TLS | Clients |
|----------|------|-----|---------|
| `{uuid}-amqp` | 5672 | No (not published on host) | Router-local only |
| `{uuid}-amqps` | **`messagingPort`** (5671) | mTLS, SASL EXTERNAL, **`router-local-server-{agent}`** | Workloads on this node |

**Interior** adds:

| Listener | Port | Role |
|----------|------|------|
| `{uuid}-edge` | **`edgeRouterPort`** (45671) | Accepts edge routers |
| `{uuid}-inter-router` | **`interRouterPort`** (55671) | Accepts interior routers |

Both use **`router-site-server-{agentName}`** and require peer certificates.

---

## Upstream connectors

One connector per upstream:

| This router | Connector role | Dials upstream port |
|-------------|----------------|---------------------|
| **edge** | `edge` | Upstream **`edgeRouterPort`** |
| **interior** | `inter-router` | Upstream **`interRouterPort`** |

Connector **host** is the upstream agent’s **`host`**. Name is upstream uuid or **`default-router`**.

Changing **`upstreamRouters`** rewrites connectors and flags the agent. An **interior** router with downstream links cannot switch to **edge** until links are removed.

---

## Running config (do not edit)

Controller writes the router microservice **`config`** JSON: **`metadata`**, **`siteConfig`**, listeners, connectors, SSL profiles, and **`bridges`**.

| Section | Purpose |
|---------|---------|
| **`bridges`** | TCP bridges from the **Service** catalog — regenerated when Services or tags change |
| **`metadata.mode`** | `edge` or `interior` |

Change agent role, ports, host, upstream list, or deploy a **Service** — Controller regenerates config and signals the agent.

Hub storage on Kubernetes control plane may use ConfigMap **`iofog-router`** / **`skrouterd.json`**; edge agents store config in the router system microservice spec. Details: [Service schema](../../service/schema.md).
