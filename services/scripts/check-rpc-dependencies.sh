#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
services_dir=$(dirname -- "$script_dir")
allow_file="$services_dir/rpc-dependencies.allow"
actual_file=$(mktemp)
allowed_file=$(mktemp)
trap 'rm -f "$actual_file" "$allowed_file"' EXIT

for context in "$services_dir"/*/internal/svc/servicecontext.go; do
  [ -f "$context" ] || continue
  service=${context%/internal/svc/servicecontext.go}
  service=${service##*/}
  rg -o 'wklive/proto/[a-z0-9_]+' "$context" |
    sed "s#wklive/proto/#$service #" || true
done | sort -u >"$actual_file"

sed '/^[[:space:]]*#/d; /^[[:space:]]*$/d' "$allow_file" | sort -u >"$allowed_file"
diff -u "$allowed_file" "$actual_file"
tsort "$actual_file" >/dev/null
echo "RPC dependency graph is allowlisted and acyclic."
