# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v3.8.0-rc.1` and later pre-releases on `develop` | Yes |
| v3.7 / legacy Java agent / `iofog-agent` paths | No |

This repository builds two CLI flavors from one tree: **iofogctl** (Eclipse upstream) and **potctl** (Datasance). Security support applies to both at the same tagged versions.

## Reporting a vulnerability

If you believe you have found a security issue in iofogctl or potctl:

1. **Do not** open a public GitHub issue for exploitable vulnerabilities.
2. Email **security@datasance.com** with:
   - A description of the issue and impact
   - Steps to reproduce (proof-of-concept if available)
   - Affected version / commit, CLI flavor (`iofogctl` or `potctl`), and platform
3. We aim to acknowledge reports within **5 business days** and provide a remediation timeline when confirmed.

For non-security bugs, use the public issue tracker or `CONTRIBUTING`.

## Security gates (maintainers)

Before release tags, run:

```bash
make security-code   # gosec on ./cmd ./internal ./pkg
make vulncheck       # govulncheck@v1.1.4 + go mod verify
```

- **gosec** is intentionally **not** in golangci-lint; static analysis is scoped to CLI module trees.
- **govulncheck** scans `./cmd/... ./internal/... ./pkg/...`. Goal: **zero** vulnerabilities affecting call paths.
- CI: [`.github/workflows/ci.yml`](.github/workflows/ci.yml) (security job), [`.github/workflows/govulncheck.yml`](.github/workflows/govulncheck.yml) (on `go.sum` push, daily cron, manual dispatch), [`.github/workflows/codeql.yml`](.github/workflows/codeql.yml).

## Build tags

Build, test, goreleaser, and govulncheck share the same Go build tags via `GOTAGS` in the Makefile (default: `containers_image_openpgp,exclude_graphdriver_btrfs`). This keeps vulnerability scanning aligned with how binaries are compiled without requiring optional cgo dependencies (`libgpgme-dev`, `libbtrfs-dev`) on CI.

## Known vulnerability exceptions

None at this time.

`make vulncheck` must pass with zero undocumented findings affecting CLI call paths.

## Exception policy

New exceptions require:

1. Entry in the table above (GO ID, CVE if any, component, rationale, fix timeline).
2. Matching ID in `scripts/vulncheck.sh` `ALLOWED_VULNS`.
3. Brief note under **Known limitations** in `CHANGELOG.md` at next release.

Undocumented findings **fail** `make vulncheck` and CI.
