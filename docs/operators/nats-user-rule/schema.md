# NatsUserRule schema

```yaml
apiVersion: datasance.com/v3
kind: NatsUserRule
metadata:
  name: <rule-name>          # 1–255 chars; must not be a reserved name
  namespace: my-ecn
spec:
  description: "Controller-only note (not a JWT claim)"
  maxPayload: 1m
  pubAllow:
    - "orders.>"
```

**User rules do not** carry account limits (`maxConnections`, imports/exports, JetStream account quotas). Use **`maxPayload`**, not **`maxMsgPayload`**.

**`metadata`** and **`spec`** are required. Validation is server-side.

---

## Limits and units

- **`-1`** = unlimited on `maxSubscriptions`, `maxData`, `maxPayload`.
- **Byte fields:** integer, `-1`, or **`1k` / `100m` / `1g` / `1t`** on `maxData` and `maxPayload`.
- **`respTtl`:** nanoseconds (`5000000000` = 5s).

---

## User `spec` fields

| Field | Type | Meaning |
|-------|------|---------|
| `description` | string | Controller-only; not stored on user JWT |
| `maxSubscriptions` | int ≥ -1 | Max subscriptions. Alias: `subs` |
| `maxData` | bytes | Max data in bytes (no `data:` alias on user rules) |
| `maxPayload` | bytes | Max message payload. Alias: `payload` |
| `bearerToken` | bool | JWT is the credential (MQTT pattern). Alias: `bearer_token` |
| `proxyRequired` | bool | Require PROXY protocol header. Alias: `proxy_required` |
| `allowedConnectionTypes` | enum[] | `STANDARD`, `WEBSOCKET`, `LEAFNODE`, `LEAFNODE_WS`, `MQTT`, `MQTT_WS`, `IN_PROCESS`. Alias: `allowed_connection_types` |
| `src` | string[] | Allowed client IP or CIDR |
| `times` | `{start,end}[]` | Daily connect windows (`hh:mm:ss`) |
| `timesLocation` | string | IANA zone for `times`. Aliases: `times_location`, `locale` |
| `pubAllow` / `pubDeny` | string[] | Publish ACL. Aliases: `pub.allow` / `pub.deny` |
| `subAllow` / `subDeny` | string[] | Subscribe ACL; queue: `"subject queue"`. Aliases: `sub.allow` / `sub.deny` |
| `respMax` / `respTtl` | int | Reply-subject publish limits. Aliases: `resp.max` / `resp.ttl` |
| `tags` | string[] | Tags on user JWT |

**Subject syntax:** `*` / `>`; **deny wins** over allow.

Omit allow lists to leave permissions unset. nsc `--allow-pubsub` maps to the **same subjects** in both pub and sub lists.

User account placement and expiry (`expiresIn` on user create) are **not** rule fields.

---

## Alternative spec shape (aliases)

```yaml
spec:
  subs: -1
  payload: 1m
  bearer_token: false
  allowed_connection_types:
    - STANDARD
    - WEBSOCKET
  pub:
    allow: ["orders.>"]
    deny: []
  sub:
    allow: ["orders.>"]
    deny: []
  resp:
    max: 100
    ttl: 5000000000
```

See [nats-handoff.md](../nats-handoff.md) for the full alias table.

---

## Deploy behavior

| Step | Behavior |
|------|----------|
| Exists | Update from YAML |
| Missing | Create from YAML |
| Delete | Microservices rebound to **`default-user`** |

---

## CLI commands

| Command | Args |
|---------|------|
| `get nats-user-rules` | `-n` |
| `describe nats-user-rule` | NAME; `--output yaml\|json\|wide` |
| `delete nats-user-rule` | NAME |

User credentials: `get nats-users`, `describe nats-user APP USER`.

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Limits, connection types, pub/sub |
| [with-microservice.yaml](examples/with-microservice.yaml) | Rule + standalone Microservice |
