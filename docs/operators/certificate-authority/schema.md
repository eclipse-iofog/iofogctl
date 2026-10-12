# CertificateAuthority schema

```yaml
apiVersion: datasance.com/v3
kind: CertificateAuthority
metadata:
  name: <ca-name>
  namespace: my-ecn
spec:
  subject: "CN=My ECN CA,O=Example"
  type: <ca-type>
  expiration: 3650          # optional; days — typical for self-signed
  secretName: <secret-name> # import types only; see below
```

**`metadata.name`** is the Controller CA catalog name. The CLI sets the API **`name`** from **`metadata.name`** (any **`spec.name`** in file is overwritten).

---

## `spec.type` values

| `type` | When to use | Prerequisites |
|--------|-------------|---------------|
| **`self-signed`** | Controller **generates** a new CA key pair and registers it in the PKI catalog | **`subject`** recommended; **`expiration`** in days (optional) |
| **`direct`** | Import CA material already stored in a **PoT `kind: Secret`** in this ECN | TLS **Secret** deployed first; **`metadata.name`** must match Secret name (see CLI rules) |
| **`k8s-secret`** | Import CA material from a **Kubernetes Secret** in the control plane cluster | K8s secret exists; **`secretName`** / naming rules below |

No other `type` strings are documented for ECN CA deploy in v3.9.x operator guides. Controller returns **400** on unknown types.

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `subject` | Recommended | X.509 subject DN (especially for **`self-signed`**) |
| `type` | **Yes** | **`self-signed`**, **`direct`**, or **`k8s-secret`** |
| `expiration` | No | CA validity in **days** (Controller semantics; common on **`self-signed`**) |
| `secretName` | Import types | For **`direct`** / **`k8s-secret`**: Secret name holding cert/key. If omitted, CLI defaults to **`metadata.name`** when calling Controller |

---

## CLI validation (import types)

| Rule | Error |
|------|-------|
| CA already exists | Updating Certificates are not allowed… |
| `type` is **`k8s-secret`** or **`direct`**, and `metadata.name` ≠ effective `secretName` | Name must match Secret name |
| No Controller | CLI error |

For **`self-signed`**, **`secretName`** is not used.

---

## `type: self-signed`

Controller creates and stores the CA. No backing Secret.

```yaml
kind: CertificateAuthority
metadata:
  name: ecn-root-ca
spec:
  subject: "CN=ECN Root CA,O=Example"
  type: self-signed
  expiration: 3650
```

Example file: [self-signed-ca.yaml](examples/self-signed-ca.yaml).

Leaf certificates can then use **`kind: Certificate`** with **`spec.ca.type: direct`** and **`ca.secretName: ecn-root-ca`**, or **`ca.type: self-signed`** for one-off leaf issuance — see [Certificate schema](../certificate/schema.md).

---

## `type: direct` (PoT Secret)

1. Deploy **`kind: Secret`** with **`spec.type: tls`** and **base64-encoded PEM** in **`data`** keys such as **`tls.crt`**, **`tls.key`**, and optionally **`ca.crt`** — see [Secret — TLS data](../secret/schema.md#tls-secrets-spec-type-tls).
2. Deploy CA — **`metadata.name`** **must equal** the Secret name:

```yaml
kind: CertificateAuthority
metadata:
  name: my-import-ca
spec:
  subject: "CN=Imported CA,O=Example"
  type: direct
```

Example: [import-direct.yaml](examples/import-direct.yaml).

Global control-plane Router/NATS CAs use the same **`direct`** + TLS Secret pattern during **`DeployGlobalCertificates`** — [securing-cluster](../controlplane/remote/securing-cluster.md).

---

## `type: k8s-secret`

Import from a Kubernetes Secret (control plane namespace). Set **`spec.secretName`** to the K8s secret name when it differs from **`metadata.name`**.

**CLI rule:** when **`secretName`** is set (or defaulted), **`metadata.name`** must match that Secret name for **`k8s-secret`** and **`direct`**.

Operators must ensure the K8s secret contains the key layout the Controller expects for TLS import (align with operator / Controller docs for your deployment).

---

## Describe output

`describe certificate NAME` for a CA:

- **`kind: CertificateAuthority`**
- **`data`**: `certificate`, `privateKey`, validity, serial (**sensitive**)

There is no separate `describe certificate-authority` subcommand.

---

## Delete

`delete certificate NAME` detects **`IsCA`** and prompts before **`DeleteCA`**.

---

## Examples

| File | Shows |
|------|--------|
| [self-signed-ca.yaml](examples/self-signed-ca.yaml) | **`type: self-signed`** |
| [import-direct.yaml](examples/import-direct.yaml) | TLS Secret (base64) + **`type: direct`** |
