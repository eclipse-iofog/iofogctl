# Registry schema

Reference for **`kind: Registry`** deploy YAML. Fields live under **`spec`** (not nested under a generic apps struct).

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3
kind: Registry
metadata:
  namespace: <optional>        # must match CLI -n if set
spec:
  # fields below
```

Strict YAML decoding rejects unknown keys and **retired v3.8 registry certificate booleans** (`requiresCert`, `isSecure`, `certificate`).

---

## Field reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | **Yes** | Registry host or URL (e.g. `registry.example.com`, `https://gcr.io`, `https://huggingface.co`) |
| `private` | bool | No | Default **`false`**. When **`true`**, credentials rules apply per `type` |
| `type` | string | No | Default **`oci`**. Allowed: **`oci`**, **`hf`** (Hugging Face hub style) |
| `username` | string | Conditional | **Private `oci`:** required with `password`. **Private `hf`:** optional |
| `password` | string | Conditional | **Private `oci`:** required with `username`. **Private `hf`:** required (token) |
| `email` | string | No | Optional for both types |
| `ca` | string | No | **Base64-encoded** PEM for a private CA (custom TLS). Not the old `certificate` boolean flag |
| `insecure` | bool | No | Skip TLS verification for registry endpoint when **`true`** |
| `id` | int | Patch only | **Create:** omit or `0`. **Update:** must be **> 0** (Controller registry id) |

### Semantics

- **`private: false`** → Controller stores a **public** registry (`IsPublic` on API).
- **`type`** is normalized to lowercase; invalid values error at CLI validation.
- **`spec.ca`:** supply the CA certificate as a **single base64 string** (standard PEM content encoded), not a file path in deploy YAML.

---

## Validation (CLI)

| Condition | Error |
|-----------|--------|
| Empty `url` | URL cannot be empty |
| `type` not `oci` or `hf` | type must be oci or hf |
| `private: true`, `oci`, missing user/pass | Username and password are required |
| `private: true`, `hf`, missing password | Password cannot be empty for private hf registry |
| Legacy `requiresCert` / `isSecure` / `certificate` | Strict unmarshal failure |

Controller may return additional **400/409** bodies; the CLI surfaces the HTTP error text.

---

## Initial deploy (create)

**Rules:**

1. Do **not** set **`spec.id`**, or set **`id: 0`**.
2. Provide **`url`**, **`type`** (or rely on default `oci`), and **`private`** as needed.
3. Run `deploy -f`, then `get registries` to read the assigned **ID**.

**Public OCI example:** [examples/oci-public.yaml](examples/oci-public.yaml)

**Private OCI example:** [examples/oci-private.yaml](examples/oci-private.yaml)

**Public HF example:** [examples/hf-public.yaml](examples/hf-public.yaml)

**Private HF example:** [examples/hf-private.yaml](examples/hf-private.yaml)

**Insecure TLS example:** [examples/oci-insecure.yaml](examples/oci-insecure.yaml)

**Private CA (base64 `ca`):** [examples/oci-private-ca.yaml](examples/oci-private-ca.yaml)

---

## Re-deploy (update / patch)

**Rules:**

1. Set **`spec.id`** to the existing registry id from `get registries` or `describe registry <ID>`.
2. Include fields you want to change; omitted pointer fields may be sent as null on update depending on Controller API — treat re-deploy as **full spec intent** for url, type, private, credentials, `ca`, and `insecure`.
3. Changing **url** on an id in use may fail at Controller with **409** if microservices or fleet artifacts still reference the old endpoint.

**Example:** [examples/patch-update.yaml](examples/patch-update.yaml)

---

## Describe output

`describe registry <ID>` emits deploy-shaped YAML under **`spec`**, including **`id`**, **`url`**, **`private`**, **`type`**, **`username`**, **`email`**, **`ca`**, **`insecure`**, and **`password`** (sensitive — handle like a secret).

There is no **`metadata.name`** in describe output for registries.

---

## Delete

```bash
potctl delete registry <ID> -n my-ecn
```

Deletion may fail if the registry is still referenced; read the Controller error body.

---

## Retired fields (do not use)

| Removed (pre–v3.9) | Replacement |
|--------------------|-------------|
| `isSecure` | `type`, `ca`, `insecure` |
| `requiresCert` | `ca` (base64 PEM) |
| `certificate` (registry flag) | `ca` (base64 PEM) |
