#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OPTION_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
VERIFIER="$SCRIPT_DIR/option-operations-input-verify.sh"
FINALIZATION_VERIFIER="$SCRIPT_DIR/option-evidence-finalization-verify.sh"
fixture_dir=$(mktemp -d "${TMPDIR:-/tmp}/option-operations-input.XXXXXX") || exit 1
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM

valid="$fixture_dir/valid.md"
launch_bundle="$fixture_dir/launch-bundle.md"
fixed_evidence="$fixture_dir/fixed-evidence.md"
printf '%s\n' 'OPTION_EVIDENCE_STATUS: APPROVED' '' '| 检查 | 结果 |' '| --- | --- |' \
  '| 固定运营材料 | PASSED |' >"$fixed_evidence"
if command -v shasum >/dev/null 2>&1; then
  fixed_evidence_sha=$(shasum -a 256 "$fixed_evidence" | awk '{print $1}')
else
  fixed_evidence_sha=$(sha256sum "$fixed_evidence" | awk '{print $1}')
fi
printf '%s\n' \
  'OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED' \
  'OPTION_LAUNCH_CONTRACT_COUNT: 1' \
  'OPTION_LAUNCH_CONTRACT: 1|OPT-TEST|/approved/checklist.md|aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|APPROVED' \
  >"$launch_bundle"
sed \
  -e 's/OPTION_OPERATIONS_INPUT_STATUS: DRAFT/OPTION_OPERATIONS_INPUT_STATUS: APPROVED/' \
  -e 's/\[ \]/[x]/g' \
  -e 's/待填写；联系方式存安全系统/approved-ref/g' \
  -e 's/待填写；未批一律false；与渲染YAML逐项一致/approved-ref/g' \
  -e 's/待填写；必须为0/0/g' \
  -e 's/| 待填写 | OPEN |/| approved-ref | APPROVED |/g' \
  -e 's/| OPEN |$/| APPROVED |/g' \
  -e "s#|OPEN||#|APPROVED|$fixed_evidence|$fixed_evidence_sha#g" \
  -e 's/| 待填写 |$/| approved-ref |/g' \
  -e 's/| 必填资料总数 | 31 + 10 × 本次合约总数；归档时必须改为整数 |/| 必填资料总数 | 41 |/' \
  -e 's/| 已批准 | approved-ref |/| 已批准 | 41 |/' \
  -e 's/| 不适用且有依据 | approved-ref |/| 不适用且有依据 | 0 |/' \
  -e 's/| OPEN\/REJECTED | approved-ref |/| OPEN\/REJECTED | 0 |/' \
  "$OPTION_DIR/docs/option-operations-input-checklist.md" >"$valid"

"$VERIFIER" "$valid" "$launch_bundle" >/dev/null || exit 1
"$FINALIZATION_VERIFIER" "$valid" >/dev/null || exit 1
printf 'PASS  valid approved operations input\n'

expect_rejected() {
  name="$1"
  candidate="$2"
  candidate_bundle="${3:-$launch_bundle}"
  if "$VERIFIER" "$candidate" "$candidate_bundle" >/dev/null 2>&1; then
    printf 'FAIL  %s was accepted\n' "$name" >&2
    exit 1
  fi
  printf 'PASS  %s rejected\n' "$name"
}

unchecked="$fixture_dir/unchecked.md"
sed 's/\[x\]/[ ]/g' "$valid" >"$unchecked"
expect_rejected "unchecked completion item" "$unchecked"

open_status="$fixture_dir/open-status.md"
sed 's/| APPROVED |/| OPEN |/g' "$valid" >"$open_status"
expect_rejected "unresolved row status" "$open_status"

placeholder="$fixture_dir/placeholder.md"
sed 's/approved-ref/待填写/g' "$valid" >"$placeholder"
expect_rejected "unresolved placeholder" "$placeholder"

missing_material="$fixture_dir/missing-material.md"
sed '/^OPTION_OPERATIONS_MATERIAL: tenant_legal_jurisdiction|/d' "$valid" >"$missing_material"
expect_rejected "missing fixed-material record" "$missing_material"

duplicate_material="$fixture_dir/duplicate-material.md"
cp "$valid" "$duplicate_material"
printf '%s\n' "OPTION_OPERATIONS_MATERIAL: tenant_legal_jurisdiction|APPROVED|$fixed_evidence|$fixed_evidence_sha" \
  >>"$duplicate_material"
expect_rejected "duplicate fixed-material record" "$duplicate_material"

open_material="$fixture_dir/open-material.md"
sed 's/tenant_legal_jurisdiction|APPROVED|/tenant_legal_jurisdiction|OPEN|/' \
  "$valid" >"$open_material"
expect_rejected "unresolved fixed-material status" "$open_material"

bad_material_hash="$fixture_dir/bad-material-hash.md"
sed "s#tenant_legal_jurisdiction|APPROVED|$fixed_evidence|$fixed_evidence_sha#tenant_legal_jurisdiction|APPROVED|$fixed_evidence|0000000000000000000000000000000000000000000000000000000000000000#" \
  "$valid" >"$bad_material_hash"
expect_rejected "fixed-material hash mismatch" "$bad_material_hash"

draft_evidence="$fixture_dir/draft-evidence.md"
printf '%s\n' 'OPTION_EVIDENCE_STATUS: DRAFT' >"$draft_evidence"
if command -v shasum >/dev/null 2>&1; then
  draft_evidence_sha=$(shasum -a 256 "$draft_evidence" | awk '{print $1}')
else
  draft_evidence_sha=$(sha256sum "$draft_evidence" | awk '{print $1}')
fi
unfinished_material="$fixture_dir/unfinished-material.md"
sed "s#$fixed_evidence|$fixed_evidence_sha#$draft_evidence|$draft_evidence_sha#g" \
  "$valid" >"$unfinished_material"
expect_rejected "unfinalized fixed-material evidence" "$unfinished_material"

not_applicable="$fixture_dir/not-applicable.md"
awk '
  !changed && /^OPTION_OPERATIONS_MATERIAL: tenant_legal_jurisdiction\|APPROVED\|/ {
    sub(/\|APPROVED\|/, "|NOT_APPLICABLE|")
    changed=1
  }
  {print}
' "$valid" | sed \
  -e 's/| 已批准 | 41 |/| 已批准 | 40 |/' \
  -e 's/| 不适用且有依据 | 0 |/| 不适用且有依据 | 1 |/' \
  >"$not_applicable"
"$VERIFIER" "$not_applicable" "$launch_bundle" >/dev/null || exit 1
printf 'PASS  evidenced not-applicable material reconciled\n'

bad_summary="$fixture_dir/bad-summary.md"
sed 's/| 已批准 | 41 |/| 已批准 | 40 |/' "$valid" >"$bad_summary"
expect_rejected "summary count mismatch" "$bad_summary"

two_contract_bundle="$fixture_dir/two-contract-launch-bundle.md"
printf '%s\n' \
  'OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED' \
  'OPTION_LAUNCH_CONTRACT_COUNT: 2' \
  'OPTION_LAUNCH_CONTRACT: 1|OPT-TEST|/approved/checklist-1.md|aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|APPROVED' \
  'OPTION_LAUNCH_CONTRACT: 2|OPT-TEST-2|/approved/checklist-2.md|bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb|APPROVED' \
  >"$two_contract_bundle"
expect_rejected "summary/launch contract-count mismatch" "$valid" "$two_contract_bundle"

draft_launch_bundle="$fixture_dir/draft-launch-bundle.md"
sed 's/OPTION_LAUNCH_CHECKLIST_STATUS: APPROVED/OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT/' \
  "$launch_bundle" >"$draft_launch_bundle"
expect_rejected "draft launch bundle" "$valid" "$draft_launch_bundle"

missing_launch_record="$fixture_dir/missing-launch-record.md"
sed '/^OPTION_LAUNCH_CONTRACT:/d' "$launch_bundle" >"$missing_launch_record"
expect_rejected "launch bundle record-count mismatch" "$valid" "$missing_launch_record"

severe="$fixture_dir/severe.md"
sed 's/| 未关闭SEV-1\/2或资金差异 | 0 |/| 未关闭SEV-1\/2或资金差异 | 1 |/' \
  "$valid" >"$severe"
expect_rejected "unresolved severe issue" "$severe"

printf '\nOption operations-input verifier self-test passed.\n'
