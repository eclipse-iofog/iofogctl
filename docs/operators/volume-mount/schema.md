# VolumeMount schema

```yaml
apiVersion: datasance.com/v3
kind: VolumeMount
metadata:
  name: <mount-name>
  namespace: my-ecn
spec:
  configMapName: <configmap-name>   # one of
  secretName: <secret-name>         # one of
```

Describe output may also include read-only **`uuid`** and **`version`** from Controller.

---

## Field reference

| Field | Required | Description |
|-------|----------|-------------|
| `configMapName` | One of | ConfigMap **`metadata.name`** in same namespace |
| `secretName` | One of | Secret **`metadata.name`** in same namespace |

CLI does not enforce mutual exclusivity; Controller should reject invalid combinations. Operator practice: set **exactly one** backing resource.

---

## Deploy behavior

| State | Action |
|-------|--------|
| Absent | `CreateVolumeMount` |
| Present | `UpdateVolumeMount` (change backing names) |

---

## Agent linking

### Controller auto-attach (microservice)

On **microservice** create/update, if **`spec.container.volumes`** includes an entry with:

- **`type: volumeMount`**
- **`hostDestination`** = existing VolumeMount **`metadata.name`**

the **Controller** **links** that VolumeMount to **`spec.agent`** automatically (edgelet on the host that runs the microservice). No CLI **`attach volume-mount`** is required for that agent when deploy is driven by the microservice manifest.

Deploy the **VolumeMount** document (and Secret/ConfigMap) **before** the microservice in multi-doc YAML.

### Manual attach / detach

| Command | Syntax |
|---------|--------|
| Attach | `attach volume-mount NAME AGENT [AGENT…]` |
| Detach | `detach volume-mount NAME AGENT [AGENT…]` |

Calls **`LinkVolumeMount`** / **`UnlinkVolumeMount`**. Use for pre-provisioning or extra agents without a referencing microservice.

---

## Microservice reference

```yaml
spec:
  agent:
    name: edge-node-1
  container:
    volumes:
      - type: volumeMount
        hostDestination: nats-creds
        containerDestination: /etc/nats/creds
        accessMode: ro
```

| Field | Value |
|-------|--------|
| **`type`** | **`volumeMount`** |
| **`hostDestination`** | VolumeMount **`metadata.name`** (not the Secret/ConfigMap name) |
| **`containerDestination`** | Path inside the container |
| **`accessMode`** | **`ro`** or **`rw`** |

See [microservice schema — container volumes](../workloads/microservice/schema.md#container-volumes).

---

## CLI commands

| Command | Args |
|---------|------|
| `get volume-mounts` | `-n` |
| `describe volume-mount` | NAME |
| `delete volume-mount` | NAME |
| `attach volume-mount` | NAME agents… |
| `detach volume-mount` | NAME agents… |

---

## Examples

| File | Shows |
|------|--------|
| [from-secret.yaml](examples/from-secret.yaml) | Secret-backed mount |
| [from-configmap.yaml](examples/from-configmap.yaml) | ConfigMap-backed mount |
| [stack.yaml](examples/stack.yaml) | Secret + VolumeMount multi-doc |
