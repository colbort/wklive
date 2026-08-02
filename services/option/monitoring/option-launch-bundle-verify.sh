#!/bin/sh
set -u

if [ "$#" -ne 3 ] && [ "$#" -ne 4 ]; then
  printf 'usage: %s BUNDLE EXPECTED_TENANT_ID EXPECTED_RELEASE_COMMIT [CONTRACT_SET_RECONCILIATION]\n' "$0" >&2
  exit 2
fi

bundle=$1
expected_tenant=$2
expected_release=$3
contract_set_reconciliation=${4:-}
failures=0
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
checklist_template="$script_dir/../docs/option-contract-launch-checklist.md"
contract_set_template="$script_dir/../docs/templates/option-contract-set-reconciliation.md"

pass() {
  printf 'PASS  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  failures=$((failures + 1))
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

document_setting_count() {
  file=$1
  name=$2
  awk -v name="$name" '
    $0 ~ "^[[:space:]]*" name "[[:space:]]*:" {count++}
    END {print count+0}
  ' "$file"
}

read_document_setting() {
  file=$1
  name=$2
  awk -v name="$name" '
    $0 ~ "^[[:space:]]*" name "[[:space:]]*:" {
      sub("^[[:space:]]*" name "[[:space:]]*:[[:space:]]*", "")
      sub(/[[:space:]]+$/, "")
      print
      exit
    }
  ' "$file"
}

canonical_checklist_body() {
  awk '
    /^## 1[.] / {inside=1}
    /^## 6[.] / {inside=0}
    inside {
      sub(/^- \[[xX ]\]/, "- [ ]")
      print
    }
  ' "$1"
}

check_document_setting() {
  file=$1
  name=$2
  expected=$3
  description=$4
  count=$(document_setting_count "$file" "$name")
  actual=$(read_document_setting "$file" "$name")
  if [ "$count" -eq 1 ] && [ "$actual" = "$expected" ]; then
    pass "$description"
  else
    fail "$description (expected one '$name: $expected', found count=$count value='$actual')"
  fi
}

check_document_setting_filled() {
  file=$1
  name=$2
  description=$3
  count=$(document_setting_count "$file" "$name")
  actual=$(read_document_setting "$file" "$name")
  case "$actual" in
    ''|0|DRAFT|PENDING|REJECTED|待填写) valid=false ;;
    *) valid=true ;;
  esac
  if [ "$count" -eq 1 ] && [ "$valid" = true ]; then
    pass "$description"
  else
    fail "$description (expected one filled $name, found count=$count value='$actual')"
  fi
}

validate_contract_source() {
  source_file=$1
  expected_sha=$2
  sha_marker_name=$3
  description=$4
  source_kind=$5
  case "$source_file" in
    /*) pass "$description path is absolute" ;;
    *) fail "$description path is absolute" ;;
  esac
  if [ ! -s "$source_file" ]; then
    fail "$description exists and is non-empty"
    return
  fi
  sha_marker_count=$(document_setting_count "$contract_set_reconciliation" "$sha_marker_name")
  case "$expected_sha" in
    *[!A-Fa-f0-9]*|'') source_sha_valid=false ;;
    *)
      if [ "${#expected_sha}" -eq 64 ] && [ "$sha_marker_count" -eq 1 ]; then
        source_sha_valid=true
      else
        source_sha_valid=false
      fi
      ;;
  esac
  if [ "$source_sha_valid" = true ]; then
    pass "$description declares one valid SHA-256"
  else
    fail "$description declares one valid SHA-256"
  fi
  actual_sha=$(sha256_file "$source_file") || {
    fail "$description SHA-256 utility is available"
    return
  }
  normalized_expected_sha=$(printf '%s\n' "$expected_sha" | tr '[:upper:]' '[:lower:]')
  if [ "$actual_sha" = "$normalized_expected_sha" ]; then
    pass "$description hash matches"
  else
    fail "$description hash matches"
  fi
  if [ "$source_kind" = target ]; then
    source_record_description="valid unique PENDING records inside the approved launch window"
    source_format_valid=$(awk -F'|' -v window_start="$contract_window_start" -v window_end="$contract_window_end" '
      NF != 4 || $1 !~ /^[1-9][0-9]*$/ || $2 == "" || $2 ~ /[[:space:]]/ ||
        $3 != 1 || $4 !~ /^[1-9][0-9]*$/ || $4 < window_start || $4 >= window_end {bad=1}
      {if (seen_id[$1]++ || seen_code[$2]++) bad=1}
      END {print bad ? "false" : "true"}
    ' "$source_file")
  else
    source_record_description="valid unique contract identities"
    source_format_valid=$(awk -F'|' '
      NF != 2 || $1 !~ /^[1-9][0-9]*$/ || $2 == "" || $2 ~ /[[:space:]]/ {bad=1}
      {if (seen_id[$1]++ || seen_code[$2]++) bad=1}
      END {print bad ? "false" : "true"}
    ' "$source_file")
  fi
  if [ "$source_format_valid" = true ]; then
    pass "$description has $source_record_description"
  else
    fail "$description has $source_record_description"
  fi
}

case "$expected_tenant" in
  ''|*[!0-9]*|0) fail "expected tenant ID is a positive integer" ;;
esac
if [ -z "$expected_release" ]; then
  fail "expected release commit is non-empty"
fi
if [ ! -s "$bundle" ]; then
  fail "launch bundle exists and is non-empty"
  exit 1
fi

check_document_setting "$bundle" OPTION_LAUNCH_CHECKLIST_STATUS APPROVED \
  "launch bundle is APPROVED"
check_document_setting "$bundle" OPTION_LAUNCH_TENANT_ID "$expected_tenant" \
  "launch bundle tenant matches the production tenant"
check_document_setting "$bundle" OPTION_LAUNCH_RELEASE_COMMIT "$expected_release" \
  "launch bundle release matches the production release"
for approval_name in \
  OPTION_LAUNCH_PRODUCT_APPROVAL_REF \
  OPTION_LAUNCH_TECH_APPROVAL_REF \
  OPTION_LAUNCH_RISK_APPROVAL_REF \
  OPTION_LAUNCH_CLEARING_APPROVAL_REF \
  OPTION_LAUNCH_OPERATIONS_APPROVAL_REF \
  OPTION_LAUNCH_COMPLIANCE_APPROVAL_REF; do
  check_document_setting_filled "$bundle" "$approval_name" \
    "launch bundle has a filled $approval_name"
done

declared_count=$(read_document_setting "$bundle" OPTION_LAUNCH_CONTRACT_COUNT)
declared_count_markers=$(document_setting_count "$bundle" OPTION_LAUNCH_CONTRACT_COUNT)
case "$declared_count" in
  ''|*[!0-9]*|0) fail "launch bundle declares a positive contract count" ;;
  *)
    if [ "$declared_count_markers" -eq 1 ]; then
      pass "launch bundle declares one positive contract count"
    else
      fail "launch bundle declares the contract count exactly once"
    fi
    ;;
esac

record_count=$(awk '/^OPTION_LAUNCH_CONTRACT:[[:space:]]*/ {count++} END {print count+0}' "$bundle")
if [ "$record_count" = "$declared_count" ]; then
  pass "launch bundle record count matches the declared count"
else
  fail "launch bundle record count matches the declared count (declared=$declared_count records=$record_count)"
fi

if awk '
  /^OPTION_LAUNCH_CONTRACT:[[:space:]]*/ {
    line=$0
    sub(/^OPTION_LAUNCH_CONTRACT:[[:space:]]*/, "", line)
    n=split(line, field, "\\|")
    if (n != 5 || field[1] == "" || field[2] == "" || field[3] == "") bad=1
    if (seen_id[field[1]]++ || seen_code[field[2]]++ || seen_path[field[3]]++) bad=1
  }
  END {exit bad ? 1 : 0}
' "$bundle"; then
  pass "launch bundle records have five fields and unique IDs, codes and paths"
else
  fail "launch bundle records have five fields and unique IDs, codes and paths"
fi

if grep -Eq '[|][[:space:]]*待填写([[:space:];|]|$)|通过/拒绝|_{4,}' "$bundle"; then
  fail "launch bundle contains no unresolved production placeholders"
else
  pass "launch bundle contains no unresolved production placeholders"
fi

records=$(sed -n 's/^OPTION_LAUNCH_CONTRACT:[[:space:]]*//p' "$bundle")
while IFS='|' read -r contract_id contract_code checklist_path checklist_sha checklist_status extra; do
  if [ -z "$contract_id$contract_code$checklist_path$checklist_sha$checklist_status${extra:-}" ]; then
    continue
  fi
  label="contract $contract_id/$contract_code"
  case "$contract_id" in
    ''|*[!0-9]*|0) fail "$label has a positive numeric contract_id" ;;
    *) pass "$label has a positive numeric contract_id" ;;
  esac
  case "$contract_code" in
    ''|*[[:space:]]*) fail "$label has a non-empty whitespace-free contract_code" ;;
    *) pass "$label has a non-empty whitespace-free contract_code" ;;
  esac
  case "$checklist_path" in
    /*) pass "$label checklist path is absolute" ;;
    *) fail "$label checklist path is absolute" ;;
  esac
  if [ -n "${extra:-}" ]; then
    fail "$label contains more than five fields"
  fi
  if [ "$checklist_status" = APPROVED ]; then
    pass "$label record status is APPROVED"
  else
    fail "$label record status is APPROVED"
  fi
  case "$checklist_sha" in
    *[!A-Fa-f0-9]*|'') fail "$label checklist SHA-256 is valid" ;;
    *)
      if [ "${#checklist_sha}" -eq 64 ]; then
        pass "$label checklist SHA-256 is valid"
      else
        fail "$label checklist SHA-256 is valid"
      fi
      ;;
  esac
  if [ ! -s "$checklist_path" ]; then
    fail "$label checklist exists and is non-empty"
    continue
  fi
  actual_sha=$(sha256_file "$checklist_path") || {
    fail "$label checklist SHA-256 utility is available"
    continue
  }
  expected_sha=$(printf '%s\n' "$checklist_sha" | tr '[:upper:]' '[:lower:]')
  if [ "$actual_sha" = "$expected_sha" ]; then
    pass "$label checklist hash matches"
  else
    fail "$label checklist hash matches"
  fi
  check_document_setting "$checklist_path" OPTION_LAUNCH_CHECKLIST_STATUS APPROVED \
    "$label checklist is APPROVED"
  check_document_setting "$checklist_path" OPTION_CONTRACT_TENANT_ID "$expected_tenant" \
    "$label checklist tenant matches"
  check_document_setting "$checklist_path" OPTION_CONTRACT_ID "$contract_id" \
    "$label checklist contract_id matches"
  check_document_setting "$checklist_path" OPTION_CONTRACT_CODE "$contract_code" \
    "$label checklist contract_code matches"
  for approval_name in \
    OPTION_PRODUCT_APPROVAL_REF \
    OPTION_TECH_APPROVAL_REF \
    OPTION_RISK_APPROVAL_REF \
    OPTION_CLEARING_APPROVAL_REF \
    OPTION_OPERATIONS_APPROVAL_REF \
    OPTION_COMPLIANCE_APPROVAL_REF; do
    check_document_setting_filled "$checklist_path" "$approval_name" \
      "$label checklist has a filled $approval_name"
  done
  required_checkbox_count=$(awk '/^- \[[ xX]\]/{count++} END {print count+0}' "$checklist_template")
  checklist_checkbox_count=$(awk '/^- \[[ xX]\]/{count++} END {print count+0}' "$checklist_path")
  unchecked_checkbox_count=$(awk '/^- \[ \]/{count++} END {print count+0}' "$checklist_path")
  if [ "$required_checkbox_count" -gt 0 ] && [ "$checklist_checkbox_count" -eq "$required_checkbox_count" ]; then
    pass "$label checklist retains all $required_checkbox_count acceptance items"
  else
    fail "$label checklist retains all template acceptance items (required=$required_checkbox_count found=$checklist_checkbox_count)"
  fi
  if [ "$unchecked_checkbox_count" -eq 0 ]; then
    pass "$label checklist has no unchecked acceptance item"
  else
    fail "$label checklist has no unchecked acceptance item (found=$unchecked_checkbox_count)"
  fi
  template_body=$(canonical_checklist_body "$checklist_template")
  checklist_body=$(canonical_checklist_body "$checklist_path")
  if [ -n "$template_body" ] && [ "$checklist_body" = "$template_body" ]; then
    pass "$label checklist retains the exact governed acceptance text"
  else
    fail "$label checklist retains the exact governed acceptance text"
  fi
  if grep -Eq '通过/拒绝|_{4,}' "$checklist_path"; then
    fail "$label checklist contains no unresolved placeholders"
  else
    pass "$label checklist contains no unresolved placeholders"
  fi
done <<EOF
$records
EOF

if [ -n "$contract_set_reconciliation" ]; then
  if [ ! -s "$contract_set_reconciliation" ]; then
    fail "contract-set reconciliation exists and is non-empty"
  else
    check_document_setting "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_RECONCILIATION_STATUS APPROVED \
      "contract-set reconciliation is APPROVED"
    check_document_setting "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_TENANT_ID "$expected_tenant" \
      "contract-set reconciliation tenant matches"
    check_document_setting "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_RELEASE_COMMIT "$expected_release" \
      "contract-set reconciliation release matches"
    check_document_setting_filled "$contract_set_reconciliation" \
      OPTION_TARGET_CONTRACT_SET_SOURCE \
      "contract-set reconciliation identifies the target-environment export"
    check_document_setting_filled "$contract_set_reconciliation" \
      OPTION_PUBLICATION_CONTRACT_SET_SOURCE \
      "contract-set reconciliation identifies the publication artifact"
    check_document_setting_filled "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_REVIEW_REF \
      "contract-set reconciliation has an independent review reference"

    contract_window_start=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_WINDOW_START_MS)
    contract_window_end=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_WINDOW_END_MS)
    window_start_count=$(document_setting_count "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_WINDOW_START_MS)
    window_end_count=$(document_setting_count "$contract_set_reconciliation" \
      OPTION_CONTRACT_SET_WINDOW_END_MS)
    case "$contract_window_start" in
      ''|*[!0-9]*|0) contract_window_valid=false ;;
      *)
        case "$contract_window_end" in
          ''|*[!0-9]*|0) contract_window_valid=false ;;
          *)
            if [ "$window_start_count" -eq 1 ] && [ "$window_end_count" -eq 1 ] && \
              [ "$contract_window_start" -lt "$contract_window_end" ]; then
              contract_window_valid=true
            else
              contract_window_valid=false
            fi
            ;;
        esac
        ;;
    esac
    if [ "$contract_window_valid" = true ]; then
      pass "contract-set reconciliation declares one valid half-open launch window"
    else
      fail "contract-set reconciliation declares one valid half-open launch window"
    fi

    target_source=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_TARGET_CONTRACT_SET_SOURCE)
    target_source_sha=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_TARGET_CONTRACT_SET_SOURCE_SHA256)
    publication_source=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_PUBLICATION_CONTRACT_SET_SOURCE)
    publication_source_sha=$(read_document_setting "$contract_set_reconciliation" \
      OPTION_PUBLICATION_CONTRACT_SET_SOURCE_SHA256)
    validate_contract_source "$target_source" "$target_source_sha" \
      OPTION_TARGET_CONTRACT_SET_SOURCE_SHA256 "target-environment contract export" target
    validate_contract_source "$publication_source" "$publication_source_sha" \
      OPTION_PUBLICATION_CONTRACT_SET_SOURCE_SHA256 "publication contract export" publication

    reconciled_count=$(read_document_setting "$contract_set_reconciliation" OPTION_CONTRACT_SET_COUNT)
    reconciled_count_markers=$(document_setting_count "$contract_set_reconciliation" OPTION_CONTRACT_SET_COUNT)
    case "$reconciled_count" in
      ''|*[!0-9]*|0) fail "contract-set reconciliation declares a positive contract count" ;;
      *)
        if [ "$reconciled_count_markers" -eq 1 ]; then
          pass "contract-set reconciliation declares one positive contract count"
        else
          fail "contract-set reconciliation declares the contract count exactly once"
        fi
        ;;
    esac

    if [ -s "$target_source" ]; then
      target_record_count=$(awk 'END {print NR+0}' "$target_source")
      target_identity_set=$(awk -F'|' '{print $1 "|" $2}' "$target_source" | LC_ALL=C sort)
    else
      target_record_count=0
      target_identity_set=
    fi
    if [ -s "$publication_source" ]; then
      publication_record_count=$(awk 'END {print NR+0}' "$publication_source")
      publication_identity_set=$(LC_ALL=C sort "$publication_source")
    else
      publication_record_count=0
      publication_identity_set=
    fi
    if [ "$target_record_count" = "$reconciled_count" ] && \
      [ "$publication_record_count" = "$reconciled_count" ] && \
      [ "$record_count" = "$reconciled_count" ]; then
      pass "approved, target and publication contract counts are equal"
    else
      fail "approved, target and publication contract counts are equal (approved=$record_count target=$target_record_count publication=$publication_record_count declared=$reconciled_count)"
    fi

    bundle_identity_set=$(printf '%s\n' "$records" | awk -F'|' 'NF {print $1 "|" $2}' | LC_ALL=C sort)
    if [ "$bundle_identity_set" = "$target_identity_set" ]; then
      pass "approved bundle exactly matches the target-environment contract set"
    else
      fail "approved bundle exactly matches the target-environment contract set"
    fi
    if [ "$bundle_identity_set" = "$publication_identity_set" ]; then
      pass "approved bundle exactly matches the publication contract set"
    else
      fail "approved bundle exactly matches the publication contract set"
    fi
    required_reconciliation_checks=$(awk '/^- \[[ xX]\]/{count++} END {print count+0}' \
      "$contract_set_template")
    actual_reconciliation_checks=$(awk '/^- \[[ xX]\]/{count++} END {print count+0}' \
      "$contract_set_reconciliation")
    unchecked_reconciliation_checks=$(awk '/^- \[ \]/{count++} END {print count+0}' \
      "$contract_set_reconciliation")
    required_reconciliation_text=$(awk '/^- \[[ xX]\]/{sub(/^- \[[xX ]\]/, "- [ ]"); print}' \
      "$contract_set_template")
    actual_reconciliation_text=$(awk '/^- \[[ xX]\]/{sub(/^- \[[xX ]\]/, "- [ ]"); print}' \
      "$contract_set_reconciliation")
    if [ "$required_reconciliation_checks" -gt 0 ] && \
      [ "$actual_reconciliation_checks" -eq "$required_reconciliation_checks" ] && \
      [ "$required_reconciliation_text" = "$actual_reconciliation_text" ]; then
      pass "contract-set reconciliation retains all governed acceptance items"
    else
      fail "contract-set reconciliation retains all governed acceptance items"
    fi
    if [ "$unchecked_reconciliation_checks" -eq 0 ]; then
      pass "contract-set reconciliation has no unchecked acceptance item"
    else
      fail "contract-set reconciliation has no unchecked acceptance item (found=$unchecked_reconciliation_checks)"
    fi
    if grep -Eq '[|][[:space:]]*待填写([[:space:];|]|$)|通过/拒绝|_{4,}' \
      "$contract_set_reconciliation"; then
      fail "contract-set reconciliation contains no unresolved production placeholders"
    else
      pass "contract-set reconciliation contains no unresolved production placeholders"
    fi
  fi
fi

if [ "$failures" -ne 0 ]; then
  printf '\nLaunch bundle verification failed with %s error(s).\n' "$failures" >&2
  exit 1
fi

printf '\nLaunch bundle verification passed for %s contract(s).\n' "$record_count"
