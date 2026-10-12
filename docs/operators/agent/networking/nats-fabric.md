# NATS fabric (messaging topology)

The **NATS fabric** is the broker mesh (servers, leaves, cluster routes, MQTT). **Application and microservice clients** get user JWTs and mounts via [NATS access](../../workloads/nats-access/README.md). **Rule templates** are [NatsAccountRule](../../nats-account-rule/README.md) / [NatsUserRule](../../nats-user-rule/README.md).

Configure the fabric on **`Agent` / `LocalAgent` `spec.config`** or **`AgentConfig` `spec`** ([schema](../agent-config/schema.md#nats-configuration)).

---

## What you declare

Default **`natsMode`** is **`leaf`** (typical edge node).

| `natsMode` | Role | Listeners |
|------------|------|-----------|
| **`leaf`** | Local NATS; dials upstream servers over TLS | Client **4222**, MQTT **8883**, monitor **8222** (no cluster/leaf **listen**) |
| **`server`** | Full server; cluster + leafnode accept | Above plus cluster **6222**, leafnode **7422** |
| **`none`** | No NATS process on this agent | None — workloads use **hub** URL ([runtime](../../workloads/nats-access/runtime.md#natsserverurl)) |

Both **leaf** and **server** publish the full port set on the system microservice; a **leaf** only **listens** on client/MQTT/monitor ports and **dials** upstreams on **7422**.

**`host`** is required unless both **`routerMode`** and **`natsMode`** are **`none`**.

**`upstreamNatsServers`:** agent name, uuid, or **`default-nats-hub`**. A **leaf** cannot be an upstream. If omitted: **default-nats-hub** + every NATS **server** on a system agent.

**Remote CP — first agent:** forced **`natsMode: server`** with system agent promotion. **Kubernetes CP:** hub is **platform NATS**; added **server** agents join cluster routes; routes containing **`nats-headless`** are preserved.

---

## What Controller builds (`leaf` or `server`)

1. Fabric CAs + server and MQTT certificates for this node.
2. JetStream encryption key → secret **`nats-jetstream-key-{agent}`** (`jsk`), injected as **`JETSTREAM_KEY`**.
3. NATS operator + system account; **leaf** also gets **`SYS-leaf-{agentName}`**.
4. System app **`system-{agentName}`** + catalog microservice **`nats`**.
5. Renders **`server.conf`** or **`leaf.conf`** → ConfigMap **`nats-server-conf-{agent}`**, mount **`/etc/nats/config`**.
6. Account-JWT resolver bundle mounted (scope differs for server vs leaf).
7. TLS secret mounts; **server** gets system-user creds.
8. **Leaf:** upstream remotes, per-account **`leaf-{agent}`** users (**`default-leaf-user`**), creds ConfigMap **`nats-leaf-creds-{agent}`**.
9. Publishes ports; health check **`http://localhost:8222/healthz`**.

Container is **not** host-networked. **`NET_RAW`** + service account.

**JetStream:** volume **`{agentName}-nats-jetstream`** at **`/home/runner/data`**, defaults **`jsStorageSize` `10g`**, **`jsMemoryStoreSize` `1g`**, **ChaChaPoly** encrypted. Domain = controller **namespace** on **server**, **agent name** on **leaf**.

Resource names slug the agent name (lowercase, max 48 chars).

---

## TLS and ports

Auto CAs **`nats-site-ca`** and **`default-nats-local-ca`** (60 months) — not ECN [CertificateAuthority](../../certificate-authority/README.md).

| Path | Protection |
|------|------------|
| Client **4222** | Operator JWT mode; workloads use user JWT creds ([NATS access](../../workloads/nats-access/README.md)) |
| Cluster **6222** / leafnode **7422** (servers) | TLS verify + handshake_first |
| Leaf remotes | **`tls://{upstreamHost}:7422`**, mTLS |
| MQTT **8883** | TLS (**`nats-mqtt-server-{agent}`**, includes **`nats.default.svc.bridge.local`**) |
| JetStream disk | ChaChaPoly per-node key |

Operator seed stays in **`nats-operator-seed`**; operator JWT is in generated config only.

---

## Server (`natsMode: server`)

Generated **`server.conf`** includes operator, system account, JetStream, **cluster** routes, **leafnodes** listener, **MQTT**, and **full resolver** dir.

Resolver bundle (account JWTs only — **not** user creds): system account, every application with **`natsAccess`**, controller relay when enabled. Hub uses ConfigMap **`iofog-nats-jwt-bundle`** at **`/tmp/nats/jwt`**.

System users: **`admin-hub`** on hub, **`admin-server-{agent}`** on other servers (creds under **`/etc/nats/creds`**).

Adding or removing **server** agents recomputes cluster routes for all servers. Single-member → multi-member cluster rebuilds the NATS container config.

More on resolver updates when apps change: [NATS access runtime](../../workloads/nats-access/runtime.md#how-account-jwts-reach-nats-servers).

---

## Leaf (`natsMode: leaf`)

**`leaf.conf`:** JetStream, MQTT, operator, resolver — **no cluster listen**. Upstream **remotes** per application with **`natsAccess`** and a microservice on this agent, plus each upstream server.

Leaf user **`leaf-{agentName}`** in each application account (**`default-leaf-user`**). Creds in **`nats-leaf-creds-{agent}`** → **`/etc/nats/creds`**.

Leaf resolver bundle **`nats-jwt-bundle-{agent}`:** **`SYS-leaf-{agent}`** + application accounts with workloads here (smaller than hub bundle).

System user **`admin-leaf-{agent}`** on the leaf system account.

---

## Mode changes and `none`

| Change | Effect |
|--------|--------|
| **leaf → server** | Remove leaf JWT/creds ConfigMaps and leaf system account; enable cluster |
| **server → leaf** | Build leaf artifacts; drop cluster routes |
| **`natsMode: none`** | Delete NATS microservice and leaf artifacts; microservice **`natsAccess`** uses hub **`NATS_SERVER_URL`** |

Host or role crossing **`none`** → reissue certs, flag volume mounts.

---

## What you do not do by hand

| Manual | From **`natsMode`** |
|--------|---------------------|
| Issue NATS/MQTT/leaf certs | **`nats-site-ca`**, **`default-nats-local-ca`** |
| Write **`server.conf`** / leaf remotes | Role + **`upstreamNatsServers`** |
| **`nsc`** operator / SYS / leaf users per site | Platform reconcile |
| Copy account JWTs to nodes | Resolver ConfigMaps |
| Pick JetStream encryption key | Per-node secret + **`JETSTREAM_KEY`** |

Default **`leaf`** is enough for most edge agents; Controller provisions process, trust, and links.
