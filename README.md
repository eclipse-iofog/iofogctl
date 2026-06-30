[![CLI CI](https://github.com/eclipse-iofog/iofogctl/actions/workflows/ci.yml/badge.svg)](https://github.com/eclipse-iofog/iofogctl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/eclipse-iofog/iofogctl?include_prereleases)](https://github.com/eclipse-iofog/iofogctl/releases)
[![Go](https://img.shields.io/badge/Go-1.26.4-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-EPL--2.0-blue.svg)](LICENSE)
[![govulncheck](https://github.com/eclipse-iofog/iofogctl/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/eclipse-iofog/iofogctl/actions/workflows/govulncheck.yml)
[![CodeQL](https://github.com/eclipse-iofog/iofogctl/actions/workflows/codeql.yml/badge.svg)](https://github.com/eclipse-iofog/iofogctl/actions/workflows/codeql.yml)

[![Linux amd64](https://img.shields.io/badge/linux--amd64-supported-2ea44f?style=flat&logo=linux&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)
[![Linux arm64](https://img.shields.io/badge/linux--arm64-supported-2ea44f?style=flat&logo=linux&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)
[![Linux armv6](https://img.shields.io/badge/linux--armv6-supported-2ea44f?style=flat&logo=linux&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)
[![Linux armv7](https://img.shields.io/badge/linux--armv7-supported-2ea44f?style=flat&logo=linux&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)

[![macOS amd64](https://img.shields.io/badge/macos--amd64-supported-2ea44f?style=flat&logo=apple&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)
[![macOS arm64](https://img.shields.io/badge/macos--arm64-supported-2ea44f?style=flat&logo=apple&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)
[![Windows amd64](https://img.shields.io/badge/windows--amd64-supported-2ea44f?style=flat&logo=windows&logoColor=white)](https://github.com/eclipse-iofog/iofogctl/releases)

Upstream: [eclipse-iofog/iofogctl](https://github.com/eclipse-iofog/iofogctl) · Development mirror: [Datasance/potctl](https://github.com/Datasance/potctl)

**iofogctl** and **potctl** are dual-flavor CLIs for installing, configuring, and operating ioFog [Edge Compute Networks](https://iofog.org/docs/2/getting-started/core-concepts.html) (ECNs). v3.8 is a greenfield release: configuration lives under `~/.iofog/v3`, and there is no in-place upgrade path from legacy potctl- or v3.7 deployments.

Release binaries ship for **linux** (amd64, arm64, armv6, armv7), **macOS** (amd64, arm64), and **Windows** (amd64). Built with **Go 1.26.4** (see `go.mod`).

## Install — iofogctl

Package repository: [iofog.datasance.com](https://iofog.datasance.com/)

**Linux (DEB or RPM):**

```bash
wget -q -O - https://iofog.datasance.com/iofogctl_installer.sh | sudo bash
sudo apt install -y iofogctl   # Debian/Ubuntu
# or
sudo yum install -y iofogctl   # RHEL/CentOS/Fedora
```

**macOS (Homebrew):**

```bash
brew tap eclipse-iofog/iofogctl
brew install iofogctl
```

## Install — potctl

Package repository: [downloads.datasance.com](https://downloads.datasance.com/)

**Linux (DEB or RPM):**

```bash
wget -q -O - https://downloads.datasance.com/potctl_installer.sh | sudo bash
sudo apt install -y potctl   # Debian/Ubuntu
# or
sudo yum install -y potctl   # RHEL/CentOS/Fedora
```

**macOS (Homebrew):**

```bash
brew tap Datasance/potctl
brew install potctl
```

## Edge node agent — edgelet

v3.8 edge nodes run **edgelet**, not the legacy Java `iofog-agent`. Deploy edge nodes with the CLI (`deploy -f` manifest) or install the edgelet binary directly:

```bash
curl -fsSL https://github.com/eclipse-iofog/edgelet/releases/download/v1.0.0-rc.8/install.sh -o install.sh
chmod +x install.sh
sudo ./install.sh --version=v1.0.0-rc.8
```

Eclipse canonical: [eclipse-iofog/edgelet](https://github.com/eclipse-iofog/edgelet/releases) · Datasance mirror: [Datasance/edgelet](https://github.com/Datasance/edgelet/releases)

## Usage

```bash
iofogctl version
iofogctl connect --help
iofogctl deploy -f ecn.yaml
# potctl is the Datasance-flavor equivalent (same commands, different binary name)
```

Documentation:

- **iofogctl:** [iofog.org](https://iofog.org/docs)
- **potctl:** [docs.datasance.com](https://docs.datasance.com)

Shell autocompletion: `iofogctl autocomplete bash` (or `zsh`), then follow the printed instructions.

## Build from source

Requires Go **1.26.4+** (matches the Go badge and `go.mod`). Build outside `$GOPATH` (Go modules):

```bash
make FLAVOR=iofog build       # iofogctl → bin/iofogctl
make FLAVOR=datasance build   # potctl → bin/potctl
# or
make iofogctl
make potctl
```

Install to `$GOPATH/bin`:

```bash
make FLAVOR=iofog install
```

## Running tests

```bash
make test
```

Unit tests (short mode):

```bash
make test-unit
```

## License

[EPL-2.0](LICENSE)
