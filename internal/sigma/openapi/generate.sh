#!/usr/bin/env bash
# Generate models and client methods from the vendored Sigma REST OpenAPI snapshot.
# Never hand-edit generated.go. Apply generator corrections only through overlay.yaml.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "${root}"

spec_file="specs/sigma-rest-api.openapi.json"
ops_file="internal/sigma/openapi/operations.yaml"
cfg_file="internal/sigma/openapi/oapi-codegen.yaml"
overlay_file="internal/sigma/openapi/overlay.yaml"
out_file="internal/sigma/openapi/generated.go"

if [ ! -f "${spec_file}" ]; then
  echo "missing vendored OpenAPI snapshot: ${spec_file}" >&2
  exit 1
fi

python3 - "${spec_file}" "${ops_file}" "${cfg_file}" "${overlay_file}" "${out_file}" <<'PY'
import json
import pathlib
import subprocess
import sys

spec_path, ops_path, cfg_path, overlay_path, out_path = map(pathlib.Path, sys.argv[1:])

spec = json.loads(spec_path.read_text())
known = set()
for methods in spec.get("paths", {}).values():
    if not isinstance(methods, dict):
        continue
    for operation in methods.values():
        if isinstance(operation, dict) and operation.get("operationId"):
            known.add(operation["operationId"])

ids = []
seen = set()
for raw_line in ops_path.read_text().splitlines():
    line = raw_line.strip()
    if line.startswith("- "):
        line = line[2:].strip()
    if not line.startswith("operation_id:"):
        continue
    operation_id = line.split(":", 1)[1].strip()
    if not operation_id or operation_id in seen:
        continue
    seen.add(operation_id)
    ids.append(operation_id)

if not ids:
    raise SystemExit(f"no operation_id entries found in {ops_path}")

missing = [operation_id for operation_id in ids if operation_id not in known]
if missing:
    raise SystemExit(
        "operations.yaml references operation IDs missing from the vendored snapshot:\n  "
        + "\n  ".join(missing)
    )

config = cfg_path.read_text()
if "include-operation-ids:" in config:
    raise SystemExit(
        f"{cfg_path} must not list include-operation-ids; operations.yaml is the source of truth"
    )

apply_overlay = any(
    line.strip().startswith("- target:") for line in overlay_path.read_text().splitlines()
)

ids_yaml = "\n".join(f"    - {operation_id}" for operation_id in ids)
merged = config.rstrip() + "\n  include-operation-ids:\n" + ids_yaml + "\n"
if apply_overlay:
    merged += (
        "  overlay:\n"
        f"    path: {overlay_path.name}\n"
        "    strict: true\n"
    )

merged_path = cfg_path.with_name(".oapi-codegen.run.yaml")
merged_path.write_text(merged)
try:
    subprocess.run(
        ["oapi-codegen", "-config", str(merged_path.resolve()), str(spec_path.resolve())],
        check=True,
        cwd=spec_path.resolve().parents[1],
    )
finally:
    merged_path.unlink(missing_ok=True)

if not out_path.is_file():
    raise SystemExit(f"oapi-codegen did not write {out_path}")
PY
