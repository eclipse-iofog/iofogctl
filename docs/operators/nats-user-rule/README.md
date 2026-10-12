# NATS user rule (`kind: NatsUserRule`)

**NatsUserRule** defines a **NATS user-level** policy template on the Controller. The Controller signs a **user JWT** (permissions, limits, connection constraints). **Microservices** attach a rule by name in **`spec.natsConfig.natsRule`** — standalone **`Microservice`** documents or entries under **`Application.spec.microservices[]`**.

Account-level policy is **NatsAccountRule** on the **Application** only. See [Nats account rule](../nats-account-rule/README.md).

| Kind | Binds to | Signed into |
|------|----------|-------------|
| **NatsAccountRule** | **Application** `spec.natsConfig.natsRule` | Account JWT |
| **NatsUserRule** | **Microservice** `spec.natsConfig.natsRule` | User JWT |

## Defaults and reserved names

When **`natsConfig.natsAccess: true`** and **`natsConfig.natsRule`** is omitted on a **Microservice**, Controller uses **`default-user`** (`STANDARD` + `WEBSOCKET`, not bearer, limits `-1`).

Reserved names cannot be created, updated, or deleted (**400**). Use **`describe nats-user-rule <name> -o yaml`** to inspect platform rules.

| Reserved name | Role |
|---------------|------|
| `default-user` | Microservice default when `natsRule` is omitted |
| `default-mqtt-user` | MQTT bearer user (`MQTT` + `STANDARD`) |
| `default-leaf-user` | Leaf connection (`LEAFNODE` + `WEBSOCKET`) |
| `controller-user` | Pub/sub only `controller.relay.v1.>` |

## What gets deployed

Same mechanics as **NatsAccountRule**:

1. Validate document (`kind`, `metadata`, **`spec`**).
2. **`UpdateNatsUserRuleFromYAML`**, else **`CreateNatsUserRuleFromYAML`**.

Identity: **`metadata.name`** (1–255 characters).

**Update** / **delete** behave like account rules: background JWT reissue; delete rebinds microservices to **`default-user`**.

## Deploy

```bash
potctl deploy -f nats-user-rule.yaml -n my-ecn
potctl get nats-user-rules -n my-ecn
potctl describe nats-user-rule checkout-user -n my-ecn
potctl delete nats-user-rule checkout-user -n my-ecn
```

## Microservice binding

```yaml
spec:
  application: orders
  natsConfig:
    natsAccess: true
    natsRule: checkout-user   # NatsUserRule metadata.name; omit → default-user
```

**`natsEnabled`** on microservice YAML is an alias of **`natsConfig.natsAccess`**.

The NATS **username** comes from the microservice (or user-create API), not from **`metadata.name`** on the rule.

Latin-1 required for JWT-bound strings (subjects, tags, `src`, `times`, etc.).

After binding on a microservice: [NATS access](../workloads/nats-access/README.md) (creds mount, **`NATS_CREDS_PATH`**, lifecycle on rule change).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/with-microservice.yaml](examples/with-microservice.yaml)
- [Microservice NATS](../workloads/microservice/schema.md#nats)
- [Nats account rule](../nats-account-rule/README.md)
