#!/usr/bin/env bash
# Fail when forbidden flavor-specific strings appear in internal/ production code.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

patterns=(
	'ghcr\.io/datasance'
	'ghcr\.io/eclipse-iofog'
	'datasance\.com/v3'
	'iofog\.org/v3'
	'downloads\.datasance\.com'
	'packagecloud\.io/iofog'
	'https://docs\.datasance\.com'
	'https://github\.com/Datasance/potctl'
)

allowlist=(
	':!internal/**/*_test.go'
	':!internal/**/testdata/**'
	':!internal/**/fixtures/**'
)

failed=0
for pattern in "${patterns[@]}"; do
	matches="$(git grep -En "$pattern" -- internal/ "${allowlist[@]}" 2>/dev/null || true)"
	if [[ -n "$matches" ]]; then
		echo "grep-gate: forbidden pattern /${pattern}/ in internal/:" >&2
		echo "$matches" >&2
		failed=1
	fi
done

if [[ $failed -ne 0 ]]; then
	echo "grep-gates: use pkg/util ldflag getters instead of hardcoded flavor strings" >&2
	exit 1
fi

echo "grep-gates: OK"
