# Kubernetes control plane networking

On a **Kubernetes** control plane the **operator** creates the default router and the NATS hub. It does not create them as Edgelet system microservices. After the Controller API is up, the operator **registers** those workloads so later agents dial the stored host and ports.

Creating an agent with **`isSystem: true`** is rejected. No agent is the hub. Edge fabric (per-node `router` / `nats` system microservices) is [agent networking](../../agent/networking/README.md).

How the operator reconciles the CR: [operator.md](operator.md). YAML fields: [schema.md](schema.md).

Omitting **`spec.nats`** means NATS is **enabled**. **`spec.nats.enabled: false`** skips NATS reconcile.

CR fields that feed these workloads: **`services.router`**, **`ingresses.router`**, **`replicas.nats`**, **`nats.jetStream`**.

---

## Parallel reconcile

While the ControlPlane condition is **`deploying`** or **`updating`**:

| Reconciler | Workload | Registers with Controller? |
|------------|----------|----------------------------|
| Router | Deployment `router`, Service `router`, ConfigMap `iofog-router`, TLS Secrets | No |
| NATS | StatefulSet `nats`, Services, ConfigMaps, bootstrap Secrets, TLS Secrets | No. It only **reads** NATS bootstrap |
| Controller | Controller Deployment | **Yes** — default router, NATS hub, and CA import |

Router and NATS pods can exist before registration finishes. Edges use addresses stored after registration succeeds.

---

## Router objects

| Kind | Name | Notes |
|------|------|--------|
| Deployment | `router` | 1 replica. A second `router-2` is not created by a standard ControlPlane |
| Service | `router` | Type from `spec.services.router.type` (default **LoadBalancer**) |
| ConfigMap | `iofog-router` | Key **`skrouterd.json`** |
| ServiceAccount / Role / RoleBinding | `router` | |

| Service port name | Port | Role |
|-------------------|------|------|
| `router-message` | **5671** | AMQPS |
| `router-interior` | **55671** | Inter-router |
| `router-edge` | **45671** | Edge listeners |

HTTP health and metrics listen on container port **9090**. That port is not on Service `router`.

The router is **interior**, site id **`default-router`**, platform **`kubernetes`**. Config env: **`QDROUTERD_CONF=/tmp/skrouterd.json`**, **`SSL_PROFILE_PATH=/etc/skupper-router-certs`**.

### Address before certificates

| Condition | Address used for SANs and registration |
|-----------|----------------------------------------|
| Service type **LoadBalancer** | LoadBalancer hostname or IP (reconcile waits) |
| Otherwise | **`spec.ingresses.router.address`** |

If the Service is not a LoadBalancer and that address is empty, reconcile fails (`missing Proxy.Router data for non LoadBalancer Router service`).

### Router certificates

The operator looks up these Secrets in the ControlPlane namespace. Missing ones are generated (RSA 2048, 5 years). If all four already exist, they are reused.

| Secret | Role |
|--------|------|
| `router-site-ca` | Site CA (self-signed when generated) |
| `default-router-local-ca` | Local CA |
| `router-site-server` | Site server cert, signed by `router-site-ca` |
| `router-local-server` | Local server cert, signed by `default-router-local-ca` |

SANs include **`router.<namespace>.svc.cluster.local`** and the external address. Hub certificate names have **no agent suffix**. Certificates Controller later issues for an Edgelet agent are **`router-site-server-{agentName}`** and **`router-local-server-{agentName}`**, signed by the same CAs after import.

Bring your own CA by creating the CA Secrets (and optionally the server Secrets) **before** reconcile. Existing Secrets are not overwritten. Rotation: [securing-cluster.md](securing-cluster.md).

The Deployment mounts only the **server** Secrets at `/etc/skupper-router-certs/router-site-server` and `…/router-local-server`, plus ConfigMap `iofog-router` at `/tmp/skrouterd.json`.

Listeners in the template: edge **45671**, inter-router **55671**, AMQPS **5671** (local server profile), in-pod AMQP **5672** (no TLS, not on the Service). SASL **EXTERNAL**, **`authenticatePeer: true`** on the TLS listeners.

If `iofog-router` already exists, the operator **merges**. Template entities (`router`, `site`, named `sslProfile` and `listener`) are replaced. Other entries are kept. That merge keeps the **`tcpListener`** and **`tcpConnector`** entries Controller adds for [Services](../../service/README.md).

### Registering the default router

After login, the operator calls **`PUT /api/v3/router`**. Controller stores one default-router record (`isDefault: true`, interior). It does **not** create a router microservice for that record.

| API field | Source |
|-----------|--------|
| `host` | LoadBalancer address of Service `router`, or `ingresses.router.address` |
| messaging / inter-router / edge ports | LoadBalancer path: **5671 / 55671 / 45671**. Ingress path: `ingresses.router.messagePort`, `interiorPort`, `edgePort` |

Edges learn the router endpoint from the Controller, not by reading the Kubernetes Service directly.

---

## NATS objects

| Kind | Name | Notes |
|------|------|--------|
| StatefulSet | `nats` | `serviceName: nats-headless`. Replicas = `spec.replicas.nats`, minimum **2** |
| Service | `nats-headless` | `ClusterIP: None`. Pod DNS `nats-0.nats-headless`, … |
| Service | `nats` | Cluster, leaf, MQTT (type from `spec.services.nats`) |
| Service | `nats-server` | Client **4222** and monitor **8222** |
| ConfigMap | `iofog-nats-config` | Key **`server.conf`** |
| ConfigMap | `iofog-nats-jwt-bundle` | Account JWT files |
| PVC template | `js-data` | JetStream file store, default **10Gi** |

| Port | Exposed on |
|------|------------|
| client **4222**, monitor **8222** | Headless and `nats-server` (not Service `nats`) |
| cluster **6222**, leaf **7422**, MQTT **8883** | Headless and Service `nats` |

In-cluster listeners stay on these ports. **`spec.ingresses.nats.*Port`** is used when **registering the hub** and for leaf **advertise**, not to retarget the Kubernetes Service.

### Bootstrap

NATS reconcile calls the Controller in-cluster (`controller.<namespace>.svc.cluster.local:51121`), then **`GET /api/v3/nats/bootstrap`**. Controller creates the operator JWT, the system account, and hub user **`admin-hub`**. The operator stores:

| Secret | Contents |
|--------|----------|
| `nats-operator-seed` | key `seed` |
| `nats-creds-sys-admin-hub` | key `admin-hub.creds` |

If the Controller is not ready yet, NATS reconcile requeues. That is expected while the three reconcilers run together.

JetStream key Secret **`nats-jetstream-key-<controlplane-name>`**, key **`jsk`**: 32 random bytes, created once, injected as **`JETSTREAM_KEY`**. Store is ChaChaPoly. Domain is the ControlPlane **namespace**. Memory and file caps come from **`spec.nats.jetStream`** (defaults **1G** / **10G**).

### NATS certificates

| Secret | Role |
|--------|------|
| `nats-site-ca` | Site CA |
| `default-nats-local-ca` | Local CA |
| `nats-site-server` | Cluster and leaf TLS |
| `nats-mqtt-server` | MQTT TLS |

SANs include headless pod names, `*.nats-headless.<namespace>.svc.cluster.local`, `nats.<namespace>.svc`, `nats-server.<namespace>.svc`, and the external address when known.

Server Secrets are annotated with the replica count. Changing **`spec.replicas.nats`** deletes and recreates the **server** certs so SANs match. CA Secrets are not deleted.

Agent certificates use different names: **`nats-server-{agent}`** and **`nats-mqtt-server-{agent}`**, signed by these CAs after import.

Mounts: site and MQTT certs under `/etc/nats/certs/…`, `server.conf` at `/etc/nats/config`, JWT bundle at `/tmp/nats/jwt`, hub creds, PVC `/home/runner/data`.

### `server.conf`

Rewritten on every NATS reconcile so replica routes stay current. Cluster routes are `nats://nats-<i>.nats-headless:6222` for each replica, **plus** any non-ordinal routes already in the ConfigMap (routes Controller added for agent NATS servers). Operator ordinals are replaced, not duplicated. Routes that contain **`nats-headless`** are kept when Controller patches agent servers in.

Resolver: full directory `/home/runner/nats/jwt`, interval **2m**. JWT ConfigMap **`iofog-nats-jwt-bundle`** starts as the system-account JWT. The operator creates it if missing and **does not replace** an existing ConfigMap. Controller fills application account keys afterward. User creds are not in this bundle. See [NATS access](../../workloads/nats-access/runtime.md).

### Registering the NATS hub

After the default router is registered, **`PUT /api/v3/nats/hub`**. Controller stores one hub record (`isHub: true`, no agent uuid).

| Condition | Host |
|-----------|------|
| `services.nats.type` is LoadBalancer | LB address of Service `nats` (waits) |
| otherwise | `spec.ingresses.nats.address` |

If the host is still empty, the hub is not registered. Default ports sent to the API: server **4222**, cluster **6222**, leaf **7422**, MQTT **8883**, monitor **8222**.

---

## Importing CAs

After hub registration, the operator imports CA Secrets so Controller can sign **agent** certificates with the same authorities. Each call is create-CA with **`type: k8s-secret`** and **`secretName`** equal to the Secret name, only if that CA is not already present.

| Secret imported | When |
|-----------------|------|
| `router-site-ca`, `default-router-local-ca` | always |
| `nats-site-ca`, `default-nats-local-ca` | NATS enabled |

Agent reconcile then issues per-node server certificates. It does not generate a second CA.

---

## What Controller writes back

The operator owns the hub processes and the listener, TLS, and replica-route sections. Controller patches three ConfigMaps afterward:

| ConfigMap | What Controller changes |
|-----------|-------------------------|
| **`iofog-router`** / `skrouterd.json` | Hub **`tcpListener`** and **`tcpConnector`** entries for [Services](../../service/README.md). Operator merge keeps names that are not in its template |
| **`iofog-nats-config`** / `server.conf` | Appends agent-server cluster routes; keeps `nats-headless` routes. Removing an agent server rolls the StatefulSet |
| **`iofog-nats-jwt-bundle`** | Full resolver bundle when application accounts change. Survives the next operator reconcile because the operator does not replace an existing ConfigMap |

A leaf dials the hub host stored by **`PUT /api/v3/nats/hub`**. An edge router dials the host stored by **`PUT /api/v3/router`**.
