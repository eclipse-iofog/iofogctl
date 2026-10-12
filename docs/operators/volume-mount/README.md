# Volume mount (`kind: VolumeMount`)

A **VolumeMount** binds a **Secret** or **ConfigMap** to a Controller-managed mount that edgelet exposes on agents. Microservices consume it via **`spec.container.volumes`** with **`type: volumeMount`** and **`hostDestination`** set to the VolumeMount **name**.

Unlike **`kind: Volume`**, VolumeMount is a **Controller API** object. Deploy creates or updates the definition; agents receive content when the mount is **linked** to them.

## What gets deployed

1. Controller reachable.
2. **`spec`** references **`secretName`** and/or **`configMapName`** (use **one** backing resource in practice).
3. **`GetVolumeMount`** — missing → **create**; exists → **update**.

Deploy registers the VolumeMount resource only — it does not copy files by itself.

## Linking to agents

### Automatic (microservice deploy)

When a **Microservice** lists a volume mapping with **`type: volumeMount`** and **`hostDestination: <volumeMount-name>`**, the **Controller** **automatically attaches** that VolumeMount to **`spec.agent`** (the edgelet host that will run the microservice). You do **not** need a separate **`attach volume-mount`** for that agent if the microservice YAML already references the mount.

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

Requirements:

- **`kind: VolumeMount`** (and backing Secret/ConfigMap) deployed **before** the microservice in the same **`deploy -f`** batch or an earlier apply.
- VolumeMount **`metadata.name`** equals **`hostDestination`**.

### Manual (`attach` / `detach`)

Use when you need the mount on agents **before** any microservice references it, or on **additional** agents without scheduling a workload:

```bash
potctl attach volume-mount nats-creds edge-node-2 -n my-ecn
potctl detach volume-mount nats-creds edge-node-2 -n my-ecn
```

Same CLI pattern as **Model**, **Knowledge**, and **RuntimeClass**. Agent **status** lists linked volume mounts (`describe agent`).

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: VolumeMount
metadata:
  name: nats-creds
  namespace: my-ecn
spec:
  secretName: nats-account-secret
```

## Prerequisites

- Backing **Secret** or **ConfigMap** deployed first.
- Target agent exists before microservice deploy (for automatic attach).

## Deploy

```bash
potctl deploy -f volume-mount.yaml -n my-ecn
potctl get volume-mounts -n my-ecn
potctl describe volume-mount nats-creds -n my-ecn
potctl delete volume-mount nats-creds -n my-ecn
```

## Deploy order

```text
Secret | ConfigMap → VolumeMount → Microservice (volumeMount mapping → Controller attach)
```

Optional: **`attach volume-mount`** for agents without a referencing microservice. See [operator index](../README.md#ecn-resource-deploy-order).

## Related documents

- [schema.md](schema.md)
- [examples/from-secret.yaml](examples/from-secret.yaml)
- [examples/from-configmap.yaml](examples/from-configmap.yaml)
- [examples/stack.yaml](examples/stack.yaml)
- [Secret](../secret/README.md) · [ConfigMap](../configmap/README.md)
- Microservice [volumes](../workloads/microservice/schema.md#container-volumes)
