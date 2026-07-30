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

is_positive_decimal() {
  decimal_value="$1"
  LC_ALL=C awk -v value="$decimal_value" '
    BEGIN {
      if (value !~ /^[0-9]+([.][0-9]+)?$/ || value + 0 <= 0) {
        exit 1
      }
      split(value, parts, ".")
      integer = parts[1]
      sub(/^0+/, "", integer)
      if (integer == "") integer = "0"
      if (length(integer) > 18 || length(parts[2]) > 18) {
        exit 1
      }
    }
  '
}

require_positive_decimal() {
  variable_value="$1"
  description="$2"
  if is_positive_decimal "$variable_value"; then
    pass "$description"
  else
    fail "$description (must be a positive DECIMAL(36,18))"
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
  PRODUCTION_PRICE_SOURCE_MARKETS=$(read_setting PRODUCTION_PRICE_SOURCE_MARKETS)
  PRICE_SOURCE_CREDENTIALS_APPROVED=$(read_setting PRICE_SOURCE_CREDENTIALS_APPROVED)
  PRICE_SOURCE_ACCESS_MODE=$(read_setting PRICE_SOURCE_ACCESS_MODE)
  PRICE_SOURCE_LICENSE_APPROVED=$(read_setting PRICE_SOURCE_LICENSE_APPROVED)
  PRODUCTION_CATEGORY_CODE=$(read_setting PRODUCTION_CATEGORY_CODE)
  PRODUCTION_MARKET=$(read_setting PRODUCTION_MARKET)
  PRODUCTION_PRICE_SYMBOL=$(read_setting PRODUCTION_PRICE_SYMBOL)
  PRODUCTION_PERPETUAL_SYMBOL=$(read_setting PRODUCTION_PERPETUAL_SYMBOL)
  PRODUCTION_DELIVERY_SYMBOL=$(read_setting PRODUCTION_DELIVERY_SYMBOL)
  INDEX_ALGORITHM=$(read_setting INDEX_ALGORITHM)
  INDEX_SOURCE_WEIGHTS=$(read_setting INDEX_SOURCE_WEIGHTS)
  INDEX_MAX_DEVIATION_BPS=$(read_setting INDEX_MAX_DEVIATION_BPS)
  INDEX_FORMULA_VERSION=$(read_setting INDEX_FORMULA_VERSION)
  PERPETUAL_PRICE_AUTHORITY=$(read_setting PERPETUAL_PRICE_AUTHORITY)
  PERPETUAL_PRICE_MARKET=$(read_setting PERPETUAL_PRICE_MARKET)
  MARK_FORMULA_VERSION=$(read_setting MARK_FORMULA_VERSION)
  MARK_MAX_BASIS_BPS=$(read_setting MARK_MAX_BASIS_BPS)
  MARK_CURRENT_WEIGHT=$(read_setting MARK_CURRENT_WEIGHT)
  MARK_PREVIOUS_WEIGHT=$(read_setting MARK_PREVIOUS_WEIGHT)
  FUNDING_FORMULA_VERSION=$(read_setting FUNDING_FORMULA_VERSION)
  PRICE_FORMULA_INTERVAL_MS=$(read_setting PRICE_FORMULA_INTERVAL_MS)
  DELIVERY_ALGORITHM=$(read_setting DELIVERY_ALGORITHM)
  DELIVERY_SOURCE_WEIGHTS=$(read_setting DELIVERY_SOURCE_WEIGHTS)
  DELIVERY_MAX_DEVIATION_BPS=$(read_setting DELIVERY_MAX_DEVIATION_BPS)
  DELIVERY_LOCK_WINDOW_MS=$(read_setting DELIVERY_LOCK_WINDOW_MS)
  DELIVERY_FORMULA_VERSION=$(read_setting DELIVERY_FORMULA_VERSION)
  HISTORICAL_REPLAY_REPORT=$(read_setting HISTORICAL_REPLAY_REPORT)
  HISTORICAL_REPLAY_REPORT_SHA256=$(read_setting HISTORICAL_REPLAY_REPORT_SHA256)
  HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF=$(read_setting HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF)
  ALERT_PLATFORM=$(read_setting ALERT_PLATFORM)
  ALERT_ONCALL_TEAM=$(read_setting ALERT_ONCALL_TEAM)
  ALERT_ESCALATION_POLICY=$(read_setting ALERT_ESCALATION_POLICY)
  ALERT_TEST_REPORT=$(read_setting ALERT_TEST_REPORT)
  ALERT_TEST_REPORT_SHA256=$(read_setting ALERT_TEST_REPORT_SHA256)
  ALERT_TEST_PRODUCTION_APPROVAL_REF=$(read_setting ALERT_TEST_PRODUCTION_APPROVAL_REF)
  CONTRACT_ONCALL_ACCOUNT=$(read_setting CONTRACT_ONCALL_ACCOUNT)
  INSURANCE_OPERATOR_ACCOUNT=$(read_setting INSURANCE_OPERATOR_ACCOUNT)
  DR_OPERATOR_ACCOUNT=$(read_setting DR_OPERATOR_ACCOUNT)
  DELIVERY_OPERATOR_ACCOUNT=$(read_setting DELIVERY_OPERATOR_ACCOUNT)
  PRODUCTION_REVIEWER_ACCOUNT=$(read_setting PRODUCTION_REVIEWER_ACCOUNT)
  PRODUCTION_APPROVER_ACCOUNT=$(read_setting PRODUCTION_APPROVER_ACCOUNT)
  PRODUCTION_TENANT_ID=$(read_setting PRODUCTION_TENANT_ID)
  PRODUCTION_SETTLEMENT_COIN=$(read_setting PRODUCTION_SETTLEMENT_COIN)
  INSURANCE_FUND_MIN_AVAILABLE=$(read_setting INSURANCE_FUND_MIN_AVAILABLE)
  FUND_ACCOUNT_PERMISSION_APPROVED=$(read_setting FUND_ACCOUNT_PERMISSION_APPROVED)
  FUND_ACCOUNT_APPROVER=$(read_setting FUND_ACCOUNT_APPROVER)
  LIQUIDATION_ENABLE_WINDOW=$(read_setting LIQUIDATION_ENABLE_WINDOW)
  LIQUIDATION_ROLLBACK_PLAN=$(read_setting LIQUIDATION_ROLLBACK_PLAN)
  LIQUIDATION_ROLLBACK_PLAN_SHA256=$(read_setting LIQUIDATION_ROLLBACK_PLAN_SHA256)
  LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF=$(read_setting LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF)
  DR_RPO_MINUTES=$(read_setting DR_RPO_MINUTES)
  DR_RTO_MINUTES=$(read_setting DR_RTO_MINUTES)
  DR_BACKUP_ENCRYPTION=$(read_setting DR_BACKUP_ENCRYPTION)
  DR_OFFSITE_LOCATION=$(read_setting DR_OFFSITE_LOCATION)
  DR_EXERCISE_REPORT=$(read_setting DR_EXERCISE_REPORT)
  DR_EXERCISE_REPORT_SHA256=$(read_setting DR_EXERCISE_REPORT_SHA256)
  DR_EXERCISE_PRODUCTION_APPROVAL_REF=$(read_setting DR_EXERCISE_PRODUCTION_APPROVAL_REF)
else
  PRODUCTION_PRICE_SOURCE_IDS=""
  PRODUCTION_PRICE_SOURCE_MARKETS=""
  PRICE_SOURCE_CREDENTIALS_APPROVED=""
  PRICE_SOURCE_ACCESS_MODE=""
  PRICE_SOURCE_LICENSE_APPROVED=""
  PRODUCTION_CATEGORY_CODE=""
  PRODUCTION_MARKET=""
  PRODUCTION_PRICE_SYMBOL=""
  PRODUCTION_PERPETUAL_SYMBOL=""
  PRODUCTION_DELIVERY_SYMBOL=""
  INDEX_ALGORITHM=""
  INDEX_SOURCE_WEIGHTS=""
  INDEX_MAX_DEVIATION_BPS=""
  INDEX_FORMULA_VERSION=""
  PERPETUAL_PRICE_AUTHORITY=""
  PERPETUAL_PRICE_MARKET=""
  MARK_FORMULA_VERSION=""
  MARK_MAX_BASIS_BPS=""
  MARK_CURRENT_WEIGHT=""
  MARK_PREVIOUS_WEIGHT=""
  FUNDING_FORMULA_VERSION=""
  PRICE_FORMULA_INTERVAL_MS=""
  DELIVERY_ALGORITHM=""
  DELIVERY_SOURCE_WEIGHTS=""
  DELIVERY_MAX_DEVIATION_BPS=""
  DELIVERY_LOCK_WINDOW_MS=""
  DELIVERY_FORMULA_VERSION=""
  HISTORICAL_REPLAY_REPORT=""
  HISTORICAL_REPLAY_REPORT_SHA256=""
  HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF=""
  ALERT_PLATFORM=""
  ALERT_ONCALL_TEAM=""
  ALERT_ESCALATION_POLICY=""
  ALERT_TEST_REPORT=""
  ALERT_TEST_REPORT_SHA256=""
  ALERT_TEST_PRODUCTION_APPROVAL_REF=""
  CONTRACT_ONCALL_ACCOUNT=""
  INSURANCE_OPERATOR_ACCOUNT=""
  DR_OPERATOR_ACCOUNT=""
  DELIVERY_OPERATOR_ACCOUNT=""
  PRODUCTION_REVIEWER_ACCOUNT=""
  PRODUCTION_APPROVER_ACCOUNT=""
  PRODUCTION_TENANT_ID=""
  PRODUCTION_SETTLEMENT_COIN=""
  INSURANCE_FUND_MIN_AVAILABLE=""
  FUND_ACCOUNT_PERMISSION_APPROVED=""
  FUND_ACCOUNT_APPROVER=""
  LIQUIDATION_ENABLE_WINDOW=""
  LIQUIDATION_ROLLBACK_PLAN=""
  LIQUIDATION_ROLLBACK_PLAN_SHA256=""
  LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF=""
  DR_RPO_MINUTES=""
  DR_RTO_MINUTES=""
  DR_BACKUP_ENCRYPTION=""
  DR_OFFSITE_LOCATION=""
  DR_EXERCISE_REPORT=""
  DR_EXERCISE_REPORT_SHA256=""
  DR_EXERCISE_PRODUCTION_APPROVAL_REF=""
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

credentials_check_deferred=false
if [ "$PRICE_SOURCE_CREDENTIALS_APPROVED" = "true" ]; then
  pass "price-source credentials approved"
elif [ "$PRICE_SOURCE_ACCESS_MODE" = "PUBLIC_NO_CREDENTIALS" ]; then
  # The database-backed check below must prove every declared Authority is a
  # PUBLIC_REST producer before this factual no-credential mode can pass.
  credentials_check_deferred=true
else
  fail "price-source credentials approved or explicitly not required"
fi
require_true "$PRICE_SOURCE_LICENSE_APPROVED" "price-source data licenses approved"
require_value "$PRODUCTION_CATEGORY_CODE" "production category code declared"
require_value "$PRODUCTION_MARKET" "production market declared"
require_value "$PRODUCTION_PRICE_SYMBOL" "production price symbol declared"
require_value "$PRODUCTION_PERPETUAL_SYMBOL" "production perpetual symbol declared"
require_value "$PRODUCTION_DELIVERY_SYMBOL" "production delivery symbol declared"

source_market_count=$(
  printf '%s\n' "$PRODUCTION_PRICE_SOURCE_MARKETS" |
    awk -F, '{
      for (i = 1; i <= NF; i++) {
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", $i)
        if ($i != "") count++
      }
    } END { print count + 0 }'
)
source_markets_valid=true
old_ifs=$IFS
IFS=,
for source_market in $PRODUCTION_PRICE_SOURCE_MARKETS; do
  source_market=$(printf '%s\n' "$source_market" | awk '{$1=$1; print}')
  if ! valid_token "$source_market"; then
    source_markets_valid=false
  fi
done
IFS=$old_ifs
if [ "$source_market_count" -eq "$source_count" ] && [ "$source_markets_valid" = "true" ]; then
  pass "source markets match declared source authorities"
else
  fail "source markets match declared source authorities"
fi

case "$INDEX_ALGORITHM" in
  WEIGHTED_MEAN)
    INDEX_ALGORITHM_NUMBER=1
    pass "INDEX algorithm approved"
    ;;
  MEDIAN)
    INDEX_ALGORITHM_NUMBER=2
    pass "INDEX algorithm approved"
    ;;
  *)
    INDEX_ALGORITHM_NUMBER=0
    fail "INDEX algorithm approved (MEDIAN or WEIGHTED_MEAN)"
    ;;
esac
index_weights_valid=$(
  printf '%s\n' "$INDEX_SOURCE_WEIGHTS" |
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
if [ "$index_weights_valid" -eq 1 ]; then
  pass "INDEX positive source weights match declared sources"
else
  fail "INDEX positive source weights match declared sources"
fi
require_positive_integer "$INDEX_MAX_DEVIATION_BPS" "INDEX maximum deviation approved"
require_value "$INDEX_FORMULA_VERSION" "INDEX immutable formula version declared"
require_value "$PERPETUAL_PRICE_AUTHORITY" "MARK perpetual source authority declared"
require_value "$PERPETUAL_PRICE_MARKET" "MARK perpetual source market declared"
require_value "$MARK_FORMULA_VERSION" "MARK immutable formula version declared"
require_positive_integer "$MARK_MAX_BASIS_BPS" "MARK maximum basis approved"
mark_weights_valid=$(
  printf '%s,%s\n' "$MARK_CURRENT_WEIGHT" "$MARK_PREVIOUS_WEIGHT" |
    awk -F, '
      BEGIN { valid = 1 }
      {
        if (NF != 2) valid = 0
        for (i = 1; i <= NF; i++) {
          gsub(/^[[:space:]]+|[[:space:]]+$/, "", $i)
          if ($i !~ /^[0-9]+([.][0-9]+)?$/ || $i + 0 <= 0) valid = 0
        }
      }
      END { print valid }
    '
)
if [ "$mark_weights_valid" -eq 1 ]; then
  pass "MARK current and previous weights approved"
else
  fail "MARK current and previous weights approved"
fi
require_value "$FUNDING_FORMULA_VERSION" "FUNDING immutable formula version declared"
require_positive_integer "$PRICE_FORMULA_INTERVAL_MS" "price formula interval approved"

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
require_value "$HISTORICAL_REPLAY_PRODUCTION_APPROVAL_REF" "historical replay production approval reference declared"

require_value "$ALERT_PLATFORM" "production alert platform declared"
require_value "$ALERT_ONCALL_TEAM" "production on-call team declared"
require_value "$ALERT_ESCALATION_POLICY" "production alert escalation policy declared"
require_evidence_file "$ALERT_TEST_REPORT" "$ALERT_TEST_REPORT_SHA256" "production alert delivery test report attached"
require_value "$ALERT_TEST_PRODUCTION_APPROVAL_REF" "alert delivery production approval reference declared"
require_value "$CONTRACT_ONCALL_ACCOUNT" "contract on-call system account declared"
require_value "$INSURANCE_OPERATOR_ACCOUNT" "insurance-fund operator system account declared"
require_value "$DR_OPERATOR_ACCOUNT" "DR operator system account declared"
require_value "$DELIVERY_OPERATOR_ACCOUNT" "delivery operator system account declared"
require_value "$PRODUCTION_REVIEWER_ACCOUNT" "production reviewer system account declared"
require_value "$PRODUCTION_APPROVER_ACCOUNT" "production approver system account declared"

require_positive_integer "$PRODUCTION_TENANT_ID" "production tenant declared"
require_value "$PRODUCTION_SETTLEMENT_COIN" "production settlement coin declared"
require_positive_decimal "$INSURANCE_FUND_MIN_AVAILABLE" "approved insurance-fund minimum available balance declared"
require_true "$FUND_ACCOUNT_PERMISSION_APPROVED" "insurance and fee account permissions approved"
require_value "$FUND_ACCOUNT_APPROVER" "fund-account approver declared"
require_value "$LIQUIDATION_ENABLE_WINDOW" "automatic-liquidation enable window declared"
require_evidence_file "$LIQUIDATION_ROLLBACK_PLAN" "$LIQUIDATION_ROLLBACK_PLAN_SHA256" "automatic-liquidation rollback plan attached"
require_value "$LIQUIDATION_ROLLBACK_PRODUCTION_APPROVAL_REF" "liquidation rollback production approval reference declared"

require_positive_integer "$DR_RPO_MINUTES" "production RPO approved"
require_positive_integer "$DR_RTO_MINUTES" "production RTO approved"
require_value "$DR_BACKUP_ENCRYPTION" "backup encryption declared"
require_value "$DR_OFFSITE_LOCATION" "offsite backup location declared"
require_evidence_file "$DR_EXERCISE_REPORT" "$DR_EXERCISE_REPORT_SHA256" "production DR exercise report attached"
require_value "$DR_EXERCISE_PRODUCTION_APPROVAL_REF" "DR exercise production approval reference declared"

printf '\nLive deploy checks\n'
if docker compose -f "$COMPOSE_FILE" ps >/dev/null 2>&1; then
  for service_name in mysql etcd market-rpc trade-rpc asset-rpc system-rpc; do
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
  "$INDEX_FORMULA_VERSION" "$PERPETUAL_PRICE_AUTHORITY" "$PERPETUAL_PRICE_MARKET" \
  "$MARK_FORMULA_VERSION" "$FUNDING_FORMULA_VERSION" "$DELIVERY_FORMULA_VERSION" \
  "$CONTRACT_ONCALL_ACCOUNT" "$INSURANCE_OPERATOR_ACCOUNT" "$DR_OPERATOR_ACCOUNT" \
  "$DELIVERY_OPERATOR_ACCOUNT" "$PRODUCTION_REVIEWER_ACCOUNT" "$PRODUCTION_APPROVER_ACCOUNT"; do
  if ! valid_token "$token"; then
    tokens_valid=false
  fi
done
if [ "$source_ids_valid" != "true" ] || [ "$source_count" -lt 3 ] ||
   [ "$source_markets_valid" != "true" ] || [ "$source_market_count" -ne "$source_count" ]; then
  tokens_valid=false
fi
case "$PRODUCTION_TENANT_ID" in
  ''|*[!0-9]*|0)
    tokens_valid=false
    ;;
esac
for numeric_value in "$INDEX_MAX_DEVIATION_BPS" "$MARK_MAX_BASIS_BPS" \
  "$PRICE_FORMULA_INTERVAL_MS" "$DELIVERY_MAX_DEVIATION_BPS" "$DELIVERY_LOCK_WINDOW_MS"; do
  case "$numeric_value" in
    ''|*[!0-9]*|0)
      tokens_valid=false
      ;;
  esac
done
if [ "$INDEX_ALGORITHM_NUMBER" -eq 0 ] || [ "$DELIVERY_ALGORITHM_NUMBER" -eq 0 ]; then
  tokens_valid=false
fi
if [ "$index_weights_valid" -ne 1 ] || [ "$mark_weights_valid" -ne 1 ] ||
   [ "$weights_valid" -ne 1 ]; then
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
    -e "READINESS_SOURCE_MARKETS=$PRODUCTION_PRICE_SOURCE_MARKETS" \
    -e "READINESS_INDEX_SOURCE_WEIGHTS=$INDEX_SOURCE_WEIGHTS" \
    -e "READINESS_SOURCE_WEIGHTS=$DELIVERY_SOURCE_WEIGHTS" \
    -e "READINESS_CATEGORY_CODE=$PRODUCTION_CATEGORY_CODE" \
    -e "READINESS_MARKET=$PRODUCTION_MARKET" \
    -e "READINESS_PRICE_SYMBOL=$PRODUCTION_PRICE_SYMBOL" \
    -e "READINESS_PERPETUAL_SYMBOL=$PRODUCTION_PERPETUAL_SYMBOL" \
    -e "READINESS_DELIVERY_SYMBOL=$PRODUCTION_DELIVERY_SYMBOL" \
    -e "READINESS_PERPETUAL_PRICE_AUTHORITY=$PERPETUAL_PRICE_AUTHORITY" \
    -e "READINESS_PERPETUAL_PRICE_MARKET=$PERPETUAL_PRICE_MARKET" \
    -e "READINESS_TENANT_ID=$PRODUCTION_TENANT_ID" \
    -e "READINESS_SETTLEMENT_COIN=$PRODUCTION_SETTLEMENT_COIN" \
    -e "READINESS_INSURANCE_FUND_MIN_AVAILABLE=$INSURANCE_FUND_MIN_AVAILABLE" \
    -e "READINESS_CONTRACT_ONCALL_ACCOUNT=$CONTRACT_ONCALL_ACCOUNT" \
    -e "READINESS_INSURANCE_OPERATOR_ACCOUNT=$INSURANCE_OPERATOR_ACCOUNT" \
    -e "READINESS_DR_OPERATOR_ACCOUNT=$DR_OPERATOR_ACCOUNT" \
    -e "READINESS_DELIVERY_OPERATOR_ACCOUNT=$DELIVERY_OPERATOR_ACCOUNT" \
    -e "READINESS_PRODUCTION_REVIEWER_ACCOUNT=$PRODUCTION_REVIEWER_ACCOUNT" \
    -e "READINESS_PRODUCTION_APPROVER_ACCOUNT=$PRODUCTION_APPROVER_ACCOUNT" \
    -e "READINESS_INDEX_ALGORITHM=$INDEX_ALGORITHM_NUMBER" \
    -e "READINESS_INDEX_FORMULA_VERSION=$INDEX_FORMULA_VERSION" \
    -e "READINESS_INDEX_MAX_DEVIATION_BPS=$INDEX_MAX_DEVIATION_BPS" \
    -e "READINESS_MARK_FORMULA_VERSION=$MARK_FORMULA_VERSION" \
    -e "READINESS_MARK_MAX_BASIS_BPS=$MARK_MAX_BASIS_BPS" \
    -e "READINESS_MARK_CURRENT_WEIGHT=$MARK_CURRENT_WEIGHT" \
    -e "READINESS_MARK_PREVIOUS_WEIGHT=$MARK_PREVIOUS_WEIGHT" \
    -e "READINESS_FUNDING_FORMULA_VERSION=$FUNDING_FORMULA_VERSION" \
    -e "READINESS_FORMULA_INTERVAL_MS=$PRICE_FORMULA_INTERVAL_MS" \
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
  if [ "$(db_number READINESS_DB_ACTIVE_SOURCE_AUTHORITIES 0)" -eq "$source_count" ] &&
     [ "$(db_number READINESS_DB_DISTINCT_SOURCE_PROVIDERS 0)" -eq "$source_count" ]; then
    pass "declared source authorities are active and provider-independent for FINAL_QUOTE"
  else
    fail "declared source authorities are active and provider-independent for FINAL_QUOTE"
  fi
  if [ "$credentials_check_deferred" = "true" ]; then
    if [ "$(db_number READINESS_DB_PUBLIC_REST_SOURCE_AUTHORITIES 0)" -eq "$source_count" ]; then
      pass "declared price sources are PUBLIC_REST and require no credentials"
    else
      fail "PUBLIC_NO_CREDENTIALS requires every declared source to be PUBLIC_REST"
    fi
  fi
  if [ "$(db_number READINESS_DB_PRICE_ENGINE_AUTHORITY 0)" -gt 0 ]; then pass "price-engine authority can publish all contract price kinds"; else fail "price-engine authority can publish all contract price kinds"; fi
  if [ "$(db_number READINESS_DB_INDEX_FORMULAS 0)" -gt 0 ]; then pass "active three-source INDEX formula"; else fail "active three-source INDEX formula"; fi
  if [ "$(db_number READINESS_DB_MARK_FORMULAS 0)" -gt 0 ]; then pass "active INDEX_BASIS MARK formula"; else fail "active INDEX_BASIS MARK formula"; fi
  if [ "$(db_number READINESS_DB_FUNDING_FORMULAS 0)" -gt 0 ]; then pass "active FUNDING formula"; else fail "active FUNDING formula"; fi
  if [ "$(db_number READINESS_DB_DELIVERY_FORMULAS 0)" -gt 0 ]; then pass "active three-source DELIVERY formula"; else fail "active three-source DELIVERY formula"; fi

  if [ "$(db_number READINESS_DB_FRESH_SOURCES 0)" -ge 3 ]; then pass "three declared FINAL_QUOTE sources are fresh"; else fail "three declared FINAL_QUOTE sources are fresh"; fi
  if [ "$(db_number READINESS_DB_FRESH_OUTPUT_KINDS 0)" -eq 4 ]; then pass "INDEX/MARK/FUNDING/DELIVERY outputs are fresh"; else fail "INDEX/MARK/FUNDING/DELIVERY outputs are fresh"; fi

  if [ "$(db_number READINESS_DB_INSURANCE_FUNDS 0)" -gt 0 ]; then pass "active INSURANCE_FUND account meets approved minimum balance"; else fail "active INSURANCE_FUND account meets approved minimum balance"; fi
  if [ "$(db_number READINESS_DB_FEE_REVENUE 0)" -gt 0 ]; then pass "active FEE_REVENUE account"; else fail "active FEE_REVENUE account"; fi

  if [ "$(db_number READINESS_DB_PERPETUAL_CONTRACTS 0)" -gt 0 ]; then pass "enabled production perpetual contract configuration"; else fail "enabled production perpetual contract configuration"; fi
  if [ "$(db_number READINESS_DB_DELIVERY_CONTRACTS 0)" -gt 0 ]; then pass "enabled future production delivery contract configuration"; else fail "enabled future production delivery contract configuration"; fi
  if [ "$(db_number READINESS_DB_INSURANCE_CONFIGS 0)" -gt 0 ]; then pass "active default contract insurance configuration"; else fail "active default contract insurance configuration"; fi
  if [ "$(db_number READINESS_DB_CONTRACT_ONCALL_ACCOUNTS 0)" -eq 1 ]; then pass "contract on-call account has its exact role and read-only alert permissions"; else fail "contract on-call account has its exact role and read-only alert permissions"; fi
  if [ "$(db_number READINESS_DB_INSURANCE_OPERATORS 0)" -eq 1 ]; then pass "insurance-fund operator has only the three approved write menus"; else fail "insurance-fund operator has only the three approved write menus"; fi
  if [ "$(db_number READINESS_DB_DR_OPERATORS 0)" -eq 1 ]; then pass "DR operator has its exact role and no admin write permission"; else fail "DR operator has its exact role and no admin write permission"; fi
  if [ "$(db_number READINESS_DB_DELIVERY_OPERATORS 0)" -eq 1 ]; then pass "delivery operator has only the contract configuration write menu"; else fail "delivery operator has only the contract configuration write menu"; fi
  if [ "$(db_number READINESS_DB_PRODUCTION_REVIEWERS 0)" -eq 1 ]; then pass "production reviewer has its exact role and no admin write permission"; else fail "production reviewer has its exact role and no admin write permission"; fi
  if [ "$(db_number READINESS_DB_PRODUCTION_APPROVERS 0)" -eq 1 ]; then pass "production approver has its exact role and no admin write permission"; else fail "production approver has its exact role and no admin write permission"; fi
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
