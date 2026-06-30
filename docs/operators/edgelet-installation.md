# Edgelet installation and custom scripts

Operator reference for how **potctl** and **iofogctl** install edgelet on remote hosts, local hosts, and control-plane system agents. Both CLIs use the same layered bootstrap logic; only `apiVersion`, image registry, and banner strings differ by build flavor.

Edgelet is installed through **embedded shell layers** staged under a transient directory, executed in a fixed order, followed by Go-side configuration and provisioning. The CLI never invokes the upstream edgelet monolith `install.sh` at runtime.

**Applies to v3.8** (greenfield edgelet platform). Retired paths (`iofog-agent`, monolith install) are out of scope.

---

## Table of contents

1. [Platform and deployment matrix](#platform-and-deployment-matrix)
2. [When installation runs](#when-installation-runs)
3. [Default bootstrap pipeline](#default-bootstrap-pipeline)
4. [Desktop container local behavior](#desktop-container-local-behavior-darwin--windows)
5. [Embedded script inventory](#embedded-script-inventory)
6. [Bootstrap environment variables](#bootstrap-environment-variables)
7. [YAML reference: scripts block](#yaml-reference-scripts-block)
8. [CustomizeProcedures merge logic](#customizeprocedures-merge-logic)
9. [Staging vs execution](#staging-vs-execution)
10. [Remote vs local execution](#remote-vs-local-execution)
11. [Go steps outside scripts](#go-steps-outside-scripts)
12. [Uninstall and delete](#uninstall-and-delete)
13. [Cookbook examples](#cookbook-examples)
14. [Troubleshooting](#troubleshooting)
15. [Code and asset map](#code-and-asset-map)

---

## Platform and deployment matrix

### Host OS detection

| Context | How OS is determined |
|---------|----------------------|
| **Remote** Agent / CP system agent | SSH `uname -s`, normalized to `linux`, `darwin`, or `windows` on first bootstrap |
| **Local** Agent / local CP | `runtime.GOOS` of the machine running the CLI |

Remote deploy assumes SSH access. Windows edge hosts use Windows path conventions inside scripts; scripts are still staged to `/tmp/edgelet-scripts` over SSH.

### Deployment combinations

| Host OS | Deploy kind | `deploymentType` | Typical use | Bootstrap runner |
|---------|-------------|------------------|-------------|------------------|
| **linux** | Remote `Agent`, CP system agent | `native` | Production edge nodes | SSH + `sudo env …` |
| **linux** | Remote `Agent`, CP system agent | `container` | docker/podman on host | SSH + `sudo env …` |
| **linux**, **darwin**, **windows** | `LocalAgent`, local CP | `native` | Local edgelet binary | Local `sh -c` (+ sudo on native linux) |
| **darwin**, **windows** | `LocalAgent`, local CP | `container` | Desktop dev with docker | Local shell, no sudo |

Defaults when omitted: `deploymentType: native`, `containerEngine: edgelet`.

### Paths by operating system

Scripts and Go agree on these layouts (see `assets/edgelet/scripts/lib/paths.sh` and `pkg/iofog/install/paths.go`):

| OS | Edgelet binary | Config directory | Config file | Script share dir (after `bundled.sh`) |
|----|----------------|------------------|-------------|---------------------------------------|
| **linux** | `/usr/local/bin/edgelet` | `/etc/edgelet` | `/etc/edgelet/config.yaml` | `/usr/share/edgelet` |
| **darwin** | `/usr/local/bin/edgelet` | `/etc/edgelet` | `/etc/edgelet/config.yaml` | `/usr/local/share/edgelet` |
| **windows** | `%ProgramFiles%\Edgelet\edgelet.exe` | `%ProgramData%\Edgelet\config` | `…\config\config.yaml` | `%ProgramData%\Edgelet\scripts` |

**Staging directory** (transient, all platforms): `/tmp/edgelet-scripts`

WASM shim staging (when enabled): `/tmp/edgelet-scripts/wasm/`

---

## When installation runs

Installation is split into **bootstrap** (scripts + Go config materialization) and **provision** (`edgelet config` / `edgelet provision`).

```text
┌─────────────────────────────────────────────────────────────────┐
│  Parse YAML → NewLocalEdgelet / NewRemoteEdgelet (defaults)     │
│  → CustomizeProcedures (optional scripts merge)                 │
│  → Airgap / WASM prep (if enabled)                              │
│  → Bootstrap(): stage scripts → pre-install → Go config         │
│     → post-install → bundled → WASM RuntimeClass (Go)           │
│  → AgentConfig via SDK (if spec.config set)                     │
│  → edgelet config / edgelet provision                           │
└─────────────────────────────────────────────────────────────────┘
```

| Deploy target | `Bootstrap()` when | Provision when | Where `scripts` lives in YAML |
|---------------|-------------------|----------------|--------------------------------|
| Remote **Agent** | `deploy -f` agent execute | Same deploy, after bootstrap | `spec.scripts` |
| **LocalAgent** | Local agent execute | Same deploy, after bootstrap | `spec.scripts` |
| Remote **ControlPlane** (per controller host) | `DeployHostEdgelet()` during CP deploy | Later in `deploySystemAgent()` | `controllers[].systemAgent.scripts` |
| **LocalControlPlane** | `installHostEdgelet()` during CP deploy | Later in `deployLocalSystemAgent()` | `systemAgent.scripts` |

**System agent note:** Edgelet is bootstrapped when the control-plane **host** is prepared. The system-agent step that follows only pushes AgentConfig to the controller API and runs provision—it does **not** run bootstrap again.

`isSystem: true` affects system-agent defaults (router/upstream, port checks, quieter progress text). It does **not** change script merge or execution.

---

## Default bootstrap pipeline

When no custom `scripts` block is present (or entrypoints are omitted—see [CustomizeProcedures](#customizeprocedures-merge-logic)), the CLI embeds and runs the full bundle from `assets/edgelet/scripts/`.

### Step reference

| # | Step | Script or Go | Purpose | Platform / skip notes |
|---|------|--------------|---------|------------------------|
| 0 | Stage scripts | Go | Write bundle to `/tmp/edgelet-scripts` | Remote: SCP + `install -m 755`; local: `os.WriteFile` |
| 0b | Prepare WASM | Go `PrepareWasm` | Resolve/extract shim binaries, set env | Skipped unless [WASM scope](#wasm-runtime-scope) allows |
| 1 | Prerequisites | `check_prereqs.sh` | Remote: passwordless sudo probe | **Local:** exits 0 immediately (`LOCAL_INSTALL=1`) |
| 2 | Detect init | `detect_init.sh` | OS, arch, init system (systemd, openrc, …) | Always runs |
| 3 | Dependencies | `install_deps.sh` | Install/check docker or podman | **No-op exit 0** when `containerEngine: edgelet`; on **darwin/windows** skips `configure_container_engine` |
| 3b | Configure engine | `configure_container_engine.sh` | Engine socket, groups, service | Only reached when deps layer runs (not edgelet engine) |
| 4 | Install edgelet | `install.sh` **or** `install_container.sh` | Binary download/install or container prep | Native → `install.sh` with `--skip-config` `--skip-start`; container → `install_container.sh` |
| 5 | WASM runtimes | `install_wasm_runtimes.sh` | Install pre-staged shims to `/usr/local/bin` | Command omitted unless WASM scope allows; script also no-ops on non-linux / container |
| — | **Materialize config** | Go `MaterializeEdgeletRuntime` | Write `config.yaml` + sample CA if missing | **Skipped** on [desktop container local](#desktop-container-local-behavior-darwin--windows) |
| 6 | Init units | `install_init_units.sh` | systemd/openrc/procd units, drop-ins | Linux-focused; desktop paths differ |
| 7 | Start | `start_edgelet.sh` | Enable/start daemon or container unit | Desktop native darwin: background daemon; windows native: platform-specific |
| 8 | Configure container | `configure_container_edgelet.sh` | Apply config inside running container | **Desktop container only**; exits 0 otherwise |
| 9 | Wait ready | `wait_edgelet_ready.sh` | Poll until edgelet API ready | Desktop container: alternate wait path |
| — | **RuntimeClass** | Go `DeployWasmRuntimeClasses` | Apply WASM RuntimeClass manifests | Only when WASM handlers configured |
| 10 | Bundled publish | `bundled.sh` | Copy scripts to canonical share dir | **Skipped** on desktop container local |
| — | **Provision** | Go `EdgeletProvisionCommands` | `edgelet config --a …`, optional cert, `edgelet provision` | After bootstrap in deploy flow |

### WASM runtime scope

WASM install runs only when **all** of the following hold (see `pkg/iofog/install/wasm/scope.go`):

| Requirement | Value |
|-------------|-------|
| Host OS | `linux` only |
| `deploymentType` | `native` (not `container`) |
| `containerEngine` | `edgelet` or `docker` (not `podman`) |
| `package.wasm` | At least one handler configured |

If `package.wasm` is set but scope fails, the CLI prints an informational skip message and continues.

WASM artifacts are **extracted on the CLI machine** (Go), never with `tar` on the remote host. Remote deploy SCPs raw shim binaries into the WASM staging subdirectory.

### Default install arguments

For **native** deploy, embedded `install.sh` receives CLI-generated flags:

- `--version=<edgelet version>`
- `--container-engine=<engine>`
- `--skip-config` and `--skip-start` (config and start are separate layers)
- Optional: `--arch=`, `--airgap`, `--bin-path=` (airgap)

For **container** deploy, `install_container.sh` receives:

- `--image=<container image>`
- `--engine=docker|podman`
- `--tz=<timezone>`

---

## Desktop container local behavior (darwin / windows)

Local install of **container** edgelet on **darwin** or **windows** sets `desktop_container_local` in scripts (`LOCAL_INSTALL=1` + `EDGELET_INSTALL_MODE=container` + desktop host OS).

| Behavior | Production linux | Desktop container local |
|----------|------------------|-------------------------|
| Sudo | Required on remote; local native linux uses sudo | **No sudo** (`maybe_sudo` runs commands directly) |
| Host `config.yaml` write (Go) | Yes | **Skipped** |
| `bundled.sh` | Publishes to share dir | **No-op** (scripts stay in stage dir) |
| `configure_container_edgelet.sh` | Exits 0 (not desktop) | Runs: applies `EDGELET_BOOTSTRAP_CONFIG_CMD` inside container |
| `wait_edgelet_ready.sh` | systemd/API poll | `wait_edgelet_api_desktop_container` |
| Extra env | Standard bootstrap env | `EDGELET_SCRIPT_STAGE_DIR`, `PATH` includes stage dir + `bin` |

Operators must **pre-install docker** (or podman where supported). The CLI does not auto-install container engines on remote or local hosts unless custom scripts do so.

---

## Embedded script inventory

All scripts live under `assets/edgelet/scripts/` and are embedded at CLI build time (`go:embed`).

### Top-level scripts

| Script | Role |
|--------|------|
| `check_prereqs.sh` | Verifies passwordless sudo on remote hosts; skipped locally |
| `detect_init.sh` | Detects OS, architecture, and init system; sourced by most other layers |
| `install_deps.sh` | Installs docker/podman when engine is not `edgelet`; sources `configure_container_engine.sh` on linux |
| `configure_container_engine.sh` | Configures docker/podman socket access; skipped on darwin/windows |
| `install.sh` | Downloads/installs edgelet binary, directories, receipt; supports airgap, upgrade, rollback flags |
| `install_container.sh` | Pulls/prepares containerized edgelet; desktop vs linux host branches |
| `install_wasm_runtimes.sh` | Installs Go-staged WASM shims; gated by `EDGELET_WASM_INSTALL=1` |
| `install_init_units.sh` | Writes systemd/openrc/procd units for native and container deployments |
| `start_edgelet.sh` | Starts or restarts edgelet daemon or container unit |
| `configure_container_edgelet.sh` | Desktop container: apply bootstrap config via edgelet CLI inside container |
| `wait_edgelet_ready.sh` | Waits for edgelet API / daemon readiness |
| `bundled.sh` | Publishes script bundle to OS share directory for OTA/helpers |
| `uninstall.sh` | Removes edgelet binary, units, data (optional `--remove-data`); OS-aware paths |

### Library scripts (`lib/`)

| Library | Provides |
|---------|----------|
| `lib/common.sh` | `die`, `info`, `maybe_sudo`, `desktop_container_local`, `is_desktop_container_host` |
| `lib/paths.sh` | OS-specific share, config, binary, runtime paths |
| `lib/receipt.sh` | Install receipt read/write for upgrade/rollback |
| `lib/binary.sh` | Binary download, checksum, airgap copy |
| `lib/service.sh` | Init-system helpers to start/stop/restart edgelet |
| `lib/container_engine.sh` | Engine detection and defaults |
| `lib/container_cli.sh` | docker/podman CLI wrappers |
| `lib/container_mounts.sh` | Bind mounts and FHS prep for container deploy |

Default `install.sh` is always invoked with `--skip-config` and `--skip-start` because Go writes config and `start_edgelet.sh` owns the start layer.

---

## Bootstrap environment variables

The CLI prefixes every script command with env from `EdgeletInstallConfig.bootstrapEnv()` (see `pkg/iofog/install/edgelet_scripts.go`).

| Variable | Meaning |
|----------|---------|
| `EDGELET_INSTALL_MODE` | `native` or `container` |
| `CONTAINER_ENGINE` | `edgelet`, `docker`, or `podman` |
| `DEPLOYMENT_TYPE` | Same as spec (`native` default) |
| `EDGELET_VERSION` | Binary/tag version (from ldflags or `package.version`) |
| `EDGELET_CONTAINER_IMAGE` | Container image when `deploymentType: container` |
| `EDGELET_TZ` | Timezone (default `UTC`) |
| `EDGELET_GITHUB_REPO` | Release base for binary download |
| `EDGELET_CONTAINER_ENGINE_URL` | Socket URL for docker/podman |
| `LOCAL_INSTALL=1` | Set for local bootstrap (skips remote sudo check) |
| `EDGELET_SCRIPT_STAGE_DIR` | Stage dir path (desktop container) |
| `PATH` | Prepended with `$EDGELET_SCRIPT_STAGE_DIR/bin` on desktop container |
| `EDGELET_BOOTSTRAP_CONFIG_CMD` | Quoted edgelet config command (desktop container) |
| `EDGELET_WASM_INSTALL=1` | Set when WASM shims are staged |
| `EDGELET_WASM_MANIFEST` | JSON manifest of staged WASM binaries |
| `EDGELET_WASM_RESTART_ENGINE=1` | Hint to restart engine after shim install |

**Remote:** commands use `sudo env VAR=… /tmp/edgelet-scripts/….sh` so sudo does not strip variables.

**Local native linux:** may wrap with `sudo env PATH=…`.

---

## YAML reference: scripts block

```yaml
scripts:
  dir: /path/to/scripts          # required when scripts block is present
  deps:
    entrypoint: install_deps.sh   # optional override
    args:
      - docker
      - native
  install:
    entrypoint: install.sh        # optional override
    args:
      - "--version=v1.0.0-rc.8"
      - "--skip-config"
      - "--skip-start"
  uninstall:
    entrypoint: uninstall.sh      # optional override
    args:
      - "--remove-data"
```

### Field rules

| Field | Required | Description |
|-------|----------|-------------|
| `dir` | Yes (if block present) | Directory on the **operator machine** (CLI reads files locally, then stages to host) |
| `*.entrypoint` | No | Filename **relative to staged script root**; empty means use embedded default for that layer |
| `*.args` | No | Arguments appended to the entrypoint command (shell-quoted) |

Only **`deps`**, **`install`**, and **`uninstall`** are YAML-overridable. Layers such as `detect_init`, `install_init_units`, `start_edgelet`, `configure_container_edgelet`, `wait_edgelet_ready`, and `bundled` always use embedded script **names** in the command list—they cannot be pointed at alternate entrypoints via YAML.

### Where to declare `scripts`

| Resource | YAML path |
|----------|-----------|
| Remote Agent | `spec.scripts` on `Agent` |
| Local Agent | `spec.scripts` on `LocalAgent` |
| Remote CP system agent | `controllers[].systemAgent.scripts` |
| Local CP system agent | `systemAgent.scripts` on `LocalControlPlane` |

Example fragment (remote agent with custom scripts):

```yaml
apiVersion: datasance.com/v3   # or iofog.org/v3 for iofogctl
kind: Agent
metadata:
  name: edge-node-1
spec:
  host: 10.0.0.5
  ssh:
    user: ubuntu
    keyFile: ~/.ssh/id_rsa
  package:
    version: 3.8.0-rc.2
  config:
    arch: amd64
    deploymentType: native
    containerEngine: edgelet
  scripts:
    dir: ./my-edgelet-scripts
    install:
      entrypoint: install.sh
      args:
        - "--version=3.8.0-rc.2"
        - "--skip-config"
        - "--skip-start"
```

---

## CustomizeProcedures merge logic

`CustomizeProcedures` in `pkg/iofog/install/edgelet_remote.go` and `edgelet_local.go` merges operator files with embedded defaults. Remote and local use **identical** merge rules.

### Algorithm

1. **Read all files** from `scripts.dir` (non-directories) into the stage list (`scriptNames` / `scriptContents`).
2. **Always append** embedded `check_prereqs.sh` (overwrites a same-named file from `dir` when staged).
3. For each overridable layer, if YAML `entrypoint` is **empty**, embed default scripts; if **set**, use custom behavior for that layer.
4. Fill missing post-install **entrypoint structs** from defaults (names and dest paths only).
5. Bind all dest paths under `/tmp/edgelet-scripts`.
6. If `scripts.install.entrypoint` was set, set `customInstall = true`.

### Merge matrix

| Layer | YAML key | Empty `entrypoint` | Set `entrypoint` |
|-------|----------|--------------------|------------------|
| Prerequisites | _(none)_ | Embedded `check_prereqs.sh` always appended | Same—always embedded |
| Detect init | _(none)_ | Embedded file in default bundle | Command still runs embedded **name**; file must exist in `dir` if install bundle not embedded |
| Deps | `scripts.deps` | Embed `install_deps.sh` + `configure_container_engine.sh` | Your script only; embedded deps **not** added |
| Install | `scripts.install` | Embed install bundle + all `lib/*.sh` | `customInstall=true`; install bundle + libs **not** embedded |
| WASM | _(none)_ | Included in default install bundle | If install custom: file must be in `dir` |
| Post-install (init, start, configure, wait, bundled) | _(none)_ | Embedded in default bundle | Commands **still run**; files must be in `dir` when install was custom |
| Uninstall | `scripts.uninstall` | Embed `uninstall.sh` | Your script; embedded uninstall **not** added |

### `customInstall` effects

When `scripts.install.entrypoint` is set:

| Behavior | With default install | With custom install |
|----------|---------------------|---------------------|
| CLI applies `package.version` to install args | Yes | **No**—pass version in `scripts.install.args` |
| CLI applies `package.container.image` to install args | Yes | **No**—pass image flags in args |
| Airgap `SetAirgap` / `--bin-path` wiring | Yes | **Yes** (still works) |
| Embedded `install.sh`, `install_container.sh`, `lib/*` staged | Yes | **No** |
| Post-install layers executed | Yes | **Yes** (unless your install script exits the whole deploy early by failing) |

### Partial override example

If only `scripts.deps.entrypoint` is set:

- Custom deps script from `dir` runs.
- Full embedded install bundle and libs are still staged and run.
- Post-install layers use embedded scripts.

If `scripts.install.entrypoint` is set but post-install scripts are **not** in `dir`, bootstrap fails when those commands run—operators must copy required scripts into `dir` or avoid setting a custom install entrypoint.

---

## Staging vs execution

Two distinct concepts:

| Concept | Mechanism | Result |
|---------|-----------|--------|
| **Staging** | `scriptNames` / `scriptContents` | Files written to `/tmp/edgelet-scripts` on the target host |
| **Execution** | `preInstallCommands`, `postInstallCommandsBeforeBundled`, `postInstallBundledCommand` | Ordered shell commands with bootstrap env prefix |

### Pre-install command order

```text
check_prereqs → detect_init → deps → install [→ install_wasm_runtimes if WASM scope]
```

### Post-install command order

```text
install_init_units → start_edgelet → configure_container_edgelet → wait_edgelet_ready → bundled
```

WASM RuntimeClass deploy (Go) runs after the four post-install shell layers and immediately before `bundled.sh` (same order on remote and local).

---

## Remote vs local execution

| Aspect | Remote (`RemoteEdgelet`) | Local (`LocalEdgelet`) |
|--------|--------------------------|------------------------|
| Stage dir | `/tmp/edgelet-scripts` via SSH | Same path on local filesystem |
| Materialize | `copyInstallScripts()` → SCP to `/tmp`, `install -m 755` | `materializeScripts()` → local write |
| Shell | SSH runs full command string | `sh -c "<command>"` |
| Sudo | Always on bootstrap commands | Native linux: sudo; desktop container: none |
| Prereqs | Passwordless sudo required | Skipped (`LOCAL_INSTALL=1`) |
| Config write | SSH `install` if file missing | Go `MaterializeEdgeletRuntime` |
| OS detect | `uname -s` over SSH | CLI `runtime.GOOS` |

---

## Go steps outside scripts

These steps are **not** replaceable via `scripts` YAML. Custom install scripts must cooperate with them (e.g. use `--skip-config` so Go can write config).

| Step | When | Description |
|------|------|-------------|
| Airgap binary transfer | `airgap: true`, native | SCP raw edgelet binary; sets `--bin-path` on install |
| Airgap image transfer | `airgap: true` | `edgelet image load` on host |
| WASM extract/cache | `package.wasm` | CLI-side tar extract; remote SCP of raw shims |
| `MaterializeEdgeletRuntime` | After pre-install, before post-install | Writes host `config.yaml` + sample CA if absent |
| `DeployWasmRuntimeClasses` | After post-install commands | Applies RuntimeClass manifests via edgelet |
| AgentConfig SDK deploy | Agent/CP deploy with `spec.config` | Controller API agent record |
| `edgelet config --a <url>` | Provision phase | Sets controller URL |
| `edgelet config cert <base64>` | Provision, if CA returned | Installs controller CA |
| `edgelet provision <key>` | Provision | Registers agent with controller |
| `edgelet deploy -f` | Control plane hosts | Deploys translated Registry / ControlPlane manifests |
| Private registry | CP with private `controller.package` | `edgelet registry` + SDK registry create |

---

## Uninstall and delete

`Uninstall(removeData)` re-stages scripts (same merge rules), then runs the uninstall entrypoint.

| Trigger | `removeData` / args |
|---------|---------------------|
| Delete Agent / CP host teardown | `uninstall.sh` with `--remove-data` |
| Detach Agent | Uninstall **not** run on host |

Custom uninstall: set `scripts.uninstall.entrypoint` and optional `args` (e.g. `--remove-data`). When entrypoint is set, embedded `uninstall.sh` is not added to the stage bundle—your script from `dir` must implement teardown.

`uninstall.sh` handles **linux**, **darwin**, and **windows** paths (services, binary, share dir, optional data wipe).

---

## Cookbook examples

### 1. Defaults only (linux production)

Omit `scripts`. Full embedded bundle runs. Set `config.arch`, `deploymentType`, `containerEngine`, and `package.version` or `package.container.image`.

### 2. Scripts dir without entrypoint overrides

```yaml
scripts:
  dir: ./vendor/edgelet-scripts
```

All files from `dir` are staged first; embedded defaults for deps/install/uninstall/post-install are **appended**. Use this to vendor and lightly patch embedded scripts while keeping default entrypoints.

### 3. Custom deps only (corporate docker mirror)

```yaml
scripts:
  dir: ./scripts
  deps:
    entrypoint: install_deps.sh
    args: ["docker", "native"]
```

Place your custom `install_deps.sh` in `dir`. Embedded install bundle and post-install scripts still run.

### 4. Custom install (full responsibility)

```yaml
scripts:
  dir: ./scripts
  install:
    entrypoint: my_install.sh
    args:
      - "--version=3.8.0-rc.2"
      - "--airgap"
      - "--bin-path=/opt/airgap/edgelet-linux-amd64"
```

**Required in `dir` at minimum** (because post-install still runs):

- `detect_init.sh` (or copy from embedded assets)
- `install_init_units.sh`
- `start_edgelet.sh`
- `configure_container_edgelet.sh`
- `wait_edgelet_ready.sh`
- `bundled.sh`
- Any `lib/*.sh` your `my_install.sh` sources

Also set install args explicitly—CLI will not inject `package.version`.

### 5. All three layers overridden

Matches `internal/resource/testdata/edgelet/remote-agent-package-registry.yaml`:

```yaml
scripts:
  dir: /tmp/my-scripts
  deps:
    entrypoint: install_deps.sh
  install:
    entrypoint: install.sh
    args: ["1.0.0-rc.8"]
  uninstall:
    entrypoint: uninstall.sh
```

Only `check_prereqs.sh` is appended from embed. **`dir` must contain every other script** the bootstrap sequence invokes.

### 6. System agent on remote control plane

```yaml
controllers:
  - name: cp-host-1
    host: 10.0.0.10
    ssh: { user: ubuntu, keyFile: ~/.ssh/id_rsa }
    systemAgent:
      config:
        arch: amd64
        deploymentType: native
        containerEngine: edgelet
      scripts:
        dir: ./cp-scripts
```

Bootstrap runs during CP host deploy; provision runs when system agent is registered.

### 7. Airgap + custom install args

Enable `airgap: true` on the agent or CP. CLI transfers the binary and calls `SetAirgap` before bootstrap. With custom install, include airgap flags in `scripts.install.args`:

```yaml
scripts:
  install:
    entrypoint: install.sh
    args:
      - "--airgap"
      - "--bin-path=/tmp/edgelet-linux-amd64"
      - "--version=3.8.0-rc.2"
      - "--skip-config"
      - "--skip-start"
```

### 8. Local darwin container dev

```yaml
kind: LocalAgent
spec:
  config:
    deploymentType: container
    containerEngine: docker
  package:
    container:
      image: ghcr.io/example/edgelet:3.8.0-rc.2
```

Expect: no sudo, no host config materialization, no `bundled.sh` publish, desktop wait/configure paths. Docker must already be installed.

### 9. Windows local native

Use `LocalAgent` on a Windows machine running the CLI. Scripts use `%ProgramData%` and `%ProgramFiles%` paths. Staging still uses `/tmp/edgelet-scripts` in the shell environment (Git Bash/MSYS). Verify init/start paths match your shell environment.

---

## Troubleshooting

| Symptom | Likely cause | Action |
|---------|--------------|--------|
| Post-install script not found after custom install | Embedded post-install scripts not staged when `customInstall=true` | Add missing scripts to `scripts.dir` |
| Version/image ignored | Custom install entrypoint set | Pass `--version=` / `--image=` in `scripts.install.args` |
| `install_deps` runs but exits immediately | `containerEngine: edgelet` | Expected no-op; not an error |
| Remote deploy fails at prereqs | Passwordless sudo missing | Fix sudoers for SSH user |
| WASM step skipped | Scope gate (OS, deployment type, engine) | See [WASM runtime scope](#wasm-runtime-scope) |
| WASM on macOS never runs | By design | WASM is linux native only |
| `bundled.sh` does nothing locally | Desktop container deploy | Expected; scripts remain in stage dir |
| Configure container skipped | Not desktop container | Expected on linux native/container server deploy |
| Duplicate script behavior | File in `dir` then overwritten by embed | `check_prereqs.sh` always wins; later duplicates in stage list overwrite earlier writes |
| System agent provision fails but edgelet installed | Bootstrap succeeded earlier | Check controller endpoint, auth, and provision key separately |

---

## Code and asset map

| Concern | Location |
|---------|----------|
| Embedded scripts | `assets/edgelet/scripts/` |
| Script names / init | `pkg/iofog/install/pkg.go` |
| Default procedures + command builders | `pkg/iofog/install/edgelet_scripts.go` |
| Merge logic (remote) | `pkg/iofog/install/edgelet_remote.go` → `CustomizeProcedures` |
| Merge logic (local) | `pkg/iofog/install/edgelet_local.go` → `CustomizeProcedures` |
| YAML types | `pkg/iofog/install/procedures.go`, `internal/resource/remote_agent.go` |
| Agent deploy wiring | `internal/deploy/agent/remote.go`, `local.go`, `edgelet.go` |
| CP host bootstrap | `internal/deploy/controlplane/remote/edgelet_host.go`, `local/edgelet_host.go` |
| WASM scope | `pkg/iofog/install/wasm/scope.go` |
| Config materialization | `pkg/iofog/install/edgelet_config.go` |
| Test fixture (scripts YAML) | `internal/resource/testdata/edgelet/remote-agent-package-registry.yaml` |
| Partial merge test | `pkg/iofog/install/edgelet_procedures_test.go` |

For internal RFC and installation resource model, see `.cursor/cli/docs/installation-resource-model.md` and `.cursor/cli/docs/04g-edgelet-script-hardening.md`.
