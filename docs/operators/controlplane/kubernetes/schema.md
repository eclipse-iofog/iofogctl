# KubernetesControlPlane schema

Reference for **`kind: KubernetesControlPlane`** deploy YAML (v3.9.x).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: KubernetesControlPlane
metadata:
  name: <logical-name>         # stored in CLI namespace config
  namespace: <optional>        # must match CLI -n if both set
spec:
  # fields below
```

---


## CLI-only fields

### `spec.config` (required)

| Property | Value |
|----------|--------|
| Type | string (file path) |
| Required | Yes |
| Description | Path to kubeconfig used to install the operator and apply the ControlPlane CR. Expanded with `~` and normalized to an absolute path. |

### `spec.ca`

| Property | Value |
|----------|--------|
| Type | string, base64-encoded PEM |
| Required | No |
| Description | CA for verifying the **Controller API** when the CLI connects over HTTPS. Stored in `~/.iofog/v3/trust/<namespace>/`. Not sent to the operator. |

See [securing-cluster.md](securing-cluster.md).

### `spec.iofogUser` (required for embedded auth workflow)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Display name |
| `surname` | string | No | Display name |
| `email` | string | **Yes** | Used when creating the Controller user after deploy |
| `password` | string | No | If set, must meet complexity (≥12 chars, 1 upper, 1 special). If omitted, CLI may prompt at deploy |

Not sent to the operator CR.

### `spec.controllerPods`

Output/hydration in stored config after deploy (pod names). Do not rely on this in initial deploy YAML.

---

## Fields translated to operator `ControlPlane` CR

CR name is fixed: **`pot`** (Datasance) or **`iofog`** (Eclipse). Deep operator semantics: [controlplane-crd.md](https://github.com/eclipse-iofog/iofog-operator/blob/main/docs/controlplane/controlplane-crd.md).

### `spec.auth` (required)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `mode` | `embedded` \| `external` | **Yes** | |
| `insecureAllowHttp` | bool | No | Dev only |
| `insecureAllowBootstrapLog` | bool | No | Dev only |
| `bootstrap.username` | string | Yes if embedded | |
| `bootstrap.password` | string | Yes if embedded | Required **in YAML** at deploy; ≥12 chars, 1 upper, 1 special |
| `issuerUrl` | string | Yes if external | |
| `issuerCa` | string | No | Standard base64 of a PEM CA or a concatenated PEM bundle. External mode, when the issuer HTTPS certificate is signed by a private CA. Operator sets `OIDC_ISSUER_TLS_BASE64_CA`. Empty uses the system trust store. Ignored for embedded mode |
| `client.id` | string | Yes if external | |
| `client.secret` | string | Yes if external | |
| `consoleClient` | string | No | |
| `consoleClientEnabled` | bool | No | |
| `rateLimit.*` | object | No | |
| `sessionStore.*` | object | No | |
| `tokenTtl.*` | object | No | |
| `oidcTtl.*` | object | No | |

Embedded: CLI runs bootstrap login and ensures **`iofogUser`** exists. External: CLI skips embedded user bootstrap.

**Retired:** Keycloak-shaped `auth.url`, realm fields.

Optional nested fields (when set, translated to operator / Controller auth env):

| Block | Fields |
|-------|--------|
| `rateLimit` | `enabled`, `maxRequestsPerWindow`, `windowMs` |
| `sessionStore` | `type`, `ttlMs`, `secret` |
| `tokenTtl` | `accessTokenTtlSeconds`, `refreshTokenTtlSeconds` |
| `oidcTtl` | `interactionTtlSeconds`, `grantTtlSeconds`, `sessionTtlSeconds`, `idTokenTtlSeconds` |

### `spec.database`

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `provider` | string | If external DB | e.g. `postgres` |
| `host` | string | If external DB | Empty → operator uses embedded SQLite (single controller) |
| `port` | int | If external DB | |
| `user` | string | If external DB | |
| `password` | string | If external DB | |
| `databaseName` | string | If external DB | |
| `ssl` | bool | No | DB TLS |
| `ca` | string | No | Base64 PEM for DB CA → `DB_SSL_CA` |

**Validation:** If `spec.replicas.controller > 1`, external database with all of host, port, user, password, databaseName, provider must be set.

### `spec.replicas`

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `controller` | int32 | 1 if 0/omitted | Controller Deployment replicas on CR |
| `nats` | int32 | operator default | If NATS enabled and set, must be **≥ 2** |

Multi-controller replicas require external database.

### `spec.controller`

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `publicUrl` | string | Derived by operator if empty | `CONTROLLER_PUBLIC_URL` — see URL resolution below |
| `trustProxy` | bool | Auto **true** in Ingress mode | `TRUST_PROXY` |
| `consoleUrl` | string | **`publicUrl`** if empty | EdgeOps Console URL |
| `consolePort` | int | 8008 | Service port 80 → this port on pod |
| `pidBaseDir` | string | `/home/runner` | |
| `logLevel` | string | `info` | |
| `https` | bool | false | Pod TLS; readiness probe uses HTTPS when true |
| `secretName` | string | — | Required when `https: true` — TLS Secret mounted on pod; for Ingress **B2**, often **same** as `ingresses.controller.secretName` |

**Controller HTTPS patterns:** LoadBalancer + pod TLS (**A**); ClusterIP + Ingress with `https: false` (**B1**); ClusterIP + Ingress + `https: true` with shared Secret (**B2**). See [securing-cluster.md](securing-cluster.md).

#### URL resolution (`publicUrl` / `consoleUrl`)

Handled by the **operator** on the translated ControlPlane CR (not by the CLI). Summary:

| Situation | Operator behavior |
|-----------|-------------------|
| `publicUrl` set, `consoleUrl` empty | `consoleUrl = publicUrl` |
| Ingress mode (ClusterIP + `ingresses.controller.host`), empty `publicUrl` | `https://<host>` if ingress `secretName` set, else `http://<host>`; default `trustProxy: true` if unset |
| LoadBalancer mode, empty `publicUrl` | Wait for LB; `publicUrl = {scheme}://<lb>:51121`, `consoleUrl = {scheme}://<lb>` |
| Derived scheme | `https` if `controller.https: true`, or Ingress mode with ingress `secretName`; else `http` |

Full rules: [operator CRD — URL resolution](https://github.com/eclipse-iofog/iofog-operator/blob/main/docs/controlplane/controlplane-crd.md#url-resolution-publicurl-and-consoleurl). After deploy, the CLI resolves the API endpoint from LB, Ingress **`controller`**, or your explicit URL.

### `spec.events`

| Field | Type | Notes |
|-------|------|-------|
| `auditEnabled` | bool | |
| `retentionDays` | int | |
| `cleanupInterval` | int | Seconds |
| `captureIpAddress` | bool | |

### `spec.images`

Operator Deployment image (CLI install) and workload images on the ControlPlane CR.

| Field | Scope | Required | Notes |
|-------|--------|----------|-------|
| `operator` | **CLI only** | No | Operator Deployment image; **not** on the ControlPlane CR. Default v3.9.x: `ghcr.io/<registry>/operator:3.9.0` |
| `pullSecret` | CR | No | `imagePullSecrets` on operator-managed pods |
| `controller` | CR | No | Default from CLI build if empty |
| `router` | CR | No | |
| `nats` | CR | No | Default e.g. `2.15.0` at v3.9 |

The CLI installs the operator using **`operator`**, then applies the CR with `pullSecret`, `controller`, `router`, and `nats` (empty strings filled from build-time defaults).

### `spec.services`

Each of `controller`, `router`, `nats`, `natsServer`:

| Field | Type | Notes |
|-------|------|-------|
| `type` | string | `LoadBalancer`, `ClusterIP`, `NodePort` |
| `address` | string | Optional `loadBalancerIP` |
| `annotations` | map | Service annotations |
| `externalTrafficPolicy` | string | LoadBalancer / NodePort |

**External traffic policy** (`externalTrafficPolicy`):

- Only applied for `LoadBalancer` and `NodePort`.
- If omitted: **LoadBalancer → Local**, **NodePort → Cluster**, **ClusterIP → unset**.

**CLI validation:**

- `services.controller.type: ClusterIP` → **`ingresses.controller.host`** and **`ingresses.controller.secretName`** required.
- `services.router.type: ClusterIP` → **`ingresses.router.address`** and message/interior/edge ports required.

### `spec.ingresses`

**`ingresses.controller`**

| Field | Notes |
|-------|-------|
| `host` | Ingress rule host |
| `secretName` | Ingress TLS Secret |
| `ingressClassName` | |
| `annotations` | e.g. cert-manager |

**`ingresses.router`**

| Field | Default (registration) |
|-------|------------------------|
| `address` | Required when Router not LB |
| `messagePort` | 5671 |
| `interiorPort` | 55671 |
| `edgePort` | 45671 |

**`ingresses.nats`**

| Field | Typical default |
|-------|-----------------|
| `address` | Required when client Service not LB |
| `serverPort` | 4222 |
| `clusterPort` | 6222 |
| `leafPort` | 7422 |
| `mqttPort` | 8883 |
| `httpPort` | 8222 |

### `spec.nats`

| Field | Type | Default |
|-------|------|---------|
| `enabled` | bool | true if block omitted |
| `jetStream.storageSize` | string | e.g. `10Gi` |
| `jetStream.memoryStoreSize` | string | e.g. `1Gi` |
| `jetStream.storageClassName` | string | optional |

If `enabled: false`, operator skips NATS resources.

### `spec.vault`

Optional Controller secrets vault. If `provider` is set, matching provider block required (hashicorp/openbao/vault, aws, azure, google variants). See operator CRD for env mapping.

---

## Validation summary

| Rule | Error trigger |
|------|----------------|
| `iofogUser.email` | Empty |
| `auth.mode` | Missing or invalid |
| Embedded bootstrap | Missing username/password in YAML |
| External auth | Missing issuer or client credentials |
| `auth.issuerCa` | Set, but not valid standard base64 |
| `replicas.controller > 1` | Incomplete external `database` |
| ClusterIP controller | Missing ingress host/secretName |
| ClusterIP router | Missing router ingress address/ports |
| NATS enabled + `replicas.nats` | Value 1 not allowed (must be ≥ 2 or 0/unset) |
| Vault provider | Missing provider-specific block |

---

## Minimal skeleton

```yaml
apiVersion: datasance.com/v3
kind: KubernetesControlPlane
metadata:
  name: prod-cp
spec:
  config: ~/.kube/config
  iofogUser:
    email: admin@example.com
  auth:
    mode: embedded
    bootstrap:
      username: admin
      password: "ChangeMe12!" 
  database:
    provider: postgres
    host: postgres.example.svc
    port: 5432
    user: controller
    password: secret
    databaseName: controller
  replicas:
    controller: 1
    nats: 2
  images:
    operator: ghcr.io/datasance/operator:3.9.0
    controller: ghcr.io/datasance/controller:3.9.0
    router: ghcr.io/datasance/router:3.9.0
    nats: ghcr.io/datasance/nats:2.15.0
```

See [examples/](examples/) for ingress and TLS patterns.
