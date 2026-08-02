#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
VERIFY="$SCRIPT_DIR/option-readiness-attestation-verify.sh"
SCHEMA="$SCRIPT_DIR/option-production-readiness.env.example"
TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/option-readiness-attestation.XXXXXX")
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

sed 's#=/absolute/path/to/[^[:space:]]*#=#' "$SCHEMA" > "$TMP_DIR/valid.env"

expect_pass() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.env" "$SCHEMA" >/dev/null 2>&1; then
    printf 'PASS  %s\n' "$case_name"
  else
    printf 'FAIL  %s should pass\n' "$case_name" >&2
    exit 1
  fi
}

expect_fail() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.env" "$SCHEMA" >/dev/null 2>&1; then
    printf 'FAIL  %s should be rejected\n' "$case_name" >&2
    exit 1
  else
    printf 'PASS  %s rejected\n' "$case_name"
  fi
}

expect_pass valid

cp "$TMP_DIR/valid.env" "$TMP_DIR/duplicate.env"
printf '\nOPTION_RELEASE_COMMIT=deadbeef\n' >> "$TMP_DIR/duplicate.env"
expect_fail duplicate

sed '/^OPTION_RELEASE_COMMIT=/d' "$TMP_DIR/valid.env" > "$TMP_DIR/missing.env"
expect_fail missing

cp "$TMP_DIR/valid.env" "$TMP_DIR/unknown.env"
printf '\nOPTION_RELEASE_COMM1T=deadbeef\n' >> "$TMP_DIR/unknown.env"
expect_fail unknown

cp "$TMP_DIR/valid.env" "$TMP_DIR/malformed.env"
printf '\nexport OPTION_RELEASE_COMMIT=deadbeef\n' >> "$TMP_DIR/malformed.env"
expect_fail malformed

cp "$TMP_DIR/valid.env" "$TMP_DIR/placeholder.env"
sed 's#^OPTION_RELEASE_SIGNOFF=.*#OPTION_RELEASE_SIGNOFF=/absolute/path/to/signoff.md#' \
  "$TMP_DIR/placeholder.env" > "$TMP_DIR/placeholder.next"
mv "$TMP_DIR/placeholder.next" "$TMP_DIR/placeholder.env"
expect_fail placeholder

printf '\nOption readiness-attestation verifier self-test passed.\n'

