# NatsAccountRule schema

```yaml
apiVersion: datasance.com/v3
kind: NatsAccountRule
metadata:
  name: <rule-name>          # 1–255 chars; must not be a reserved name
  namespace: my-ecn          # potctl ECN; optional on Controller API body
spec:
  description: "Human-readable account description"
  # Account limits, JetStream, default_permissions, imports/exports — see below
```

**`metadata`** and **`spec`** are required. Validation is **server-side**; use **`describe nats-account-rule NAME -o yaml`** on your ECN for keys your Controller version accepts.

Read-only on describe (not in deploy spec): **`id`**, **`isSystem`**, **`name`** (deploy identity is **`metadata.name`**).

---

## Limits and units

- **`-1`** on numeric limits means **unlimited** (same as `nsc edit account`).
- **Byte fields** accept an integer, `-1`, or a 1024-based suffix: **`1k`**, **`100m`**, **`1g`**, **`1t`** (single letter only; `10mb` is rejected). Applies to: `maxData`, `maxMsgPayload`, `memStorage`, `diskStorage`, `memMaxStreamBytes`, `diskMaxStreamBytes`.
- **`respTtl`** is in **nanoseconds** (`5s` → `5000000000`). Prefer **`respMax`** + **`respTtl`** for reply permissions; account JWT signing uses those, not `responsePermissions`.
- Setting any **JetStream** limit field causes the signer to copy the whole JetStream group — set the group together and use `-1` where you want unlimited.

---

## Account `spec` fields

| Field | Type | Meaning |
|-------|------|---------|
| `description` | string | Human-readable account description (JWT `description`) |
| `infoUrl` | string | Link for more information. Alias: `info_url` |
| `maxConnections` | int ≥ -1 | Max client connections. Alias: `limits.conn` |
| `maxLeafNodeConnections` | int ≥ -1 | Max leaf connections. Alias: `limits.leaf` |
| `maxData` | bytes | Max bytes account may carry. Alias: `limits.data` |
| `maxMsgPayload` | bytes | Max message payload. Alias: `limits.payload` |
| `maxSubscriptions` | int ≥ -1 | Max subscriptions. Alias: `limits.subs` |
| `maxImports` | int ≥ -1 | Max **count** of imports (not the `imports` list). Alias: `limits.imports` |
| `maxExports` | int ≥ -1 | Max **count** of exports. Alias: `limits.exports` |
| `exportsAllowWildcards` | bool | Export subjects may use `*` / `>`. Alias: `limits.wildcards` |
| `disallowBearer` | bool | Reject bearer user JWTs. Alias: `limits.disallow_bearer` |
| `pubAllow` / `pubDeny` | string[] | Default publish ACL for users without tighter user rules. Aliases: `default_permissions.pub.allow` / `.deny` |
| `subAllow` / `subDeny` | string[] | Default subscribe ACL. Queue form: `"<subject> <queue>"`. Aliases: `default_permissions.sub.*` |
| `respMax` | int ≥ 0 | Max publishes to reply subjects. Alias: `default_permissions.resp.max` |
| `respTtl` | int (ns) ≥ 0 | Reply permission lifetime. Alias: `default_permissions.resp.ttl` |
| `memStorage` / `diskStorage` | bytes | JetStream caps; `-1` unlimited, `0` disabled |
| `streams` / `consumer` | int ≥ -1 | Max streams / consumers |
| `maxAckPending` | int ≥ -1 | Consumer ack pending cap |
| `memMaxStreamBytes` / `diskMaxStreamBytes` | bytes | Per-stream size limits |
| `maxBytesRequired` | bool | Require `max_bytes` on new streams |
| `tieredLimits` | object | Per-tier JetStream limits. Alias: `tiered_limits` |
| `imports` | object[] | Stream/service imports from other accounts |
| `exports` | object[] | Stream/service exports to other accounts |

**Subject syntax:** `*` is one token, `>` is the remainder. **Deny overrides allow.**

Account lifetime, signing keys, and `--public-key` / `--sk` style nsc flags are **not** set on this kind; Controller issues accounts when applications enable NATS.

---

## Import object

| Field | Meaning |
|-------|---------|
| `name` | Import name |
| `subject` | Remote subject |
| `type` | `stream` or `service` |
| `account` | Exporting account **public key** |
| `local_subject` | Local subject alias |
| `token` | Activation token when required |
| `share` | Share connection with other accounts |

---

## Export object

| Field | Meaning |
|-------|---------|
| `name` | Export name |
| `subject` | Exported subject (wildcards need `exportsAllowWildcards: true`) |
| `type` | `stream` or `service` |
| `description` | Text for importers |
| `info_url` | Link for importers |
| `token_req` | Importer must present activation token |
| `response_type` | Service only: `Singleton`, `Stream`, or `Chunked` |
| `account_token_position` | 1-based token index for account-scoped monitoring exports |

---

## Alternative spec shape (aliases)

Do **not** mix camelCase and alias keys for the same value. Example snake_case layout:

```yaml
spec:
  info_url: https://example.com/orders
  limits:
    conn: -1
    payload: 1m
    wildcards: true
  default_permissions:
    pub:
      allow: ["orders.>"]
      deny: ["orders.secret"]
    sub:
      allow: ["orders.>"]
      deny: []
    resp:
      max: 1
      ttl: 5000000000
  tiered_limits: {}
```

Full alias list: [nats-handoff.md](../nats-handoff.md).

---

## Deploy behavior

| Step | Behavior |
|------|----------|
| Update | PATCH YAML to account-rules API |
| Not found | Create from same YAML |
| Invalid spec / reserved name | Controller **400** |
| Delete | Consumers rebound to **`default-account`**; JWT reissue in background |

---

## CLI commands

| Command | Args |
|---------|------|
| `get nats-account-rules` | `-n` |
| `describe nats-account-rule` | NAME; `--output yaml\|json\|wide` |
| `delete nats-account-rule` | NAME |

Related introspection: `get nats-accounts`, `describe nats-account APP`.

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Starter account limits and default permissions |
| [with-application.yaml](examples/with-application.yaml) | Account rule + Application `natsConfig` |
| [import-export.yaml](examples/import-export.yaml) | Cross-account export/import pair (placeholder account key) |

Copy production policies from **`describe nats-account-rule default-account -o yaml`** before authoring custom rules.
