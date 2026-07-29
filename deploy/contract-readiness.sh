#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
READINESS_FILE="${1:-$SCRIPT_DIR/production-readiness.env}"
failures=0

pass() {
  printf 'PASS  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1"
  failures=$((failures + 1))
}

require_value() {
  variable_value="$1"
  description="$2"
  if [ -n "$variable_value" ]; then
    pass "$description"
  else
    fail "$description"
  fi
}

require_true() {
  variable_value="$1"
  description="$2"
  if [ "$variable_value" = "true" ]; then
    pass "$description"
  else
    fail "$description (must be true)"
  fi
}

require_positive_integer() {
  variable_value="$1"
  description="$2"
  case "$variable_value" in
    ''|*[!0-9]*|0)
      fail "$description (must be a positive integer)"
      ;;
    *)
      pass "$description"
      ;;
  esac
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
    *[!A-Fa-f0-9]*|'')
      fail "$description (invalid SHA-256)"
      return
      ;;
  esac
  if [ "${#expected_sha256}" -ne 64 ]; then
    fail "$description (invalid SHA-256)"
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    actual_sha256=$(shasum -a 256 "$evidence_file" | awk '{print $1}')
  elif command -v sha256sum >/dev/null 2>&1; then
    actual_sha256=$(sha256sum "$evidence_file" | awk '{print $1}')
  else
    fail "$description (SHA-256 utility unavailable)"
    return
  fi
  expected_sha256=$(printf '%s\n' "$expected_sha256" | tr '[:upper:]' '[:lower:]')
  if [ "$actual_sha256" = "$expected_sha256" ]; then
    pass "$description"
  else
    fail "$description (SHA-256 mismatch)"
  fi
}

valid_token() {
  case "$1" in
    ''|*[!A-Za-z0-9._-]*)
      return 1
      ;;
    *)
      return 0
      ;;
  esac
}

read_setting() {
  setting_name="$1"
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

db_value() {
  key="$1"
  printf '%s\n' "$db_output" |
    awk -F= -v key="$key" '$1 == key {sub(/^[^=]*=/, ""); print; exit}'
}

db_number() {
  key="$1"
  fallback="$2"
  value=$(db_value "$key")
  case "$value" in
    ''|*[!0-9]*)
      printf '%s\n' "$fallback"
      ;;
    *)
      printf '%s\n' "$value"
      ;;
  esac
}

check_service() {
  service_name="$1"
  container_id=$(docker compose -f "$COMPOSE_FILE" ps -q "$service_name" 2>/dev/null)
  if [ -z "$container_id" ]; then
    fail "Compose service $service_name is healthy"
    return
  fi
  service_status=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null)
  if [ "$service_status" = "healthy" ]; then
    pass "Compose service $service_name is healthy"
  else
    fail "Compose service $service_name is healthy (status=${service_status:-unknown})"
  fi
}

check_completed_service() {
  service_name="$1"
  container_id=$(docker compose -f "$COMPOSE_FILE" ps -aq "$service_name" 2>/dev/null)
  if [ -z "$container_id" ]; then
    fail "Compose initializer $service_name completed successfully"
    return
  fi
  service_status=$(docker inspect --format '{{.State.Status}}:{{.State.ExitCode}}' "$container_id" 2>/dev/null)
  if [ "$service_status" = "exited:0" ]; then
    pass "Compose initializer $service_name completed successfully"
  else
    fail "Compose initializer $service_name completed successfully (status=${service_status:-unknown})"
  fi
}

printf 'Contract production readiness (read-only)\n'
printf 'Attestation file: %s\n\n' "$READINESS_FILE"

if [ -f "$READINESS_FILE" ]; then
  pass "production readiness attestation exists"
else
  fail "production readiness attestation exists"
fi

if [ -f "$READINESS_FILE" ]; then
  PRODUCTION_PRICE_SOURCE_IDS=$(read_setting PRODUCTION_PRICE_SOURCE_IDS)
  PRICE_SOURCE_CREDENTIALS_APPROVED=$(read_setting PRICE_SOURCE_CREDENTIALS_APPROVED)
  PRICE_SOURCE_LICENSE_APPROVED=$(read_setting PRICE_SOURCE_LICENSE_APPROVED)
  PRODUCTION_CATEGORY_CODE=$(read_setting PRODUCTION_CATEGORY_CODE)
  PRODUCTION_MARKET=$(read_setting PRODUCTION_MARKET)
  PRODUCTION_PRICE_SYMBOL=$(read_setting PRODUCTION_PRICE_SYMBOL)
  PRODUCTION_PERPETUAL_SYMBOL=$(read_setting PRODUCTION_PERPETUAL_SYMBOL)
  PRODUCTION_DELIVERY_SYMBOL=$(read_setting PRODUCTION_DELIVERY_SYMBOL)
  DELIVERY_ALGORITHM=$(read_setting DELIVERY_ALGORITHM)
  DELIVERY_SOURCE_WEIGHTS=$(read_setting DELIVERY_SOURCE_WEIGHTS)
  DELIVERY_MAX_DEVIATION_BPS=$(read_setting DELIVERY_MAX_DEVIATION_BPS)
  DELIVERY_LOCK_WINDOW_MS=$(read_setting DELIVERY_LOCK_WINDOW_MS)
  DELIVERY_FORMULA_VERSION=$(read_setting DELIVERY_FORMULA_VERSION)
  HISTORICAL_REPLAY_REPORT=$(read_setting HISTORICAL_REPLAY_REPORT)
  HISTORICAL_REPLAY_REPORT_SHA256=$(read_setting HISTORICAL_REPLAY_REPORT_SHA256)
  ALERT_PLATFORM=$(read_setting ALERT_PLATFORM)
  ALERT_ONCALL_TEAM=$(read_setting ALERT_ONCALL_TEAM)
  ALERT_ESCALATION_POLICY=$(read_setting ALERT_ESCALATION_POLICY)
  ALERT_TEST_REPORT=$(read_setting ALERT_TEST_REPORT)
  ALERT_TEST_REPORT_SHA256=$(read_setting ALERT_TEST_REPORT_SHA256)
  PRODUCTION_TENANT_ID=$(read_setting PRODUCTION_TENANT_ID)
  PRODUCTION_SETTLEMENT_COIN=$(read_setting PRODUCTION_SETTLEMENT_COIN)
  FUND_ACCOUNT_PERMISSION_APPROVED=$(read_setting FUND_ACCOUNT_PERMISSION_APPROVED)
  FUND_ACCOUNT_APPROVER=$(read_setting FUND_ACCOUNT_APPROVER)
  LIQUIDATION_ENABLE_WINDOW=$(read_setting LIQUIDATION_ENABLE_WINDOW)
  LIQUIDATION_ROLLBACK_PLAN=$(read_setting LIQUIDATION_ROLLBACK_PLAN)
  LIQUIDATION_ROLLBACK_PLAN_SHA256=$(read_setting LIQUIDATION_ROLLBACK_PLAN_SHA256)
  DR_RPO_MINUTES=$(read_setting DR_RPO_MINUTES)
  DR_RTO_MINUTES=$(read_setting DR_RTO_MINUTES)
  DR_BACKUP_ENCRYPTION=$(read_setting DR_BACKUP_ENCRYPTION)
  DR_OFFSITE_LOCATION=$(read_setting DR_OFFSITE_LOCATION)
  DR_EXERCISE_REPORT=$(read_setting DR_EXERCISE_REPORT)
  DR_EXERCISE_REPORT_SHA256=$(read_setting DR_EXERCISE_REPORT_SHA256)
else
  PRODUCTION_PRICE_SOURCE_IDS=""
  PRICE_SOURCE_CREDENTIALS_APPROVED=""
  PRICE_SOURCE_LICENSE_APPROVED=""
  PRODUCTION_CATEGORY_CODE=""
  PRODUCTION_MARKET=""
  PRODUCTION_PRICE_SYMBOL=""
  PRODUCTION_PERPETUAL_SYMBOL=""
  PRODUCTION_DELIVERY_SYMBOL=""
  DELIVERY_ALGORITHM=""
  DELIVERY_SOURCE_WEIGHTS=""
  DELIVERY_MAX_DEVIATION_BPS=""
  DELIVERY_LOCK_WINDOW_MS=""
  DELIVERY_FORMULA_VERSION=""
  HISTORICAL_REPLAY_REPORT=""
  HISTORICAL_REPLAY_REPORT_SHA256=""
  ALERT_PLATFORM=""
  ALERT_ONCALL_TEAM=""
  ALERT_ESCALATION_POLICY=""
  ALERT_TEST_REPORT=""
  ALERT_TEST_REPORT_SHA256=""
  PRODUCTION_TENANT_ID=""
  PRODUCTION_SETTLEMENT_COIN=""
  FUND_ACCOUNT_PERMISSION_APPROVED=""
  FUND_ACCOUNT_APPROVER=""
  LIQUIDATION_ENABLE_WINDOW=""
  LIQUIDATION_ROLLBACK_PLAN=""
  LIQUIDATION_ROLLBACK_PLAN_SHA256=""
  DR_RPO_MINUTES=""
  DR_RTO_MINUTES=""
  DR_BACKUP_ENCRYPTION=""
  DR_OFFSITE_LOCATION=""
  DR_EXERCISE_REPORT=""
  DR_EXERCISE_REPORT_SHA256=""
fi

source_count=$(
  printf '%s\n' "$PRODUCTION_PRICE_SOURCE_IDS" |
    awk -F, '{
      for (i = 1; i <= NF; i++) {
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", $i)
        if ($i != "") unique[$i] = 1
      }
    } END {
      for (source in unique) count++
      print count + 0
    }'
)
source_ids_valid=true
seen_sources=","
old_ifs=$IFS
IFS=,
for source_id in $PRODUCTION_PRICE_SOURCE_IDS; do
  source_id=$(printf '%s\n' "$source_id" | awk '{$1=$1; print}')
  if ! valid_token "$source_id"; then
    source_ids_valid=false
    continue
  fi
  case "$seen_sources" in
    *,"$source_id",*)
      continue
      ;;
  esac
  seen_sources="${seen_sources}${source_id},"
done
IFS=$old_ifs

if [ "$source_count" -ge 3 ] && [ "$source_ids_valid" = "true" ]; then
  pass "at least three declared independent production price sources"
else
  fail "at least three declared independent production price sources"
fi

require_true "$PRICE_SOURCE_CREDENTIALS_APPROVED" "price-source credentials approved"
require_true "$PRICE_SOURCE_LICENSE_APPROVED" "price-source data licenses approved"
require_value "$PRODUCTION_CATEGORY_CODE" "production category code declared"
require_value "$PRODUCTION_MARKET" "production market declared"
require_value "$PRODUCTION_PRICE_SYMBOL" "production price symbol declared"
require_value "$PRODUCTION_PERPETUAL_SYMBOL" "production perpetual symbol declared"
require_value "$PRODUCTION_DELIVERY_SYMBOL" "production delivery symbol declared"

case "$DELIVERY_ALGORITHM" in
  WEIGHTED_MEAN)
    DELIVERY_ALGORITHM_NUMBER=1
    pass "DELIVERY algorithm approved"
    ;;
  MEDIAN)
    DELIVERY_ALGORITHM_NUMBER=2
    pass "DELIVERY algorithm approved"
    ;;
  *)
    DELIVERY_ALGORITHM_NUMBER=0
    fail "DELIVERY algorithm approved (MEDIAN or WEIGHTED_MEAN)"
    ;;
esac
weights_valid=$(
  printf '%s\n' "$DELIVERY_SOURCE_WEIGHTS" |
    awk -F, -v expected="$source_count" '
      BEGIN { valid = (expected >= 3) }
      {
        if ($0 == "") valid = 0
        if (NF != expected) valid = 0
        for (i = 1; i <= NF; i++) {
          gsub(/^[[:space:]]+|[[:space:]]+$/, "", $i)
          if ($i !~ /^[0-9]+([.][0-9]+)?$/ || $i + 0 <= 0) valid = 0
        }
      }
      END { print valid }
    '
)
if [ "$weights_valid" -eq 1 ]; then
  pass "DELIVERY positive source weights match declared sources"
else
  fail "DELIVERY positive source weights match declared sources"
fi
require_positive_integer "$DELIVERY_MAX_DEVIATION_BPS" "DELIVERY maximum deviation approved"
case "$DELIVERY_LOCK_WINDOW_MS" in
  ''|*[!0-9]*|0)
    fail "DELIVERY lock window approved (must be positive whole seconds)"
    ;;
  *)
    if [ $((DELIVERY_LOCK_WINDOW_MS % 1000)) -eq 0 ]; then
      pass "DELIVERY lock window approved"
    else
      fail "DELIVERY lock window approved (must be positive whole seconds)"
    fi
    ;;
esac
require_value "$DELIVERY_FORMULA_VERSION" "DELIVERY immutable formula version declared"
require_evidence_file "$HISTORICAL_REPLAY_REPORT" "$HISTORICAL_REPLAY_REPORT_SHA256" "production historical replay report attached"

require_value "$ALERT_PLATFORM" "production alert platform declared"
require_value "$ALERT_ONCALL_TEAM" "production on-call team declared"
require_value "$ALERT_ESCALATION_POLICY" "production alert escalation policy declared"
require_evidence_file "$ALERT_TEST_REPORT" "$ALERT_TEST_REPORT_SHA256" "production alert delivery test report attached"

require_positive_integer "$PRODUCTION_TENANT_ID" "production tenant declared"
require_value "$PRODUCTION_SETTLEMENT_COIN" "production settlement coin declared"
require_true "$FUND_ACCOUNT_PERMISSION_APPROVED" "insurance and fee account permissions approved"
require_value "$FUND_ACCOUNT_APPROVER" "fund-account approver declared"
require_value "$LIQUIDATION_ENABLE_WINDOW" "automatic-liquidation enable window declared"
require_evidence_file "$LIQUIDATION_ROLLBACK_PLAN" "$LIQUIDATION_ROLLBACK_PLAN_SHA256" "automatic-liquidation rollback plan attached"

require_positive_integer "$DR_RPO_MINUTES" "production RPO approved"
require_positive_integer "$DR_RTO_MINUTES" "production RTO approved"
require_value "$DR_BACKUP_ENCRYPTION" "backup encryption declared"
require_value "$DR_OFFSITE_LOCATION" "offsite backup location declared"
require_evidence_file "$DR_EXERCISE_REPORT" "$DR_EXERCISE_REPORT_SHA256" "production DR exercise report attached"

printf '\nLive deploy checks\n'
if docker compose -f "$COMPOSE_FILE" ps >/dev/null 2>&1; then
  for service_name in mysql etcd itick-rpc trade-rpc asset-rpc system-rpc; do
    check_service "$service_name"
  done
  for service_name in db-init config-seed kafka-init; do
    check_completed_service "$service_name"
  done
else
  fail "Compose runtime can be inspected"
fi

trade_config=$(
  docker compose -f "$COMPOSE_FILE" exec -T etcd \
    etcdctl --endpoints=http://127.0.0.1:2379 \
    get /wklive/trade-rpc/config --print-value-only 2>/dev/null || true
)
automatic_liquidation=$(
  printf '%s\n' "$trade_config" |
    awk '/^AutomaticLiquidation:/{getline; print $2; exit}'
)
cross_margin=$(
  printf '%s\n' "$trade_config" |
    awk '/^CrossMarginTrading:/{getline; print $2; exit}'
)
if [ "$automatic_liquidation" = "false" ]; then
  pass "AutomaticLiquidation remains closed during preflight"
else
  fail "AutomaticLiquidation remains closed during preflight"
fi
if [ "$cross_margin" = "false" ]; then
  pass "CrossMarginTrading remains closed during preflight"
else
  fail "CrossMarginTrading remains closed during preflight"
fi

tokens_valid=true
for token in "$PRODUCTION_CATEGORY_CODE" "$PRODUCTION_MARKET" "$PRODUCTION_PRICE_SYMBOL" \
  "$PRODUCTION_PERPETUAL_SYMBOL" "$PRODUCTION_DELIVERY_SYMBOL" "$PRODUCTION_SETTLEMENT_COIN" \
  "$DELIVERY_FORMULA_VERSION"; do
  if ! valid_token "$token"; then
    tokens_valid=false
  fi
done
if [ "$source_ids_valid" != "true" ] || [ "$source_count" -lt 3 ]; then
  tokens_valid=false
fi
case "$PRODUCTION_TENANT_ID" in
  ''|*[!0-9]*|0)
    tokens_valid=false
    ;;
esac
for numeric_value in "$DELIVERY_MAX_DEVIATION_BPS" "$DELIVERY_LOCK_WINDOW_MS"; do
  case "$numeric_value" in
    ''|*[!0-9]*|0)
      tokens_valid=false
      ;;
  esac
done
if [ "$DELIVERY_ALGORITHM_NUMBER" -eq 0 ]; then
  tokens_valid=false
fi
if [ "$weights_valid" -ne 1 ]; then
  tokens_valid=false
fi
case "$DELIVERY_LOCK_WINDOW_MS" in
  ''|*[!0-9]*)
    tokens_valid=false
    ;;
  *)
    if [ $((DELIVERY_LOCK_WINDOW_MS % 1000)) -ne 0 ]; then
      tokens_valid=false
    fi
    ;;
esac

if [ "$tokens_valid" = "true" ]; then
  readiness_detailed=true
else
  readiness_detailed=false
fi

db_output=$(
  docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps \
    --entrypoint /usr/local/bin/contract-readiness \
    -e "READINESS_DETAILED=$readiness_detailed" \
    -e "READINESS_SOURCE_AUTHORITIES=$PRODUCTION_PRICE_SOURCE_IDS" \
    -e "READINESS_SOURCE_WEIGHTS=$DELIVERY_SOURCE_WEIGHTS" \
    -e "READINESS_CATEGORY_CODE=$PRODUCTION_CATEGORY_CODE" \
    -e "READINESS_MARKET=$PRODUCTION_MARKET" \
    -e "READINESS_PRICE_SYMBOL=$PRODUCTION_PRICE_SYMBOL" \
    -e "READINESS_PERPETUAL_SYMBOL=$PRODUCTION_PERPETUAL_SYMBOL" \
    -e "READINESS_DELIVERY_SYMBOL=$PRODUCTION_DELIVERY_SYMBOL" \
    -e "READINESS_TENANT_ID=$PRODUCTION_TENANT_ID" \
    -e "READINESS_SETTLEMENT_COIN=$PRODUCTION_SETTLEMENT_COIN" \
    -e "READINESS_DELIVERY_ALGORITHM=$DELIVERY_ALGORITHM_NUMBER" \
    -e "READINESS_FORMULA_VERSION=$DELIVERY_FORMULA_VERSION" \
    -e "READINESS_MAX_LOOKBACK_MS=$DELIVERY_LOCK_WINDOW_MS" \
    -e "READINESS_MAX_DEVIATION_BPS=$DELIVERY_MAX_DEVIATION_BPS" \
    db-init
)
db_status=$?
if [ "$db_status" -eq 0 ]; then
  pass "model-backed database readiness inspection succeeded"
else
  fail "model-backed database readiness inspection succeeded"
  db_output=""
fi

if [ "$tokens_valid" = "true" ]; then
  if [ "$(db_value READINESS_DB_DETAILED)" = "true" ]; then
    pass "price, contract, and fund configuration can be inspected"
  else
    fail "price, contract, and fund configuration can be inspected"
  fi
  if [ "$(db_number READINESS_DB_ACTIVE_SOURCE_AUTHORITIES 0)" -eq "$source_count" ]; then pass "declared source authorities are active for FINAL_QUOTE"; else fail "declared source authorities are active for FINAL_QUOTE"; fi
  if [ "$(db_number READINESS_DB_PRICE_ENGINE_AUTHORITY 0)" -gt 0 ]; then pass "price-engine authority can publish all contract price kinds"; else fail "price-engine authority can publish all contract price kinds"; fi
  if [ "$(db_number READINESS_DB_INDEX_FORMULAS 0)" -gt 0 ]; then pass "active three-source INDEX formula"; else fail "active three-source INDEX formula"; fi
  if [ "$(db_number READINESS_DB_MARK_FORMULAS 0)" -gt 0 ]; then pass "active INDEX_BASIS MARK formula"; else fail "active INDEX_BASIS MARK formula"; fi
  if [ "$(db_number READINESS_DB_FUNDING_FORMULAS 0)" -gt 0 ]; then pass "active FUNDING formula"; else fail "active FUNDING formula"; fi
  if [ "$(db_number READINESS_DB_DELIVERY_FORMULAS 0)" -gt 0 ]; then pass "active three-source DELIVERY formula"; else fail "active three-source DELIVERY formula"; fi

  if [ "$(db_number READINESS_DB_FRESH_SOURCES 0)" -ge 3 ]; then pass "three declared FINAL_QUOTE sources are fresh"; else fail "three declared FINAL_QUOTE sources are fresh"; fi
  if [ "$(db_number READINESS_DB_FRESH_OUTPUT_KINDS 0)" -eq 4 ]; then pass "INDEX/MARK/FUNDING/DELIVERY outputs are fresh"; else fail "INDEX/MARK/FUNDING/DELIVERY outputs are fresh"; fi

  if [ "$(db_number READINESS_DB_INSURANCE_FUNDS 0)" -gt 0 ]; then pass "funded active INSURANCE_FUND account"; else fail "funded active INSURANCE_FUND account"; fi
  if [ "$(db_number READINESS_DB_FEE_REVENUE 0)" -gt 0 ]; then pass "active FEE_REVENUE account"; else fail "active FEE_REVENUE account"; fi

  if [ "$(db_number READINESS_DB_PERPETUAL_CONTRACTS 0)" -gt 0 ]; then pass "enabled production perpetual contract configuration"; else fail "enabled production perpetual contract configuration"; fi
  if [ "$(db_number READINESS_DB_DELIVERY_CONTRACTS 0)" -gt 0 ]; then pass "enabled future production delivery contract configuration"; else fail "enabled future production delivery contract configuration"; fi
  if [ "$(db_number READINESS_DB_INSURANCE_CONFIGS 0)" -gt 0 ]; then pass "active default contract insurance configuration"; else fail "active default contract insurance configuration"; fi
else
  fail "live formula and fund-account dimensions are valid"
fi

if [ "$db_status" -eq 0 ]; then
  pass "operational backlog can be inspected"
else
  fail "operational backlog can be inspected"
fi
if [ "$(db_number READINESS_DB_UNHEALTHY_OUTBOX 1)" -eq 0 ]; then
  pass "Snapshot Outbox is healthy (fresh in-flight records are allowed)"
else
  fail "Snapshot Outbox is healthy (failed, manual, or older than 60 seconds)"
fi
if [ "$(db_number READINESS_DB_OPEN_RECONCILIATION 1)" -eq 0 ]; then pass "contract reconciliation has no open issue"; else fail "contract reconciliation has no open issue"; fi
if [ "$(db_number READINESS_DB_OPEN_SETTLEMENT 1)" -eq 0 ]; then pass "settlement instructions have no open record"; else fail "settlement instructions have no open record"; fi

printf '\n'
if [ "$failures" -eq 0 ]; then
  printf 'READY: all prerequisites passed; enabling production gates still requires the approved release procedure.\n'
  exit 0
fi

printf 'NOT READY: %s prerequisite(s) failed. Production gates must remain closed.\n' "$failures"
exit 1
