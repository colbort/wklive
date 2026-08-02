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

check_sha256_manifest() {
  manifest_file="$1"
  description="$2"
  if [ ! -s "$manifest_file" ]; then
    fail "$description (manifest missing or empty)"
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    if (cd "$REPO_ROOT" && shasum -a 256 -c "$manifest_file" >/dev/null); then
      pass "$description"
    else
      fail "$description (hash mismatch)"
    fi
  elif command -v sha256sum >/dev/null 2>&1; then
    if (cd "$REPO_ROOT" && sha256sum -c "$manifest_file" >/dev/null); then
      pass "$description"
    else
      fail "$description (hash mismatch)"
    fi
  else
    fail "$description (SHA-256 utility unavailable)"
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
if [ "$operations_count" -eq 57 ]; then
  pass "57 Option operations alert rules are present"
else
  fail "57 Option operations alert rules are present (found $operations_count)"
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
if [ "$severity_count" -eq 61 ] && [ "$catalog_count" -eq 61 ]; then
  pass "all 61 Option alerts declare severity and catalog_id"
else
  fail "all 61 Option alerts declare severity and catalog_id (severity=$severity_count catalog=$catalog_count)"
fi
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'portfolio_liquidation_duplicate_open' \
  "operations metrics expose duplicate open portfolio liquidations"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'portfolio_liquidation_evidence_invalid' \
  "operations metrics expose invalid portfolio liquidation evidence"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'portfolio_liquidation_cancel_streak' \
  "operations metrics expose consecutive stale portfolio cancellations"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionPortfolioLiquidationDuplicateOpen' \
  "duplicate open portfolio liquidations trigger a critical alert"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionPortfolioLiquidationEvidenceInvalid' \
  "invalid portfolio liquidation evidence triggers a critical alert"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionPortfolioLiquidationCancelStreak' \
  "three stale portfolio cancellations trigger a critical alert"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'insurance_takeover_inventory' \
  "operations metrics expose open insurance takeover inventory"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'insurance_takeover_underlying_quantity' \
  "operations metrics expose insurance takeover underlying quantity"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'insurance_takeover_mark_value' \
  "operations metrics expose insurance takeover marked option value"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'insurance_takeover_abs_delta' \
  "operations metrics expose insurance takeover absolute Delta"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionInsuranceTakeoverInventoryOpen' \
  "open insurance takeover inventory triggers a dedicated alert"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionInsuranceTakeoverExpiryDue' \
  "insurance takeover inventory near expiry triggers a critical alert"
check_contains "$OPTION_DIR/monitoring/option-operations-alert-rules.yml" 'OptionInsuranceTakeoverMarketInvalid' \
  "unreliable insurance takeover valuation triggers a critical alert"

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
idx_option_physical_delivery_monitor
idx_option_liquidation_portfolio_monitor'
baseline_migration="$OPTION_DIR/migrations/20260731_zr_option_operations_monitoring_indexes.sql"
incremental_migration="$OPTION_DIR/migrations/20260731_zs_option_time_sensitive_monitoring_indexes.sql"
portfolio_liquidation_monitoring_migration="$OPTION_DIR/migrations/20260801_option_portfolio_liquidation_monitoring.sql"
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
    idx_option_liquidation_portfolio_monitor)
      migration_file="$portfolio_liquidation_monitoring_migration"
      ;;
  esac
  if ! grep -q "$index_name" $OPTION_SCHEMA_FILES ||
     ! grep -q "$index_name" "$migration_file"; then
    fail "monitoring index $index_name exists in schema and idempotent migration"
    index_ok=false
  fi
done
if [ "$index_ok" = "true" ]; then
  pass "all 18 Option monitoring indexes exist in schema and migration"
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

greeks_freshness_migration="$OPTION_DIR/migrations/20260731_zu_option_greeks_freshness.sql"
check_contains "$OPTION_DIR/option.sql" 'greeks_max_age_seconds.*DEFAULT 0' \
  "Option schema declares an explicit unconfigured Greeks threshold"
check_contains "$greeks_freshness_migration" 'ADD COLUMN greeks_max_age_seconds' \
  "Greeks freshness migration adds the contract threshold idempotently"
check_contains "$greeks_freshness_migration" 'NEW[.]greeks_max_age_seconds <= 0' \
  "database gate rejects unconfigured TRADING contracts"
check_contains "$OPTION_DIR/models/toptioncontractmodel_gen.go" 'GreeksMaxAgeSeconds.*greeks_max_age_seconds' \
  "make gen-model synchronized the Greeks threshold"
check_contains "$REPO_ROOT/proto/option/model.proto" 'greeks_max_age_seconds = 55' \
  "Option RPC contract exposes the Greeks threshold"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'greeks_snapshot_time < [?]-contract[.]greeks_max_age_seconds' \
  "operations metrics apply each contract Greeks threshold"

portfolio_order_config_migration="$OPTION_DIR/migrations/20260731_zv_option_order_portfolio_config.sql"
portfolio_config_lineage_migration="$OPTION_DIR/migrations/20260802_option_portfolio_risk_config_lineage.sql"
check_contains "$OPTION_DIR/option.sql" 'portfolio_risk_config_id.*DEFAULT 0' \
  "Option schema records order admission portfolio config ID"
check_contains "$OPTION_DIR/option.sql" 'portfolio_risk_config_version.*DEFAULT 0' \
  "Option schema records order admission portfolio config version"
check_contains "$portfolio_order_config_migration" 'trg_option_order_portfolio_config_insert' \
  "portfolio order evidence migration validates new orders"
check_contains "$portfolio_order_config_migration" 'trg_option_order_portfolio_config_update' \
  "portfolio order evidence migration protects immutable evidence"
check_contains "$portfolio_order_config_migration" 'config[.]effective_from<=NEW[.]create_times' \
  "portfolio order evidence is checked at order creation time"
check_contains "$OPTION_DIR/models/toptionordermodel_gen.go" 'PortfolioRiskConfigId.*portfolio_risk_config_id' \
  "make gen-model synchronized order portfolio config evidence"
check_contains "$REPO_ROOT/proto/option/model.proto" 'portfolio_risk_config_id = 34' \
  "Option RPC exposes order admission portfolio config evidence"
check_contains "$OPTION_DIR/internal/logic/app/placeorderlogic.go" 'order[.]PortfolioRiskConfigId = portfolioResult[.]configID' \
  "order admission persists the resolved portfolio config atomically"
check_contains "$OPTION_DIR/option.sql" 'source_config_id.*DEFAULT 0' \
  "portfolio risk config schema persists copied-source lineage"
check_contains "$portfolio_config_lineage_migration" 'NEW[.]effective_from <= NEW[.]create_times' \
  "portfolio config lineage migration rejects non-future drafts"
check_contains "$portfolio_config_lineage_migration" 'NEW[.]effective_from <= NEW[.]reviewed_at' \
  "portfolio config lineage migration rejects retroactive approval"
check_contains "$portfolio_config_lineage_migration" 'NEW[.]source_config_id <> OLD[.]source_config_id' \
  "portfolio config lineage migration protects immutable rollback provenance"
check_contains "$portfolio_config_lineage_migration" 'source[.]id = NEW[.]source_config_id' \
  "portfolio config lineage migration validates the copied source"
check_contains "$OPTION_DIR/models/toptionportfolioriskconfigmodel_gen.go" 'SourceConfigId.*source_config_id' \
  "make gen-model synchronized portfolio rollback provenance"
check_contains "$REPO_ROOT/proto/option/model.proto" 'source_config_id = 24' \
  "Option RPC exposes copied-source lineage"
portfolio_version_test="$OPTION_DIR/internal/logic/task/p1_portfolio_risk_version_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP1PortfolioRiskVersionGovernance' \
  "real Asset RPC gate invokes portfolio version switching and rollback"
check_contains "$portfolio_version_test" 'assertP1PortfolioPhase.*v1' \
  "portfolio version gate checks the V1 order and risk phase"
check_contains "$portfolio_version_test" 'SourceConfigId != v1[.]Id' \
  "portfolio rollback gate requires copied-source provenance"
check_contains "$portfolio_version_test" 'database allowed portfolio approval at or after effective boundary' \
  "portfolio version gate rejects retroactive approval at the database layer"
check_contains "$OPTION_DIR/models/toptionriskaccountmodel.go" 'cacheTOptionRiskAccountTenantIdUserIdAccountIdSettleCoinPrefix' \
  "first portfolio admission derives the wallet identity cache key"
check_contains "$OPTION_DIR/models/toptionriskaccountmodel.go" '[}], identityKey' \
  "first portfolio admission invalidates the wallet identity negative cache"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260802_option_portfolio_risk_config_lineage.sql' \
  "real Asset RPC gate installs portfolio lineage guards twice"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'portfolio_version_governance=' \
  "real Asset RPC gate reports portfolio version and rollback evidence"

assignment_pagination_migration="$OPTION_DIR/migrations/20260731_zw_option_assignment_pagination.sql"
check_contains "$OPTION_DIR/option.sql" 'idx_option_position_assignment_fifo.*tenant_id.*contract_id.*side.*status.*create_times.*id' \
  "Option schema indexes deterministic FIFO assignment pagination"
check_contains "$assignment_pagination_migration" 'idx_option_position_assignment_fifo' \
  "American assignment pagination index has an idempotent migration"
check_contains "$OPTION_DIR/models/toptionpositionmodel.go" 'FindAssignableShortsPage' \
  "American assignment model exposes bounded candidate pages"
check_contains "$OPTION_DIR/models/toptionpositionmodel.go" 'create_times > [?].*create_times = [?].*id > [?]' \
  "American assignment model uses a stable FIFO keyset cursor"
check_contains "$OPTION_DIR/models/toptionpositionmodel.go" 'FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate' \
  "position model exposes a scoped locking read for close-order mutations"
check_contains "$OPTION_DIR/models/toptionpositionmodel.go" 'LIMIT 1 FOR UPDATE' \
  "scoped position mutation read takes a database row lock"
check_contains "$OPTION_DIR/internal/logic/app/option_business_helpers.go" 'FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate' \
  "public close-order freeze and release use the position row lock"
check_contains "$OPTION_DIR/internal/logic/task/option_business_helpers.go" 'FindOneByTenantIdUserIdAccountIdContractIdSideForUpdate' \
  "assignment-side close-order release uses the position row lock"
check_contains "$OPTION_DIR/internal/logic/task/processexerciseslogic.go" 'const assignmentPageSize int64 = 500' \
  "American assignment processing bounds each FIFO candidate page"
check_contains "$OPTION_DIR/models/toptionassetinstructionmodel.go" 'AllSucceededByBizNo' \
  "American assignment completion uses an aggregate asset-instruction barrier"
check_contains "$OPTION_DIR/internal/logic/task/processassetinstructionslogic.go" 'AllSucceededByBizNo' \
  "American assignment completion avoids reloading every instruction per success"
check_contains "$OPTION_DIR/internal/logic/task/processassetinstructionslogic.go" 'item[.]StepNo < 2' \
  "American assignment only evaluates its full completion barrier on terminal steps"
check_contains "$OPTION_DIR/internal/logic/app/exerciselogic.go" 'allowsExerciseSubmission[(]currentContract[.]Status[)]' \
  "American exercise rechecks the listed lifecycle under a contract lock"
check_contains "$OPTION_DIR/internal/logic/app/setexerciseinstructionlogic.go" 'allowsExerciseSubmission[(]currentContract[.]Status[)]' \
  "expiry instructions recheck the listed lifecycle under a contract lock"
check_contains "$OPTION_DIR/internal/logic/task/processexerciseslogic.go" 'MaintenanceMargin[.]Sub[(]allocatedMaintenance[)]' \
  "American assignment reduces maintenance margin with assigned quantity"
exercise_governance_migration="$OPTION_DIR/migrations/20260730_zd_option_exercise_governance.sql"
check_contains "$exercise_governance_migration" 'status may only transition from active to superseded' \
  "database rejects exercise-instruction status rollback"
check_contains "$exercise_governance_migration" 'trg_option_exercise_instruction_no_delete' \
  "database preserves all exercise-instruction history"

margin_coin_migration="$OPTION_DIR/migrations/20260731_zx_option_margin_coin_evidence.sql"
check_contains "$margin_coin_migration" 'UPDATE `t_option_order` order_item' \
  "margin coin migration backfills deterministic order evidence"
check_contains "$margin_coin_migration" 'trg_option_order_margin_coin_insert' \
  "database rejects new orders with incorrect collateral coin"
check_contains "$margin_coin_migration" 'trg_option_margin_lot_coin_insert' \
  "database rejects new margin lots with incorrect collateral coin"
check_contains "$margin_coin_migration" 'trg_option_asset_instruction_coin_insert' \
  "database rejects asset instructions without a coin"
check_contains "$OPTION_DIR/internal/logic/app/app_order_match.go" 'return strings[.]TrimSpace[(]order[.]MarginCoin[)]' \
  "runtime never guesses order collateral from fee coin"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'margin_coin_invalid' \
  "operations metrics expose unresolved collateral coin evidence"

asset_freeze_logic="$REPO_ROOT/services/asset/internal/logic/asset/freezeassetlogic.go"
check_contains "$asset_freeze_logic" 'PrepareAssetIdempotent' \
  "Asset freeze uses the durable business idempotency ledger"
check_contains "$asset_freeze_logic" 'assetFreezeReplayMatches' \
  "Asset freeze rejects idempotency-key economic mismatches"
check_contains "$asset_freeze_logic" 'duplicate freeze evidence' \
  "Asset freeze fails closed on ambiguous historical evidence"
asset_freeze_migration="$REPO_ROOT/services/asset/migrations/20260801_option_freeze_idempotency_evidence.sql"
check_contains "$REPO_ROOT/services/asset/asset.sql" 'idx_asset_freeze_option_business_key' \
  "Asset schema indexes Option freeze business-key evidence"
check_contains "$asset_freeze_migration" 'dbinit:baseline-safe' \
  "Asset Option freeze evidence migration runs on baseline databases"
check_contains "$asset_freeze_migration" 'HAVING COUNT[(][*][)]=1' \
  "Asset migration only adopts unambiguous legacy freeze evidence"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" 'asset_freeze_duplicate' \
  "operations metrics expose duplicate Option freeze business keys"
check_contains "$operations_rules" 'OptionAssetFreezeBusinessKeyDuplicate' \
  "duplicate Option freeze business keys trigger a dedicated alert"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP0AssetRPCEndToEnd' \
  "Option repository provides a repeatable real Asset RPC P0 gate"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP0MarginCoinRelease' \
  "real Asset RPC gate covers physical Call and Put collateral release"
physical_order_coin_test="$OPTION_DIR/internal/logic/task/p0_physical_order_coin_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'physical Call Put ordinary order coin lifecycle' \
  "real Asset RPC gate invokes the physical order coin lifecycle"
check_contains "$physical_order_coin_test" 'ORDER_TYPE_LIMIT' \
  "physical order coin gate covers LIMIT"
check_contains "$physical_order_coin_test" 'ORDER_TYPE_MARKET' \
  "physical order coin gate covers MARKET"
check_contains "$physical_order_coin_test" 'ORDER_TYPE_POST_ONLY' \
  "physical order coin gate covers POST_ONLY"
check_contains "$physical_order_coin_test" 'ORDER_TYPE_IOC' \
  "physical order coin gate covers IOC"
check_contains "$physical_order_coin_test" 'ORDER_TYPE_FOK' \
  "physical order coin gate covers FOK"
check_contains "$physical_order_coin_test" 'OPTION_TYPE_CALL' \
  "physical order coin gate covers Call BTC collateral"
check_contains "$physical_order_coin_test" 'OPTION_TYPE_PUT' \
  "physical order coin gate covers Put USDT collateral"
check_contains "$physical_order_coin_test" 'cancelOptionSystemOrder' \
  "physical order coin gate covers liquidation-side order release"
check_contains "$physical_order_coin_test" 'NewForceCancelContractOrdersLogic' \
  "physical order coin gate covers governed admin force cancel"
check_contains "$physical_order_coin_test" 'expireContractOrders' \
  "physical order coin gate covers expiry release"
check_contains "$physical_order_coin_test" 'wrongCoins != 0' \
  "physical order coin gate rejects any release in the wrong coin"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'physical_order_coin=' \
  "real Asset RPC gate reports physical order coin evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260731_zx_option_margin_coin_evidence.sql' \
  "real Asset RPC gate installs the collateral-coin database guards"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP0CashSettlement' \
  "real Asset RPC gate covers confirmed-price cash settlement"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertP0SettlementAssetConservation' \
  "cash settlement gate proves wallet and Asset-flow conservation"
cash_expiry_capacity_test="$OPTION_DIR/internal/logic/task/p0_cash_expiry_capacity_rpc_integration_test.go"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP0CashExpiryCapacityAssetRPC' \
  "repeatable acceptance runs the cash-expiry capacity and economics boundary"
check_contains "$cash_expiry_capacity_test" 'longCount  = 501' \
  "cash-expiry capacity gate crosses the 500-position page boundary"
check_contains "$cash_expiry_capacity_test" 'partialQty := decimal.RequireFromString' \
  "cash-expiry gate covers a partial contract quantity"
check_contains "$cash_expiry_capacity_test" 'accountCount != 502' \
  "cash-expiry gate proves distinct account evidence"
check_contains "$cash_expiry_capacity_test" 'feeCredits != 501' \
  "cash-expiry gate proves per-position exercise fees"
check_contains "$cash_expiry_capacity_test" 'walletTotal != "12000.000000000000000000"' \
  "cash-expiry gate proves complete wallet conservation"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" 'const pageSize int64 = 100' \
  "cash settlement uses the normalized position-page boundary"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" 'cursor, pageSize' \
  "cash settlement traverses every normalized position page"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'cash_expiry_capacity=501' \
  "real Asset RPC gate reports cash-expiry capacity evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260731_zy_option_settlement_price_evidence.sql' \
  "real Asset RPC gate installs the settlement-price evidence guards"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertMissingPriceBlocksSettlement' \
  "real Asset RPC gate proves missing-price settlement fails closed"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP0CashSettlementFailureRecovery' \
  "real Asset RPC gate covers a failed settlement debit and recovery"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertP0SettlementDebitFailureBarrier' \
  "failed settlement debit gate blocks every later Asset step"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertP0RecoveredDebitIdentity' \
  "settlement debit recovery preserves the original business identity"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'failAfterCommit' \
  "settlement gate covers an Asset debit committed before response loss"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'settlement_failure_recovery_contract=' \
  "real Asset RPC gate reports settlement failure-recovery evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP0CashSettlementStaleProcessingRecovery' \
  "real Asset RPC gate covers a stale processing instruction after committed debit"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertP0StaleProcessingBarrier' \
  "stale processing gate blocks every later Asset step before recovery"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'stale_processing_recovery_contract=' \
  "real Asset RPC gate reports stale-processing recovery evidence"
manual_retry_helper="$OPTION_DIR/internal/logic/admin/manual_asset_instruction_retry.go"
check_contains "$REPO_ROOT/proto/option/option.pb.go" 'Reason.*protobuf:.*3,opt,name=reason' \
  "Option retry-asset RPC requires generated reason evidence"
check_contains "$REPO_ROOT/admin-api/api/option.api" 'reason.*validate:.*min=1,max=64' \
  "Admin API bounds manual retry reasons to 1-64 characters"
check_contains "$manual_retry_helper" 'TransactCtx' \
  "manual asset reset and audit append share one database transaction"
check_contains "$manual_retry_helper" 'ASSET_INSTRUCTION_MANUAL_RETRY' \
  "manual asset retries use a stable immutable audit event type"
check_contains "$manual_retry_helper" 'InsertOptionTradingControlEvent' \
  "manual asset retries append audit evidence without cache dependency"
check_contains "$OPTION_DIR/internal/logic/admin/retryassetinstructionlogic.go" 'GetUserIdFromMd' \
  "manual asset retry requires an authenticated operator"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'assertP0ManualRetryRejected' \
  "real Asset RPC gate rejects missing retry reason and operator"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'fromRetryCount=20' \
  "real Asset RPC gate preserves the 20-failure manual-review evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'manual retry audit event update unexpectedly succeeded' \
  "real Asset RPC gate proves manual retry events cannot be updated"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'manual retry audit event delete unexpectedly succeeded' \
  "real Asset RPC gate proves manual retry events cannot be deleted"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260730_ze_option_trading_controls.sql' \
  "real Asset RPC gate installs immutable trading-control audit guards"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'insufficient_balance_manual_retry_events=' \
  "real Asset RPC gate reports balance-topup manual recovery evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'physical delivery Call Put failure isolation and recovery' \
  "real Asset RPC gate invokes Call and Put physical delivery acceptance"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'failOnceSubAvailableClient' \
  "physical delivery acceptance injects a committed-debit response loss"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'failOnceDeductFrozenClient' \
  "physical delivery acceptance injects a committed collateral-debit response loss"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'expirePhysicalDeliveryUnit' \
  "physical delivery acceptance exercises cure expiry and default"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'NewRetryPhysicalDeliveryUnitLogic' \
  "physical delivery acceptance recovers the original unit through the governed admin entry"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'const workers = 20' \
  "physical delivery acceptance races twenty governed admin recovery requests"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_rpc_integration_test.go" 'success != 1 \|\| rejected != workers-1' \
  "physical delivery acceptance requires exactly one admin recovery winner"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'physical_delivery_contract=' \
  "real Asset RPC gate reports physical delivery unit and instruction evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP1PhysicalDeliveryProcessKillTakeover' \
  "repeatable acceptance runs physical delivery process-kill takeover"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_process_kill_rpc_integration_test.go" 'Process.Kill' \
  "physical delivery process gate sends a real SIGKILL after Asset credit commits"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_process_kill_rpc_integration_test.go" 'waitP0TaskLeaseExpiry' \
  "physical delivery process gate waits for the killed worker lease to expire naturally"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_process_kill_rpc_integration_test.go" 'remainingStatus.*ASSET_INSTRUCTION_STATUS_PENDING' \
  "physical delivery process gate proves the remaining credit is blocked before takeover"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'physical_delivery_process_kill=' \
  "real Asset RPC gate reports physical delivery process-takeover evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP1PhysicalDeliveryCapacityAssetRPC' \
  "repeatable acceptance runs the 501-unit physical delivery capacity boundary"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_capacity_rpc_integration_test.go" 'p1PhysicalDeliveryCapacityUnits = 501' \
  "physical delivery capacity gate crosses the 100-position page boundary"
check_contains "$OPTION_DIR/internal/logic/task/p1_physical_delivery_capacity_rpc_integration_test.go" 'p1PhysicalDeliveryCapacityUnits[*]4' \
  "physical delivery capacity gate requires all 2004 Asset legs"
check_contains "$OPTION_DIR/models/toptionassetinstructionmodel.go" 'SummarizeByBizNo' \
  "settlement completion aggregates instruction progress without loading the full batch"
check_contains "$OPTION_DIR/models/toptionphysicaldeliveryunitmodel.go" 'FindExceptionByBatch' \
  "physical settlement finds exceptional units with a bounded query"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'physical_delivery_capacity=' \
  "real Asset RPC gate reports physical delivery capacity evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260730_zd_option_exercise_governance.sql' \
  "real Asset RPC gate installs exercise-instruction governance guards"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'American early exercise concurrency and FIFO' \
  "real Asset RPC gate invokes concurrent American exercise acceptance"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'results := make[(]chan exerciseResult, 20[)]' \
  "American exercise acceptance submits twenty concurrent idempotent requests"
exercise_window_helper="$OPTION_DIR/internal/logic/app/exercise_submission_window.go"
check_contains "$exercise_window_helper" 'nowMillis < cutoffMillis' \
  "exercise cutoff is an exclusive millisecond boundary"
check_contains "$OPTION_DIR/internal/logic/app/exerciselogic_test.go" 'nowMillis: 199999, want: true' \
  "exercise cutoff gate accepts cutoff minus one millisecond"
check_contains "$OPTION_DIR/internal/logic/app/exerciselogic_test.go" 'nowMillis: 200000, want: false' \
  "exercise cutoff gate rejects the exact cutoff instant"
check_contains "$OPTION_DIR/internal/logic/app/exerciselogic.go" 'exerciseSubmissionWindowOpen[(]currentContract, txNowMillis[)]' \
  "American exercise rechecks the millisecond cutoff under the contract lock"
check_contains "$OPTION_DIR/internal/logic/app/setexerciseinstructionlogic.go" 'exerciseSubmissionWindowOpen[(]currentContract, txNowMillis[)]' \
  "expiry instruction rechecks the millisecond cutoff under the contract lock"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'exercise cutoff and lifecycle row-lock races' \
  "real Asset RPC gate invokes exercise cutoff and lifecycle races"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'SELECT id FROM t_option_contract WHERE id=[?] FOR UPDATE' \
  "exercise race gate waits on the real contract row lock"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'time.UnixMilli[(]cutoffMillis[)]' \
  "exercise race gate crosses the published cutoff in real wall-clock time"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'CONTRACT_STATUS_EXPIRED' \
  "expiry instruction race commits a lifecycle transition while the request waits"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'P0-INSTRUCTION-PAUSED-ACCEPT' \
  "expiry instruction gate preserves PAUSED holder rights"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'exercise_cutoff_race=' \
  "real Asset RPC gate reports exercise cutoff race evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'expiry instruction different-key replacement race' \
  "real Asset RPC gate invokes different-key instruction replacement"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'concurrency[[:space:]]*=[[:space:]]*20' \
  "instruction replacement gate submits twenty different client keys"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'assertP0ExerciseInstructionVersionChain' \
  "instruction replacement gate verifies the immutable version chain"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'old, now-superseded key returns that exact historical version' \
  "instruction replacement gate replays historical idempotency identities"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'exercise_instruction_replacement_race=' \
  "real Asset RPC gate reports instruction replacement evidence"
exercise_process_kill="$OPTION_DIR/internal/logic/task/p0_exercise_process_kill_rpc_integration_test.go"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP0ExerciseProcessKillTakeover' \
  "repeatable acceptance runs American exercise process-kill takeover"
check_contains "$exercise_process_kill" 'Process[.]Kill[(][)]' \
  "American exercise process gate sends a real SIGKILL"
check_contains "$exercise_process_kill" 'waitP0TaskLeaseExpiry' \
  "American exercise process gate waits for natural task-lease expiry"
check_contains "$exercise_process_kill" 'startP0AssetWorker[(]t, takeoverProxy[.]address[(][)][)]' \
  "American exercise process gate starts competing takeover processes"
check_contains "$exercise_process_kill" 'assertP0ExerciseProcessKillCompleted' \
  "American exercise process gate proves terminal business and Asset evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'american_exercise_process_kill=' \
  "real Asset RPC gate reports American exercise process-takeover evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'American exercise races short close orders' \
  "real Asset RPC gate invokes the American assignment versus close-order race"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_close_race_rpc_integration_test.go" 'iteration <= 10' \
  "American assignment acceptance repeats the concurrent close-order race"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_close_race_rpc_integration_test.go" 'AMERICAN_EXERCISE_ASSIGNMENT' \
  "American assignment acceptance verifies deterministic close-order cancellation"
check_contains "$OPTION_DIR/internal/logic/task/p0_assignment_capacity_rpc_integration_test.go" 'shortCount != 501 && shortCount != 5000' \
  "real Asset RPC gate fixes the 501 and 5000 short capacity boundaries"
check_contains "$OPTION_DIR/internal/logic/task/p0_assignment_capacity_rpc_integration_test.go" 'processAssetInstructions[(]t, ctx, serviceCtx[)]' \
  "American capacity gate executes clearing through the real Asset instruction path"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'for capacity_shorts in 501 5000' \
  "repeatable acceptance runs both American assignment capacity boundaries"
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" 'testP0ExpiryAutoDNEActualAssignment' \
  "real Asset RPC gate covers mixed AUTO and DNE expiry allocation"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'american_exercise=' \
  "real Asset RPC gate reports American exercise evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'american_exercise_close_race=' \
  "real Asset RPC gate reports American close-order race evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'american_assignment_capacity=' \
  "real Asset RPC gate reports American assignment capacity evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'expiry_auto_dne=' \
  "real Asset RPC gate reports AUTO and DNE expiry evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_partial_close_rpc_integration_test.go" 'testP0PartialCloseTradeAccounting' \
  "real Asset RPC gate covers matched partial-close accounting"
check_contains "$OPTION_DIR/internal/logic/task/option_business_helpers.go" 'pos.UnrealizedPnl = pos.MarkPrice.Sub[(]pos.OpenAvgPrice[)]' \
  "partial close recalculates remaining long unrealized PnL"
check_contains "$OPTION_DIR/internal/logic/task/p0_partial_close_rpc_integration_test.go" '"3.7"' \
  "partial close gate reconciles trade PnL, fees, and total return"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'partial_close_trade=' \
  "real Asset RPC gate reports partial-close evidence"
order_admission_test="$OPTION_DIR/internal/logic/task/p0_order_admission_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'full order admission to risk accounting' \
  "real Asset RPC gate invokes the full order-admission scenario"
check_contains "$order_admission_test" 'NewPlaceOrderLogic' \
  "order-admission gate enters through the public PlaceOrder logic"
check_contains "$order_admission_test" 'seedP0OpenTradingCalendar' \
  "order-admission gate uses an approved open trading calendar"
check_contains "$order_admission_test" 'MarginAmount[.]Equal[(]decimal[.]NewFromInt[(]50[)][)]' \
  "seller admission freezes the expected isolated margin"
check_contains "$order_admission_test" 'MarginAmount[.]Equal[(]decimal[.]RequireFromString[(]"10[.]4"[)][)]' \
  "buyer admission freezes premium plus taker fee"
check_contains "$order_admission_test" 'processP0TradeEvents' \
  "order-admission gate crosses the outbox and inbox position barrier"
check_contains "$order_admission_test" 'processP0OrderRiskAccounts' \
  "order-admission gate refreshes the affected wallet-scope risk accounts"
check_contains "$order_admission_test" 'ClientOrderId: "P0-ORDER-ADMISSION-BUYER"' \
  "order-admission gate replays the client order identity"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'order_admission=' \
  "real Asset RPC gate reports full order-admission evidence"
trading_calendar_test="$OPTION_DIR/internal/logic/task/p0_trading_calendar_rpc_integration_test.go"
trading_calendar_migration="$OPTION_DIR/migrations/20260731_zk_option_trading_calendar.sql"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" \
  'trading calendar future switch and manual halt release barrier' \
  "real Asset RPC gate invokes future calendar switching and manual halt"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260731_zk_option_trading_calendar[.]sql' \
  "real Asset RPC gate installs the calendar governance migration"
check_contains "$trading_calendar_test" 'P2_CALENDAR_FUTURE_SWITCH' \
  "calendar gate covers an exact future-version switch"
check_contains "$trading_calendar_test" 'P2-CALENDAR-AT-SWITCH-REJECTED' \
  "calendar gate requires a zero-side-effect boundary rejection"
check_contains "$trading_calendar_test" 'ORDER_STATUS_CANCELING' \
  "manual-halt gate reaches the in-flight Asset release state"
check_contains "$trading_calendar_test" 'resumed before releases' \
  "manual-halt gate rejects recovery before Asset releases finish"
check_contains "$trading_calendar_test" 'failOnceUnfreezeAssetClient' \
  "manual-halt gate injects a committed Asset unfreeze response loss"
check_contains "$trading_calendar_test" 'TRADING_HALT_ASSET_RESPONSE_LOSS_CONFIRMED' \
  "manual-halt response-loss recovery requires a governed operator reason"
check_contains "$trading_calendar_test" 'resumed after committed release response loss' \
  "manual-halt gate keeps recovery closed after a committed release response loss"
check_contains "$trading_calendar_test" 'rounds[[:space:]]*= 20' \
  "calendar gate repeats the concurrent halt/order race twenty times"
check_contains "$trading_calendar_test" 'accepted == 0 [|][|] rejected == 0' \
  "calendar race gate requires both legal serializations to execute"
check_contains "$trading_calendar_test" 'sync[.]WaitGroup' \
  "calendar gate starts public orders and halts from one concurrency barrier"
check_contains "$trading_calendar_test" 'instructions != 0' \
  "calendar race gate forbids Asset side effects from close-order admission"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'HasUnsafeContractResumeOrders' \
  "contract recovery uses an explicit unsafe-order barrier"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'ORDER_STATUS_EXPIRING' \
  "contract recovery barrier includes expiry funding"
check_contains "$trading_calendar_migration" 'trading calendar economic fields are immutable' \
  "calendar governance migration protects approved version evidence"
check_contains "$trading_calendar_migration" 'trading halt identity is immutable' \
  "calendar governance migration protects halt identity evidence"
check_contains "$OPTION_DIR/docs/option-p2-001-trading-calendar-repository-acceptance.md" \
  'CAL-PRE-001' \
  "calendar repository acceptance records remaining production blockers"
check_contains "$OPTION_DIR/docs/templates/trading-halt-record.md" \
  'CANCELING/EXPIRING' \
  "trading-halt operations record blocks recovery during Asset release"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  'trading_calendar_halt_order_race=' \
  "real Asset RPC gate reports the halt/order concurrency evidence"
corporate_action_test="$OPTION_DIR/internal/logic/task/p2_corporate_action_rpc_integration_test.go"
corporate_action_gate="$OPTION_DIR/models/optioncorporateactionexecutionmodel.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" \
  'corporate action 5001-position restart and frozen-asset conservation' \
  "real Asset RPC gate invokes corporate-action capacity and conservation"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260731_zl_option_corporate_action[.]sql' \
  "real Asset RPC gate installs corporate-action governance twice"
check_contains "$corporate_action_test" 'p2CorporateActionPositionCount int64 = 5001' \
  "corporate-action gate crosses the 5000-position boundary"
check_contains "$corporate_action_test" 'CORPORATE_ACTION_CONTRACT_STATUS_EXECUTING[)], 100, 0' \
  "corporate-action gate proves the first batch commits exactly one hundred"
check_contains "$corporate_action_test" 'SET pending_margin=0' \
  "corporate-action gate clears a persisted failed-position blocker"
check_contains "$corporate_action_test" 'NewProcessCorporateActionsLogic[(]ctx, serviceCtx[)]' \
  "corporate-action gate recreates task logic while resuming the durable cursor"
check_contains "$corporate_action_test" 'corporate action migration active' \
  "corporate-action gate requires concurrent risk fail-closed behavior"
check_contains "$corporate_action_test" 'freezes != 2 [|][|] flows != 2 [|][|] instructions != 0' \
  "corporate-action gate preserves two real freezes without new instructions"
check_contains "$OPTION_DIR/internal/logic/admin/reviewcorporateactionlogic.go" 'ORDER_STATUS_EXPIRING' \
  "corporate-action review blocks expiry funding in flight"
check_contains "$corporate_action_gate" 'LEFT JOIN t_option_order ai_order' \
  "corporate-action execution resolves instructions through their real relations"
check_not_contains "$corporate_action_gate" 't_option_asset_instruction[[:space:]]+WHERE.*contract_id' \
  "corporate-action execution does not query a nonexistent instruction contract column"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'migrationCache' \
  "risk scan caches active corporate-action checks by tenant and contract"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'migrationContracts' \
  "risk scan deduplicates corporate-action failures per wallet and contract"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'corporate_action_capacity=' \
  "real Asset RPC gate reports corporate-action capacity and Asset evidence"
check_contains "$OPTION_DIR/docs/option-p2-004-corporate-action-repository-acceptance.md" \
  '生产前剩余验收' \
  "corporate-action repository acceptance records remaining production blockers"
lifecycle_migration="$OPTION_DIR/migrations/20260802_option_last_trade_time.sql"
check_contains "$OPTION_DIR/option.sql" 'last_trade_time.*最后可交易时间' \
  "Option schema declares an independent last-trade boundary"
check_contains "$OPTION_DIR/models/toptioncontractmodel_gen.go" 'LastTradeTime.*last_trade_time' \
  "make gen-model output contains the independent last-trade field"
check_contains "$REPO_ROOT/proto/option/model.proto" 'last_trade_time = 56' \
  "Option protocol exposes the contract last-trade boundary"
check_contains "$REPO_ROOT/admin-api/api/option.api" 'lastTradeTime' \
  "Admin API exposes the last-trade boundary"
check_contains "$REPO_ROOT/admin-ui/src/views/option/contracts.vue" 'lastTradeTime' \
  "Admin UI edits and displays the last-trade boundary"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260802_option_last_trade_time[.]sql' \
  "real Asset RPC gate installs the lifecycle migration twice"
check_contains "$lifecycle_migration" 'chk_option_contract_lifecycle_times' \
  "lifecycle migration enforces the five-time ordering"
check_contains "$lifecycle_migration" 'trg_option_contract_last_trade_default_insert' \
  "lifecycle rollout supports legacy internal inserts without hiding new API input"
check_contains "$lifecycle_migration" \
  'OLD[.]`last_trade_time` <=> NEW[.]`last_trade_time`' \
  "generated-series contract guard protects last-trade economics"
check_contains "$OPTION_DIR/internal/delayqueue/queue.go" 'ActionCloseContractTrading' \
  "delay queue declares an independent close-trading action"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" \
  'func [(]l [*]ProcessContractLifecycleLogic[)] closeTradingContracts' \
  "periodic lifecycle scan compensates missed last-trade messages"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" \
  'CONTRACT_LAST_TRADE_ENDED' \
  "last-trade closure uses a stable audited cancellation reason"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" \
  'HasUnsafeContractResumeOrders' \
  "expiry waits for last-trade order releases to reach terminal states"
last_trade_guard_files="
$OPTION_DIR/internal/logic/app/placeorderlogic.go
$OPTION_DIR/internal/logic/app/app_order_match.go
$OPTION_DIR/internal/logic/app/trading_control.go
$OPTION_DIR/internal/logic/app/combo_order_funding.go
$OPTION_DIR/internal/logic/app/combo_order_match.go
$OPTION_DIR/internal/logic/task/processassetinstructionslogic.go
$OPTION_DIR/internal/logic/admin/resumecontracttradinglogic.go
"
for guard_file in $last_trade_guard_files; do
  check_contains "$guard_file" 'LastTradeTime' \
    "$(basename "$guard_file") guards the independent last-trade boundary"
done
check_contains "$OPTION_DIR/internal/logic/task/p0_exercise_rpc_integration_test.go" \
  'P0-LAST-TRADE-INDEPENDENT' \
  "real Asset RPC gate exercises five independent lifecycle times"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'last_trade_boundary=' \
  "real Asset RPC gate reports machine-readable last-trade evidence"
check_contains "$OPTION_DIR/docs/option-p2-002-p2-003-lifecycle-repository-acceptance.md" \
  'REPOSITORY_PASSED / PREPROD_BLOCKED' \
  "lifecycle acceptance preserves outstanding preproduction blockers"
contract_series_test="$OPTION_DIR/internal/logic/task/p2_contract_series_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" \
  'contract series concurrent generation and delayed launch gates' \
  "real Asset RPC gate invokes contract-series concurrency and listing controls"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260731_zn_option_contract_series[.]sql' \
  "real Asset RPC gate installs contract-series schema twice"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260802_option_contract_series_contract_guard[.]sql' \
  "real Asset RPC gate installs generated-contract immutability twice"
check_contains "$contract_series_test" 'p2SeriesContractNum int64 = 500' \
  "contract-series gate exercises the exact five-hundred-contract boundary"
check_contains "$contract_series_test" 'createConcurrency = 20' \
  "contract-series gate starts twenty identical concurrent creates"
check_contains "$contract_series_test" 'testP2ContractSeriesCollisionRollback' \
  "contract-series gate requires full rollback on one code collision"
check_contains "$contract_series_test" 'testP2ContractSeriesOversizeRejected' \
  "contract-series gate rejects the 502-contract input before writes"
check_contains "$contract_series_test" 'handleDelayMessage' \
  "contract-series gate exercises the real delayed listing handler"
check_contains "$contract_series_test" 'generated contract economics were mutable' \
  "contract-series gate requires database-level economic immutability"
check_contains "$OPTION_DIR/internal/logic/admin/createcontractserieslogic.go" \
  'contractSeriesCreateMaxAttempts = 5' \
  "contract-series create has a bounded lock-victim retry budget"
check_contains "$OPTION_DIR/internal/logic/admin/createcontractserieslogic.go" \
  'mysqlErr[.]Number == 1213 [|][|] mysqlErr[.]Number == 1205' \
  "contract-series create retries only MySQL deadlock and lock timeout"
check_contains "$OPTION_DIR/models/toptioncontractseriesmodel.go" \
  'FindOneByTenantIdRequestKeyNoCache' \
  "contract-series idempotency bypasses stale negative cache entries"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" \
  'listContractIfEligible' \
  "contract lifecycle centralizes transactional listing eligibility"
check_contains "$OPTION_DIR/internal/logic/task/delay_queue.go" \
  'listContractIfEligible' \
  "delayed listing uses the centralized runtime gate"
check_not_contains "$OPTION_DIR/internal/logic/task/delay_queue.go" \
  'contract[.]Status = int64[(]option[.]ContractStatus_CONTRACT_STATUS_TRADING' \
  "delayed listing cannot assign TRADING directly"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'contract_series_capacity=' \
  "real Asset RPC gate reports contract-series capacity and rollback evidence"
check_contains "$OPTION_DIR/docs/option-p2-005-contract-series-repository-acceptance.md" \
  '生产前剩余验收' \
  "contract-series repository acceptance records remaining production blockers"
public_market_test="$OPTION_DIR/internal/logic/task/p2_public_market_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" \
  'public chain book statistics OI isolation and capacity' \
  "real Asset RPC gate invokes public-market consistency and capacity acceptance"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260731_zo_option_public_market[.]sql' \
  "real Asset RPC gate installs public-market indexes twice"
check_contains "$OPTION_DIR/internal/logic/app/public_market_snapshot.go" \
  'Isolation: sql[.]LevelRepeatableRead' \
  "public-market response explicitly uses repeatable-read isolation"
check_contains "$OPTION_DIR/internal/logic/app/public_market_snapshot.go" \
  'ReadOnly:[[:space:]]+true' \
  "public-market response uses a read-only transaction"
check_contains "$OPTION_DIR/models/toptionordermodel.go" \
  'order_type IN [(][?],[?][)]' \
  "public book admits only explicit resting limit order types"
check_contains "$OPTION_DIR/models/toptionordermodel.go" \
  'ORDER_TYPE_POST_ONLY' \
  "public book includes valid post-only resting quotes"
check_contains "$public_market_test" 'capacitySpecs := make[(][[]]p2PublicContractSpec, 0, 500[)]' \
  "public-market gate builds the exact five-hundred-contract chain"
check_contains "$public_market_test" 'DepthLimit: 101' \
  "public-market gate rejects the one-hundred-and-one-level request"
check_contains "$public_market_test" 'const readers = 24' \
  "public-market gate races snapshot readers against atomic book/trade changes"
check_contains "$public_market_test" 'const concurrency = 16' \
  "public-market gate runs concurrent maximum-capacity chain and book reads"
check_contains "$public_market_test" 'p2PublicOtherTenantID' \
  "public-market gate injects cross-tenant contracts and facts"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'public_market_capacity=' \
  "real Asset RPC gate reports public-market capacity and isolation evidence"
check_contains "$OPTION_DIR/docs/option-p2-006-public-market-repository-acceptance.md" \
  '生产前剩余验收' \
  "public-market repository acceptance records remaining production blockers"
combo_order_test="$OPTION_DIR/internal/logic/task/p2_combo_order_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" \
  'testP2ComboOrderAcceptance' \
  "real Asset RPC gate invokes complex-order acceptance"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260731_zp_option_combo_order[.]sql' \
  "real Asset RPC gate installs complex-order schema and guards twice"
check_contains "$combo_order_test" 'parallel[[:space:]]*=[[:space:]]*50' \
  "complex-order gate starts fifty identical concurrent creates"
check_contains "$OPTION_DIR/models/toptioncomboordermodel.go" \
  'FindOneByTenantIdUserIdClientComboIdNoCache' \
  "complex-order idempotency exposes an authoritative no-cache lookup"
check_contains "$OPTION_DIR/internal/logic/app/placecomboorderlogic.go" \
  'comboOrderCreateMaxAttempts[[:space:]]*=[[:space:]]*5' \
  "complex-order create has a bounded lock-victim retry budget"
check_contains "$OPTION_DIR/internal/logic/app/placecomboorderlogic.go" \
  'mysqlErr[.]Number == 1213 [|][|] mysqlErr[.]Number == 1205' \
  "complex-order create retries only MySQL deadlock and lock timeout"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_funding.go" \
  'COMBO_USER_KILL_SWITCH_AFTER_FUNDING' \
  "complex-order activation rechecks the user kill switch after funding"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_funding.go" \
  'COMBO_STALE_MARK_AFTER_FUNDING' \
  "complex-order activation rechecks mark freshness after funding"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_funding.go" \
  'COMBO_STALE_UNDERLYING_AFTER_FUNDING' \
  "complex-order activation rechecks underlying freshness after funding"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_funding.go" \
  'COMBO_SELL_MARGIN_INSUFFICIENT_AFTER_FUNDING' \
  "complex-order activation rechecks seller margin after funding"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_match.go" \
  'func transitionComboToCancellation' \
  "complex-order cancellation centralizes the guarded parent transition"
check_contains "$OPTION_DIR/internal/logic/app/combo_order_match.go" \
  'COMBO_ORDER_STATUS_CANCELING' \
  "complex-order cancellation traverses CANCELING before its terminal state"
check_contains "$OPTION_DIR/internal/logic/app/cancelcomboorderlogic.go" \
  'transitionComboToCancellation' \
  "user complex-order cancellation uses the guarded parent transition"
check_contains "$combo_order_test" 'COMBO-003 injected freeze response loss after commit' \
  "complex-order gate injects a committed Asset freeze response loss"
check_contains "$combo_order_test" 'trg_accept_combo_leg2_failure' \
  "complex-order gate injects a real second-leg database failure"
check_contains "$combo_order_test" 'FOK_NOT_FILLED' \
  "complex-order gate proves single-maker FOK zero-leg behavior"
check_contains "$combo_order_test" 'SELF_TRADE_PREVENTED' \
  "complex-order gate proves cross-account self-trade prevention"
check_contains "$combo_order_test" 'combo debit barrier pending-outbox/positions' \
  "complex-order gate blocks every leg position behind the debit barrier"
check_contains "$combo_order_test" 'cross-tenant combo detail was not rejected' \
  "complex-order gate proves admin detail tenant isolation"
check_contains "$combo_order_test" 'reasonless combo force cancel was not rejected' \
  "complex-order gate requires an operator reason for whole-group cancellation"
check_contains "$combo_order_test" 'ADMIN_FORCE_CANCEL:COMBO_009_ACCEPTANCE' \
  "complex-order gate proves authorized whole-group cancellation"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'combo_acceptance=' \
  "real Asset RPC gate reports complex-order atomicity and settlement evidence"
check_contains "$OPTION_DIR/docs/option-p2-007-complex-order-repository-acceptance.md" \
  '生产前剩余验收' \
  "complex-order repository acceptance records remaining production blockers"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'wallet restriction propagation and cross-account STP' \
  "real Asset RPC gate invokes wallet restriction propagation and cross-account STP"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" '7101' \
  "wallet restriction gate checks the first business account"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" '7102' \
  "wallet restriction gate checks a second business account"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'seller[.]AccountId == buyer[.]AccountId' \
  "cross-account STP gate proves the maker and taker use different accounts"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'SELF_TRADE_PREVENTED' \
  "cross-account STP gate requires cancel-taker evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'wallet_scope_restriction=' \
  "real Asset RPC gate reports wallet-level restriction evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'cross_account_stp=' \
  "real Asset RPC gate reports cross-account STP evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'sync[.]WaitGroup' \
  "portfolio admission gate starts concurrent cross-account orders"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'AccountId: 8301' \
  "portfolio admission gate uses the first business account"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'AccountId: 8302' \
  "portfolio admission gate uses a second business account"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'riskAccounts != 1 \|\| walletAccounts != 1' \
  "portfolio admission gate requires one wallet-scope risk lock"
check_contains "$OPTION_DIR/internal/logic/task/p0_wallet_scope_stp_rpc_integration_test.go" 'margin[.]Equal[(]expectedMargin[)]' \
  "portfolio admission gate matches aggregate governed risk"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'portfolio_cross_account_concurrency=' \
  "real Asset RPC gate reports concurrent cross-account portfolio evidence"
limit_concurrency_test="$OPTION_DIR/internal/logic/task/p0_trading_limit_concurrency_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'concurrent user and open-interest limits' \
  "real Asset RPC gate invokes concurrent user and OI limits"
check_contains "$limit_concurrency_test" 'sync[.]WaitGroup' \
  "trading-limit gate starts concurrent public orders"
check_contains "$limit_concurrency_test" 'P0-CONCURRENT-USER-LIMIT-%02d' \
  "user-limit gate submits twenty distinct client identities"
check_contains "$limit_concurrency_test" 'P0-CONCURRENT-OI-LIMIT-%02d' \
  "OI-limit gate submits twenty distinct users"
check_contains "$limit_concurrency_test" 'USER_LONG_LIMIT' \
  "user-limit gate requires the exact rejection reason"
check_contains "$limit_concurrency_test" 'OPEN_INTEREST_LIMIT' \
  "OI-limit gate requires the exact rejection reason"
check_contains "$limit_concurrency_test" 'wantAccepted, wantRejected' \
  "trading-limit gate fixes accepted and rejected cardinality"
check_contains "$limit_concurrency_test" 'instructions != accepted[*]2' \
  "trading-limit gate requires freeze and release for every accepted order"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'concurrent_trading_limit=' \
  "real Asset RPC gate reports concurrent trading-limit evidence"
emergency_controls_test="$OPTION_DIR/internal/logic/task/p0_emergency_controls_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'kill switch races matching and circuit breaker cancels batches' \
  "real Asset RPC gate invokes kill-switch races and circuit-breaker batch cancellation"
check_contains "$emergency_controls_test" 'ORDER_STATUS_FUNDING' \
  "kill-switch gate includes a FUNDING order"
check_contains "$emergency_controls_test" 'ORDER_STATUS_PENDING' \
  "kill-switch gate includes a PENDING order"
check_contains "$emergency_controls_test" 'ORDER_STATUS_PART_FILLED' \
  "kill-switch gate includes a PART_FILLED order"
check_contains "$emergency_controls_test" 'kill switch release did not block non-terminal orders' \
  "kill-switch release remains fail-closed until control releases finish"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'ORDER_STATUS_CANCELING' \
  "kill-switch release barrier treats cancellation funding as non-terminal"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'ORDER_STATUS_EXPIRING' \
  "kill-switch release barrier treats expiry funding as non-terminal"
check_contains "$emergency_controls_test" 'FOR UPDATE' \
  "kill-switch match-race gate uses a real MySQL row lock"
check_contains "$emergency_controls_test" 'trades != 0 [|][|] activeOrders != 0' \
  "kill-switch match-race gate requires zero trades and zero active orders"
check_contains "$OPTION_DIR/internal/logic/app/control_cancel.go" 'controlCancelMaxAttempts = 5' \
  "control cancellation has a bounded deadlock retry budget"
check_contains "$OPTION_DIR/internal/logic/app/control_cancel.go" 'mysqlErr[.]Number == 1213 [|][|] mysqlErr[.]Number == 1205' \
  "control cancellation retries only MySQL deadlock and lock-timeout victims"
check_contains "$emergency_controls_test" 'batchSize.*= 101' \
  "circuit-breaker gate crosses the one-hundred-order page boundary"
check_contains "$emergency_controls_test" 'haltTotal != int64[(]batchSize[)] [|][|] haltSuccess != int64[(]batchSize[)]' \
  "circuit-breaker gate requires exact per-order halt counters"
check_contains "$emergency_controls_test" 'instructions != int64[(]batchSize[*]2[)]' \
  "circuit-breaker gate requires freeze and release for all 101 orders"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'kill_switch_release_barrier=' \
  "real Asset RPC gate reports kill-switch release-barrier evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'kill_switch_match_race=' \
  "real Asset RPC gate reports kill-switch match-race evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'circuit_breaker_batch=' \
  "real Asset RPC gate reports circuit-breaker batch evidence"
mmp_test="$OPTION_DIR/internal/logic/task/p1_mmp_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'MMP batch cancellation release barrier and response-loss recovery' \
  "real Asset RPC gate invokes MMP batch, barrier and response-loss recovery"
check_contains "$mmp_test" 'makerCount.*= 102' \
  "MMP gate leaves 101 group releases and crosses the one-hundred-order page boundary"
check_contains "$mmp_test" 'failOnceUnfreezeAssetClient' \
  "MMP gate loses a response after a committed Asset unfreeze"
check_contains "$mmp_test" 'MMP committed release flows=%d want=1' \
  "MMP gate proves committed response loss retains one authoritative flow"
check_contains "$mmp_test" 'results := make[(]chan resetResult, 20[)]' \
  "MMP gate starts twenty concurrent manual recoveries"
check_contains "$mmp_test" 'successfulResets != 1' \
  "MMP gate requires exactly one concurrent recovery winner"
check_contains "$mmp_test" 'MMP manual reset crossed release barrier' \
  "MMP manual recovery remains fail-closed while Asset release is non-terminal"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'FindFirstUnsafeMMPOrderForUpdate' \
  "MMP recovery locks the first trading-or-release unsafe order"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'ORDER_STATUS_CANCELING' \
  "MMP recovery treats cancellation funding as non-terminal"
check_contains "$OPTION_DIR/models/toptionordermodel.go" 'ORDER_STATUS_EXPIRING' \
  "MMP recovery treats expiry funding as non-terminal"
check_contains "$OPTION_DIR/internal/logic/app/app_order_match.go" 'controlReasonMMPTriggered, true' \
  "MMP trigger continues canceling later quotes after an individual failure"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260730_zg_option_mmp[.]sql' \
  "repeatable acceptance installs MMP state and immutability guards"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'mmp=' \
  "real Asset RPC gate reports MMP batch, release and audit evidence"
trade_correction_test="$OPTION_DIR/internal/logic/task/p1_trade_correction_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'erroneous trade cash correction debit barrier and recovery' \
  "real Asset RPC gate invokes erroneous-trade cash correction"
check_contains "$trade_correction_test" 'results := make[(]chan bool, 20[)]' \
  "trade-correction gate starts twenty independent concurrent reviews"
check_contains "$trade_correction_test" 'winners != 1' \
  "trade-correction gate requires exactly one review winner"
check_contains "$trade_correction_test" 'failOnceSubAvailableClient' \
  "trade-correction gate loses a response after a committed debit"
check_contains "$trade_correction_test" 'SELLER-CURE' \
  "trade-correction gate cures an insufficient debit balance"
check_contains "$trade_correction_test" 'originalTradeHash' \
  "trade-correction gate proves the original trade is unchanged"
check_contains "$trade_correction_test" 'correctionNet != "0[.]000000000000000000"' \
  "trade-correction gate requires exact Asset debit-credit conservation"
check_contains "$trade_correction_test" 'instructions != 3 [|][|] success != 3 [|][|] reconciled != 3 [|][|] flows != 3' \
  "trade-correction gate requires three unique reconciled money legs"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260730_zf_option_trade_correction[.]sql' \
  "repeatable acceptance installs trade-correction immutability guards"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'trade_correction=' \
  "real Asset RPC gate reports trade-correction evidence"
order_cancel_test="$OPTION_DIR/internal/logic/task/p0_order_cancel_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'user cancel IOC and FOK funding lifecycle' \
  "real Asset RPC gate invokes user-cancel and immediate-order scenarios"
check_contains "$order_cancel_test" 'NewCancelOrderLogic' \
  "order lifecycle gate enters through the public CancelOrder logic"
check_contains "$order_cancel_test" 'ASSET_INSTRUCTION_STATUS_CANCELED' \
  "pre-funding cancel proves the pending freeze is canceled before Asset execution"
check_contains "$order_cancel_test" 'ORDER_TYPE_IOC' \
  "order lifecycle gate covers an IOC partial fill"
check_contains "$order_cancel_test" 'IMMEDIATE_REMAINDER_CANCELED' \
  "IOC remainder reaches the explicit canceled terminal reason"
check_contains "$order_cancel_test" 'releaseAmount[.]Equal[(]decimal[.]RequireFromString[(]"10[.]4"[)][)]' \
  "IOC gate releases exactly the unused premium and fee reservation"
check_contains "$order_cancel_test" 'ORDER_TYPE_FOK' \
  "order lifecycle gate covers an insufficient-liquidity FOK"
check_contains "$order_cancel_test" 'FOK_NOT_FILLED' \
  "FOK gate proves the all-or-none terminal reason"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'user_cancel=' \
  "real Asset RPC gate reports user-cancel evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'ioc_partial=' \
  "real Asset RPC gate reports IOC partial-fill evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'fok_insufficient=' \
  "real Asset RPC gate reports FOK all-or-none evidence"
market_post_only_test="$OPTION_DIR/internal/logic/task/p0_order_market_postonly_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'market and post-only funding lifecycle' \
  "real Asset RPC gate invokes MARKET and POST_ONLY scenarios"
check_contains "$market_post_only_test" 'ProtectionPrice: "10", MaxTurnover: "12"' \
  "MARKET gate uses explicit protection price and spend cap"
check_contains "$market_post_only_test" 'RELEASE-REMAINDER' \
  "MARKET gate checks release of the unused spend reservation"
check_contains "$market_post_only_test" 'releaseAmount[.]Equal[(]decimal[.]RequireFromString[(]"1[.]6"[)][)]' \
  "MARKET gate proves the exact unused reservation amount"
check_contains "$market_post_only_test" 'POST_ONLY_WOULD_TAKE' \
  "POST_ONLY crossing order proves maker-only cancellation"
check_contains "$market_post_only_test" 'POST_ONLY non-crossing order did not rest' \
  "POST_ONLY non-crossing order must enter the book before user cancel"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'market_protection=' \
  "real Asset RPC gate reports MARKET protection evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'post_only=' \
  "real Asset RPC gate reports POST_ONLY maker-only evidence"
cancel_race_test="$OPTION_DIR/internal/logic/task/p0_order_cancel_concurrency_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'admin force cancel and funding race' \
  "real Asset RPC gate invokes administrative force-cancel and funding-race scenarios"
check_contains "$OPTION_DIR/internal/logic/app/cancelorderlogic.go" 'FindOneForUpdate' \
  "public cancellation revalidates the order under a database row lock"
check_contains "$OPTION_DIR/internal/logic/app/control_cancel.go" 'CancelOrderByControlWithAudit' \
  "control cancellation exposes the transactional immutable-audit path"
check_contains "$OPTION_DIR/internal/logic/admin/forcecancelcontractorderslogic.go" 'CancelOrderByControlWithAudit' \
  "administrative force cancel reuses the unified control-cancel state machine"
check_contains "$cancel_race_test" 'ADMIN_FORCE_CANCEL_ORDER' \
  "administrative force cancel records a per-order operator audit event"
check_contains "$cancel_race_test" 'for round := 0; round < rounds; round[+][+]' \
  "cancel-versus-funding acceptance repeats the concurrent race"
check_contains "$cancel_race_test" 'cancelSuccess != 1 [|][|] cancelRejected != 1' \
  "two concurrent cancellation requests have exactly one winner"
check_contains "$cancel_race_test" 'duplicateInstructionNos != 0' \
  "cancel-versus-funding acceptance rejects duplicate instruction identities"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'admin_force_cancel=' \
  "real Asset RPC gate reports administrative force-cancel evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'cancel_funding_race=' \
  "real Asset RPC gate reports cancel-versus-funding race evidence"
multi_instance_test="$OPTION_DIR/internal/logic/task/p0_asset_multi_instance_kill_rpc_integration_test.go"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'TestP0AssetMultiInstanceKillTakeover' \
  "real Asset RPC gate invokes the independent-process kill/takeover scenario"
check_contains "$multi_instance_test" 'Process[.]Kill[(][)]' \
  "multi-instance gate sends a real SIGKILL to the committed worker process"
check_contains "$multi_instance_test" 'waitP0TaskLeaseExpiry' \
  "multi-instance gate waits for the killed worker lease to expire naturally"
check_contains "$multi_instance_test" 'first := startP0AssetWorker' \
  "multi-instance gate starts the first fresh takeover worker"
check_contains "$multi_instance_test" 'second := startP0AssetWorker' \
  "multi-instance gate starts a competing fresh takeover worker"
check_contains "$multi_instance_test" 'takeoverProxy[.]calls[.]Load[(][)] != 1' \
  "only one competing takeover worker may reach Asset"
check_contains "$multi_instance_test" 'retryCount != 0' \
  "stale recovery must preserve the original retry count"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'multi_instance_kill=' \
  "real Asset RPC gate reports independent-process takeover evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'testP0IsolatedShortLiquidationAccounting' \
  "real Asset RPC gate covers isolated short liquidation accounting"
check_contains "$OPTION_DIR/internal/logic/task/processassetinstructionslogic.go" 'takeover.MaintenanceMargin = takeover.MaintenanceMargin.Add' \
  "liquidation takeover preserves maintenance margin evidence"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'allocateIsolatedLiquidationLots' \
  "isolated partial liquidation allocates margin lots proportionally"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'selectIsolatedLiquidationCandidate' \
  "risk scan selects an automatic isolated partial-liquidation quantity"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'RunTaskWithLock.*process_risk_accounts' \
  "full-tenant and replay risk scans retain one mutual-exclusion domain"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'results\[in[.]TenantId\].*RiskScanTenantResult' \
  "tenant-scoped risk scans publish an explicit zero-group result"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'must not advance the' \
  "post-evaluation execution failure cannot advance risk scan completion"
risk_isolation_test="$OPTION_DIR/internal/logic/task/p1_risk_scan_isolation_rpc_integration_test.go"
check_contains "$OPTION_DIR/internal/logic/task/p0_asset_rpc_integration_test.go" 'testP1RiskScanFailureIsolation' \
  "real Asset RPC gate invokes risk-wallet and cross-tenant failure isolation"
check_contains "$risk_isolation_test" 'injected per-wallet Asset failure' \
  "risk isolation gate injects a single-wallet Asset dependency failure"
check_contains "$risk_isolation_test" 'OptionTaskReq[{]TenantId: 0[}]' \
  "risk isolation gate uses the production full-tenant scan scope"
check_contains "$risk_isolation_test" 'RISK_ACCOUNT_STATUS_RESTRICTED, false' \
  "risk isolation gate requires failed wallets to remain uncalculated and restricted"
check_contains "$OPTION_DIR/internal/observability/risk_metrics_test.go" 'WithoutClearingOthers' \
  "tenant-scoped risk metric publication cannot clear an unrelated tenant"
check_contains "$OPTION_DIR/internal/observability/risk_metrics_test.go" 'GlobalScanClearsMissingTenant' \
  "full-tenant risk metric publication clears disappeared tenant series"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'risk_failure_isolation=' \
  "real Asset RPC gate reports wallet and tenant risk-isolation evidence"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'Div[(]reliefPerStep[)][.]Floor[(][)][.]Add' \
  "partial-liquidation sizing strictly crosses the maintenance boundary"
check_contains "$OPTION_DIR/models/toptionliquidationmodel.go" 'FindOpenByWallet' \
  "liquidations are serialized across the whole settlement wallet"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'isInsuranceInventoryPosition' \
  "insurance takeover inventory cannot recursively enter customer liquidation"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'position[.]FrozenQty = position[.]FrozenQty[.]Add[(]plan[.]quantity[)]' \
  "liquidation reserves source position quantity before funding"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'isolated_liquidation=' \
  "real Asset RPC gate reports isolated liquidation evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'partial_liquidation=' \
  "real Asset RPC gate reports proportional partial-liquidation evidence"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'testP0LiquidationDeficitFailureRecovery' \
  "real Asset RPC gate covers insurance and platform-backstop deficit recovery"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'insurance response loss after commit' \
  "liquidation gate injects committed insurance-cover response loss"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'platform backstop response loss after commit' \
  "liquidation gate injects committed platform-backstop response loss"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'failAfterCommit: true' \
  "liquidation gate injects committed collateral-debit response loss"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'deficit_liquidation=' \
  "real Asset RPC gate reports liquidation deficit-recovery evidence"
portfolio_liquidation_migration="$OPTION_DIR/migrations/20260801_option_portfolio_liquidation.sql"
check_contains "$OPTION_DIR/option.sql" 'liquidation_scope.*组合钱包' \
  "Option schema records isolated versus portfolio-wallet liquidation scope"
check_contains "$OPTION_DIR/models/toptionliquidationmodel_gen.go" 'PortfolioMaintenanceBefore.*portfolio_maintenance_before' \
  "make gen-model synchronized portfolio liquidation evidence"
check_contains "$REPO_ROOT/proto/option/model.proto" 'portfolio_collateral_after = 34' \
  "Option RPC exposes portfolio liquidation collateral proof"
check_contains "$portfolio_liquidation_migration" 'trg_option_liquidation_evidence_insert' \
  "portfolio liquidation migration validates evidence on insert"
check_contains "$portfolio_liquidation_migration" 'liquidation identity and portfolio evidence are immutable' \
  "portfolio liquidation migration protects immutable evidence"
check_contains "$OPTION_DIR/schema/90_constraints.sql" 'status.*IN [(]1,2,3,4,5,6,7[)]' \
  "canonical liquidation constraint permits stale-snapshot CANCELED status"
check_contains "$portfolio_liquidation_migration" 'status.*IN [(]1,2,3,4,5,6,7[)]' \
  "portfolio liquidation migration upgrades the CANCELED status constraint"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'selectPortfolioLiquidationCandidate' \
  "risk scan deterministically selects a risk-reducing portfolio position"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'selectPortfolioLiquidationQuantity' \
  "portfolio liquidation selects quantity with the governed scenario model"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'equityAfter[.]GreaterThan[(]maintenanceAfterTotal[)]' \
  "portfolio partial liquidation strictly restores wallet maintenance health"
check_not_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'partial portfolio liquidation is not supported' \
  "portfolio liquidation execution accepts a governed partial quantity"
check_contains "$OPTION_DIR/internal/logic/task/processriskaccountslogic.go" 'collateralAfter[.]LessThan[(]initialAfter[.]Requirement[)]' \
  "portfolio liquidation creation protects residual initial requirement"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'errPortfolioLiquidationSnapshotStale' \
  "portfolio liquidation cancels and rebuilds stale risk snapshots"
check_contains "$OPTION_DIR/models/toptionoutboxmodel.go" 'HasIncompletePortfolioForWallet' \
  "portfolio liquidation trade-event barrier is scoped to the affected wallet"
check_contains "$OPTION_DIR/models/toptionassetinstructionmodel.go" 'HasIncompleteForWallet' \
  "portfolio liquidation asset barrier covers all affected wallet instructions"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'wallet recovered above current maintenance requirement' \
  "portfolio liquidation stops when current wallet equity has recovered"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" 'residualCollateralFloor' \
  "portfolio liquidation rechecks the higher current collateral floor inside preparation"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'current residual requirement did not conservatively reduce collateral use' \
  "real Asset RPC gate covers a post-trigger market change and higher residual requirement"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'wallet recovered above current maintenance requirement' \
  "real Asset RPC gate covers canceling liquidation after wallet recovery"
check_contains "$OPTION_DIR/models/toptionliquidationmodel.go" 'ExecCtx[(]ctx' \
  "stale portfolio liquidation cancellation invalidates cached status"
check_contains "$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go" 'testP0PortfolioLiquidationSequential' \
  "real Asset RPC gate covers sequential portfolio liquidation"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'portfolio_liquidation_sequential=' \
  "real Asset RPC gate reports sequential portfolio-liquidation evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'portfolio_partial_liquidation=' \
  "real Asset RPC gate reports portfolio partial-liquidation evidence"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'portfolio_liquidation_recovery_cancel=' \
  "real Asset RPC gate reports wallet-recovery cancellation evidence"
check_contains "$OPTION_DIR/docs/templates/option-exercise-expiry-control-record.md" '空头 gross 扣款 = 多头净入账 [+] 行权费' \
  "operations pack includes an exercise and expiry conservation record"

insurance_flow_migration="$OPTION_DIR/migrations/20260802_option_insurance_fund_flow_semantics.sql"
insurance_flow_test="$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go"
check_contains "$OPTION_DIR/option.sql" '业务绝对金额，方向由flow_type确定' \
  "Option schema declares positive insurance-flow magnitudes"
check_contains "$OPTION_DIR/models/toptioninsurancefundflowmodel_gen.go" \
  '业务绝对金额，方向由flow_type确定' \
  "make gen-model output matches insurance-flow magnitude semantics"
check_contains "$insurance_flow_migration" 'Existing signed rows are intentionally not rewritten' \
  "insurance-flow rollout preserves historical evidence"
check_contains "$insurance_flow_migration" 'trg_option_insurance_fund_flow_validate_insert' \
  "insurance-flow migration validates new economic evidence"
check_contains "$insurance_flow_migration" 'trg_option_insurance_fund_flow_no_update' \
  "insurance-flow migration rejects economic updates"
check_contains "$insurance_flow_migration" 'trg_option_insurance_fund_flow_no_delete' \
  "insurance-flow migration rejects economic deletes"
check_contains "$OPTION_DIR/models/toptioninsurancefundflowmodel.go" \
  'optionInsuranceFundSignedAmountSQL.*flow_type IN [(]2,4[)].*-ABS[(]amount[)]' \
  "insurance-flow readers derive direction from flow type"
check_contains "$OPTION_DIR/models/toptioninsurancefundflowmodel.go" \
  'amount must be a positive magnitude' \
  "insurance-flow writer rejects signed and zero new rows"
check_contains "$OPTION_DIR/models/optionoperationsmodel.go" \
  'optionInsuranceFundSignedAmountSQL' \
  "operations overview uses normalized insurance-flow direction"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" \
  'optionInsuranceFundSignedAmountSQL' \
  "Prometheus metrics use normalized insurance-flow direction"
check_contains "$insurance_flow_test" 'signedAmount.Equal[(]decimal.NewFromInt[(]-15[)][)]' \
  "real Asset RPC gate proves a deficit cover is an outflow"
check_contains "$insurance_flow_test" 'insurance fund flow update was not rejected' \
  "real Asset RPC gate proves insurance-flow update immutability"
check_contains "$insurance_flow_test" 'insurance fund flow delete was not rejected' \
  "real Asset RPC gate proves insurance-flow delete immutability"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" \
  '20260802_option_insurance_fund_flow_semantics[.]sql' \
  "real Asset RPC gate installs insurance-flow semantics twice"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'insurance_ledger_semantics=' \
  "real Asset RPC gate reports machine-readable insurance-ledger evidence"
check_contains "$OPTION_DIR/docs/option-p1-008-insurance-fund-ledger-repository-acceptance.md" \
  'REPOSITORY_PASSED / PREPROD_BLOCKED' \
  "insurance-ledger acceptance preserves production approval blockers"
check_not_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  '当前未发现新的仓库级资金正确性阻断' \
  "current status removes the stale no-new-funding-blocker conclusion"
check_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  'P2-008 RFQ/大宗/做市义务保持 `DEFERRED`' \
  "current status keeps deferred institutional capabilities closed"
operations_pack="$OPTION_DIR/docs/option-product-operations-pack.md"
launch_checklist="$OPTION_DIR/docs/option-contract-launch-checklist.md"
check_contains "$operations_pack" \
  'OPTION_OPERATIONS_PACK_STATUS: INTERNAL_TEMPLATE_NOT_APPROVED_FOR_PUBLICATION' \
  "operations pack is explicitly an internal unpublished template"
check_contains "$operations_pack" \
  '最后交易时间、主动行权/到期指令截止、到期时间和结算/交割时间是独立边界' \
  "operations pack explains independent lifecycle boundaries"
check_contains "$operations_pack" \
  '不是用户余额' \
  "operations pack does not present insurance funds as a user guarantee"
check_contains "$operations_pack" \
  '发布负责人必须对最终发布副本执行占位符检查' \
  "operations pack makes unresolved placeholders a publication blocker"
check_contains "$launch_checklist" \
  'OPTION_LAUNCH_CHECKLIST_STATUS: DRAFT' \
  "contract launch checklist defaults to draft"
check_contains "$OPTION_DIR/monitoring/option-production-readiness.env.example" \
  '^OPTION_LAUNCH_CHECKLIST=' \
  "production readiness requires an immutable contract launch checklist"
check_contains "$OPTION_DIR/monitoring/option-production-readiness.sh" \
  'OPTION_LAUNCH_CHECKLIST_STATUS.*APPROVED' \
  "production gate rejects a draft contract launch checklist"
check_contains "$launch_checklist" \
  'option-release-scope[.]sh --release-clean' \
  "contract launch checklist requires a clean reviewed release identity"
check_contains "$OPTION_DIR/monitoring/option-release-scope.sh" \
  '20260802_asset_backstop_policy_limits[.]sql' \
  "release scope includes the Asset platform-backstop schema dependency"
check_contains "$OPTION_DIR/monitoring/option-release-scope.sh" \
  'proto/asset/asset[.]proto' \
  "release scope includes the Asset protocol dependency"
check_contains "$OPTION_DIR/monitoring/option-release-scope.sh" \
  'asset_platform_backstop_policy_permissions[.]sql' \
  "release scope includes the platform-backstop RBAC dependency"
check_contains "$launch_checklist" \
  'make gen-model.*make gen' \
  "contract launch checklist requires both model and protocol generation"
check_contains "$launch_checklist" \
  '18 项 Option 监控索引和 61 条 Option 告警规则' \
  "contract launch checklist uses current monitoring counts"
check_contains "$launch_checklist" \
  '正绝对金额[+]类型方向' \
  "contract launch checklist includes insurance-ledger production approval"
check_not_contains "$launch_checklist" \
  '产品依赖项均为 `DONE`' \
  "contract launch checklist does not confuse repository status with production approval"
insurance_ledger_approval="$OPTION_DIR/docs/templates/option-insurance-fund-ledger-production-approval.md"
check_contains "$insurance_ledger_approval" \
  'OPTION_INSURANCE_LEDGER_APPROVAL_STATUS: DRAFT' \
  "insurance-ledger production approval defaults to draft"
check_contains "$insurance_ledger_approval" \
  'CASE WHEN flow_type IN [(]2,4[)] THEN -ABS[(]amount[)] ELSE ABS[(]amount[)] END' \
  "insurance-ledger production approval embeds the canonical signed expression"
check_contains "$insurance_ledger_approval" \
  't_asset_platform_flow[.]biz_no' \
  "insurance-ledger production approval records the Asset business-identity bridge"
check_contains "$insurance_ledger_approval" \
  '下游消费者盘点' \
  "insurance-ledger production approval requires downstream consumer inventory"
check_contains "$insurance_ledger_approval" \
  '禁止修改/删除原保险流水' \
  "insurance-ledger production approval prohibits history rewriting"
check_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  'option-insurance-fund-ledger-production-approval[.]md' \
  "current blockers link the insurance-ledger production approval"
check_contains "$launch_checklist" \
  'option-insurance-fund-ledger-production-approval[.]md' \
  "contract launch checklist requires approved insurance-ledger evidence"
check_contains "$OPTION_DIR/docs/templates/option-production-readiness-signoff.md" \
  'option-insurance-fund-ledger-production-approval[.]md.*APPROVED' \
  "production signoff requires approved insurance-ledger evidence"
physical_default_policy="$OPTION_DIR/docs/templates/option-physical-delivery-default-policy-approval.md"
physical_default_record="$OPTION_DIR/docs/templates/option-physical-delivery-default-record.md"
check_contains "$physical_default_policy" \
  'OPTION_PHYSICAL_DEFAULT_POLICY_STATUS: DRAFT' \
  "physical-default policy approval defaults to draft"
check_contains "$physical_default_policy" \
  '多头应交资产扣可用.*空头担保扣冻结.*多头应收资产入账.*空头应收资产入账' \
  "physical-default policy preserves the four-leg debit barrier"
check_contains "$physical_default_policy" \
  '当前代码没有自动退款、罚款、拍卖、平台承接或放弃行权路径' \
  "physical-default policy does not claim unsupported disposition paths"
check_contains "$physical_default_policy" \
  'D1 补资后继续原交割' \
  "physical-default policy requires an explicit disposition decision"
check_contains "$physical_default_policy" \
  '真实编排器下强杀Option/Asset/队列容器' \
  "physical-default policy requires orchestrator failure acceptance"
check_contains "$physical_default_record" \
  'OPTION_PHYSICAL_DEFAULT_RECORD_STATUS: OPEN' \
  "physical-default incident record starts open"
check_contains "$physical_default_record" \
  'option-physical-delivery-default-policy-approval[.]md' \
  "physical-default incident record links the approved policy"
check_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  'option-physical-delivery-default-policy-approval[.]md' \
  "current blockers link the physical-default policy approval"
check_contains "$launch_checklist" \
  'option-physical-delivery-default-policy-approval[.]md.*APPROVED' \
  "contract launch checklist requires approved physical-default policy"
check_contains "$OPTION_DIR/docs/templates/option-production-readiness-signoff.md" \
  'option-physical-delivery-default-policy-approval[.]md.*APPROVED' \
  "production signoff requires approved physical-default policy"
check_contains "$operations_pack" \
  'option-physical-delivery-default-policy-approval[.]md' \
  "operations pack binds physical-default notices to approved policy"
database_audit_template="$OPTION_DIR/docs/templates/option-database-security-audit-acceptance.md"
check_contains "$database_audit_template" \
  'OPTION_DATABASE_AUDIT_STATUS: DRAFT' \
  "database-security audit template defaults to draft"
check_contains "$database_audit_template" \
  '失败事务中的应用审计写入会一起回滚' \
  "database-security audit requires transaction-external evidence"
check_contains "$database_audit_template" \
  'DB-AUD-012' \
  "database-security audit covers direct DML and DDL bypass attempts"
check_contains "$database_audit_template" \
  'DB-AUD-N05' \
  "database-security audit distinguishes direct SQL from application rejection"
check_contains "$database_audit_template" \
  '任何未授权写入成功、审计事件缺失、身份无法归属' \
  "database-security audit treats successful bypass and missing evidence as incidents"
check_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  'option-database-security-audit-acceptance[.]md' \
  "current blockers link the database-security audit acceptance"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  'option-database-security-audit-acceptance[.]md' \
  "preproduction evidence pack requires the database-security audit acceptance"
check_contains "$OPTION_DIR/docs/templates/option-production-readiness-signoff.md" \
  'option-database-security-audit-acceptance[.]md.*APPROVED' \
  "production signoff requires approved database-security audit evidence"
platform_backstop_template="$OPTION_DIR/docs/templates/option-platform-backstop-policy-approval.md"
platform_backstop_e2e_template="$OPTION_DIR/docs/templates/option-platform-backstop-e2e.md"
check_contains "$platform_backstop_template" \
  'OPTION_PLATFORM_BACKSTOP_APPROVAL_STATUS: DRAFT' \
  "platform-backstop policy template defaults to draft"
check_contains "$platform_backstop_template" '无限负余额.*不得批准' \
  "platform-backstop policy prohibits unlimited negative funding"
check_contains "$platform_backstop_template" 'BST-012' \
  "platform-backstop policy covers limits, concurrency, replay and isolation"
check_contains "$platform_backstop_e2e_template" \
  'OPTION_PLATFORM_BACKSTOP_E2E_STATUS: DRAFT' \
  "platform-backstop target-environment E2E template defaults to draft"
check_contains "$platform_backstop_e2e_template" 'BST-012' \
  "platform-backstop target-environment E2E template covers final isolation"
check_contains "$OPTION_DIR/monitoring/option-production-readiness.env.example" \
  'OPTION_PLATFORM_BACKSTOP_POLICY_ID=0' \
  "production readiness example binds an exact Asset policy ID"
check_contains "$OPTION_DIR/monitoring/option-production-readiness.env.example" \
  'OPTION_PLATFORM_BACKSTOP_POLICY_VERSION=0' \
  "production readiness example binds an exact Asset policy version"
check_contains "$OPTION_DIR/internal/config/config.go" 'PlatformBackstop struct' \
  "Option config has an independent platform-backstop runtime gate"
check_contains "$OPTION_DIR/etc/option.yaml" 'PlatformBackstop:' \
  "Option runtime example declares the platform-backstop gate"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" \
  'PlatformBackstop[.]Enabled' \
  "liquidation execution requires the platform-backstop runtime gate"
check_contains "$OPTION_DIR/internal/logic/task/processliquidationslogic.go" \
  'platform backstop runtime gate is disabled; manual deficit resolution required' \
  "disabled platform backstop routes unresolved deficits to manual review"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" \
  'platformBackstopRuntimeEnabled' \
  "contract listing refuses a disabled platform-backstop policy"
check_contains "$OPTION_DIR/internal/logic/admin/resumecontracttradinglogic.go" \
  'PlatformBackstop[.]Enabled' \
  "contract resume refuses a disabled platform-backstop policy"
check_contains "$OPTION_DIR/docs/option-current-status-and-production-blockers.md" \
  'OPT-P0-007' \
  "current status tracks the platform-backstop repository and production boundary"
check_contains "$OPTION_DIR/docs/option-p0-007-repository-acceptance.md" \
  'REPOSITORY_ACCEPTANCE_STATUS: PASSED / PREPROD_BLOCKED' \
  "platform-backstop repository acceptance passed without claiming production readiness"
check_contains "$REPO_ROOT/services/asset/internal/logic/asset/coverplatformbackstopdeficitlogic.go" \
  'SubAvailableWithFloor' \
  "Asset backstop runtime enforces an atomic approved balance floor"
check_not_contains "$REPO_ROOT/services/asset/models/tassetplatformaccountmodel.go" \
  'SubAvailableAllowNegative' \
  "Asset platform-account model has no unlimited-negative balance method"
check_not_contains "$REPO_ROOT/services/asset/internal/logic/asset/coverplatformbackstopdeficitlogic.go" \
  'SubAvailableAllowNegative' \
  "Asset backstop cover path does not call an unlimited-negative balance method"
check_contains "$OPTION_DIR/acceptance/run-platform-backstop-rpc-acceptance.sh" \
  'TestPlatformBackstopRPCLimitsMySQL' \
  "platform-backstop acceptance runs real RPC boundary and concurrency cases"
blocker_material_matrix="$OPTION_DIR/docs/option-production-blocker-evidence-matrix.md"
check_contains "$blocker_material_matrix" \
  'MATRIX_STATUS: REPOSITORY_MATERIALS_READY / PRODUCTION_BLOCKED' \
  "production blocker material matrix remains explicitly blocked"
check_contains "$blocker_material_matrix" 'OPT-P0-007' \
  "production blocker matrix tracks the platform-backstop external approval gap"
check_contains "$blocker_material_matrix" 'ORCHESTRATOR_TAKEOVER_REPORT' \
  "production blocker matrix tracks the orchestrator evidence gap"
check_contains "$blocker_material_matrix" 'BEANSTALK_CAPACITY_REPORT' \
  "production blocker matrix tracks the native-capacity evidence gap"
completion_audit="$OPTION_DIR/docs/option-completion-audit.md"
check_contains "$completion_audit" \
  '^AUDIT_STATUS: REPOSITORY_ACTIONS_COMPLETE / PREPROD_BLOCKED$' \
  "completion audit separates repository completion from production readiness"
check_contains "$completion_audit" 'make gen-model' \
  "completion audit makes model generation part of DDL acceptance"
check_contains "$completion_audit" 'option-production-evidence-report[.]md' \
  "completion audit links the generic target-environment evidence template"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  'option-platform-backstop-policy-approval[.]md' \
  "preproduction evidence pack separates disabled and enabled backstop evidence"
check_contains "$OPTION_DIR/docs/templates/option-production-readiness-signoff.md" \
  'option-platform-backstop-policy-approval[.]md.*APPROVED' \
  "production signoff requires approved platform-backstop policy when enabled"
orchestrator_report_template="$OPTION_DIR/docs/templates/option-orchestrator-takeover-report.md"
check_contains "$orchestrator_report_template" \
  'OPTION_ORCHESTRATOR_TAKEOVER_STATUS: DRAFT' \
  "orchestrator takeover report defaults to draft"
check_contains "$orchestrator_report_template" 'ORC-010' \
  "orchestrator takeover report covers the complete failure matrix"
check_contains "$orchestrator_report_template" '不删除或缩短Redis/Etcd租约' \
  "orchestrator takeover report requires natural lease expiry"
beanstalk_capacity_template="$OPTION_DIR/docs/templates/option-beanstalk-capacity-rto-report.md"
check_contains "$beanstalk_capacity_template" \
  'OPTION_BEANSTALK_CAPACITY_STATUS: DRAFT' \
  "Beanstalk capacity report defaults to draft"
check_contains "$beanstalk_capacity_template" 'BS-010' \
  "Beanstalk capacity report covers native architecture, WAL, kill and long-run cases"
check_contains "$beanstalk_capacity_template" '是否使用模拟/emulation（必须否）' \
  "Beanstalk capacity report prohibits emulated architecture evidence"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  'option-orchestrator-takeover-report[.]md' \
  "preproduction evidence pack requires the orchestrator takeover report"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  'option-beanstalk-capacity-rto-report[.]md' \
  "preproduction evidence pack requires the Beanstalk capacity report"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  '18 项监控索引存在' \
  "preproduction evidence pack uses the current monitoring-index count"
check_not_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  '17 项监控索引' \
  "preproduction evidence pack removed the stale monitoring-index count"
generic_evidence_template="$OPTION_DIR/docs/templates/option-production-evidence-report.md"
check_contains "$generic_evidence_template" \
  '^OPTION_EVIDENCE_STATUS:[[:space:]]*DRAFT$' \
  "generic production-evidence template defaults to draft"
check_contains "$OPTION_DIR/monitoring/option-production-readiness.env.example" \
  '^OPTION_INSURANCE_LEDGER_APPROVAL=' \
  "production readiness declares the seller insurance-ledger approval"
check_contains "$OPTION_DIR/docs/templates/option-alert-delivery-test.md" \
  '^OPTION_ALERT_DELIVERY_STATUS:[[:space:]]*DRAFT$' \
  "alert-delivery template defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-daily-fund-reconciliation.md" \
  '^OPTION_DAILY_RECONCILIATION_STATUS:[[:space:]]*DRAFT$' \
  "daily-reconciliation template defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-market-freshness-approval.md" \
  '^OPTION_MARKET_FRESHNESS_APPROVAL_STATUS:[[:space:]]*DRAFT$' \
  "market-freshness approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/trading-calendar-approval.md" \
  '^OPTION_TRADING_CALENDAR_APPROVAL_STATUS:[[:space:]]*DRAFT$' \
  "trading-calendar approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/trading-calendar-annual-review.md" \
  '^OPTION_TRADING_CALENDAR_ANNUAL_REVIEW_STATUS:[[:space:]]*DRAFT$' \
  "trading-calendar annual review defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-portfolio-risk-validation-record.md" \
  '^OPTION_PORTFOLIO_MODEL_VALIDATION_STATUS:[[:space:]]*DRAFT$' \
  "portfolio-model validation defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-insurance-inventory-exit-approval.md" \
  '^OPTION_INSURANCE_INVENTORY_EXIT_APPROVAL_STATUS:[[:space:]]*DRAFT$' \
  "insurance-inventory exit approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-insurance-inventory-exit-execution-record.md" \
  '^OPTION_INSURANCE_INVENTORY_EXIT_EXECUTION_STATUS:[[:space:]]*DRAFT$' \
  "insurance-inventory exit execution record defaults to draft"
check_contains "$OPTION_DIR/docs/templates/complex-order-readiness.md" \
  '^OPTION_COMPLEX_ORDER_READINESS_STATUS:[[:space:]]*DRAFT$' \
  "complex-order readiness record defaults to draft"
check_contains "$OPTION_DIR/docs/templates/public-market-readiness.md" \
  '^OPTION_PUBLIC_MARKET_READINESS_STATUS:[[:space:]]*DRAFT$' \
  "public-market readiness record defaults to draft"
check_contains "$OPTION_DIR/docs/templates/contract-series-approval.md" \
  '^OPTION_CONTRACT_SERIES_APPROVAL_STATUS:[[:space:]]*DRAFT$' \
  "contract-series approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/contract-series-approval.md" \
  '上市 UTC.*最后交易 UTC.*行权截止 UTC.*到期 UTC.*交割 UTC' \
  "contract-series approval exposes all five lifecycle times"
check_not_contains "$OPTION_DIR/docs/templates/contract-series-approval.md" \
  '到期/最后交易 UTC|list < exercise_cutoff' \
  "contract-series approval does not collapse last-trade and expiry"
check_contains "$OPTION_DIR/docs/templates/settlement-price-approval.md" \
  '^OPTION_SETTLEMENT_PRICE_APPROVAL_STATUS:[[:space:]]*DRAFT$' \
  "settlement-price approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/option-exercise-expiry-control-record.md" \
  '^OPTION_EXERCISE_EXPIRY_CONTROL_STATUS:[[:space:]]*DRAFT$' \
  "exercise-expiry control record defaults to draft"
check_contains "$OPTION_DIR/docs/templates/cash-settlement-topup-recovery-approval.md" \
  '^OPTION_CASH_SETTLEMENT_TOPUP_STATUS:[[:space:]]*DRAFT$' \
  "cash-settlement top-up approval defaults to draft"
check_contains "$OPTION_DIR/docs/templates/corporate-action-case.md" \
  '^OPTION_CORPORATE_ACTION_CASE_STATUS:[[:space:]]*DRAFT$' \
  "corporate-action case defaults to draft"
check_contains "$OPTION_DIR/docs/templates/daily-reconciliation.md" \
  '^OPTION_DAILY_OPERATIONS_RECONCILIATION_STATUS:[[:space:]]*DRAFT$' \
  "daily operations reconciliation defaults to draft"
check_contains "$OPTION_DIR/docs/templates/incident-report.md" \
  '^OPTION_INCIDENT_STATUS:[[:space:]]*OPEN$' \
  "incident report starts open"
check_contains "$OPTION_DIR/docs/templates/institutional-market-readiness.md" \
  '^OPTION_INSTITUTIONAL_MARKET_STATUS:[[:space:]]*DEFERRED$' \
  "institutional market readiness remains deferred"
check_contains "$OPTION_DIR/docs/templates/risk-parameter-change.md" \
  '^OPTION_RISK_PARAMETER_CHANGE_STATUS:[[:space:]]*DRAFT$' \
  "risk-parameter change record defaults to draft"
check_contains "$OPTION_DIR/docs/templates/trading-halt-record.md" \
  '^OPTION_TRADING_HALT_STATUS:[[:space:]]*OPEN$' \
  "trading-halt record starts open"
check_contains "$OPTION_DIR/docs/templates/option-preproduction-evidence-pack.md" \
  '^OPTION_PREPRODUCTION_EVIDENCE_PACK_STATUS:[[:space:]]*DRAFT$' \
  "preproduction evidence pack defaults to draft"
alert_delivery_template="$OPTION_DIR/docs/templates/option-alert-delivery-test.md"
for alert_id in 005 007 008 009 011 015 016 020 026 029 030; do
  check_contains "$alert_delivery_template" "OPT-A$alert_id" \
    "alert-delivery template uses canonical OPT-A$alert_id catalog identity"
done
consistency_docs="
$OPTION_DIR/acceptance/README.md
$OPTION_DIR/docs/option-acceptance-test-plan.md
$OPTION_DIR/docs/option-design-review.md
$OPTION_DIR/docs/option-remediation-plan.md
$OPTION_DIR/docs/option-p2-002-p2-003-lifecycle-repository-acceptance.md
"
for consistency_doc in $consistency_docs; do
  check_not_contains "$consistency_doc" '118[.]520|1m41[.]080' \
    "$(basename "$consistency_doc") removed superseded acceptance timings"
done
current_status_docs="
$OPTION_DIR/docs/option-design-review.md
$OPTION_DIR/docs/option-p1-004-repository-acceptance.md
$OPTION_DIR/docs/option-p1-005-insurance-inventory-exit-design.md
$OPTION_DIR/docs/option-p2-001-trading-calendar-repository-acceptance.md
$OPTION_DIR/docs/option-operations-runbook.md
"
for current_status_doc in $current_status_docs; do
  check_not_contains "$current_status_doc" 'VERIFYING' \
    "$(basename "$current_status_doc") uses the unified repository/preproduction status"
done
check_contains "$OPTION_DIR/docs/option-design-review.md" \
  '当前共有 195 个顶层 `Test[*]` 测试' \
  "design review records the current top-level Go test count"
check_not_contains "$OPTION_DIR/docs/option-design-review.md" \
  '费用/部分数量/多账户和容量仍待验收|当前缺口赔付写入正数|表定义要求入金为正、出金为负' \
  "design review removed superseded cash-expiry and insurance-ledger gaps"
check_contains "$OPTION_DIR/docs/option-p2-001-trading-calendar-repository-acceptance.md" \
  'instructions=9277 success=9270 canceled=7 reconciled=9270' \
  "calendar acceptance references the latest full repository gate"
check_not_contains "$OPTION_DIR/docs/option-p1-008-insurance-fund-ledger-repository-acceptance.md" \
  '新增17项检查' \
  "insurance-ledger acceptance avoids a stale readiness-check count"
repository_evidence="$OPTION_DIR/docs/evidence/option-repository-technical-evidence-20260802.md"
repository_evidence_manifest="$OPTION_DIR/docs/evidence/option-repository-technical-evidence-20260802.sha256"
check_contains "$repository_evidence" 'EVIDENCE_SCOPE: REPOSITORY_ONLY' \
  "repository evidence cannot be presented as preproduction evidence"
check_contains "$repository_evidence" 'EVIDENCE_STATUS: PASSED_NOT_RELEASE_CANDIDATE' \
  "repository evidence records the missing immutable release identity"
check_contains "$OPTION_DIR/docs/templates/option-production-readiness-signoff.md" \
  '仓库技术基线（不计生产放行）' \
  "production signoff excludes repository-only evidence from release approval"
check_sha256_manifest "$repository_evidence_manifest" \
  "repository technical evidence manifest matches its reviewed artifacts"

insurance_exit_migration="$OPTION_DIR/migrations/20260802_option_insurance_inventory_exit.sql"
insurance_exit_permissions="$REPO_ROOT/services/system/migrations/20260802_option_insurance_inventory_exit_permissions.sql"
insurance_exit_test="$OPTION_DIR/internal/logic/task/p0_liquidation_rpc_integration_test.go"
check_contains "$OPTION_DIR/option.sql" 'CREATE TABLE `t_option_insurance_inventory_exit`' \
  "Option schema declares governed insurance inventory exits"
check_contains "$OPTION_DIR/option.sql" 'uk_option_insurance_exit_active.*tenant_id.*active_key' \
  "Option schema serializes active exits per insurance position"
check_contains "$insurance_exit_migration" 'invalid insurance inventory exit active key' \
  "insurance-exit migration protects the active-position unique key"
check_contains "$insurance_exit_migration" 'insurance inventory exit economic fields are immutable' \
  "insurance-exit migration protects immutable request economics"
check_contains "$insurance_exit_migration" 'insurance inventory exit requires independent review' \
  "insurance-exit migration enforces four-eyes review"
check_contains "$insurance_exit_migration" 'submitted insurance inventory exit requires order evidence' \
  "insurance-exit migration requires a linked order before submission"
check_contains "$OPTION_DIR/models/toptioninsuranceinventoryexitmodel_gen.go" 'ActiveKey.*active_key' \
  "make gen-model synchronized the insurance-exit active key"
check_contains "$REPO_ROOT/proto/option/option.proto" 'rpc CreateInsuranceInventoryExit' \
  "Option RPC exposes insurance-exit creation"
check_contains "$REPO_ROOT/proto/option/option.proto" 'rpc ReviewInsuranceInventoryExit' \
  "Option RPC exposes independent insurance-exit review"
check_contains "$REPO_ROOT/proto/option/option.proto" 'rpc ExecuteInsuranceInventoryExit' \
  "Option RPC exposes approved insurance-exit execution"
check_contains "$REPO_ROOT/proto/option/model.proto" 'OrderStatus order_status = 23' \
  "insurance-exit responses expose live linked-order status"
check_contains "$OPTION_DIR/internal/logic/admin/listinsuranceinventoryexitslogic.go" 'ApplyInsuranceInventoryExitOrder' \
  "insurance-exit list derives fill progress from the linked order"
check_contains "$OPTION_DIR/internal/logic/app/placeorderlogic.go" 'NewAdministrativePlaceOrderLogic' \
  "insurance exits reuse the normal order path with ADMIN source"
check_contains "$OPTION_DIR/internal/logic/admin/executeinsuranceinventoryexitlogic.go" 'ORDER_TYPE_IOC' \
  "insurance exits submit an IOC order"
check_contains "$OPTION_DIR/internal/logic/admin/executeinsuranceinventoryexitlogic.go" 'YES_NO_YES' \
  "insurance exits are reduce-only"
check_contains "$OPTION_DIR/internal/config/config.go" 'InsuranceInventoryExit struct' \
  "Option runtime declares a fail-closed insurance-exit feature switch"
check_contains "$OPTION_DIR/etc/option.yaml" 'InsuranceInventoryExit:' \
  "Option runtime configuration defaults insurance exits explicitly"
check_contains "$OPTION_DIR/etc/option.yaml" 'Enabled: false' \
  "Option runtime keeps insurance exits disabled by default"
check_contains "$OPTION_DIR/internal/config/config.go" 'MaxQuantityPerRequest string' \
  "insurance exits declare a hard per-request quantity limit"
check_contains "$OPTION_DIR/internal/config/config.go" 'MaxPremiumPerRequest.*string' \
  "insurance exits declare a hard per-request premium limit"
check_contains "$OPTION_DIR/internal/config/config.go" 'MaxDailyQuantity.*string' \
  "insurance exits declare a hard UTC daily quantity limit"
check_contains "$OPTION_DIR/internal/config/config.go" 'MaxMarkDeviationRatio.*string' \
  "insurance exits declare a hard mark-deviation limit"
check_contains "$OPTION_DIR/internal/config/config.go" 'MinOrderBookQuantity.*string' \
  "insurance exits declare a minimum executable-depth limit"
check_contains "$OPTION_DIR/models/toptioninsuranceinventoryexitmodel.go" 'SumReservedQuantity' \
  "insurance exits aggregate open reservations and UTC-day submissions"
check_contains "$OPTION_DIR/internal/logic/admin/executeinsuranceinventoryexitlogic.go" 'validateInsuranceInventoryExitOrderBookDepth' \
  "insurance exits recheck executable depth before submitting IOC"
check_contains "$OPTION_DIR/internal/logic/admin/createinsuranceinventoryexitlogic.go" 'insuranceInventoryExitRuntimeLimits' \
  "insurance-exit creation fails closed on the runtime switch and limits"
check_contains "$OPTION_DIR/internal/logic/admin/reviewinsuranceinventoryexitlogic.go" 'insuranceInventoryExitRuntimeLimits' \
  "insurance-exit review fails closed on the runtime switch and limits"
check_contains "$OPTION_DIR/internal/logic/admin/executeinsuranceinventoryexitlogic.go" 'insuranceInventoryExitRuntimeLimits' \
  "insurance-exit execution fails closed on the runtime switch and limits"
check_contains "$insurance_exit_test" 'concurrent insurance exit execution response' \
  "real Asset RPC gate executes the same approved exit concurrently"
check_contains "$insurance_exit_test" 'instructionFlows != 4' \
  "insurance-exit gate proves instruction-to-Asset-flow cardinality"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" '20260802_option_insurance_inventory_exit.sql' \
  "repeatable acceptance installs insurance-exit guards twice"
check_contains "$OPTION_DIR/acceptance/run-p0-asset-rpc-e2e.sh" 'insurance_inventory_exit=' \
  "real Asset RPC gate reports insurance inventory exit evidence"
check_contains "$REPO_ROOT/admin-api/api/option.api" '/risk/insurance-inventory-exits/execute' \
  "Admin API exposes the governed insurance-exit execution route"
check_contains "$REPO_ROOT/admin-api/internal/logic/option/createinsuranceinventoryexitlogic.go" 'OptionCli.CreateInsuranceInventoryExit' \
  "Admin API proxies insurance-exit creation to Option RPC"
check_contains "$REPO_ROOT/admin-api/internal/logic/option/reviewinsuranceinventoryexitlogic.go" 'OptionCli.ReviewInsuranceInventoryExit' \
  "Admin API proxies insurance-exit review to Option RPC"
check_contains "$REPO_ROOT/admin-api/internal/logic/option/executeinsuranceinventoryexitlogic.go" 'OptionCli.ExecuteInsuranceInventoryExit' \
  "Admin API proxies insurance-exit execution to Option RPC"
check_contains "$REPO_ROOT/admin-api/internal/logic/option/listinsuranceinventoryexitslogic.go" 'OptionCli.ListInsuranceInventoryExits' \
  "Admin API proxies insurance-exit listing to Option RPC"
check_contains "$REPO_ROOT/admin-ui/src/views/option/risk.vue" 'executeInsuranceExit' \
  "Admin UI provides the governed insurance-exit workbench"
check_contains "$REPO_ROOT/admin-ui/src/views/option/risk.vue" 'optionOrderStatusLabel' \
  "Admin UI displays insurance-exit order status and fill progress"
check_contains "$insurance_exit_permissions" 'option:insurance-inventory-exit:create' \
  "System permissions separate insurance-exit creation"
check_contains "$insurance_exit_permissions" 'option:insurance-inventory-exit:review' \
  "System permissions separate insurance-exit review"
check_contains "$insurance_exit_permissions" 'option:insurance-inventory-exit:execute' \
  "System permissions separate insurance-exit execution"
check_contains "$OPTION_DIR/docs/templates/option-insurance-inventory-exit-execution-record.md" '四眼复核' \
  "operations pack includes an insurance-exit execution record"

settlement_evidence_migration="$OPTION_DIR/migrations/20260731_zy_option_settlement_price_evidence.sql"
check_contains "$OPTION_DIR/option.sql" 'idx_option_settlement_snapshot_evidence' \
  "Option schema indexes immutable settlement snapshot evidence"
check_contains "$settlement_evidence_migration" 'idx_option_settlement_snapshot_evidence' \
  "settlement evidence migration adds its lookup index idempotently"
check_contains "$settlement_evidence_migration" 'automatic settlement price does not match snapshot median' \
  "database verifies automatic settlement price against immutable snapshot median"
check_contains "$settlement_evidence_migration" 'manual settlement correction requires a creator' \
  "database distinguishes governed manual settlement corrections"
check_contains "$settlement_evidence_migration" 'settlement price evidence cannot be deleted' \
  "database preserves all settlement price evidence versions"
check_contains "$OPTION_DIR/internal/logic/task/processcontractlifecyclelogic.go" 'validateSettlementPriceForUse' \
  "settlement lifecycle fail-closes at the confirmed-price use point"
check_contains "$OPTION_DIR/internal/logic/helpers/settlement_price_evidence.go" 'creator cannot confirm the same version' \
  "settlement price evidence enforces independent confirmation"
check_contains "$OPTION_DIR/models/optionoperationsmetricsmodel.go" "price_source=BINARY 'manual-correction'" \
  "operations metrics recognize governed manual settlement corrections"

beanstalk_readiness="$REPO_ROOT/deploy/beanstalk-readiness.sh"
beanstalk_resilience="$REPO_ROOT/deploy/beanstalk-resilience-smoke.sh"
if [ -x "$beanstalk_resilience" ]; then
  pass "repository provides the executable isolated Beanstalkd resilience gate"
else
  fail "repository provides the executable isolated Beanstalkd resilience gate"
fi
if [ ! -x "$beanstalk_readiness" ]; then
  fail "repository provides the executable native Beanstalkd readiness gate"
elif [ "$MODE" = "repository" ]; then
  if "$beanstalk_readiness" --repository-only >/dev/null; then
    pass "repository Beanstalkd image, WAL and resilience artifacts are pinned and guarded"
  else
    fail "repository Beanstalkd image, WAL and resilience artifacts are pinned and guarded"
  fi
elif "$beanstalk_readiness" >/dev/null; then
  pass "production Beanstalkd instances are healthy and match the host architecture"
else
  fail "production Beanstalkd instances are healthy and match the host architecture"
fi

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
OPTION_PLATFORM_BACKSTOP_ENABLED=$(read_setting OPTION_PLATFORM_BACKSTOP_ENABLED)
OPTION_PORTFOLIO_MARGIN_ENABLED=$(read_setting OPTION_PORTFOLIO_MARGIN_ENABLED)
OPTION_PHYSICAL_DELIVERY_ENABLED=$(read_setting OPTION_PHYSICAL_DELIVERY_ENABLED)
OPTION_COMPLEX_ORDERS_ENABLED=$(read_setting OPTION_COMPLEX_ORDERS_ENABLED)
OPTION_PUBLIC_MARKET_ENABLED=$(read_setting OPTION_PUBLIC_MARKET_ENABLED)
OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=$(read_setting OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED)
OPTION_INSURANCE_INVENTORY_EXIT_ENABLED=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_ENABLED)
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

evidence_path=$(read_setting OPTION_GREEKS_THRESHOLD_APPROVAL)
evidence_sha=$(read_setting OPTION_GREEKS_THRESHOLD_APPROVAL_SHA256)
require_evidence_file "$evidence_path" "$evidence_sha" \
  "contract Greeks freshness threshold approval attached and hash-matched"
check_contains "$evidence_path" \
  '^OPTION_MARKET_FRESHNESS_APPROVAL_STATUS:[[:space:]]*APPROVED$' \
  "contract Greeks freshness threshold approval is explicitly APPROVED"

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

for feature_flag in OPTION_SELLER_TRADING_ENABLED OPTION_PLATFORM_BACKSTOP_ENABLED OPTION_PORTFOLIO_MARGIN_ENABLED \
  OPTION_PHYSICAL_DELIVERY_ENABLED OPTION_COMPLEX_ORDERS_ENABLED \
  OPTION_PUBLIC_MARKET_ENABLED OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED \
  OPTION_INSURANCE_INVENTORY_EXIT_ENABLED; do
  feature_value=$(read_setting "$feature_flag")
  require_boolean "$feature_value" "$feature_flag explicitly declares release scope"
done

evidence_names='OPTION_RELEASE_SIGNOFF
OPTION_LAUNCH_CHECKLIST
OPTION_MIGRATION_REPORT
OPTION_ASSET_E2E_REPORT
OPTION_FAILURE_INJECTION_REPORT
OPTION_CAPACITY_REPORT
OPTION_ORCHESTRATOR_TAKEOVER_REPORT
OPTION_BEANSTALK_CAPACITY_REPORT
OPTION_ALERT_DELIVERY_REPORT
OPTION_DAILY_RECONCILIATION_REPORT
OPTION_DATABASE_AUDIT_REPORT
OPTION_ONCALL_ROSTER_REPORT
OPTION_TRADING_CALENDAR_APPROVAL
OPTION_TRADING_CALENDAR_ANNUAL_REVIEW
OPTION_TRADING_CALENDAR_PREPRODUCTION_REPORT'
for evidence_name in $evidence_names; do
  evidence_path=$(read_setting "$evidence_name")
  evidence_sha=$(read_setting "${evidence_name}_SHA256")
  require_evidence_file "$evidence_path" "$evidence_sha" "$evidence_name attached and hash-matched"
done

generic_approved_evidence_names='OPTION_MIGRATION_REPORT
OPTION_ASSET_E2E_REPORT
OPTION_FAILURE_INJECTION_REPORT
OPTION_CAPACITY_REPORT
OPTION_ONCALL_ROSTER_REPORT
OPTION_TRADING_CALENDAR_PREPRODUCTION_REPORT'
for evidence_name in $generic_approved_evidence_names; do
  evidence_path=$(read_setting "$evidence_name")
  check_contains "$evidence_path" '^OPTION_EVIDENCE_STATUS:[[:space:]]*APPROVED$' \
    "$evidence_name is explicitly APPROVED"
done

OPTION_ALERT_DELIVERY_REPORT=$(read_setting OPTION_ALERT_DELIVERY_REPORT)
check_contains "$OPTION_ALERT_DELIVERY_REPORT" \
  '^OPTION_ALERT_DELIVERY_STATUS:[[:space:]]*APPROVED$' \
  "production alert-delivery report is explicitly APPROVED"
OPTION_DAILY_RECONCILIATION_REPORT=$(read_setting OPTION_DAILY_RECONCILIATION_REPORT)
check_contains "$OPTION_DAILY_RECONCILIATION_REPORT" \
  '^OPTION_DAILY_RECONCILIATION_STATUS:[[:space:]]*APPROVED$' \
  "production daily reconciliation report is explicitly APPROVED"
OPTION_TRADING_CALENDAR_APPROVAL=$(read_setting OPTION_TRADING_CALENDAR_APPROVAL)
check_contains "$OPTION_TRADING_CALENDAR_APPROVAL" \
  '^OPTION_TRADING_CALENDAR_APPROVAL_STATUS:[[:space:]]*APPROVED$' \
  "production trading calendar is explicitly APPROVED"
OPTION_TRADING_CALENDAR_ANNUAL_REVIEW=$(read_setting OPTION_TRADING_CALENDAR_ANNUAL_REVIEW)
check_contains "$OPTION_TRADING_CALENDAR_ANNUAL_REVIEW" \
  '^OPTION_TRADING_CALENDAR_ANNUAL_REVIEW_STATUS:[[:space:]]*APPROVED$' \
  "production trading-calendar annual review is explicitly APPROVED"

OPTION_DATABASE_AUDIT_REPORT=$(read_setting OPTION_DATABASE_AUDIT_REPORT)
check_contains "$OPTION_DATABASE_AUDIT_REPORT" \
  '^OPTION_DATABASE_AUDIT_STATUS:[[:space:]]*APPROVED$' \
  "production database-security audit report is explicitly APPROVED"
OPTION_ORCHESTRATOR_TAKEOVER_REPORT=$(read_setting OPTION_ORCHESTRATOR_TAKEOVER_REPORT)
check_contains "$OPTION_ORCHESTRATOR_TAKEOVER_REPORT" \
  '^OPTION_ORCHESTRATOR_TAKEOVER_STATUS:[[:space:]]*APPROVED$' \
  "production orchestrator takeover report is explicitly APPROVED"
OPTION_BEANSTALK_CAPACITY_REPORT=$(read_setting OPTION_BEANSTALK_CAPACITY_REPORT)
check_contains "$OPTION_BEANSTALK_CAPACITY_REPORT" \
  '^OPTION_BEANSTALK_CAPACITY_STATUS:[[:space:]]*APPROVED$' \
  "production Beanstalk capacity report is explicitly APPROVED"

OPTION_RELEASE_SIGNOFF=$(read_setting OPTION_RELEASE_SIGNOFF)
check_contains "$OPTION_RELEASE_SIGNOFF" '^OPTION_EVIDENCE_STATUS:[[:space:]]*APPROVED$' \
  "Option release signoff is explicitly APPROVED"
OPTION_LAUNCH_CHECKLIST=$(read_setting OPTION_LAUNCH_CHECKLIST)
check_contains "$OPTION_LAUNCH_CHECKLIST" \
  '^OPTION_LAUNCH_CHECKLIST_STATUS:[[:space:]]*APPROVED$' \
  "every contract in the release has an explicitly APPROVED launch checklist"

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
  evidence_path=$(read_setting OPTION_INSURANCE_LEDGER_APPROVAL)
  evidence_sha=$(read_setting OPTION_INSURANCE_LEDGER_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance ledger production approval attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_INSURANCE_LEDGER_APPROVAL_STATUS:[[:space:]]*APPROVED$' \
    "insurance ledger production policy is explicitly APPROVED"
  evidence_path=$(read_setting OPTION_INSURANCE_SIGN_RESOLUTION_REPORT)
  evidence_sha=$(read_setting OPTION_INSURANCE_SIGN_RESOLUTION_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance sign resolution report attached and hash-matched"
  check_contains "$evidence_path" '^OPTION_EVIDENCE_STATUS:[[:space:]]*APPROVED$' \
    "insurance sign resolution report is explicitly APPROVED"
fi

if [ "$OPTION_PLATFORM_BACKSTOP_ENABLED" = "true" ]; then
  require_true "$OPTION_SELLER_TRADING_ENABLED" \
    "platform backstop cannot be enabled while seller trading is disabled"
  require_true "$OPTION_BACKSTOP_LIMIT_APPROVED" \
    "platform backstop monetary limits approved before runtime enablement"
  platform_backstop_policy_id=$(read_setting OPTION_PLATFORM_BACKSTOP_POLICY_ID)
  platform_backstop_policy_version=$(read_setting OPTION_PLATFORM_BACKSTOP_POLICY_VERSION)
  require_positive_integer "$platform_backstop_policy_id" \
    "platform backstop exact Asset policy ID declared"
  require_positive_integer "$platform_backstop_policy_version" \
    "platform backstop exact Asset policy version declared"
  evidence_path=$(read_setting OPTION_PLATFORM_BACKSTOP_APPROVAL)
  evidence_sha=$(read_setting OPTION_PLATFORM_BACKSTOP_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "platform backstop policy approval attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_PLATFORM_BACKSTOP_APPROVAL_STATUS:[[:space:]]*APPROVED$' \
    "platform backstop policy is explicitly APPROVED"
  evidence_path=$(read_setting OPTION_PLATFORM_BACKSTOP_E2E_REPORT)
  evidence_sha=$(read_setting OPTION_PLATFORM_BACKSTOP_E2E_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "platform backstop BST-001 through BST-012 report attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_PLATFORM_BACKSTOP_E2E_STATUS:[[:space:]]*APPROVED$' \
    "platform backstop target-environment E2E is explicitly APPROVED"
  check_contains "$evidence_path" 'BST-012' \
    "platform backstop E2E report includes the final isolation case"
  check_contains "$evidence_path" \
    "Asset policy ID.*${platform_backstop_policy_id}" \
    "platform backstop E2E report binds the declared Asset policy ID"
  check_contains "$evidence_path" \
    "Asset policy version.*${platform_backstop_policy_version}" \
    "platform backstop E2E report binds the declared Asset policy version"
  evidence_path=$(read_setting OPTION_PLATFORM_BACKSTOP_RUNTIME_CONFIG)
  evidence_sha=$(read_setting OPTION_PLATFORM_BACKSTOP_RUNTIME_CONFIG_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "platform backstop rendered runtime config attached and hash-matched"
  if awk '
    /^PlatformBackstop:[[:space:]]*$/ { section=1; next }
    section && /^[^[:space:]]/ { section=0 }
    section && /^[[:space:]]+Enabled:[[:space:]]*true[[:space:]]*$/ { found=1 }
    END { exit(found ? 0 : 1) }
  ' "$evidence_path"; then
    pass "rendered Option config explicitly enables the platform-backstop section"
  else
    fail "rendered Option config explicitly enables the platform-backstop section"
  fi
fi

if [ "$OPTION_PORTFOLIO_MARGIN_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_MODEL_VALIDATION_REPORT)
  evidence_sha=$(read_setting OPTION_MODEL_VALIDATION_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "independent portfolio model validation attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_PORTFOLIO_MODEL_VALIDATION_STATUS:[[:space:]]*APPROVED$' \
    "portfolio model validation is explicitly APPROVED"
  evidence_path=$(read_setting OPTION_PORTFOLIO_VERSION_SWITCH_E2E_REPORT)
  evidence_sha=$(read_setting OPTION_PORTFOLIO_VERSION_SWITCH_E2E_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "portfolio version-switch order/scan E2E attached and hash-matched"
  check_contains "$evidence_path" '^OPTION_EVIDENCE_STATUS:[[:space:]]*APPROVED$' \
    "portfolio version-switch E2E is explicitly APPROVED"
fi
if [ "$OPTION_INSURANCE_INVENTORY_EXIT_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_APPROVAL)
  evidence_sha=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance inventory exit limits and four-eyes approval attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_INSURANCE_INVENTORY_EXIT_APPROVAL_STATUS:[[:space:]]*APPROVED$' \
    "insurance inventory exit policy is explicitly APPROVED"
  evidence_path=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_E2E_REPORT)
  evidence_sha=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_E2E_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance inventory exit preproduction Asset E2E attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_INSURANCE_INVENTORY_EXIT_EXECUTION_STATUS:[[:space:]]*APPROVED$' \
    "insurance inventory exit preproduction E2E is explicitly APPROVED"
  evidence_path=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_RUNTIME_CONFIG)
  evidence_sha=$(read_setting OPTION_INSURANCE_INVENTORY_EXIT_RUNTIME_CONFIG_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "insurance inventory exit runtime limits attached and hash-matched"
  check_contains "$evidence_path" 'InsuranceInventoryExit:' \
    "insurance inventory exit runtime config contains its governed section"
  check_contains "$evidence_path" 'Enabled:[[:space:]]*true' \
    "insurance inventory exit runtime config explicitly enables execution"
  for limit_name in MaxQuantityPerRequest MaxPremiumPerRequest MaxDailyQuantity \
    MaxMarkDeviationRatio MinOrderBookQuantity; do
    check_contains "$evidence_path" "$limit_name:[[:space:]]*[\"'\'']*[0-9]*[1-9]" \
      "insurance inventory exit runtime config sets positive $limit_name"
  done
fi
if [ "$OPTION_PHYSICAL_DELIVERY_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_PHYSICAL_DELIVERY_APPROVAL)
  evidence_sha=$(read_setting OPTION_PHYSICAL_DELIVERY_APPROVAL_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "physical-delivery default and legal approval attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_PHYSICAL_DEFAULT_POLICY_STATUS:[[:space:]]*APPROVED$' \
    "physical-delivery default policy is explicitly APPROVED"
fi
if [ "$OPTION_COMPLEX_ORDERS_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_COMPLEX_ORDER_E2E_REPORT)
  evidence_sha=$(read_setting OPTION_COMPLEX_ORDER_E2E_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "complex-order concurrency/Asset E2E attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_COMPLEX_ORDER_READINESS_STATUS:[[:space:]]*APPROVED$' \
    "complex-order target-environment E2E is explicitly APPROVED"
fi
if [ "$OPTION_PUBLIC_MARKET_ENABLED" = "true" ]; then
  evidence_path=$(read_setting OPTION_PUBLIC_MARKET_PROBE_REPORT)
  evidence_sha=$(read_setting OPTION_PUBLIC_MARKET_PROBE_REPORT_SHA256)
  require_evidence_file "$evidence_path" "$evidence_sha" \
    "public market cross-tenant/TTL/SLA probe attached and hash-matched"
  check_contains "$evidence_path" \
    '^OPTION_PUBLIC_MARKET_READINESS_STATUS:[[:space:]]*APPROVED$' \
    "public-market target-environment probe is explicitly APPROVED"
fi
printf '\n'
if [ "$failures" -eq 0 ]; then
  printf 'READY: Option production prerequisites passed; release still follows the approved change window.\n'
  exit 0
fi
printf 'NOT READY: %s Option prerequisite(s) failed. Trading gates must remain closed.\n' "$failures"
exit 1
