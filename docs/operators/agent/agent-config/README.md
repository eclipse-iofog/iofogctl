# AgentConfig (`kind: AgentConfig`)

Use **`kind: AgentConfig`** to create or update an agent record on the **Controller API** without running edgelet bootstrap on a host.

## When to use

| Scenario | Approach |
|----------|----------|
| Agent deploy with host install | Usually omit AgentConfig; CLI **synthesizes** config from `Agent` / `LocalAgent` **`spec.config`** |
| Config-only change | Deploy standalone AgentConfig (agent must exist or will be created) |
| Explicit ordering in multi-doc YAML | Separate AgentConfig document before agents that depend on upstream router names |
| System agent on CP | CP deploy builds AgentConfig internally; not a standalone `kind: Agent` |

## Document shape

```yaml
apiVersion: datasance.com/v3
kind: AgentConfig
metadata:
  name: <agent-name>    # must match Controller agent name
  namespace: <optional>
  tags: [<optional>]
spec:
  # AgentConfiguration fields (see schema.md)
```

**`metadata.name`** is the agent name. Fields under **`spec`** map to **`AgentConfiguration`** (SDK + CLI wrappers).

## Deploy behavior

1. Requires control plane with Controller API in the namespace.
2. **Validate** router/NATS rules for edge vs system agents.
3. **`Process`:** resolve **`upstreamRouters`** and **`upstreamNatsServers`** from agent **names** to UUIDs; set **host** for interior routers when needed.
4. **Create** agent if missing, else **PATCH** configuration.
5. **Wait** for agent platform ready (unless system agent with quiet progress).

Controller **platform reconcile** then creates **`system-{agentName}`** with **`router`** and **`nats`** system microservices, fabric CAs, and upstream links — [Networking topology](../networking/README.md).

### Pairing with `Agent` / `LocalAgent`

In one `deploy -f` file:

- AgentConfig executors run in a **batch before** agent bootstrap kinds.
- If an AgentConfig document shares **`metadata.name`** with an agent, the CLI merges **host** (from agent spec / config host) and **tags**.

After bootstrap, the agent facade may **update** AgentConfig again when `spec.config` is present on the agent document.

## Describe output

`potctl describe agent-config NAME` emits nested **`routerConfig`** and **`natsConfig`** blocks aligned with Controller storage — use as reference for round-trip YAML.

## Related

- [Networking topology](../networking/README.md)
- [schema.md](schema.md)
- [examples/minimal-edge.yaml](examples/minimal-edge.yaml)
- [examples/multi-doc-with-agent.yaml](examples/multi-doc-with-agent.yaml)
- [README — deploy sequence](../README.md#deploy-sequence)
