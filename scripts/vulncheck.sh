#!/usr/bin/env bash
# Run govulncheck on CLI module code paths and allow documented exceptions
# listed in SECURITY.md (sync ALLOWED_VULNS with that file).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Documented upstream exceptions — see SECURITY.md § Known vulnerability exceptions.
ALLOWED_VULNS="GO-2026-5932"

GOTAGS="${GOTAGS:-containers_image_openpgp,exclude_graphdriver_btrfs}"

out="$(mktemp)"
trap 'rm -f "$out"' EXIT

set +e
govulncheck -tags="${GOTAGS}" -format=text ./cmd/... ./internal/... ./pkg/... >"$out" 2>&1
status=$?
set -e

cat "$out"

if [[ $status -eq 0 ]]; then
	echo "govulncheck: no vulnerabilities affecting call paths"
	exit 0
fi

if [[ $status -ne 3 ]]; then
	echo "govulncheck: unexpected exit status $status" >&2
	exit "$status"
fi

if [[ -z "$ALLOWED_VULNS" ]]; then
	echo "govulncheck: vulnerabilities found (no exceptions configured)" >&2
	exit 3
fi

found="$(grep -oE 'GO-[0-9]{4}-[0-9]+' "$out" | sort -u || true)"
if [[ -z "$found" ]]; then
	echo "govulncheck: failed but no GO-* IDs parsed; see output above" >&2
	exit 3
fi

unexpected=""
while IFS= read -r id; do
	[[ -z "$id" ]] && continue
	allowed=false
	for a in $ALLOWED_VULNS; do
		if [[ "$id" == "$a" ]]; then
			allowed=true
			break
		fi
	done
	if [[ "$allowed" == false ]]; then
		unexpected="${unexpected} ${id}"
	fi
done <<<"$found"

if [[ -n "${unexpected// /}" ]]; then
	echo "govulncheck: unexpected vulnerabilities (not in SECURITY.md exceptions):${unexpected}" >&2
	exit 3
fi

echo "govulncheck: only documented exceptions remain; see SECURITY.md"
exit 0
