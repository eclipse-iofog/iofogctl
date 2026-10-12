# Certificate schema

```yaml
apiVersion: datasance.com/v3
kind: Certificate
metadata:
  name: <cert-name>
  namespace: my-ecn
spec:
  subject: "CN=api.example.com,O=Example"
  hosts: "api.example.com,api.internal"
  expiration: 365
  ca:
    type: <issuer-type>
    secretName: <ca-name>   # when required by type
```

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `subject` | **Yes** | X.509 subject DN |
| `hosts` | **Yes** | Comma-separated SAN hostnames / IPs |
| `expiration` | No | Leaf validity in **days** |
| `ca.type` | **Yes** | Issuing CA mode (see table) |
| `ca.secretName` | Conditional | CA catalog name or secret reference |

---

## `spec.ca.type` values

| `ca.type` | Meaning |
|-----------|---------|
| **`self-signed`** | Controller issues the leaf using **self-signed** PKI semantics (no named CA in **`secretName`**) |
| **`direct`** | Issue from an existing **CA catalog entry** / PoT secret reference — set **`ca.secretName`** to the **CertificateAuthority `metadata.name`** (or matching import secret name) |

Other values may be rejected by Controller with **400**; the leaf **`ca`** block mirrors **`CertificateCreateCA`** on the API.

### Issue against a CA catalog entry

After **`kind: CertificateAuthority`** with **`metadata.name: ecn-root-ca`**:

```yaml
ca:
  type: direct
  secretName: ecn-root-ca
```

### Self-signed leaf

```yaml
ca:
  type: self-signed
```

Example: [self-signed-leaf.yaml](examples/self-signed-leaf.yaml).

---

## Deploy behavior

| Case | CLI |
|------|-----|
| Name not found | `CreateCertificate` |
| Name exists | Error: updating not allowed |

Rotation: **`delete certificate`** then deploy again.

---

## Describe output

- **`kind: Certificate`**
- **`data`**: leaf cert, private key, chain metadata, expiry helpers (**sensitive**)

---

## Examples

| File | Shows |
|------|--------|
| [minimal.yaml](examples/minimal.yaml) | Leaf signed by named CA (`direct`) |
| [self-signed-leaf.yaml](examples/self-signed-leaf.yaml) | **`ca.type: self-signed`** |
