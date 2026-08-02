#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
VERIFY="$SCRIPT_DIR/option-evidence-finalization-verify.sh"
TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/option-evidence-finalization.XXXXXX")
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

write_case() {
  case_name=$1
  case_body=$2
  printf '%s\n' "$case_body" > "$TMP_DIR/$case_name.md"
}

expect_pass() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.md" >/dev/null 2>&1; then
    printf 'PASS  %s\n' "$case_name"
  else
    printf 'FAIL  %s should pass\n' "$case_name" >&2
    exit 1
  fi
}

expect_fail() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.md" >/dev/null 2>&1; then
    printf 'FAIL  %s should be rejected\n' "$case_name" >&2
    exit 1
  else
    printf 'PASS  %s rejected\n' "$case_name"
  fi
}

write_case valid 'OPTION_EVIDENCE_STATUS: APPROVED

本说明保留规则文字：“待填写”只能出现在说明正文，不能出现在最终表格。

| 项目 | 最终值 | 状态 |
| --- | --- | --- |
| 资金差异 | 0 | PASSED |

- [x] 原始附件、哈希和复核人均已归档。'
expect_pass valid

write_case draft_status 'OPTION_EVIDENCE_STATUS: DRAFT'
expect_fail draft_status

write_case table_placeholder 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 值 |
| --- | --- |
| release | 待填写 |'
expect_fail table_placeholder

write_case bracket_placeholder 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 值 |
| --- | --- |
| release | [RELEASE_SHA] |'
expect_fail bracket_placeholder

write_case prose_placeholder 'OPTION_EVIDENCE_STATUS: APPROVED
原始附件及哈希：[EVIDENCE]'
expect_fail prose_placeholder

write_case unresolved_decision 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 结论 |
| --- | --- |
| 守恒 | 通过/拒绝 |'
expect_fail unresolved_decision

write_case empty_cell 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 结论 |
| --- | --- |
| 守恒 |  |'
expect_fail empty_cell

write_case open_row 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 状态 |
| --- | --- |
| 差异案件 | OPEN |'
expect_fail open_row

write_case unchecked 'OPTION_EVIDENCE_STATUS: APPROVED
- [ ] 告警送达已验证。'
expect_fail unchecked

write_case unchecked_table 'OPTION_EVIDENCE_STATUS: APPROVED
| 项目 | 完成 |
| --- | --- |
| 通知送达 | [ ] |'
expect_fail unchecked_table

write_case deployment_placeholder 'OPTION_EVIDENCE_STATUS: APPROVED
receiver: REPLACE_BEFORE_DEPLOY'
expect_fail deployment_placeholder

printf '\nOption evidence-finalization verifier self-test passed.\n'
