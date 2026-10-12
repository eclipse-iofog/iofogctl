# RuntimeClass schema

Reference for **Controller** **`kind: RuntimeClass`** deploy YAML.

**Important:** **`handler` is a top-level field**, sibling to `metadata` — not under `spec`.

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: RuntimeClass
metadata:
  name: <runtime-class-name>   # required; lowercase alphanumeric
  namespace: <optional>
handler: <handler-key>         # required — see handler list below
```

Describe output follows the same shape: **`handler`** at root, **no `spec:`** block.

---

## Field reference

| Field | Location | Required | Description |
|-------|----------|----------|-------------|
| `metadata.name` | metadata | **Yes** | Controller resource name |
| `handler` | **document root** | **Yes** | Runtime key consumed by edgelet / microservice `container.runtime` |

The CLI accepts **`handler` in `spec`** only as an internal fallback when dispatch strips the root field — **author manifests with root `handler`** for round-trip with `describe`.

---

## Supported handler keys (edgelet WASM shims)

Align with agent bootstrap **`package.wasm.<handler>`** keys:

| Handler | Typical shim |
|---------|----------------|
| `spin` | Spin |
| `slight` | Slight |
| `lunatic` | Lunatic |
| `wws` | Wasmer WWS |
| `wasmedge` | WasmEdge |
| `wasmer` | Wasmer |
| `wasmtime` | Wasmtime |
| `edgelet-wasmtime` | Edgelet Wasmtime |

Deploying a RuntimeClass **does not** install binaries — bootstrap must place the shim on the agent host first.

---

## Validation (CLI)

| Condition | Error |
|-----------|--------|
| Missing `metadata.name` | Name must be specified |
| Invalid name | Lowercase alphanumeric |
| Missing `handler` | handler must be specified |

Controller may reject unknown handlers or link requests with **400/409**.

---

## Initial deploy (create)

1. Install WASM shim on target agents ([edgelet-bootstrap](../agent/edgelet-bootstrap.md)).
2. Deploy RuntimeClass with root **`handler`**.
3. **`attach runtimeclass`** to each edgelet agent that should expose the runtime.

**Examples:** [examples/minimal.yaml](examples/minimal.yaml), [examples/wasmtime.yaml](examples/wasmtime.yaml)

---

## Re-deploy (update / patch)

Same **`metadata.name`** → **UpdateRuntimeClass** with a new **`handler`** value.

Changing handler on linked agents can affect running microservices; expect Controller-side validation.

---

## Describe output

```yaml
kind: RuntimeClass
metadata:
  name: spin
  namespace: my-ecn
handler: spin
status:
  linkedAgents:
    - edge-a
```

---

## Attach / detach constraints

- **Edgelet agents only** for link API.
- Agent arguments are **names**, resolved to fog UUIDs.
- **409** responses may list blocking microservice UUIDs — CLI prints the body.

---

## Delete

```bash
potctl delete runtimeclass <name> -n my-ecn
```

Detach agents first if Controller requires no active links.

---

## Contrast with edgelet RuntimeClass manifest

Edgelet manifests use **`apiVersion: edgelet.iofog.org/v1`**, applied with **`edgelet deploy -f`** during control plane or agent install — not via `potctl deploy -f` user YAML for Controller RuntimeClass.

Do not conflate the two layers when troubleshooting “runtime not found” on an agent.
