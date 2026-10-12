# Secret schema

Reference for **`kind: Secret`** deploy YAML.

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Secret
metadata:
  name: <secret-name>          # required; lowercase alphanumeric
  namespace: my-ecn            # must match CLI -n
spec:
  type: <string>               # optional; passed to Controller on create
data:
  <key>: <string value>
  ...
```

Strict decoding applies to **`spec`**; **`data`** is a separate YAML document section parsed as `map[string]string`.

---

## Field reference

| Field | Location | Required | Description |
|-------|----------|----------|-------------|
| `metadata.name` | metadata | **Yes** | Controller secret name; CLI validates lowercase alphanumeric |
| `spec.type` | spec | No | Secret type on **create** (e.g. **`Opaque`**, **`tls`**). Re-deploy updates **`data`** only in CLI |
| `data` | root | **Yes** (deploy) | Map of data keys → string values (encoding depends on **`spec.type`** — see below) |

---

## TLS secrets (`spec.type: tls`)

Used for **CertificateAuthority `type: direct`** and similar Controller TLS imports. Each **`data`** value must be a **single-line base64 string** of the **raw PEM file bytes** (standard base64, no PEM headers inside the YAML string).

Common keys:

| Key | Content (before base64) |
|-----|-------------------------|
| **`tls.crt`** | Certificate PEM |
| **`tls.key`** | Private key PEM |
| **`ca.crt`** | CA / bundle PEM (optional; may match cert for CA secrets) |

Encode on the workstation:

```bash
base64 -i tls.crt.pem | tr -d '\n'
base64 -i tls.key.pem | tr -d '\n'
```

Example:

```yaml
spec:
  type: tls
data:
  tls.crt: LS0tLS1CRUdJTiBDRVJUSU...
  tls.key: LS0tLS1CRUdJTiBSU0Eg...
```

Do **not** paste multiline `-----BEGIN CERTIFICATE-----` blocks into **`data`** for **`type: tls`** — Controller expects base64-wrapped PEM.

## Opaque and general secrets

For **`spec.type: Opaque`** (or omitted type), **`data`** values are **plain strings** as stored in the Controller (passwords, tokens, JWT text, etc.) — not base64 unless you intentionally store base64 text.

Microservice **`valueFromSecret`** reads these keys as-is after Controller resolution.

---

## Deploy behavior

| State | CLI action |
|-------|------------|
| Secret absent | `CreateSecret` with `name`, `type`, `data` |
| Secret present | `UpdateSecret` with new `data` map |

Controller must exist in namespace config; otherwise deploy errors.

---

## Consumption

### Microservice environment

```yaml
container:
  env:
    - key: DB_PASSWORD
      valueFromSecret: postgres-creds/password
```

Format: **`secret-name/dataKey`**. Secret and microservice share **`metadata.namespace`**.

### VolumeMount

```yaml
kind: VolumeMount
spec:
  secretName: postgres-creds
```

See [volume-mount/schema.md](../volume-mount/schema.md).

---

## CLI commands

| Command | Argument |
|---------|----------|
| `get secrets` | namespace |
| `describe secret` | **name** |
| `delete secret` | **name** |

`describe` round-trips **`spec.type`** and **`data`**.

---

## Validation (CLI)

| Condition | Result |
|-----------|--------|
| Empty `metadata.name` | Error |
| Invalid name characters | `IsLowerAlphanumeric` error |
| No Controller in namespace | Error |
| Invalid YAML in `spec` or `data` | Unmarshal error |

Controller may return **400/404** for API-level validation.

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Generic key/value secret |
| [tls-material.yaml](examples/tls-material.yaml) | **`type: tls`**, base64 **`data`** for CA **`direct`** import |
