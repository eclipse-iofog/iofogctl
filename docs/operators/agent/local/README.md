# Local agent (`kind: LocalAgent`)

Use **`kind: LocalAgent`** to install **edgelet** on the **same machine** that runs **potctl** / **iofogctl** and register it with the Controller.

Typical uses: development, CI, and desktop testing alongside a local or remote control plane.

## What gets deployed

Same pipeline as [remote agents](../remote/README.md), except:

- No SSH; host OS is **`runtime.GOOS`** of the CLI process.
- Default agent name **`local`** when `metadata.name` is omitted.
- **Port 54321** must be free for non-system local agents (conflicts with default local system agent layout).

## Prerequisites

- Control plane deployed in the namespace.
- For **native** linux local install, sudo may be required for bootstrap layers.
- For **container** edgelet on darwin/windows, docker (or supported engine) must already be installed — the CLI does not auto-install engines.

## Deploy

```bash
potctl deploy -f local-agent.yaml -n my-ecn
```

## Host identity

| Field | Precedence for API / provision |
|-------|--------------------------------|
| `spec.config.host` | Preferred Controller **host** when set |
| `spec.host` | Fallback |
| _(default)_ | `localhost` |

## Desktop container behavior

When `spec.config.deploymentType: container` on **darwin** or **windows**, bootstrap skips host config materialization and uses desktop-specific wait/configure paths. See [edgelet-bootstrap — desktop container](../edgelet-bootstrap.md#desktop-container-local-behavior-darwin--windows).

## Related documents

- [schema.md](schema.md)
- [examples/minimal.yaml](examples/minimal.yaml)
- [examples/desktop-container.yaml](examples/desktop-container.yaml)
- [lifecycle.md](../lifecycle.md)
