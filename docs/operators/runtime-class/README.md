# RuntimeClass (`kind: RuntimeClass`)

Use **`kind: RuntimeClass`** (Controller user API) to register a **WASM or alternate container runtime handler** name on the Controller. Link runtime classes to **edgelet agents** so microservices can set **`spec.container.runtime`**.

This is **not** the edgelet install manifest:

| Layer | apiVersion | When |
|-------|------------|------|
| **Controller (this doc)** | `<group>/v3` | Operator **`deploy -f`** — fleet + microservice runtime selection |
| **Edgelet bootstrap** | `edgelet.iofog.org/v1` | Agent install — WASM shim binaries on the host |

Edgelet install-time RuntimeClass and **`package.wasm`** shims: [edgelet-bootstrap](../agent/edgelet-bootstrap.md).

## What gets deployed

1. Controller API available.
2. CLI reads **`handler` from the document root** (not `spec.handler`).
3. Upsert by **`metadata.name`**: create or update handler string.

## Fleet linking

Runtime classes attach only to **edgelet** agents (not arbitrary container-only fogs):

```bash
potctl attach runtimeclass spin edge-a -n my-ecn
potctl detach runtimeclass spin edge-a -n my-ecn
potctl describe runtimeclass spin -n my-ecn
```

If Controller returns **400** for a non-edgelet agent, the CLI prefixes:

`cannot attach runtimeclass: RuntimeClass can only be linked to edgelet agents`

## WASM handler names

The **`handler`** value must match a runtime installed on the agent (edgelet bootstrap / shim list). Canonical keys:

| Handler | Notes |
|---------|--------|
| `spin` | Spin shim |
| `slight` | Slight shim |
| `lunatic` | Lunatic shim |
| `wws` | Wasmer WWS |
| `wasmedge` | WasmEdge |
| `wasmer` | Wasmer |
| `wasmtime` | Wasmtime |
| `edgelet-wasmtime` | Edgelet-bundled Wasmtime |

Installing shims is **linux native** edgelet/docker scope — see bootstrap doc.

## Prerequisites

- **`metadata.name`**: lowercase alphanumeric (often same as handler, not required).
- **`handler`**: required at **YAML root**.

## Deploy order

```text
Registry → Model → Knowledge → RuntimeClass → MicroserviceTemplate → CatalogItem → Application → Microservice
```

## Deploy

```bash
potctl deploy -f runtimeclass.yaml -n my-ecn
potctl get runtimeclass -n my-ecn
potctl delete runtimeclass spin -n my-ecn
```

## Microservice usage

Set container runtime on a microservice (see [Microservice](../workloads/microservice/schema.md)):

```yaml
spec:
  container:
    runtime: spin
```

The runtime string must match a **RuntimeClass** name/handler you deployed and attached to the target agent.

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/wasmtime.yaml](examples/wasmtime.yaml)
- [Agent bootstrap](../agent/edgelet-bootstrap.md)

## Mirror flavors

| | **Datasance (potctl)** | **Eclipse (iofogctl)** |
|---|------------------------|-------------------------|
| Deploy YAML `apiVersion` | `datasance.com/v3` | `iofog.org/v3` |
