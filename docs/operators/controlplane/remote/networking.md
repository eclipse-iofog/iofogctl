# Remote control plane networking

On a **remote** control plane there is no operator and no Kubernetes ConfigMap named `iofog-router` or `iofog-nats-config` for the hub process. Controller runs on an **Edgelet** node. That node is the **system agent**, and it **is** the default router and the NATS hub.

**Local** control plane uses the same Edgelet system-agent path. See [local control plane](../local/README.md).

Kubernetes hubs (Deployment `router`, StatefulSet `nats`, operator registration) are [kubernetes/networking.md](../kubernetes/networking.md). Per-node edge fabric after the hub exists is [agent networking](../../agent/networking/README.md).

---

## Who creates the hub

Edgelet registers Controller on the system node with **`POST /api/v3/agent/controller/register`**. The call is accepted only when the agent is **`isSystem`**. Controller stores a system microservice named **`controller`** in application **`system-{agentName}`**.

On an empty remote cluster, creating the **first** agent promotes it to **`isSystem`**, **`routerMode: interior`**, and **`natsMode: server`**, including when the request asked for another role. **potctl** does this through **`spec.controllers[0].systemAgent`** during control plane deploy (serial system-agent step after the Controller API is up). See [remote README](README.md) and [multi-controller HA](multi-controller-ha.md).

Controller then provisions the **`router`** and **`nats`** system microservices on that same node:

- The router is marked **default** when no default router exists yet (`default-router`).
- The NATS server is marked **hub** when no hub exists yet (`default-nats-hub`).

Later agents attach to them. A system agent must stay **`routerMode: interior`** and **`natsMode: server`**. An interior router that still has downstream routers cannot be switched to **`edge`**.

```text
Edgelet on the system node
  registers microservice "controller"
        |
        +--> Controller provisions microservice "router"
        |      config on the microservice (listeners, connectors, bridges)
        |      certs router-site-server-{name}, router-local-server-{name}
        |      this router row is default-router
        |
        +--> Controller provisions microservice "nats"
               server.conf in ConfigMap nats-server-conf-{name}
               this NATS row is the hub
               leaves and extra servers dial its host
```

The router container on this node uses **host networking** so **45671** and **55671** bind on the host. The NATS container does not.

---

## Certificates

Certificates, listeners, leaf remotes, and JetStream for that node follow the same provisioning path as any other agent: [router fabric](../../agent/networking/router-fabric.md), [NATS fabric](../../agent/networking/nats-fabric.md).

Names are **per agent** (`router-site-server-{agentName}`, `nats-server-{agent}`), not the unsuffixed hub names the operator uses on Kubernetes (`router-site-server`, `nats-site-server`).

Controller creates **`router-site-ca`**, **`default-router-local-ca`**, **`nats-site-ca`**, and **`default-nats-local-ca`** itself, self-signed, valid for **60 months**, the first time that node is reconciled — unless you supplied those CAs on remote deploy ([securing-cluster.md](securing-cluster.md)).

These are not the ECN [CertificateAuthority](../../certificate-authority/README.md) catalog.

---

## How updates are applied

| Change | Where it is written | How the node picks it up |
|--------|---------------------|--------------------------|
| Router listeners, connectors, TLS profiles | `config` on the **`router`** system microservice (`bridges.tcpListeners` / `tcpConnectors` for Services) | Edgelet change list, microservice config |
| Service bridge on the default router | Same microservice config | Same change list |
| NATS `server.conf` or `leaf.conf` | ConfigMap **`nats-server-conf-{agent}`**, key `server.conf`, mounted at `/etc/nats/config` | Volume mount on the **`nats`** system microservice |
| Account JWT bundle on the hub | ConfigMap **`iofog-nats-jwt-bundle`**, mounted at `/tmp/nats/jwt` | Same volume-mount path |
| Cluster membership | Every server’s `server.conf` is rendered again. Growing from one server to many rebuilds the local NATS container | Change list, and rebuild when the template changes |
| Host or role change | Certificates reissued, secrets remounted | Change list, `volumeMounts` |

There is no operator rewrite to preserve. Controller owns these objects directly.

Observe the hub workloads:

```bash
potctl get system-microservices -n my-ecn
potctl describe system-microservice system-<controller-agent>/router -n my-ecn
potctl describe system-microservice system-<controller-agent>/nats -n my-ecn
```

Service TCP bridges on this hub are written into that router microservice config, not into a cluster ConfigMap. See [Service](../../service/README.md). Application account JWTs in the hub bundle: [NATS access](../../workloads/nats-access/runtime.md).
