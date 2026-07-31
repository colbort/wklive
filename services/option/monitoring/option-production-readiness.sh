#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OPTION_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
OPTION_SCHEMA_FILES="$OPTION_DIR/option.sql"
REPO_ROOT=$(CDPATH= cd -- "$OPTION_DIR/../.." && pwd)
MODE=production
if [ "${1:-}" = "--repository-only" ]; then
  MODE=repository
  shift
fi
READINESS_FILE="${1:-$SCRIPT_DIR/option-production-readiness.env}"
failures=0

pass() {
  printf 'PASS  %s\n' "$1"
}

skip() {
  printf 'SKIP  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1"
  failures=$((failures + 1))
}

read_setting() {
  setting_name="$1"
  if [ ! -f "$READINESS_FILE" ]; then
    return
  fi
  awk -v setting_name="$setting_name" '
    $0 ~ "^[[:space:]]*" setting_name "[[:space:]]*=" {
      sub("^[[:space:]]*" setting_name "[[:space:]]*=[[:space:]]*", "")
      sub(/[[:space:]]+$/, "")
      if (($0 ~ /^".*"$/) || ($0 ~ /^'\''.*'\''$/)) {
        print substr($0, 2, length($0) - 2)
      } else {
        print
      }
      exit
    }
  ' "$READINESS_FILE"
}

require_value() {
  if [ -n "$1" ]; then
    pass "$2"
  else
    fail "$2"
  fi
}

require_true() {
  if [ "$1" = "true" ]; then
    pass "$2"
  else
    fail "$2 (must be true)"
  fi
}

require_boolean() {
  case "$1" in
    true|false) pass "$2" ;;
    *) fail "$2 (must be true or false)" ;;
  esac
}

require_positive_integer() {
  case "$1" in
    ''|*[!0-9]*|0) fail "$2 (must be a positive integer)" ;;
    *) pass "$2" ;;
  esac
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

require_evidence_file() {
  evidence_file="$1"
  expected_sha256="$2"
  description="$3"
  if [ -z "$evidence_file" ] || [ ! -s "$evidence_file" ]; then
    fail "$description (file not found or empty)"
    return
  fi
  case "$expected_sha256" in
    ''|*[!A-Fa-f0-9]*) fail "$description (invalid SHA-256)"; return ;;
  esac
  if [ "${#expected_sha256}" -ne 64 ]; then
    fail "$description (invalid SHA-256)"
    return
  fi
  actual_sha256=$(sha256_file "$evidence_file") || {
    fail "$description (SHA-256 utility unavailable)"
    return
  }
  expected_sha256=$(printf '%s\n' "$expected_sha256" | tr '[:upper:]' '[:lower:]')
  if [ "$actual_sha256" = "$expected_sha256" ]; then
    pass "$description"
  else
    fail "$description (SHA-256 mismatch)"
  fi
}

check_contains() {
  file="$1"
  pattern="$2"
  description="$3"
  if [ -s "$file" ] && grep -Eq "$pattern" "$file"; then
    pass "$description"
  else
    fail "$description"
  fi
}

check_not_contains() {
  file="$1"
  pattern="$2"
  description="$3"
  if [ -s "$file" ] && ! grep -Eq "$pattern" "$file"; then
    pass "$description"
  else
    fail "$description"
  fi
}

alert_count() {
  awk '/^[[:space:]]*- alert:/{count++} END{print count+0}' "$1"
}

printf 'Option production readiness (%s mode)\n\n' "$MODE"

operations_rules="$SCRIPT_DIR/option-operations-alert-rules.yml"
combo_rules="$SCRIPT_DIR/option-alert-rules.yml"
operations_count=$(alert_count "$operations_rules")
combo_count=$(alert_count "$combo_rules")
if [ "$operations_count" -eq 49 ]; then
  pass "49 Option operations alert rules are present"
else
  fail "49 Option operations alert rules are present (found $operations_count)"
fi
if [ "$combo_count" -eq 4 ]; then
  pass "4 Option combo alert rules are present"
else
  fail "4 Option combo alert rules are present (found $combo_count)"
fi

duplicate_alerts=$(
  awk '/^[[:space:]]*- alert:/{print $3}' "$operations_rules" "$combo_rules" |
    sort | uniq -d
)
if [ -z "$duplicate_alerts" ]; then
  pass "Option alert names are unique"
else
  fail "Option alert names are unique (duplicates: $duplicate_alerts)"
fi

severity_count=$(awk '$1=="severity:"{count++} END{print count+0}' "$operations_rules" "$combo_rules")
catalog_count=$(awk '$1=="catalog_id:"{count++} END{print count+0}' "$operations_rules" "$combo_rules")
if [ "$severity_count" -eq 53 ] && [ "$catalog_count" -eq 53 ]; then
  pass "all 53 Option alerts declare severity and catalog_id"
else
  fail "all 53 Option alerts declare severity and catalog_id (severity=$severity_count catalog=$catalog_count)"
fi

required_indexes='idx_option_contract_monitor
idx_option_contract_lifecycle_monitor
idx_option_public_chain_monitor
idx_option_user_control_monitor
idx_option_risk_account_portfolio_monitor
idx_option_portfolio_config_monitor
idx_option_control_event_monitor
idx_option_position_monitor
idx_option_settlement_price_monitor
idx_option_trade_correction_monitor
idx_corporate_action_monitor
idx_corporate_action_contract_monitor
idx_option_contract_series_monitor
idx_option_mmp_monitor
idx_option_exercise_monitor
idx_option_asset_instruction_control_monitor
idx_option_physical_delivery_monitor'
baseline_migration="$OPTION_DIR/migrations/20260731_zr_option_operations_monitoring_indexes.sql"
incremental_migration="$OPTION_DIR/migrations/20260731_zs_option_time_sensitive_monitoring_indexes.sql"
baseline_sha256=$(sha256_file "$baseline_migration") || baseline_sha256=""
if [ "$baseline_sha256" = "5a0614ceed65d6c17259b49c3e1958f52960b19a1378229a75d28e5812739942" ]; then
  pass "recorded 20260731_zr monitoring migration checksum is unchanged"
else
  fail "recorded 20260731_zr monitoring migration checksum is unchanged"
fi
index_ok=true
for index_name in $required_indexes; do
  migration_file="$baseline_migration"
  case "$index_name" in
    idx_option_exercise_monitor|idx_option_asset_instruction_control_monitor|idx_option_physical_delivery_monitor)
      migration_file="$incremental_migration"
      ;;
  esac
  if ! grep -q "$index_name" $OPTION_SCHEMA_FILES ||
     ! grep -q "$index_name" "$migration_file"; then
    fail "monitoring index $index_name exists in schema and idempotent migration"
    index_ok=false
  fi
done
if [ "$index_ok" = "true" ]; then
  pass "all 17 Option monitoring indexes exist in schema and migration"
fi

daily_reconciliation_migration="$OPTION_DIR/migrations/20260731_zt_option_daily_reconciliation_run.sql"
daily_reconciliation_job_migration="$REPO_ROOT/services/system/migrations/20260731_zt_option_daily_reconciliation_job.sql"
check_contains "$OPTION_DIR/option.sql" 'CREATE TABLE `t_option_reconciliation_run`' \
  "Option schema declares immutable daily reconciliation runs"
check_contains "$daily_reconciliation_migration" 'trg_option_reconciliation_run_no_update' \
  "daily reconciliation run migration protects immutable evidence"
check_contains "$daily_reconciliation_migration" 'DECIMAL\(36,18\)' \
  "Option wallet mirror preserves Asset 18-decimal precision"
check_contains "$daily_reconciliation_migration" 'idx_option_reconciliation_run_monitor' \
  "daily reconciliation heartbeat query has a scope/status/time index"
check_contains "$OPTION_DIR/option.sql" 'CREATE TABLE `t_option_reconciliation_run_detail`' \
  "Option schema declares immutable scope-2 reconciliation details"
check_contains "$daily_reconciliation_migration" 'fk_option_reconciliation_detail_run' \
  "scope-2 reconciliation details enforce run tenant/date/scope identity"
check_contains "$daily_reconciliation_migration" 'expected_closing`=`opening_amount`\+`external_net`\+`option_net`\+`manual_net' \
  "scope-2 reconciliation details enforce the conservation formula"
check_contains "$daily_reconciliation_migration" 'trg_option_reconciliation_detail_no_update' \
  "scope-2 reconciliation details preserve immutable evidence"
asset_reconciliation_index_migration="$REPO_ROOT/services/asset/migrations/20260731_option_daily_reconciliation_index.sql"
check_contains "$REPO_ROOT/services/asset/asset.sql" 'idx_asset_flow_option_reconciliation' \
  "Asset schema indexes Option scope-2 flow cutoffs"
check_contains "$asset_reconciliation_index_migration" 'idx_asset_flow_option_reconciliation' \
  "Asset migration adds the Option scope-2 flow index idempotently"
check_contains "$REPO_ROOT/services/asset/asset.sql" 'idx_platform_flow_option_reconciliation' \
  "Asset schema indexes Option scope-2 platform-flow cutoffs"
check_contains "$asset_reconciliation_index_migration" 'idx_platform_flow_option_reconciliation' \
  "Asset migration adds the Option scope-2 platform-flow index idempotently"
check_contains "$daily_reconciliation_job_migration" 'option.ProcessDailyReconciliation' \
  "System migration schedules Option daily reconciliation"

if command -v promtool >/dev/null 2>&1; then
  if promtool check rules "$operations_rules" "$combo_rules" >/dev/null; then
    pass "promtool validates Option alert rules"
  else
    fail "promtool validates Option alert rules"
  fi
elif [ "$MODE" = "repository" ]; then
  skip "promtool is unavailable; production mode requires it"
else
  fail "promtool is installed for production validation"
fi

if [ "$MODE" = "repository" ]; then
  printf '\n'
  if [ "$failures" -eq 0 ]; then
    printf 'READY: repository-owned Option monitoring artifacts passed.\n'
    exit 0
  fi
  printf 'NOT READY: %s repository check(s) failed.\n' "$failures"
  exit 1
fi

if [ -f "$READINESS_FILE" ]; then
  pass "Option production readiness attestation exists"
else
  fail "Option production readiness attestation exists ($READINESS_FILE)"
fi

OPTION_PRODUCTION_TENANT_ID=$(read_setting OPTION_PRODUCTION_TENANT_ID)
OPTION_RELEASE_COMMIT=$(read_setting OPTION_RELEASE_COMMIT)
OPTION_PRODUCT_APPROVAL_REF=$(read_setting OPTION_PRODUCT_APPROVAL_REF)
OPTION_RISK_APPROVAL_REF=$(read_setting OPTION_RISK_APPROVAL_REF)
OPTION_CLEARING_APPROVAL_REF=$(read_setting OPTION_CLEARING_APPROVAL_REF)
OPTION_COMPLIANCE_APPROVAL_REF=$(read_setting OPTION_COMPLIANCE_APPROVAL_REF)
OPTION_ZERO_DIFF_HEARTBEAT_ENABLED=$(read_setting OPTION_ZERO_DIFF_HEARTBEAT_ENABLED)
OPTION_PRODUCTION_METRICS_TARGET_VERIFIED=$(read_setting OPTION_PRODUCTION_METRICS_TARGET_VERIFIED)
OPTION_METRICS_URL=$(read_setting OPTION_METRICS_URL)
OPTION_SELLER_TRADING_ENABLED=$(read_setting OPTION_SELLER_TRADING_ENABLED)
OPTION_PORTFOLIO_MARGIN_ENABLED=$(read_setting OPTION_PORTFOLIO_MARGIN_ENABLED)
OPTION_PHYSICAL_DELIVERY_ENABLED=$(read_setting OPTION_PHYSICAL_DELIVERY_ENABLED)
OPTION_COMPLEX_ORDERS_ENABLED=$(read_setting OPTION_COMPLEX_ORDERS_ENABLED)
OPTION_PUBLIC_MARKET_ENABLED=$(read_setting OPTION_PUBLIC_MARKET_ENABLED)
OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=$(read_setting OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED)
OPTION_INSURANCE_FLOW_SIGN_RESOLVED=$(read_setting OPTION_INSURANCE_FLOW_SIGN_RESOLVED)
OPTION_BACKSTOP_LIMIT_APPROVED=$(read_setting OPTION_BACKSTOP_LIMIT_APPROVED)

require_positive_integer "$OPTION_PRODUCTION_TENANT_ID" "production Option tenant declared"
if printf '%s\n' "$OPTION_RELEASE_COMMIT" |
   grep -Eq '^([A-Fa-f0-9]{7,64}|sha256:[A-Fa-f0-9]{64})$'; then
  pass "Option release commit/image digest declared"
else
  fail "Option release commit/image digest declared (must be commit hash or sha256 digest)"
fi
require_value "$OPTION_PRODUCT_APPROVAL_REF" "Option product approval reference declared"
require_value "$OPTION_RISK_APPROVAL_REF" "Option risk approval reference declared"
require_value "$OPTION_CLEARING_APPROVAL_REF" "Option clearing approval reference declared"
require_value "$OPTION_COMPLIANCE_APPROVAL_REF" "Option compliance approval reference declared"
require_true "$OPTION_ZERO_DIFF_HEARTBEAT_ENABLED" "Asset daily zero-difference heartbeat enabled"
require_true "$OPTION_PRODUCTION_METRICS_TARGET_VERIFIED" "production Option metrics target verified"

case "$OPTION_METRICS_URL" in
  http://*|https://*) pass "production Option metrics URL declared" ;;
  *) fail "production Option metrics URL declared (must be http or https)" ;;
esac
metrics_payload=""
if command -v curl >/dev/null 2>&1; then
  metrics_payload=$(curl --fail --silent --show-error --max-time 10 "$OPTION_METRICS_URL" 2>/dev/null) ||
    metrics_payload=""
  if [ -n "$metrics_payload" ]; then
    pass "production Option metrics endpoint is reachable"
  else
    fail "production Option metrics endpoint is reachable"
  fi
else
  fail "curl is installed for production metrics verification"
fi

expected_metric_families='wklive_option_operations_sample_success
wklive_option_operations_last_success_timestamp_seconds
wklive_option_operations_count
wklive_option_operations_oldest_timestamp_seconds
wklive_option_operations_amount
wklive_option_risk_scan_groups
wklive_option_risk_scan_failed_groups
wklive_option_risk_scan_failure_ratio
wklive_option_combo_isolation_violation_total
wklive_option_combo_debit_barrier_violation_total'
metric_families_ok=true
for metric_name in $expected_metric_families; do
  if ! printf '%s\n' "$metrics_payload" | grep -q "^# HELP $metric_name "; then
    metric_families_ok=false
  fi
done
if [ "$metric_families_ok" = "true" ]; then
  pass "production endpoint exposes required Option metric families"
else
  fail "production endpoint exposes required Option metric families"
fi

sample_success=$(
  printf '%s\n' "$metrics_payload" |
    awk '$1=="wklive_option_operations_sample_success"{print $2; exit}'
)
if [ "$sample_success" = "1" ]; then
  pass "latest Option operations sample succeeded"
else
  fail "latest Option operations sample succeeded"
fi
last_success=$(
  printf '%s\n' "$metrics_payload" |
    awk '$1=="wklive_option_operations_last_success_timestamp_seconds"{print $2; exit}'
)
metrics_fresh=$(
  LC_ALL=C awk -v now="$(date +%s)" -v last="$last_success" '
    BEGIN {
      if (last ~ /^[0-9]+([.][0-9]+)?$/ && last > 0 && now-last >= 0 && now-last <= 45) {
        print "true"
      } else {
        print "false"
      }
    }'
)
if [ "$metrics_fresh" = "true" ]; then
  pass "Option operations sample is no older than 45 seconds"
else
  fail "Option operations sample is no older than 45 seconds"
fi

mirror_last_success=$(
  printf '%s\n' "$metrics_payload" |
    awk -v tenant="$OPTION_PRODUCTION_TENANT_ID" '
      $1 ~ /^wklive_option_operations_oldest_timestamp_seconds\{/ &&
      $1 ~ ("tenant_id=\"" tenant "\"") &&
      $1 ~ /category="daily_mirror_reconciliation_heartbeat"/ {print $2; exit}
    '
)
mirror_heartbeat_fresh=$(
  LC_ALL=C awk -v now="$(date +%s)" -v last="$mirror_last_success" '
    BEGIN {
      if (last ~ /^[0-9]+([.][0-9]+)?$/ && last > 0 && now-last >= 0 && now-last <= 129600) {
        print "true"
      } else {
        print "false"
      }
    }'
)
if [ "$mirror_heartbeat_fresh" = "true" ]; then
  pass "production tenant wallet-mirror reconciliation succeeded within 36 hours"
else
  fail "production tenant wallet-mirror reconciliation succeeded within 36 hours"
fi

conservation_last_success=$(
  printf '%s\n' "$metrics_payload" |
    awk -v tenant="$OPTION_PRODUCTION_TENANT_ID" '
      $1 ~ /^wklive_option_operations_oldest_timestamp_seconds\{/ &&
      $1 ~ ("tenant_id=\"" tenant "\"") &&
      $1 ~ /category="daily_conservation_heartbeat"/ {print $2; exit}
    '
)
conservation_heartbeat_fresh=$(
  LC_ALL=C awk -v now="$(date +%s)" -v last="$conservation_last_success" '
    BEGIN {
      if (last ~ /^[0-9]+([.][0-9]+)?$/ && last > 0 && now-last >= 0 && now-last <= 129600) {
        print "true"
      } else {
        print "false"
      }
    }'
)
if [ "$conservation_heartbeat_fresh" = "true" ]; then
  pass "production tenant full-funds reconciliation succeeded within 36 hours"
else
  fail "production tenant full-funds reconciliation succeeded within 36 hours"
fi

for feature_flag in OPTION_SELLER_TRADING_ENABLED OPTION_PORTFOLIO_MARGIN_ENABLED \
  OPTION_PHYSICAL_DELIVERY_ENABLED OPTION_COMPLEX_ORDERS_ENABLED \
  OPTION_PUBLIC_MARKET_ENABLED OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED; do
  feature_value=$(read_setting "$feature_flag")
  require_boolean "$feature_value" "$feature_flag explicitly declares release scope"
done

evidence_names='OPTION_RELEASE_SIGNOFF
OPTION_MIGRATION_REPORT
OPTION_ASSET_E2E_REPORT
OPTION_FAILURE_INJECTION_REPORT
OPTION_CAPACITY_REPORT
OPTION_ALERT_DELIVERY_REPORT
OPTION_DAILY_RECONCILIATION_REPORT
OPTION_DATABASE_AUDIT_REPORT
OPTION_ONCALL_ROSTER_REPORT'
for evidence_name in $evidence_names; do
  evidence_path=$(read_setting "$evidence_name")
  evidence_sha=$(read_setting "${evidence_name}_SHA256")
  require_evidence_file "$evidence_path" "$evidence_sha" "$evidence_name attached and hash-matched"
done

OPTION_RELEASE_SIGNOFF=$(read_setting OPTION_RELEASE_SIGNOFF)
check_contains "$OPTION_RELEASE_SIGNOFF" '^OPTION_EVIDENCE_STATUS:[[:space:]]*APPROVED$' \
  "Option release signoff is explicitly APPROVED"

OPTION_PROMETHEUS_CONFIG=$(read_setting OPTION_PROMETHEUS_CONFIG)
OPTION_PROMETHEUS_CONFIG_SHA256=$(read_setting OPTION_PROMETHEUS_CONFIG_SHA256)
OPTION_ALERTMANAGER_CONFIG=$(read_setting OPTION_ALERTMANAGER_CONFIG)
OPTION_ALERTMANAGER_CONFIG_SHA256=$(read_setting OPTION_ALERTMANAGER_CONFIG_SHA256)
require_evidence_file "$OPTION_PROMETHEUS_CONFIG" "$OPTION_PROMETHEUS_CONFIG_SHA256" \
  "production Prometheus config attached and hash-matched"
require_evidence_file "$OPTION_ALERTMANAGER_CONFIG" "$OPTION_ALERTMANAGER_CONFIG_SHA256" \
  "production Alertmanager config attached and hash-matched"
check_contains "$OPTION_PROMETHEUS_CONFIG" 'option-operations-alert-rules[.]yml' \
  "production Prometheus loads Option operations rules"
check_contains "$OPTION_PROMETHEUS_CONFIG" 'option-alert-rules[.]yml' \
  "production Prometheus loads Option combo rules"
check_contains "$OPTION_PROMETHEUS_CONFIG" '9105' \
  "production Prometheus scrapes Option metrics port 9105"
check_contains "$OPTION_ALERTMANAGER_CONFIG" 'severity.*critical|critical.*severity' \
  "Alertmanager routes Option critical alerts"
check_contains "$OPTION_ALERTMANAGER_CONFIG" 'severity.*warning|warning.*severity' \
  "Alertmanager routes Option warning alerts"
check_contains "$OPTION_ALERTMANAGER_CONFIG" 'severity.*info|info.*severity' \
  "Alertmanager routes Option info alerts"
check_not_contains "$OPTION_ALERTMANAGER_CONFIG" 'REPLACE_BEFORE_DEPLOY' \
  "Alertmanager example placeholders have been replaced"

if command -v promtool >/dev/null 2>&1; then
  if promtool check config "$OPTION_PROMETHEUS_CONFIG" >/dev/null; then
    pass "promtool validates production Prometheus config"
  else
    fail "promtool validates production Prometheus config"
  fi
fi
if command -v amtool >/dev/null 2>&1; then
  if amtool check-config "$OPTION_ALERTMANAGER_CONFIG" >/dev/null; then
    pass "amtool validates production Alertmanager config"
  else
    fail "amtool validates production Alertmanager config"
  fi
else
  fail "amtool is installed for production validation"
fi

if [ "$OPTION_SELLER_TRADING_ENABLED" = "true" ]; then
  require_true "$OPTION_INSURANCE_FLOW_SIGN_RESOLVED" \
    "insurance payout sign mismatch resolved before seller trading"
  require_true "$OPTION_BACKSTOP_LIMIT_APPROVED" \
    "platform backstop limit approved before seller trading"
  evidence_path=$(read_setting OPTION_INSURANCE_SIGN_RESOLUTION_REPORT)
  evidence_sha=$(read_setting OPTION_INSURANCE_SIGN_RESOLUTION_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance sign resolution report attached and hash-matched"
fi

if [ "$OPTION_PORTFOLIO_MARGIN_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_MODEL_VALIDATION_REPORT)
  evidence_sha=$(read_setting OPTION_MODEL_VALIDATION_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "independent portfolio model validation attached and hash-matched"
fi
if [ "$OPTION_PHYSICAL_DELIVERY_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_PHYSICAL_DELIVERY_APPROVAL)
  evidence_sha=$(read_setting OPTION_PHYSICAL_DELIVERY_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "physical-delivery default and legal approval attached and hash-matched"
fi
if [ "$OPTION_COMPLEX_ORDERS_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_COMPLEX_ORDER_E2E_REPORT)
  evidence_sha=$(read_setting OPTION_COMPLEX_ORDER_E2E_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "complex-order concurrency/Asset E2E attached and hash-matched"
fi
if [ "$OPTION_PUBLIC_MARKET_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_PUBLIC_MARKET_PROBE_REPORT)
  evidence_sha=$(read_setting OPTION_PUBLIC_MARKET_PROBE_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "public market cross-tenant/TTL/SLA probe attached and hash-matched"
fi
if [ "$OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_GREEKS_THRESHOLD_APPROVAL)
  evidence_sha=$(read_setting OPTION_GREEKS_THRESHOLD_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "Greeks freshness threshold approval attached and hash-matched"
fi

printf '\n'
if [ "$failures" -eq 0 ]; then
  printf 'READY: Option production prerequisites passed; release still follows the approved change window.\n'
  exit 0
fi
printf 'NOT READY: %s Option prerequisite(s) failed. Trading gates must remain closed.\n' "$failures"
exit 1
