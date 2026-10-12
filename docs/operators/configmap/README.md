# ConfigMap (`kind: ConfigMap`)

**ConfigMaps** store non-secret configuration strings in the **Controller API**. Microservices read values via **`valueFromConfigMap`**, and **VolumeMount** resources can expose a ConfigMap as files on agents.

## What gets deployed

1. Controller must be reachable in the namespace.
2. CLI unmarshals **`spec`** flags (`immutable`, `useVault`) and top-level **`data`**.
3. **`GetConfigMap`** — missing → **create**; exists → **update** (data and selective spec fields).

Identity is **`metadata.name`** (lowercase alphanumeric).

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: ConfigMap
metadata:
  name: platform-config
  namespace: my-ecn
spec:
  immutable: false    # optional
  useVault: false     # optional
data:
  features: '{"darkMode":true}'
  log_level: info
```

Payload keys live in root-level **`data`**, not `spec.data`.

## Prerequisites

- Controller in namespace.
- Deploy before **VolumeMount** / microservice references.

## Deploy

```bash
potctl deploy -f configmap.yaml -n my-ecn
potctl get configmaps -n my-ecn
potctl describe configmap platform-config -n my-ecn
potctl delete configmap platform-config -n my-ecn
```

`describe` may emit multiline values using YAML literal block style for readability.

## Deploy order

**ConfigMap** runs after **Secret** / **Certificate** kinds and before **VolumeMount** and workloads. See [operator index](../README.md#ecn-resource-deploy-order).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/immutable.yaml](examples/immutable.yaml)
- Microservice [env](../workloads/microservice/schema.md#container-environment-env)
- [Volume mount](../volume-mount/README.md)
