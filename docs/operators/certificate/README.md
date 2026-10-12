# Certificate (`kind: Certificate`)

**Certificate** requests a **leaf TLS certificate** signed by an existing **CertificateAuthority** in the same ECN namespace. Certificates are stored in the Controller PKI service.

## What gets deployed

1. Controller reachable.
2. CLI sends **`CreateCertificate`** with subject, SAN hosts, expiration, and issuing **`ca`** reference.
3. If a certificate with the same **name** already exists → deploy **fails** (no update). Delete and recreate to rotate.

## Prerequisites

- A CA already registered (`get certificates` shows `IsCA` entries, or deploy **`kind: CertificateAuthority`** first).
- **`spec.ca`**: **`self-signed`** for Controller-issued leaf without a named CA, or **`direct`** + **`secretName`** for a CA from **`kind: CertificateAuthority`** (see [schema.md](schema.md)).

## Deploy

```bash
potctl deploy -f cert.yaml -n my-ecn
potctl get certificates -n my-ecn
potctl describe certificate api-tls -n my-ecn
potctl delete certificate api-tls -n my-ecn
```

`describe` includes **private keys** in **`data`** — protect output files.

## Deploy order

**Certificate** after **CertificateAuthority** (and any **Secret** used for CA import) in multi-doc YAML.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/self-signed-leaf.yaml](examples/self-signed-leaf.yaml)
- [Certificate authority](../certificate-authority/README.md)
