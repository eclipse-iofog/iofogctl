# Remote ControlPlane schema

Reference for **`kind: ControlPlane`** remote deploy YAML (v3.9.x).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: ControlPlane
metadata:
  name: <ecn-name>
  namespace: <optional>
spec:
  # global fields
  controllers:
    - name: ...
      host: ...
      ssh: ...
      systemAgent: {}   # required block
```

Alias at parse: **`RemoteController`** document kind maps to **`Controller`** for add-on deploys only.

---

## Global spec fields

### `spec.endpoint`

| Type | Required | Notes |
|------|----------|-------|
| string (URL) | Recommended | CLI API wait; must match `controller.publicUrl` if both set |

### `spec.ca`

Base64 PEM — CLI Controller API trust. Not sent to Edgelet. See [securing-cluster.md](securing-cluster.md).

### `spec.iofogUser`

| Field | Required |
|-------|----------|
| `email` | **Yes** |
| `password` | No (complexity if set; may prompt) |

### `spec.controller`

Same as local **`LocalControllerSpec`**: `publicUrl`, `trustProxy`, `consoleUrl`, `consolePort`, `logLevel`, `pidBaseDir`, **`package`** for image and private registry.

Required for **Controller add-on** eligibility (stored CP must include this block).

### `spec.auth` (required)

Global auth for the ECN (same block on every host’s translated Edgelet manifest).

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `mode` | `embedded` \| `external` | **Yes** | |
| `insecureAllowHttp` | bool | No | Dev only |
| `insecureAllowBootstrapLog` | bool | No | Dev only |
| `bootstrap.username` | string | Yes if embedded | |
| `bootstrap.password` | string | Yes if embedded | Required **in YAML** at deploy; ≥12 chars, 1 upper, 1 special |
| `issuerUrl` | string | Yes if external | |
| `client.id` | string | Yes if external | |
| `client.secret` | string | Yes if external | |
| `consoleClient` | string | No | |
| `consoleClientEnabled` | bool | No | |
| `rateLimit.*` | object | No | See nested fields below |
| `sessionStore.*` | object | No | |
| `tokenTtl.*` | object | No | |
| `oidcTtl.*` | object | No | |

Translated to Edgelet **`spec.auth`** on each **`edgelet deploy -f`**. Env mapping: [Edgelet — spec.auth](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md#spec-auth).

| Mode | CLI (initial ControlPlane deploy) |
|------|-----------------------------------|
| `embedded` | After all hosts: **`EnsureIofogUserEmbedded`** once (bootstrap + **`iofogUser`**) |
| `external` | Skips embedded user bootstrap |

**Controller add-on** does not re-run auth or embedded user setup; it reuses stored global **`spec.auth`**.

**Retired:** Keycloak-shaped `auth.url`, realm fields.

Optional nested fields:

| Block | Fields |
|-------|--------|
| `rateLimit` | `enabled`, `maxRequestsPerWindow`, `windowMs` |
| `sessionStore` | `type`, `ttlMs`, `secret` |
| `tokenTtl` | `accessTokenTtlSeconds`, `refreshTokenTtlSeconds` |
| `oidcTtl` | `interactionTtlSeconds`, `grantTtlSeconds`, `sessionTtlSeconds`, `idTokenTtlSeconds` |

### `spec.database`

| Situation | Rule |
|-----------|------|
| One controller | Optional — SQLite on Edgelet when omitted |
| Multiple controllers | **Required** full external DB (`postgres` / `mysql`) |

Fields: `provider`, `host`, `port`, `user`, `password`, `databaseName`, optional `ssl`, `ca` (base64 PEM).

### Global CA blocks (`SiteCertificate`)

Each block:

| Field | Format |
|-------|--------|
| `tlsCert` | base64 PEM |
| `tlsKey` | base64 PEM |

| YAML key | Controller secret / CA name |
|----------|----------------------------|
| `routerSiteCA` | `router-site-ca` |
| `routerLocalCA` | `default-router-local-ca` |
| `natsSiteCA` | `nats-site-ca` |
| `natsLocalCA` | `default-nats-local-ca` |

Uploaded once after deploy via Controller API when any block is non-nil. See [securing-cluster.md](securing-cluster.md).

### `spec.tls` (global)

| Field | Format |
|-------|--------|
| `ca`, `cert`, `key` | base64 PEM; cert and key must appear together |

Default Controller listener TLS for **all** hosts unless overridden per controller.

### `spec.systemMicroservices`

Arch → image map for **`router`** and **`nats`** (keys: `amd64`, `arm64`, `arm`, `riscv64`). CLI defaults missing entries to build-time router/nats images.

### `spec.nats`

| Field | Notes |
|-------|-------|
| `enabled` | Deploy NATS system microservice when true |

### `spec.events` / `spec.vault`

Same semantics as local/Kubernetes (vault provider blocks validated when enabled).

### `spec.airgap`

When true, every controller with airgap (global or per-host) needs **`systemAgent.config.arch`**.

---

## `spec.controllers[]` (required, ≥1)

| Field | Required | Description |
|-------|----------|-------------|
| `name` | **Yes** | DNS-like lowercase alphanumeric unique per CP |
| `host` | **Yes** | SSH target |
| `ssh.user` | **Yes** | |
| `ssh.keyFile` | **Yes** | Path to private key |
| `ssh.port` | No | Default 22 |
| `systemAgent` | **Yes** | Block must exist (`{}` allowed) |
| `tls` | No | Overrides global `spec.tls` for this host only |
| `airgap` | No | Per-host airgap (inherits global when false) |
| `endpoint` | No | Populated/used post-deploy for host-specific API URL |

### `controllers[].systemAgent`

| Field | Notes |
|-------|-------|
| `config.arch` | Required when `config` is set; required for airgap |
| `config.host` | Agent address registered in Controller (defaults from controller API host) |
| `config.deploymentType` | `native` or `container` |
| `package` / `scripts` | Edgelet agent install options |

Remote validation: if **`systemAgent.config`** is present, **`arch`** is required and must be valid.

---

## Retired metadata

**`metadata.controlPlaneType`** — rejected; use **`kind: ControlPlane`**.

---

## Translation to Edgelet

- **`metadata.name`** on Edgelet manifest comes from deploy options (CLI namespace name / CP metadata).
- Global **`spec.tls`** or per-controller override → Edgelet **`spec.tls.base64`**.
- Global CA blocks **do not** appear in Edgelet YAML; they are Controller API objects only.

Env mapping: [Edgelet control-plane.md](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md).

---

## Validation summary

| Rule | |
|------|--|
| At least one controller | |
| Unique controller names | |
| SSH user, keyFile, host per controller | |
| `systemAgent` non-nil per controller | |
| `endpoint` / `publicUrl` URL validity and match | |
| Multi-controller ⇒ database | |
| Private `controller.package` ⇒ registry + username + password | |
| Base64 validity on `ca`, CA blocks, `tls` | |
| Same-file **`ControlPlane`** + **`Controller`** name/host collision | Rejected |

---

## Examples

- [examples/single-controller.yaml](examples/single-controller.yaml)
- [examples/multi-controller.yaml](examples/multi-controller.yaml)
- [examples/with-global-cas.yaml](examples/with-global-cas.yaml)
