# Securing the local control plane

This guide covers **`LocalControlPlane`** TLS and trust for **potctl** / **iofogctl** (v3.9.x). The Controller container is managed by **Edgelet**; messaging TLS for Router and NATS follows Controller catalog and system microservice deployment rules.

Edgelet listener TLS details: [Edgelet control plane — spec.tls](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md#spec-tls).

## Overview

| Mechanism | Purpose | Applied on local deploy? |
|-----------|---------|---------------------------|
| `spec.ca` | CLI trust for Controller **API** HTTPS | **Yes** — namespace trust store |
| `spec.tls` | Controller **listener** TLS (cert/key/ca) | **Yes** — translated to Edgelet `spec.tls.base64` |
| `routerSiteCA`, `routerLocalCA`, `natsSiteCA`, `natsLocalCA` | Bring-your-own messaging CAs (Controller secrets + CA catalog) | **Validated only** — **not** uploaded today |
| Edgelet `spec.tls.path` | Host directory mount for certs | **Not** set by CLI (CLI uses base64 block only) |

### Future parity

Remote **`ControlPlane`** deploy calls the Controller API to create TLS secrets and CA entries for the four global CA blocks when present. **LocalControlPlane** will gain the same behavior in a **future release**. For production BYO messaging CAs, use **remote** deploy or manual Controller API import until then.

## CLI API trust: `spec.ca`

Same as Kubernetes and remote:

- **Format:** base64-encoded PEM (single line).
- **Storage:** `~/.iofog/v3/trust/<namespace>/ca.pem` after deploy.
- **Use:** HTTPS verification when CLI and SDK call the Controller API.

```bash
# macOS
base64 -i controller-ca.pem | tr -d '\n'
```

`connect --ca` remains a connect-time override.

## Controller listener TLS: `spec.tls`

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `ca` | base64 PEM | No | Intermediate / root for clients |
| `cert` | base64 PEM | Yes* | Server certificate |
| `key` | base64 PEM | Yes* | Private key |

\* **`cert` and `key` must both be set** if either is set.

The CLI maps this block to Edgelet **`spec.tls.base64`** (not `spec.tls.path`). Edgelet sets `SERVER_DEV_MODE=false` and `TLS_BASE64_*` env vars. See Edgelet env table in [control-plane.md](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md).

**Align with URLs:** Set `spec.controller.publicUrl` and optional `spec.console.url` to `https://` schemes that match your certificate SANs.

## Global messaging CAs (YAML blocks)

Each block uses **`SiteCertificate`** shape:

| Field | Format |
|-------|--------|
| `tlsCert` | base64-encoded PEM (CA or cert material) |
| `tlsKey` | base64-encoded PEM private key |

| YAML block | Controller secret name (remote deploy) |
|------------|----------------------------------------|
| `routerSiteCA` | `router-site-ca` |
| `routerLocalCA` | `default-router-local-ca` |
| `natsSiteCA` | `nats-site-ca` |
| `natsLocalCA` | `default-nats-local-ca` |

On **remote** deploy, the CLI creates/update-idempotently TLS secrets and CA catalog entries with these names. On **local** deploy today, you may include these blocks for forward-compatible manifests, but they **have no effect** until parity ships or you import via Controller API.

Edgelet **forbids** legacy `spec.siteCA` / `spec.localCA` in the Edgelet manifest — import via Controller REST after deploy ([Edgelet rules](https://github.com/Datasance/edgelet/blob/main/docs/edgelet/control-plane.md)).

## Database TLS

When `spec.database.provider` is set:

| Field | Edgelet → Controller env |
|-------|----------------------------|
| `ssl` | `DB_USE_SSL` |
| `ca` | `DB_SSL_CA` (base64 PEM in YAML) |

## Edge agents and provision

After the system agent provisions:

- If the provision key response includes **`caCert`**, Edgelet receives **`edgelet config cert <base64>`** so the agent trusts the Controller messaging CAs.
- Agents must trust the same CAs you publish in the Controller catalog for Router (5671) and NATS (4222 / 8883).

## Checklist (local test)

- [ ] Set `spec.ca` if the Controller API uses a private CA.
- [ ] Set `spec.tls` cert/key (and optional ca) for HTTPS listener testing.
- [ ] Do not rely on `routerSiteCA` blocks locally until parity release — use remote CP or manual CA import.
- [ ] Keep bootstrap and DB passwords out of Git (dev inline passwords only in isolated labs).

## Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| CLI TLS error to API | Missing/wrong `spec.ca` |
| Browser/CLI HTTPS to controller works but agent fails messaging | Messaging CAs not in Controller catalog |
| Validation error on tls | cert/key mismatch or invalid base64 |
| `routerSiteCA` in YAML but no effect | Expected on local until future CLI parity |
