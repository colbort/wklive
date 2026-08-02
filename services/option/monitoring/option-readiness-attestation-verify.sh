#!/bin/sh
set -u

if [ "$#" -ne 1 ] && [ "$#" -ne 2 ]; then
  printf 'usage: %s READINESS_ENV [READINESS_ENV_SCHEMA]\n' "$0" >&2
  exit 2
fi

READINESS_FILE=$1
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
SCHEMA_FILE=${2:-$SCRIPT_DIR/option-production-readiness.env.example}

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  exit 1
}

[ -s "$READINESS_FILE" ] || fail "readiness attestation is missing or empty"
[ -s "$SCHEMA_FILE" ] || fail "readiness attestation schema is missing or empty"

expected_keys=$(awk -F '=' '
  /^[[:space:]]*OPTION_[A-Z0-9_]+[[:space:]]*=/ {
    key=$1
    gsub(/[[:space:]]/, "", key)
    if (seen[key]++) duplicate=1
    print key
  }
  END { if (duplicate) exit 2 }
' "$SCHEMA_FILE") || fail "readiness attestation schema contains duplicate keys"
[ -n "$expected_keys" ] || fail "readiness attestation schema declares no OPTION keys"

actual_keys=$(awk -F '=' '
  /^[[:space:]]*($|#)/ {next}
  /^[[:space:]]*OPTION_[A-Z0-9_]+[[:space:]]*=/ {
    key=$1
    gsub(/[[:space:]]/, "", key)
    print key
    next
  }
  {bad=1}
  END {if (bad) exit 2}
' "$READINESS_FILE") || fail "attestation contains a malformed or non-OPTION assignment"

duplicate_keys=$(printf '%s\n' "$actual_keys" | awk 'seen[$0]++ {print $0}')
[ -z "$duplicate_keys" ] || fail "attestation contains duplicate keys: $(printf '%s' "$duplicate_keys" | tr '\n' ' ')"

unknown_keys=$(printf '%s\n' "$actual_keys" | while IFS= read -r actual_key; do
  [ -n "$actual_key" ] || continue
  if ! printf '%s\n' "$expected_keys" | grep -Fqx "$actual_key"; then
    printf '%s\n' "$actual_key"
  fi
done)
[ -z "$unknown_keys" ] || fail "attestation contains unknown keys: $(printf '%s' "$unknown_keys" | tr '\n' ' ')"

missing_keys=$(printf '%s\n' "$expected_keys" | while IFS= read -r expected_key; do
  [ -n "$expected_key" ] || continue
  if ! printf '%s\n' "$actual_keys" | grep -Fqx "$expected_key"; then
    printf '%s\n' "$expected_key"
  fi
done)
[ -z "$missing_keys" ] || fail "attestation is missing required keys: $(printf '%s' "$missing_keys" | tr '\n' ' ')"

placeholder_lines=$(awk '
  /^[[:space:]]*($|#)/ {next}
  {
    value=$0
    sub(/^[^=]*=[[:space:]]*/, "", value)
    sub(/[[:space:]]+$/, "", value)
    if ((value ~ /^".*"$/) || (value ~ /^\047.*\047$/)) {
      value=substr(value, 2, length(value)-2)
    }
    if (value ~ /\/absolute\/path\/to\// ||
        value ~ /REPLACE_BEFORE_DEPLOY|待填写|TBD|TODO/ ||
        value ~ /^<[^>]+>$/ || value ~ /^\[[^]]+\]$/) {
      print NR ":" $0
    }
  }
' "$READINESS_FILE")
[ -z "$placeholder_lines" ] || {
  printf 'FAIL  attestation contains unresolved placeholder values\n' >&2
  printf '%s\n' "$placeholder_lines" | sed 's/^/      /' >&2
  exit 1
}

printf 'PASS  readiness attestation has the exact unique key set and no placeholder values\n'

