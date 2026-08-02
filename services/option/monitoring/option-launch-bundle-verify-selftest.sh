#!/bin/sh
set -u

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
option_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
verifier="$script_dir/option-launch-bundle-verify.sh"
checklist_template="$option_dir/docs/option-contract-launch-checklist.md"
bundle_template="$option_dir/docs/templates/option-contract-launch-bundle.md"
reconciliation_template="$option_dir/docs/templates/option-contract-set-reconciliation.md"
fixture_dir=$(mktemp -d /tmp/option-launch-bundle-selftest.XXXXXX) || exit 1
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM

tenant=900101
contract_id=990101
contract_code=BTC-USDT-TEST-C
release=7d81ffcb43b2dc4d1adab70dca3548e4ad8ad191
window_start=1785711600000
list_time=1785715200000
window_end=1785718800000
failures=0

sha256_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    return 1
  fi
}

build_checklist() {
  output=$1
  checkbox=$2
  altered=$3
  sed \
    -e 's/^OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT$/OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED/' \
    -e "s/^OPTION_CONTRACT_TENANT_ID: 0$/OPTION_CONTRACT_TENANT_ID: $tenant/" \
    -e "s/^OPTION_CONTRACT_ID: 0$/OPTION_CONTRACT_ID: $contract_id/" \
    -e "s/^OPTION_CONTRACT_CODE: DRAFT$/OPTION_CONTRACT_CODE: $contract_code/" \
    -e 's/^OPTION_PRODUCT_APPROVAL_REF: DRAFT$/OPTION_PRODUCT_APPROVAL_REF: PRODUCT-APPROVED/' \
    -e 's/^OPTION_TECH_APPROVAL_REF: DRAFT$/OPTION_TECH_APPROVAL_REF: TECH-APPROVED/' \
    -e 's/^OPTION_RISK_APPROVAL_REF: DRAFT$/OPTION_RISK_APPROVAL_REF: RISK-APPROVED/' \
    -e 's/^OPTION_CLEARING_APPROVAL_REF: DRAFT$/OPTION_CLEARING_APPROVAL_REF: CLEARING-APPROVED/' \
    -e 's/^OPTION_OPERATIONS_APPROVAL_REF: DRAFT$/OPTION_OPERATIONS_APPROVAL_REF: OPERATIONS-APPROVED/' \
    -e 's/^OPTION_COMPLIANCE_APPROVAL_REF: DRAFT$/OPTION_COMPLIANCE_APPROVAL_REF: COMPLIANCE-APPROVED/' \
    -e "s/^- \[ \]/- [$checkbox]/" \
    -e "s/合约代码：________　租户：________　计划开放时间：________　负责人：________/合约代码：${contract_code}　租户：${tenant}　计划开放时间：2026-08-03T00:00:00Z　负责人：operator-1/" \
    -e 's/|  | 通过\/拒绝 |  |  |/| approver | 通过 | 2026-08-02T00:00:00Z | APPROVED |/g' \
    "$checklist_template" > "$output"
  if [ "$altered" = true ]; then
    sed '
      s/合约说明书包含标的/合约说明书包含标的（未授权改写）/
    ' "$output" > "$output.tmp"
    mv "$output.tmp" "$output"
  fi
}

build_bundle() {
  output=$1
  checklist=$2
  declared_count=$3
  checklist_sha=$(sha256_file "$checklist") || exit 1
  sed \
    -e 's/^OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT$/OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED/' \
    -e "s/^OPTION_LAUNCH_TENANT_ID: 0$/OPTION_LAUNCH_TENANT_ID: $tenant/" \
    -e "s/^OPTION_LAUNCH_RELEASE_COMMIT: DRAFT$/OPTION_LAUNCH_RELEASE_COMMIT: $release/" \
    -e "s/^OPTION_LAUNCH_CONTRACT_COUNT: 0$/OPTION_LAUNCH_CONTRACT_COUNT: $declared_count/" \
    -e 's/^OPTION_LAUNCH_PRODUCT_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_PRODUCT_APPROVAL_REF: PRODUCT-APPROVED/' \
    -e 's/^OPTION_LAUNCH_TECH_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_TECH_APPROVAL_REF: TECH-APPROVED/' \
    -e 's/^OPTION_LAUNCH_RISK_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_RISK_APPROVAL_REF: RISK-APPROVED/' \
    -e 's/^OPTION_LAUNCH_CLEARING_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_CLEARING_APPROVAL_REF: CLEARING-APPROVED/' \
    -e 's/^OPTION_LAUNCH_OPERATIONS_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_OPERATIONS_APPROVAL_REF: OPERATIONS-APPROVED/' \
    -e 's/^OPTION_LAUNCH_COMPLIANCE_APPROVAL_REF: DRAFT$/OPTION_LAUNCH_COMPLIANCE_APPROVAL_REF: COMPLIANCE-APPROVED/' \
    -e 's/待填写/APPROVED-VALUE/g' \
    -e 's/通过\/拒绝/通过/g' \
    "$bundle_template" > "$output"
  printf '\nOPTION_LAUNCH_CONTRACT: %s|%s|%s|%s|APPROVED\n' \
    "$contract_id" "$contract_code" "$checklist" "$checklist_sha" >> "$output"
}

build_reconciliation() {
  output=$1
  target_source=$2
  publication_source=$3
  target_sha=$(sha256_file "$target_source") || exit 1
  publication_sha=$(sha256_file "$publication_source") || exit 1
  sed \
    -e 's/^OPTION_CONTRACT_SET_RECONCILIATION_STATUS: DRAFT$/OPTION_CONTRACT_SET_RECONCILIATION_STATUS: APPROVED/' \
    -e "s/^OPTION_CONTRACT_SET_TENANT_ID: 0$/OPTION_CONTRACT_SET_TENANT_ID: $tenant/" \
    -e "s/^OPTION_CONTRACT_SET_RELEASE_COMMIT: DRAFT$/OPTION_CONTRACT_SET_RELEASE_COMMIT: $release/" \
    -e 's/^OPTION_CONTRACT_SET_COUNT: 0$/OPTION_CONTRACT_SET_COUNT: 1/' \
    -e "s/^OPTION_CONTRACT_SET_WINDOW_START_MS: 0$/OPTION_CONTRACT_SET_WINDOW_START_MS: $window_start/" \
    -e "s/^OPTION_CONTRACT_SET_WINDOW_END_MS: 0$/OPTION_CONTRACT_SET_WINDOW_END_MS: $window_end/" \
    -e "s#^OPTION_TARGET_CONTRACT_SET_SOURCE: DRAFT#OPTION_TARGET_CONTRACT_SET_SOURCE: $target_source#" \
    -e "s/^OPTION_TARGET_CONTRACT_SET_SOURCE_SHA256: DRAFT$/OPTION_TARGET_CONTRACT_SET_SOURCE_SHA256: $target_sha/" \
    -e "s#^OPTION_PUBLICATION_CONTRACT_SET_SOURCE: DRAFT#OPTION_PUBLICATION_CONTRACT_SET_SOURCE: $publication_source#" \
    -e "s/^OPTION_PUBLICATION_CONTRACT_SET_SOURCE_SHA256: DRAFT$/OPTION_PUBLICATION_CONTRACT_SET_SOURCE_SHA256: $publication_sha/" \
    -e 's/^OPTION_CONTRACT_SET_REVIEW_REF: DRAFT$/OPTION_CONTRACT_SET_REVIEW_REF: REVIEW-APPROVED/' \
    -e 's/^- \[ \]/- [x]/' \
    -e 's/待填写/APPROVED-VALUE/g' \
    -e 's/通过\/拒绝/通过/g' \
    "$reconciliation_template" > "$output"
}

expect_pass() {
  name=$1
  bundle=$2
  reconciliation=$3
  if "$verifier" "$bundle" "$tenant" "$release" "$reconciliation" \
    > "$fixture_dir/$name.log" 2>&1; then
    printf 'PASS  %s\n' "$name"
  else
    printf 'FAIL  %s\n' "$name" >&2
    sed 's/^/      /' "$fixture_dir/$name.log" >&2
    failures=$((failures + 1))
  fi
}

expect_reject() {
  name=$1
  bundle=$2
  reconciliation=$3
  if "$verifier" "$bundle" "$tenant" "$release" "$reconciliation" \
    > "$fixture_dir/$name.log" 2>&1; then
    printf 'FAIL  %s (unexpected pass)\n' "$name" >&2
    failures=$((failures + 1))
  else
    printf 'PASS  %s rejected\n' "$name"
  fi
}

target="$fixture_dir/target.txt"
publication="$fixture_dir/publication.txt"
printf '%s|%s|1|%s\n' "$contract_id" "$contract_code" "$list_time" > "$target"
printf '%s|%s\n' "$contract_id" "$contract_code" > "$publication"

valid_checklist="$fixture_dir/checklist-valid.md"
valid_bundle="$fixture_dir/bundle-valid.md"
valid_reconciliation="$fixture_dir/reconciliation-valid.md"
build_checklist "$valid_checklist" x false
build_bundle "$valid_bundle" "$valid_checklist" 1
build_reconciliation "$valid_reconciliation" "$target" "$publication"
expect_pass valid "$valid_bundle" "$valid_reconciliation"

bad_count_bundle="$fixture_dir/bundle-bad-count.md"
build_bundle "$bad_count_bundle" "$valid_checklist" 2
expect_reject count-mismatch "$bad_count_bundle" "$valid_reconciliation"

unchecked_checklist="$fixture_dir/checklist-unchecked.md"
unchecked_bundle="$fixture_dir/bundle-unchecked.md"
build_checklist "$unchecked_checklist" ' ' false
build_bundle "$unchecked_bundle" "$unchecked_checklist" 1
expect_reject unchecked-item "$unchecked_bundle" "$valid_reconciliation"

altered_checklist="$fixture_dir/checklist-altered.md"
altered_bundle="$fixture_dir/bundle-altered.md"
build_checklist "$altered_checklist" x true
build_bundle "$altered_bundle" "$altered_checklist" 1
expect_reject altered-requirement "$altered_bundle" "$valid_reconciliation"

mismatched_publication="$fixture_dir/publication-mismatch.txt"
mismatched_reconciliation="$fixture_dir/reconciliation-mismatch.md"
printf '990102|BTC-USDT-TEST-P\n' > "$mismatched_publication"
build_reconciliation "$mismatched_reconciliation" "$target" "$mismatched_publication"
expect_reject publication-mismatch "$valid_bundle" "$mismatched_reconciliation"

trading_target="$fixture_dir/target-trading.txt"
trading_reconciliation="$fixture_dir/reconciliation-trading.md"
printf '%s|%s|2|%s\n' "$contract_id" "$contract_code" "$list_time" > "$trading_target"
build_reconciliation "$trading_reconciliation" "$trading_target" "$publication"
expect_reject premature-trading "$valid_bundle" "$trading_reconciliation"

if [ "$failures" -ne 0 ]; then
  printf '\nOption launch-bundle self-test failed with %s error(s).\n' "$failures" >&2
  exit 1
fi

printf '\nOption launch-bundle self-test passed.\n'
