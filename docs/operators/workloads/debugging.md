# Workload debugging — logs and exec

Stream **container logs** and open **interactive shells** for **user microservices** and **system microservices**. For **edgelet / fog daemon** logs and **node debug exec**, see [agent/debugging.md](../agent/debugging.md).

Commands use the plural **`logs`** subcommand:

```bash
potctl logs microservice APPLICATION/MSVC_NAME -n my-ecn
potctl exec microservice APPLICATION/MSVC_NAME -n my-ecn
```

---

## Fully qualified names

| Pattern | Example | Lookup |
|---------|---------|--------|
| User app | `myapp/api` | `GetMicroserviceByName` |
| System app | `system-edge01/debug` | Falls back to `GetSystemMicroserviceByName` when user app missing |

Use **`get microservices`** vs **`get system-microservices`** to discover names.

**Describe:**

| Resource | Command |
|----------|---------|
| User microservice | `describe microservice myapp/api` |
| System microservice | `describe system-microservice system-edge01/debug` |

`describe microservice` **does not** emit YAML for legacy **system catalog** microservices (catalog ids ≤ 3). Use **`describe system-microservice`** for those.

---

## `logs microservice`

### Purpose

Tail **stdout/stderr of the workload container** on the agent hosting the microservice.

### Flags

Same as agent fog logs (remote path):

| Flag | Default | Range / format |
|------|---------|----------------|
| `--tail` | `100` | 1–5000 |
| `--follow` | `true` | Stream until stop |
| `--since` | — | RFC3339 |
| `--until` | — | RFC3339 |

### How it works

1. Resolve microservice (user app, else **system** app).
2. Require status **`RUNNING`** — otherwise CLI errors.
3. Load agent from namespace config + Controller.
4. **Remote agent:** WebSocket **`DialMicroserviceLogs`** or **`DialSystemMicroserviceLogs`** with tail options.
5. **LocalAgent:** Read docker/podman logs for container **`iofog_<microservice-uuid>`** (one-shot dump; tail/follow flags **not** applied on local engine path).

### Examples

```bash
potctl logs microservice wasm/wasm -n my-ecn
potctl logs microservice system-lima/debug -n my-ecn --follow=false --tail=200
```

---

## `exec microservice`

### Purpose

Interactive **WebSocket exec** into the **running workload container**. No **`attach exec`** step (unlike **`exec agent`**, which may provision a debug system microservice).

### Requirements

- Microservice **RUNNING**
- Controller auth with **Developer** role (microservice exec) — see error table below
- **System microservices:** Controller may require **SRE** — CLI surfaces *Only SRE can access system microservices*

### How it works

1. **`lookupMicroservice`** — user API, then system API on “Invalid application id”.
2. Verify **RUNNING**.
3. **`DialMicroserviceExecWithOptions`** or **`DialSystemMicroserviceExecWithOptions`**.
4. Local **interactive terminal** (stdin/stdout/stderr over WebSocket).
5. Session close prints success message.

**Quota:** up to **3** concurrent exec sessions per microservice (Controller **409**).

### Examples

```bash
potctl exec microservice myapp/api -n my-ecn
potctl exec microservice system-edge01/debug -n my-ecn
```

---

## Compare: workload vs agent debugging

| | **`logs agent`** | **`logs microservice`** |
|---|------------------|-------------------------|
| Target | Edgelet / fog daemon | App or system **container** |
| API | `DialFogLogs` | `DialMicroserviceLogs` / system variant |

| | **`exec agent`** | **`exec microservice`** |
|---|------------------|-------------------------|
| Target | Debug **system** container on node | **Your** (or system) workload container |
| Provision | May call `AttachExecToAgent` | None |
| Typical role | **SRE** (node exec) | **Developer** (microservice exec) |

---

## System microservice rebuild

Operators may **`rebuild system-microservice`** after image or config changes on Controller-managed system workloads:

```bash
potctl rebuild system-microservice system-edge01/debug -n my-ecn
```

User **`kind: Microservice`** deploy **cannot create** system microservices (SDK returns error).

---

## Troubleshooting

| Message | Likely cause |
|---------|----------------|
| Microservice is not currently running | Status not RUNNING — wait or fix deploy |
| Insufficient permissions… Developer | RBAC — need Developer for user MS exec |
| Only SRE can access system microservices | Use SRE account or avoid exec on system MS |
| Maximum of 3 concurrent exec sessions | Close other exec sessions |
| Log session relay unavailable | Controller HA / NATS relay — retry |
| Timeout waiting for agent connection | Agent offline or network path broken |

---

## Related

- [Microservice schema](microservice/schema.md) — `status`, `execStatus` on **`describe`**
- [Agent debugging](../agent/debugging.md) — fog logs and node shell
