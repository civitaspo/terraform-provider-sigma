#!/usr/bin/env bash
set -euo pipefail
# Fail when handwritten internal/provider or internal/sigma statement coverage is below 80%.
# Generated OpenAPI code lives in internal/sigma/openapi and is excluded.

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"
if [ ! -f go.mod ]; then
  echo "go.mod not present yet; skipping coverage"
  exit 0
fi
profile="$(mktemp)"
trap 'rm -f "$profile"' EXIT

go test -race -covermode=atomic -coverprofile="$profile" ./...

python3 - "$profile" <<'PY'
import sys
from collections import defaultdict

path = sys.argv[1]
statements = defaultdict(int)
covered = defaultdict(int)
with open(path) as handle:
    for line in handle:
        if line.startswith("mode:"):
            continue
        line = line.strip()
        if not line:
            continue
        loc, rest = line.split(" ", 1)
        nstmt, count = rest.split()
        nstmt = int(nstmt)
        count = int(count)
        file = loc.rsplit(":", 1)[0]
        if "/internal/sigma/openapi/" in file or "/internal/provider/testutil/" in file:
            continue
        if "/internal/provider/" in file:
            pkg = "internal/provider"
        elif "/internal/sigma/" in file:
            pkg = "internal/sigma"
        else:
            continue
        statements[pkg] += nstmt
        if count > 0:
            covered[pkg] += nstmt

failed = False
for pkg in ("internal/provider", "internal/sigma"):
    total = statements[pkg]
    hit = covered[pkg]
    if total == 0:
        print(f"{pkg}: no statements in coverage profile", file=sys.stderr)
        failed = True
        continue
    pct = 100.0 * hit / total
    print(f"{pkg}: {pct:.1f}% ({hit}/{total} statements)")
    if pct < 80.0:
        print(f"{pkg} coverage {pct:.1f}% is below 80%", file=sys.stderr)
        failed = True
sys.exit(1 if failed else 0)
PY
