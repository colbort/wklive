#!/bin/sh
set -u

if [ "$#" -ne 2 ]; then
  printf 'usage: %s APPROVED_OPERATIONS_INPUT_CHECKLIST APPROVED_LAUNCH_BUNDLE\n' "$0" >&2
  exit 2
fi

input_file="$1"
launch_bundle="$2"
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  exit 1
}

sha256_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    return 1
  fi
}

[ -s "$input_file" ] || fail "operations input checklist is missing or empty"
[ -s "$launch_bundle" ] || fail "approved launch bundle is missing or empty"

contract_count=$(awk '
  /^OPTION_LAUNCH_CONTRACT_COUNT:[[:space:]]*/ {
    count++
    value=$0
    sub(/^OPTION_LAUNCH_CONTRACT_COUNT:[[:space:]]*/, "", value)
    sub(/[[:space:]]+$/, "", value)
  }
  END {
    if (count != 1 || value !~ /^[1-9][0-9]*$/) exit 2
    print value
  }
' "$launch_bundle") || fail "launch bundle must declare one positive contract count"
launch_status_count=$(grep -Ec '^OPTION_LAUNCH_CHECKLIST_STATUS:[[:space:]]*APPROVED[[:space:]]*$' \
  "$launch_bundle" || true)
[ "$launch_status_count" -eq 1 ] || fail "launch bundle must be APPROVED exactly once"
launch_record_count=$(grep -Ec '^OPTION_LAUNCH_CONTRACT:[[:space:]]*' "$launch_bundle" || true)
[ "$launch_record_count" -eq "$contract_count" ] ||
  fail "launch bundle record count must match its declared contract count"
if ! sed -n 's/^OPTION_LAUNCH_CONTRACT:[[:space:]]*//p' "$launch_bundle" | awk -F '|' '
  {
    if (NF != 5 || $1 == "" || $2 == "" || $5 != "APPROVED" || seen_id[$1]++ || seen_code[$2]++) bad=1
  }
  END {exit bad ? 1 : 0}
'; then
  fail "launch bundle records must have unique identities and APPROVED status"
fi

status_count=$(grep -Ec '^OPTION_OPERATIONS_INPUT_STATUS:[[:space:]]*APPROVED[[:space:]]*$' "$input_file" || true)
[ "$status_count" -eq 1 ] || fail "operations input status must be APPROVED exactly once"

if grep -Eq '[|][[:space:]]*待填写([[:space:];|]|$)' "$input_file"; then
  fail "operations input checklist contains an unresolved table placeholder"
fi
if grep -Eq '^[[:space:]]*-[[:space:]]*\[[[:space:]]\]' "$input_file"; then
  fail "operations input checklist contains an unchecked completion item"
fi
if grep -Eq '[|][[:space:]]*(OPEN|REJECTED|DRAFT)[[:space:]]*[|][[:space:]]*$' "$input_file"; then
  fail "operations input checklist contains an unresolved row status"
fi

checked_count=$(grep -Ec '^[[:space:]]*-[[:space:]]*\[[xX]\]' "$input_file" || true)
[ "$checked_count" -eq 8 ] || fail "operations input checklist must retain and pass all eight completion items"

require_label() {
  label="$1"
  count=$(awk -F '|' -v label="$label" '
    function trim(value) {
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      return value
    }
    NF >= 3 && trim($2) == label { count++ }
    END { print count+0 }
  ' "$input_file")
  [ "$count" -eq 1 ] || fail "required operations row is missing or duplicated: $label"
}

required_labels='tenant_id / 法律实体 / 司法辖区
目标市场 / 官方规则版本
release commit / 全部image digest
变更单 / 计划窗口 / 回滚窗口
产品、技术、风控、清算财务、运营、合规负责人
本次合约总数 / 系列ID版本
合约上市合集及SHA-256
目标环境与公告/客户端合约集导出、SHA及三方对账
上市合集逐合约校验输出
Option渲染YAML / SHA-256 / `ProductScope`逐项布尔值
身份
经济参数
五时间
费用
卖方担保
交易控制
结算
行权
行情
条件功能
年度日历
日历版本
临时休市
公司行动
系列政策
保险账本
日终守恒
平台兜底
保险库存退出
组合保证金
MMP
风险参数变更
实物违约
公共行情
复杂订单
美式提前行权
Beanstalk
编排故障
监控通知
用户材料
值班资料
必填资料总数
已批准
不适用且有依据
OPEN/REJECTED
未关闭SEV-1/2或资金差异'

old_ifs=$IFS
IFS='
'
for required_label in $required_labels; do
  require_label "$required_label"
done
IFS=$old_ifs

expected_material_ids='tenant_legal_jurisdiction
target_market_rules
release_identity
change_window
accountable_roles
contract_series_scope
launch_bundle
contract_set_reconciliation
launch_bundle_verification
product_scope_runtime
annual_calendar
calendar_version
temporary_halt
corporate_action
series_policy
insurance_ledger
daily_conservation
platform_backstop
insurance_inventory_exit
portfolio_margin
mmp
risk_parameter_change
physical_default
public_market
complex_orders
american_exercise
beanstalk_capacity
orchestrator_takeover
alert_delivery
user_materials
oncall_roster'

material_records=$(sed -n 's/^OPTION_OPERATIONS_MATERIAL:[[:space:]]*//p' "$input_file")
material_record_count=$(printf '%s\n' "$material_records" | awk 'NF {count++} END {print count+0}')
[ "$material_record_count" -eq 31 ] || fail "operations input checklist must declare exactly 31 fixed-material records"

if ! printf '%s\n' "$material_records" | awk -F '|' '
  NF {
    if (NF != 4 || $1 == "" || seen[$1]++) bad=1
  }
  END {exit bad ? 1 : 0}
'; then
  fail "fixed-material records must have four fields and unique non-empty identities"
fi

fixed_approved=0
fixed_not_applicable=0
old_ifs=$IFS
IFS='
'
for material_id in $expected_material_ids; do
  matching_record=$(printf '%s\n' "$material_records" | awk -F '|' -v material_id="$material_id" '
    $1 == material_id {count++; record=$0}
    END {if (count != 1) exit 2; print record}
  ') || fail "fixed-material record is missing or duplicated: $material_id"
  material_status=$(printf '%s\n' "$matching_record" | awk -F '|' '{print $2}')
  material_path=$(printf '%s\n' "$matching_record" | awk -F '|' '{print $3}')
  material_sha=$(printf '%s\n' "$matching_record" | awk -F '|' '{print $4}')

  case "$material_status" in
    APPROVED) fixed_approved=$((fixed_approved + 1)) ;;
    NOT_APPLICABLE) fixed_not_applicable=$((fixed_not_applicable + 1)) ;;
    *) fail "fixed-material status must be APPROVED or NOT_APPLICABLE: $material_id" ;;
  esac
  case "$material_path" in
    /*) ;;
    *) fail "fixed-material evidence path must be absolute: $material_id" ;;
  esac
  [ -s "$material_path" ] || fail "fixed-material evidence is missing or empty: $material_id"
  case "$material_sha" in
    ''|*[!A-Fa-f0-9]*) fail "fixed-material SHA-256 is invalid: $material_id" ;;
  esac
  [ "${#material_sha}" -eq 64 ] || fail "fixed-material SHA-256 is invalid: $material_id"
  actual_material_sha=$(sha256_file "$material_path") || fail "SHA-256 utility is unavailable"
  normalized_material_sha=$(printf '%s\n' "$material_sha" | tr '[:upper:]' '[:lower:]')
  [ "$actual_material_sha" = "$normalized_material_sha" ] || fail "fixed-material hash mismatch: $material_id"
  "$script_dir/option-evidence-finalization-verify.sh" "$material_path" >/dev/null 2>&1 ||
    fail "fixed-material evidence is not finalized: $material_id"
done
IFS=$old_ifs

summary_value() {
  label="$1"
  awk -F '|' -v label="$label" '
    function trim(value) {
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      return value
    }
    NF >= 3 && trim($2) == label { count++; value=trim($3) }
    END {
      if (count != 1 || value !~ /^[0-9]+$/) exit 2
      print value
    }
  ' "$input_file"
}

total=$(summary_value '必填资料总数') || fail "required-material total must be one integer"
approved=$(summary_value '已批准') || fail "approved count must be one integer"
not_applicable=$(summary_value '不适用且有依据') || fail "not-applicable count must be one integer"
open_count=$(summary_value 'OPEN/REJECTED') || fail "OPEN/REJECTED count must be one integer"
severe_count=$(summary_value '未关闭SEV-1/2或资金差异') || fail "unresolved severe issue count must be one integer"

[ "$total" -gt 0 ] || fail "required-material total must be positive"
expected_total=$((31 + 10 * contract_count))
[ "$total" -eq "$expected_total" ] ||
  fail "required-material total must equal 31 plus ten parameter groups per approved contract"
expected_approved=$((fixed_approved + 10 * contract_count))
[ "$approved" -eq "$expected_approved" ] ||
  fail "approved count must equal approved fixed-material records plus ten groups per launch contract"
[ "$not_applicable" -eq "$fixed_not_applicable" ] ||
  fail "not-applicable count must equal fixed-material NOT_APPLICABLE records"
[ $((approved + not_applicable)) -eq "$total" ] ||
  fail "derived approved plus evidenced not-applicable counts must equal the required-material total"
[ "$open_count" -eq 0 ] || fail "OPEN/REJECTED count must be zero"
[ "$severe_count" -eq 0 ] || fail "unresolved SEV-1/2 or fund-difference count must be zero"

printf 'PASS  approved operations input checklist is complete and internally reconciled\n'
