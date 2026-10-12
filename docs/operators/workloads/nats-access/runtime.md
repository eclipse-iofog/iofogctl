# NATS access — runtime

What Controller materializes when **`natsConfig.natsAccess: true`** is set on applications and microservices. Policy content comes from bound **[NatsAccountRule](../../nats-account-rule/README.md)** / **[NatsUserRule](../../nats-user-rule/README.md)** documents.

Lifecycle (updates, revoke, delete): [lifecycle.md](lifecycle.md).

---

## Who signs what

Controller keeps **one operator** for the deployment. Each application with NATS gets **one account**. Each microservice (or extra user) with NATS gets **one user** inside that account.

| Object | Name | Who signs the JWT | Where the secret lives |
|--------|------|-------------------|-------------------------|
| Operator | `{controllerName}-operator` | Operator signs its own JWT | Seed secret **`nats-operator-seed`**. Not mounted into workloads. |
| Account | Application name (`orders`) | Operator | Seed secret **`nats-account-seed-{application}`**. Not mounted into workloads. |
| User | Microservice name (`checkout`) or extra user name | Account | Creds secret **`nats-creds-{application}-{user}`** — mounted into the container for microservice users |

Account and user names are **slugified** in secret names and creds file paths. JWT **`name`** claims stay the application or microservice name.

A **creds file** is the user JWT plus the user seed. Secret data key: **`{account}/{user}.creds`** (example: **`orders/checkout.creds`**).

When no rule name is set, the application uses **`default-account`** and the microservice uses **`default-user`**.

Inspect accounts: **`potctl get nats-accounts`**, **`potctl describe nats-account APP`**.

---

## What the microservice receives

Enabling **`natsConfig.natsAccess`** on a microservice runs these steps in one transaction:

1. Verify the application account exists (application **`natsAccess`** must be **true**).
2. Create the user, sign the user JWT with the account seed, write the creds secret.
3. Create a **VolumeMount** for that secret and link it to the agent running the microservice (sets agent volume-mount change flag).
4. Add a read-only **volume mapping** on the microservice (Controller-managed — you do not author this mapping in deploy YAML when using **`natsAccess: true`**):

| Mapping field | Value |
|---------------|-------|
| `type` | `volumeMount` |
| `hostDestination` | creds secret name |
| `containerDestination` | `/etc/nats/creds` |
| `accessMode` | `ro` |

5. Set environment variables:

| Variable | Value |
|----------|-------|
| **`NATS_CREDS_PATH`** | `/etc/nats/creds/{account}/{user}.creds` |
| **`NATS_SERVER_URL`** | Where this container should connect (see below) |

The agent pulls the microservice spec and linked volume mount. The workload connects using **`NATS_CREDS_PATH`**. It never receives the account seed or operator seed.

Microservice users are **not** bearer tokens and have **no expiry** on the JWT.

---

## `NATS_SERVER_URL`

| Agent has local NATS | Container network | URL |
|----------------------|-------------------|-----|
| Yes | host network | `nats://localhost:{serverPort}` |
| Yes | bridge network | `nats://nats.default.svc.bridge.local:{serverPort}` |
| No | either | `nats://{hubHost}:{hubServerPort}` |

Default port **`4222`**. If the agent has no local NATS and no hub exists, enabling NATS fails validation.

When a microservice **moves to another agent**, Controller recomputes **`NATS_SERVER_URL`** for the new placement.

---

## How account JWTs reach NATS servers

Clients present a **user JWT**. The server must already hold the **account JWT** signed by the trusted operator. Controller maintains a **full resolver** directory (`{accountPublicKey}.jwt` files).

| NATS role | Bundle contents | Delivery |
|-----------|-----------------|----------|
| Server (hub) | System account, every application with **`natsAccess`**, controller relay account when NATS enabled | ConfigMap **`iofog-nats-jwt-bundle`** on NATS microservice (e.g. `/tmp/nats/jwt`) |
| Leaf | Leaf system account + application accounts that have a microservice on that fog | Per-fog ConfigMap, same mount pattern |

Leaf nodes also get creds for **`leaf-{fogName}`** users (**`default-leaf-user`**) for upstream authentication — not workload users.

After an account JWT changes, Controller enqueues **NATS reconcile** and refreshes bundles on:

- every NATS server (non-leaf), and  
- every agent that runs a microservice of the affected application.

NATS rereads the resolver directory on its configured interval. **User creds are not in the bundle** — only on the workload via secret mount.

How NATS **servers and leaves** are provisioned on agents: [NATS fabric](../../agent/networking/nats-fabric.md). The hub JWT ConfigMap is created by the Kubernetes operator and then patched by Controller ([Kubernetes networking](../../controlplane/kubernetes/networking.md)); on remote/local it is a Controller ConfigMap on the system agent ([remote networking](../../controlplane/remote/networking.md)). Control plane global CAs: [securing cluster](../../controlplane/remote/securing-cluster.md).
