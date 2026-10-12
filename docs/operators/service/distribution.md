# Service distribution tags

Edge agents expose a **local TCP listener** on **`bridgePort`** only when their **agent tags** select the Service. That is separate from **hub** provisioning (`provisioningStatus: ready`) and from **Kubernetes Service** annotations.

## Agent tag prefix

Default prefix: **`service.iofog.org/tag`** (Controller `SERVICE_ANNOTATION_TAG` / flavor `serviceAnnotationTag`).

The prefix is stored on the **agent**. The **Service** stores the **bare token** in **`metadata.tags`**.

| Agent tag | Matches services |
|-----------|------------------|
| `service.iofog.org/tag: checkout` | Services whose `metadata.tags` include **`checkout`** |
| `service.iofog.org/tag: all` | **Every** Service |
| `checkout` (bare on agent) | **Nothing** — prefix required |
| `service.iofog.org/tag:checkout` (no space) | **Nothing** — stored value includes `": "` after the prefix |

Set tags on **`Agent`** / **`LocalAgent`** / **`AgentConfig`** at install or update:

```yaml
metadata:
  name: plant-a
  tags:
    - "service.iofog.org/tag: checkout"
    - generic-agent-tag
```

See [stack-with-agent-tags.yaml](examples/stack-with-agent-tags.yaml) and [Agent service distribution tags](../agent/README.md#service-distribution-tags).

## Service `metadata.tags`

| Tag in Service YAML | Used for distribution? | Used for K8s annotations? |
|---------------------|------------------------|---------------------------|
| `checkout` | yes — matches agents with `service.iofog.org/tag: checkout` | yes — `checkout: ""` |
| `service.beta.kubernetes.io/aws-load-balancer-type: nlb` | **no** (contains `:`) | yes — full key/value |
| tag with `=` | **no** | per Controller rules |

On **service create/update**, Controller:

1. Includes agents tagged **`service.iofog.org/tag: all`**.
2. For each service tag **without** `:` or `=`, includes agents tagged **`service.iofog.org/tag: {token}`**.
3. Enqueues fog-platform reconcile for each matching agent.

On **fog reconcile**, the agent’s router **`tcpListeners`** named `*-listener` are rebuilt. Each matched service gets a listener:

```json
{
  "name": "checkout-listener",
  "port": "10024",
  "address": "checkout"
}
```

`address` is the Service name; the router bridges to the **`tcpConnector`** with the same `address` (on the backend agent or default router per [schema.md](schema.md)).

Agents with router mode **`none`** do not get listeners.

## Example

| Agent | Tags | Listeners after reconcile |
|-------|------|---------------------------|
| plant-a | `service.iofog.org/tag: checkout` | `checkout-listener` only |
| plant-b | `service.iofog.org/tag: all` | listener for **every** service |
| plant-c | `checkout` only | **none** |

The **connector** for a microservice is written on the agent that **runs** the microservice whether or not distribution tags match. **Listeners** on that same agent appear only if its tags select the service.

Local clients on a tagged agent connect to **`bridgePort`**; the router forwards via shared **`address`** to the connector.

More detail: [service-handoff.md](../service-handoff.md) (Tags, Distribution, router storage).
