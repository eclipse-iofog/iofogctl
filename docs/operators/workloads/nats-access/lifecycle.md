# NATS access — lifecycle

Controller behavior when you create, update, disable, or delete applications and microservices with **`natsConfig`**. Runtime materialization: [runtime.md](runtime.md).

---

## Application

### Create with `natsAccess: true`

After the application row is stored with **`natsAccess`** and the resolved rule id, Controller:

1. Creates the account if missing (operator signs the account JWT).
2. Re-signs the account JWT with the current **NatsAccountRule** (keeps existing user **revocations**).
3. For each microservice in the application with **`natsAccess`**, ensures user + creds mount.
4. Enqueues resolver distribution for NATS servers and agents running this application’s microservices.

Same YAML with embedded **`microservices[]`** still requires **application** **`natsAccess: true`** before any microservice **`natsAccess`**. Microservice create fails closed if the application flag is off.

### Change `natsRule`

Account **keys do not change**. Controller re-signs the account JWT with the new rule and keeps revocations. Then for each microservice in the application:

- **`natsAccess: true`** — re-issue user JWT if that user’s rule binding changed; ensure creds mount remains.
- **`natsAccess` false** but user or mount still present — revoke user and detach mount.

Servers and hosting agents receive the updated account JWT via the resolver bundle.

### Turn `natsAccess` on

Same as create: ensure account, sign, attach creds for microservices that already requested NATS, push account JWT.

### Turn `natsAccess` off

Revoke every microservice user in the application and remove creds mounts. Delete account row, account seed secret, and remaining user creds secrets. Resolver reconcile drops that account JWT from bundles.

### Delete the application

Microservices delete first; each user is revoked (public key on account JWT, then user + creds secret removed). Then remaining account users (MQTT, extra users) are removed, account seed deleted, account deleted. Resolver removes the account from servers.

System applications cannot be deleted through the user application API.

YAML reference: [Application schema](../application/schema.md).

---

## Microservice

### Create or enable `natsAccess`

Requires application **`natsConfig.natsAccess: true`**. Runs [runtime](runtime.md) mount steps. User JWT signed by account seed.

Controller stores **`natsAccountId`**, **`natsUserId`**, **`natsCredsSecretName`** on the microservice row.

### Change `natsRule`

Account unchanged. **Old user key is not reused:**

1. Current user public key added to account JWT **`revocations`**; account JWT re-signed.
2. New user keypair; new user JWT with the new **NatsUserRule**.
3. Same creds **secret overwritten** with new creds file.
4. Resolver reconcile pushes updated account JWT.

**`NATS_CREDS_PATH`** stays the same; file contents change. Agent receives updated secret and microservice spec via change tracking.

If the **rule id** did not change, Controller does not mint a new user key. Editing the **NatsUserRule document** itself reissues every user bound to that rule (including microservices on **`default-user`** when that rule was edited) — same revoke + new creds path.

### Turn `natsAccess` off

Remove volume mapping and **`NATS_CREDS_PATH`** / **`NATS_SERVER_URL`**. Unlink volume mount from agent. Revoke user (public key on account JWT), delete creds secret and user row. **Account remains** for other microservices in the application.

### Delete the microservice

Clear NATS foreign keys, revoke user like disabling **`natsAccess`**. Application account is **not** deleted.

### Move to another agent

Creds volume mount links to the new agent when NATS is ensured again. **`NATS_SERVER_URL`** recomputed.

YAML reference: [Microservice schema](../microservice/schema.md#nats).

---

## Users that are not microservices

Clients that are **not** a Controller-deployed container — laptop, CI, MQTT device — get users on the application **account** via API. Controller signs JWTs and stores creds secrets; it **does not** mount them into a microservice.

The application must have an account (application **`natsAccess: true`**, or ensure via API). Fetch creds and give them to the client; delete the user when access should stop.

### Standard NATS user

| Field | Meaning |
|-------|---------|
| `name` | User name inside the application account |
| `natsRule` | **NatsUserRule** name; omit → **`default-user`** |
| `expiresIn` | Optional lifetime (`h` / `d` / `m` on API). Omitted → no JWT **`exp`**. Microservice users have no expiry. |

Not a bearer user; connection types and subjects come from the rule.

**potctl:**

```bash
potctl nats users create orders laptop --nats-rule checkout-user --expires-in 604800 -n my-ecn
potctl nats users creds orders laptop -o ./laptop.creds -n my-ecn
potctl describe nats-user orders laptop -n my-ecn
potctl nats users delete orders laptop -n my-ecn
```

**`--expires-in`** on the CLI is **seconds** (example: `604800` = 7 days). Use **`get nats-users`** to list microservice-linked users (`microserviceUuid` set) and extra users (`microserviceUuid` null).

**Delete** is rejected while the user name is still linked to an existing microservice with NATS enabled — disable microservice **`natsAccess`** or delete the microservice first. System-account and leaf-system users cannot be deleted here.

### MQTT bearer user

Default rule **`default-mqtt-user`** (bearer, MQTT + STANDARD). JWT sets bearer authentication without a connect challenge.

```bash
potctl nats users create-mqtt-bearer orders sensor-7 --expires-in 2592000 -n my-ecn
potctl nats users creds orders sensor-7 -o ./sensor-7.creds -n my-ecn
potctl nats users delete-mqtt-bearer orders sensor-7 -n my-ecn
```

Creating the same name again returns the existing user (no key rotation). Delete and create again to rotate.

Rule fields for MQTT and bearer: [Nats user rule](../../nats-user-rule/README.md).

---

## Advanced CLI

| Command | Use |
|---------|-----|
| `potctl nats operator describe -n ECN` | Operator metadata / JWT |
| `potctl nats accounts ensure APP --nats-rule RULE -n ECN` | Ensure account without a full application redeploy (API equivalent) |

Primary path for edge workloads remains **`deploy -f`** with **`natsConfig`** on Application and Microservice.
