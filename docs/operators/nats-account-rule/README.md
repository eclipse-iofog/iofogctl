# NATS account rule (`kind: NatsAccountRule`)

**NatsAccountRule** defines a **NATS account-level** policy template on the Controller. The Controller signs an **account JWT** from the rule (`limits`, `default_permissions`, imports/exports). **Applications** attach a rule by name in **`spec.natsConfig.natsRule`**.

For **user-level** policies (microservice JWTs), see [Nats user rule](../nats-user-rule/README.md).

| Kind | Binds to | Signed into |
|------|----------|-------------|
| **NatsAccountRule** | **Application** `spec.natsConfig.natsRule` | Account JWT |
| **NatsUserRule** | **Microservice** `spec.natsConfig.natsRule` | User JWT |

The static NATS server [`authorization`](https://docs.nats.io/reference/config/authorization/) block (`username`, `password`, `users`, etc.) is a different mechanism. Controller does **not** emit that block from these kinds.

## Defaults and reserved names

When **`natsConfig.natsAccess: true`** and **`natsConfig.natsRule`** is omitted on an **Application**, Controller uses the reserved rule **`default-account`**.

You **cannot** create, update, or delete reserved rules (Controller returns **400**). Inspect them with **`describe nats-account-rule <name> -o yaml`**.

| Reserved name | Role |
|---------------|------|
| `default-system-account` | NATS system account on edge agents |
| `default-account` | Application default when `natsRule` is omitted |
| `controller-account` | Controller relay account |

Do not deploy YAML documents whose **`metadata.name`** matches a reserved name.

## What gets deployed

1. Control plane / Controller with NATS available.
2. CLI validates the document (`kind`, `metadata`, **`spec`** required). **`apiVersion`** is not validated for this kind on the Controller; still use your flavor’s `apiVersion` in potctl YAML.
3. **`UpdateNatsAccountRuleFromYAML`**; on **404** → **`CreateNatsAccountRuleFromYAML`**.

Identity: **`metadata.name`** (1–255 characters). **`metadata.namespace`** in examples is for potctl **`-n`** / ECN context.

**Update** returns immediately; affected JWTs are reissued in the background. **Delete** rebinds applications to **`default-account`** and reissues in the background.

## Prerequisites

- Controller in namespace (control plane present).
- Deploy the rule **before** (or in the same multi-doc file **before**) an Application sets **`natsConfig.natsAccess: true`** with this **`natsRule`** name.

## Deploy

```bash
potctl deploy -f nats-account-rule.yaml -n my-ecn
potctl get nats-account-rules -n my-ecn
potctl describe nats-account-rule orders-account -n my-ecn
potctl delete nats-account-rule orders-account -n my-ecn
```

Describe supports **`--output yaml|json|wide`**. Round-trip: describe a working rule, edit, redeploy.

Strings that land in a JWT must be **Latin-1** (ASCII is safe). Invalid characters on subjects, tags, imports/exports, etc. return **400**.

## Application binding

```yaml
spec:
  natsConfig:
    natsAccess: true
    natsRule: orders-account   # NatsAccountRule metadata.name; omit → default-account
```

Per-microservice NATS uses **NatsUserRule** on embedded or standalone microservices — not an account rule. See [Application schema](../workloads/application/schema.md) and [Microservice NATS](../workloads/microservice/schema.md#nats).

After binding: [NATS access](../workloads/nats-access/README.md) (account JWT, resolver, microservice creds mount).

## Related documents

- [schema.md](schema.md) — field reference, aliases, imports/exports
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/with-application.yaml](examples/with-application.yaml)
- [examples/import-export.yaml](examples/import-export.yaml)
- [Nats user rule](../nats-user-rule/README.md)
