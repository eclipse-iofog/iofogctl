# Secret (`kind: Secret`)

**Secrets** store sensitive string data in the **Controller API** for an ECN namespace. Microservices inject values via **`valueFromSecret`**, **VolumeMount** resources can mount a Secret, and **CertificateAuthority** import flows may reference PoT Secrets.

Secrets are **not** Kubernetes `Secret` objects unless you use a CA **`type: k8s-secret`** import path documented under [Certificate authority](../certificate-authority/README.md).

## What gets deployed

1. Namespace must have a reachable **Controller** (`connect` or control plane deploy).
2. CLI unmarshals **`spec.type`** (optional) and top-level **`data`** (required map of keys → strings). For **`type: tls`**, values are **base64-encoded PEM** — see [schema — TLS secrets](schema.md#tls-secrets-spec-type-tls).
3. **`GetSecret`** — missing → **create**; exists → **update** (data only).

Identity is **`metadata.name`** (lowercase alphanumeric). Names are unique per Controller namespace.

## Document shape

Unlike most kinds, payload bytes live in a sibling **`data`** block (same pattern as Kubernetes), not under `spec.data`:

```yaml
apiVersion: datasance.com/v3
kind: Secret
metadata:
  name: postgres-creds
  namespace: my-ecn
spec:
  type: Opaque   # optional
data:
  password: changeme
  username: app
```

## Prerequisites

- Controller online in target namespace.
- Deploy **before** microservices that reference `valueFromSecret: <this-name>/<key>` or VolumeMounts that set **`secretName`**.

## Deploy

```bash
potctl deploy -f secret.yaml -n my-ecn
potctl get secrets -n my-ecn
potctl describe secret postgres-creds -n my-ecn
potctl delete secret postgres-creds -n my-ecn
```

Re-deploying the same **`metadata.name`** **updates** keys in **`data`**. Keys omitted from a new deploy file are **not** automatically deleted unless Controller merge rules say otherwise — prefer explicit **`delete secret`** when rotating names.

## Deploy order

In multi-document YAML, **`Secret`** runs early (before **CertificateAuthority**, **Certificate**, **ConfigMap**, **VolumeMount**, and **Microservice**). See [operator index — deploy order](../README.md#ecn-resource-deploy-order).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/tls-material.yaml](examples/tls-material.yaml)
- Microservice [environment variables](../workloads/microservice/schema.md#container-environment-env)
- [Volume mount](../volume-mount/README.md)
- [Certificate authority](../certificate-authority/README.md)

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
