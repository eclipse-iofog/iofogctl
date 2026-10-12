# NATS access (applications and microservices)

Turn on messaging by setting **`natsConfig`** on the **Application** and on each **Microservice** that should connect. Controller creates the operator chain, signs JWTs, stores credentials, mounts them into the container, and pushes **account** JWTs to every NATS server that must accept the connection. There is no `nsc` step, no hand-copied `.creds` file, and no per-node resolver edit.

**Policy templates** (limits, pub/sub, imports) live in deploy kinds **[NatsAccountRule](../../nats-account-rule/README.md)** and **[NatsUserRule](../../nats-user-rule/README.md)**. This tree documents **what happens after you bind those rules** through `natsConfig`.

| Document | Contents |
|----------|----------|
| [runtime.md](runtime.md) | Signing chain, creds mount, env vars, resolver bundles |
| [lifecycle.md](lifecycle.md) | Create / update / disable / delete; extra (non-container) users |
| [examples/orders-with-nats.yaml](examples/orders-with-nats.yaml) | Application + embedded microservice |

Workload field reference: [Application schema](../application/schema.md), [Microservice schema](../microservice/schema.md#nats).

---

## Declare access in YAML

```yaml
apiVersion: datasance.com/v3
kind: Application
metadata:
  name: orders
spec:
  natsConfig:
    natsAccess: true
    natsRule: orders-account    # NatsAccountRule metadata.name; omit → default-account
  microservices:
    - name: checkout
      natsConfig:
        natsAccess: true
        natsRule: checkout-user # NatsUserRule metadata.name; omit → default-user
```

Rules:

- **`natsAccess`** and **`natsRule`** belong under **`natsConfig`**. A top-level **`natsAccess`** is rejected.
- A microservice can enable NATS only when its **application** already has **`natsConfig.natsAccess: true`**.
- In one multi-doc file, deploy **[NatsAccountRule](../../nats-account-rule/README.md)** and **[NatsUserRule](../../nats-user-rule/README.md)** before **Application** / **Microservice** (see [operators deploy order](../../README.md#ecn-resource-deploy-order)).

Standalone **`kind: Microservice`** uses the same **`natsConfig`** block; the application must already exist with NATS enabled.

---

## Binding rules to workloads

| Workload | Field | Rule kind | Default if `natsRule` omitted |
|----------|-------|-----------|-------------------------------|
| **Application** | `spec.natsConfig.natsRule` | **NatsAccountRule** | **`default-account`** |
| **Microservice** | `spec.natsConfig.natsRule` | **NatsUserRule** | **`default-user`** |

Microservice YAML may use **`natsEnabled`** as an alias of **`natsConfig.natsAccess`**.

JWT limits and permissions come from the bound rule. Editing a **rule document** re-signs JWTs for every workload bound to that rule (see [lifecycle.md](lifecycle.md)).

---

## What you do not do by hand

| Manual NATS | With potctl / Controller |
|-------------|---------------------------|
| `nsc add operator`, configure every server | Control plane deploy; operator seed in Controller secrets |
| `nsc add account` per application, push account JWT to resolver | **`natsConfig.natsAccess`** on the application |
| `nsc add user`, copy `.creds` into each container | **`natsConfig.natsAccess`** on the microservice — creds secret, mount, **`NATS_CREDS_PATH`**, **`NATS_SERVER_URL`** |
| Re-copy creds when policy changes | Change **`natsRule`** or edit the rule; Controller re-signs and overwrites creds (revokes old user key when user policy changes) |
| Revoke deleted clients manually | Delete microservice, turn **`natsAccess`** off, or delete extra users via **`nats users delete`** |

Details: [runtime.md](runtime.md), [lifecycle.md](lifecycle.md).

---

## CLI overview

| Task | Command |
|------|---------|
| Deploy rules + app + microservice | `potctl deploy -f … -n ECN` |
| List NATS users on accounts | `potctl get nats-users -n ECN` |
| Describe a user | `potctl describe nats-user APP USER -n ECN` |
| Laptop / CI user (no container mount) | `potctl nats users create …` — [lifecycle.md#users-that-are-not-microservices](lifecycle.md#users-that-are-not-microservices) |
| Fetch creds file | `potctl nats users creds APP USER -o ./user.creds -n ECN` |
| Inspect operator / ensure account (advanced) | `potctl nats operator describe`, `potctl nats accounts ensure APP` |

Replace **`potctl`** with **`iofogctl`** on the Eclipse flavor.

---

## Related documents

- [Nats account rule](../../nats-account-rule/README.md) · [Nats user rule](../../nats-user-rule/README.md)
- [Workloads](../README.md)
