#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
COMPOSE_FILE=${OPTION_P0_E2E_COMPOSE_FILE:-$REPO_ROOT/deploy/docker-compose.yml}
DATABASE=${OPTION_P0_E2E_DATABASE:-wklive_option_p0_asset_e2e}
MYSQL_PASSWORD=${OPTION_P0_E2E_MYSQL_PASSWORD:-123456}
ASSET_PORT=${OPTION_P0_E2E_ASSET_PORT:-18084}
KEEP=${OPTION_P0_E2E_KEEP:-0}

case "$DATABASE" in
  wklive_option_p0_asset_e2e|wklive_option_p0_asset_e2e_[A-Za-z0-9_]*) ;;
  *)
    echo "refusing unsafe acceptance database name: $DATABASE" >&2
    exit 1
    ;;
esac

MYSQL_CONTAINER=${OPTION_P0_E2E_MYSQL_CONTAINER:-$(docker compose -f "$COMPOSE_FILE" ps -q mysql)}
ETCD_CONTAINER=${OPTION_P0_E2E_ETCD_CONTAINER:-$(docker compose -f "$COMPOSE_FILE" ps -q etcd)}
if [ -z "$MYSQL_CONTAINER" ] || [ -z "$ETCD_CONTAINER" ]; then
  echo "mysql and etcd from deploy/docker-compose.yml must already be running" >&2
  exit 1
fi
if nc -z 127.0.0.1 "$ASSET_PORT" >/dev/null 2>&1; then
  echo "Asset acceptance port is already in use: $ASSET_PORT" >&2
  exit 1
fi

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/wklive-option-p0-asset-e2e.XXXXXX")
REDIS_CONTAINER="wklive-option-p0-asset-e2e-redis-$$"
COMMON_KEY="/wklive/acceptance/option-p0-asset-e2e-$$/common"
ASSET_KEY="/wklive/acceptance/option-p0-asset-e2e-$$/asset"
ASSET_PID=""
FAILED=1

mysql_cli() {
  docker exec "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$@"
}

cleanup() {
  if [ -n "$ASSET_PID" ]; then
    kill "$ASSET_PID" >/dev/null 2>&1 || true
    wait "$ASSET_PID" >/dev/null 2>&1 || true
  fi
  docker exec "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 del "$COMMON_KEY" >/dev/null 2>&1 || true
  docker exec "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 del "$ASSET_KEY" >/dev/null 2>&1 || true
  docker rm -f "$REDIS_CONTAINER" >/dev/null 2>&1 || true
  if [ "$KEEP" != "1" ]; then
    mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`;" >/dev/null 2>&1 || true
  fi
  if [ "$FAILED" = "1" ] && [ -f "$WORK_DIR/asset-rpc.log" ]; then
    echo "Asset RPC log:" >&2
    tail -120 "$WORK_DIR/asset-rpc.log" >&2 || true
  fi
  case "$WORK_DIR" in
    "${TMPDIR:-/tmp}"/wklive-option-p0-asset-e2e.*) rm -rf -- "$WORK_DIR" ;;
  esac
}
trap cleanup EXIT INT TERM

echo "creating isolated database $DATABASE"
mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`; CREATE DATABASE \`$DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/asset/asset.sql"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/option/option.sql"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/option/schema/90_constraints.sql"
# Simulate an upgraded database: the canonical schema already contains the
# index, so remove it once and prove that the baseline-safe migration creates it.
mysql_cli "$DATABASE" -e "ALTER TABLE t_asset_freeze DROP INDEX idx_asset_freeze_option_business_key;"
mysql_cli "$DATABASE" -e "
INSERT INTO t_asset_freeze (
  freeze_no,tenant_id,user_id,wallet_type,coin,biz_type,scene_type,biz_id,biz_no,
  amount,used_amount,unfreeze_amount,remain_amount,status,expire_time,remark,
  create_times,update_times
) VALUES
  ('P0-MIGRATION-UNIQUE-FREEZE',996030,1,5,'USDT','option','place_order',1,
   'P0-MIGRATION-UNIQUE',10,0,0,10,1,0,'unique legacy evidence',600,600),
  ('P0-MIGRATION-DUP-FREEZE-A',996032,2,5,'USDT','option','place_order',2,
   'P0-MIGRATION-DUP',20,0,0,20,1,0,'duplicate legacy evidence A',700,700),
  ('P0-MIGRATION-DUP-FREEZE-B',996032,2,5,'USDT','option','place_order',2,
   'P0-MIGRATION-DUP',20,0,0,20,1,0,'duplicate legacy evidence B',800,800),
  ('P0-MIGRATION-TRADE-DUP-A',996033,3,3,'USDT','trade','place_order',3,
   'P0-MIGRATION-TRADE-DUP',30,0,0,30,1,0,'non-option duplicate A',500,500),
  ('P0-MIGRATION-TRADE-DUP-B',996033,3,3,'USDT','trade','place_order',3,
   'P0-MIGRATION-TRADE-DUP',30,0,0,30,1,0,'non-option duplicate B',550,550);
"
for migration in \
  "$REPO_ROOT/services/asset/migrations/20260801_option_freeze_idempotency_evidence.sql"
do
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$migration"
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$migration"
done
for migration in \
	"$REPO_ROOT/services/option/migrations/20260730_y_option_exercise_idempotency.sql" \
	"$REPO_ROOT/services/option/migrations/20260730_zd_option_exercise_governance.sql" \
	"$REPO_ROOT/services/option/migrations/20260730_ze_option_trading_controls.sql" \
	"$REPO_ROOT/services/option/migrations/20260730_zf_option_trade_correction.sql" \
	"$REPO_ROOT/services/option/migrations/20260730_zg_option_mmp.sql" \
	"$REPO_ROOT/services/option/migrations/20260730_zh_option_portfolio_risk_governance.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_x_settlement_price_rule.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_z_option_settlement_price_approval.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_zb_option_wallet_account_mirror.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_zc_option_risk_net_value.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zk_option_trading_calendar.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zl_option_corporate_action.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zn_option_contract_series.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zo_option_public_market.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zp_option_combo_order.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zx_option_margin_coin_evidence.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zy_option_settlement_price_evidence.sql" \
	"$REPO_ROOT/services/option/migrations/20260731_zv_option_order_portfolio_config.sql" \
	"$REPO_ROOT/services/option/migrations/20260801_option_portfolio_liquidation.sql" \
	"$REPO_ROOT/services/option/migrations/20260802_option_portfolio_risk_config_lineage.sql" \
	"$REPO_ROOT/services/option/migrations/20260802_option_contract_series_contract_guard.sql" \
	"$REPO_ROOT/services/option/migrations/20260802_option_insurance_inventory_exit.sql" \
	"$REPO_ROOT/services/option/migrations/20260802_option_last_trade_time.sql" \
	"$REPO_ROOT/services/option/migrations/20260802_option_insurance_fund_flow_semantics.sql"
do
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$migration"
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$migration"
done

echo "starting isolated Redis"
docker run -d --rm --name "$REDIS_CONTAINER" --label wklive.acceptance=option-p0-asset-e2e \
  -p 127.0.0.1::6379 redis:7.4-alpine >/dev/null
REDIS_ADDR=$(docker port "$REDIS_CONTAINER" 6379/tcp | sed -n '1p')
if [ -z "$REDIS_ADDR" ]; then
  echo "failed to resolve isolated Redis port" >&2
  exit 1
fi

cat > "$WORK_DIR/common.yaml" <<EOF
Log:
  Mode: console
  Encoding: plain
  Stat: false
  Level: error
Mysql:
  DataSource: root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local
CacheRedis:
  - Host: $REDIS_ADDR
    Type: node
Redis:
  Key: option-p0-asset-e2e
  Host: $REDIS_ADDR
  Type: node
EOF
cat > "$WORK_DIR/asset.yaml" <<EOF
Name: asset.p0.acceptance.rpc
ListenOn: 127.0.0.1:$ASSET_PORT
Mode: test
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: asset.p0.acceptance.rpc
EOF
docker exec -i "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 put "$COMMON_KEY" < "$WORK_DIR/common.yaml" >/dev/null
docker exec -i "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 put "$ASSET_KEY" < "$WORK_DIR/asset.yaml" >/dev/null

echo "building and starting the real Asset gRPC service"
(
  cd "$REPO_ROOT/services/asset"
  GOCACHE="$WORK_DIR/go-build-cache" go build -o "$WORK_DIR/asset-rpc" .
)
"$WORK_DIR/asset-rpc" -etcd 127.0.0.1:2379 -common "$COMMON_KEY" -config "$ASSET_KEY" >"$WORK_DIR/asset-rpc.log" 2>&1 &
ASSET_PID=$!
attempt=0
until nc -z 127.0.0.1 "$ASSET_PORT" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if ! kill -0 "$ASSET_PID" >/dev/null 2>&1; then
    echo "Asset RPC exited before becoming ready" >&2
    exit 1
  fi
  if [ "$attempt" -ge 100 ]; then
    echo "Asset RPC readiness timeout" >&2
    exit 1
  fi
  sleep 0.1
done

echo "running Option -> Asset real RPC acceptance"
(
  cd "$REPO_ROOT/services/option"
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./models -run '^TestOptionAssetFreezeDuplicateMetricsMySQL$' -count=1
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP0AssetRPCEndToEnd$' -count=1
  for capacity_shorts in 501 5000; do
    OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
    OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
    OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
    OPTION_P0_ASSET_CAPACITY_SHORTS="$capacity_shorts" \
    GOCACHE="$WORK_DIR/go-build-cache" \
      go test ./internal/logic/task -run '^TestP0AmericanAssignmentCapacityAssetRPC$' \
        -count=1 -timeout=25m -v
  done
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  OPTION_P1_PHYSICAL_CAPACITY_UNITS=501 \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP1PhysicalDeliveryCapacityAssetRPC$' \
      -count=1 -timeout=25m -v
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP0CashExpiryCapacityAssetRPC$' \
      -count=1 -timeout=25m -v
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP0AssetMultiInstanceKillTakeover$' -count=1 -timeout=90s
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP0ExerciseProcessKillTakeover$' -count=1 -timeout=100s -v
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP1PhysicalDeliveryProcessKillTakeover$' -count=1 -timeout=100s -v
)

echo "verifying evidence and cleanup scope"
mysql_cli -N "$DATABASE" -e "
SELECT CONCAT('risk_accounts=',COUNT(*),' account_id_sum=',COALESCE(SUM(account_id),0))
FROM t_option_risk_account WHERE tenant_id=996031;
SELECT CONCAT('risk_failure_isolation=',
  ' fault_accounts=',(SELECT COUNT(*) FROM t_option_risk_account WHERE tenant_id=996041),
  ' restricted=',(SELECT COUNT(*) FROM t_option_risk_account WHERE tenant_id=996041 AND status=5),
  ' restricted_uncalculated=',(SELECT COUNT(*) FROM t_option_risk_account
    WHERE tenant_id=996041 AND status=5 AND last_calc_time=0),
  ' same_tenant_healthy=',(SELECT COUNT(*) FROM t_option_risk_account
    WHERE tenant_id=996041 AND user_id=9103 AND status=1 AND last_calc_time>0),
  ' other_tenant_healthy=',(SELECT COUNT(*) FROM t_option_risk_account
    WHERE tenant_id=996042 AND user_id=9201 AND status=1 AND last_calc_time>0));
SELECT CONCAT('portfolio_version_governance=',
  ' configs=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config WHERE tenant_id=996051),
  ' approved=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config WHERE tenant_id=996051 AND status=2),
  ' superseded=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config WHERE tenant_id=996051 AND status=4),
  ' rejected=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config WHERE tenant_id=996051 AND status=3),
  ' rollback_source=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config target
    JOIN t_option_portfolio_risk_config source ON source.id=target.source_config_id
    WHERE target.tenant_id=996051 AND target.version=3 AND source.version=1
      AND target.model_method=source.model_method
      AND target.initial_shock_rate=source.initial_shock_rate
      AND target.maintenance_shock_rate=source.maintenance_shock_rate
      AND target.scenario_shocks=source.scenario_shocks
      AND target.concentration_threshold=source.concentration_threshold
      AND target.concentration_addon_rate=source.concentration_addon_rate
      AND target.liquidity_addon_rate=source.liquidity_addon_rate),
  ' closed_intervals=',(SELECT COUNT(*) FROM t_option_portfolio_risk_config previous
    JOIN t_option_portfolio_risk_config successor ON successor.supersedes_id=previous.id
    WHERE previous.tenant_id=996051 AND previous.status=4
      AND previous.effective_until=successor.effective_from),
  ' versioned_orders=',(SELECT COUNT(*) FROM t_option_order
    WHERE tenant_id=996051 AND portfolio_risk_config_version IN (1,2,3)),
  ' distinct_order_versions=',(SELECT COUNT(DISTINCT portfolio_risk_config_version)
    FROM t_option_order WHERE tenant_id=996051),
  ' final_risk_v3=',(SELECT COUNT(*) FROM t_option_risk_account risk
    JOIN t_option_portfolio_risk_config config
      ON config.id=risk.portfolio_risk_config_id AND config.tenant_id=risk.tenant_id
    WHERE risk.tenant_id=996051 AND config.version=3
      AND risk.portfolio_risk_config_version=3));
SELECT CONCAT('trading_calendar_halt_order_race=',
  ' rounds=',(SELECT COUNT(*) FROM t_option_contract contract
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' accepted=',(SELECT COUNT(*) FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' rejected=',20-(SELECT COUNT(*) FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'
      AND orders.status=4),
  ' client_keys=',(SELECT COUNT(*) FROM t_option_client_order_key client_key
    JOIN t_option_order orders ON orders.tenant_id=client_key.tenant_id
      AND orders.id=client_key.order_id
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id
      AND orders.id=instruction.order_id
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' halts=',(SELECT COUNT(*) FROM t_option_trading_halt halt
    JOIN t_option_contract contract ON contract.tenant_id=halt.tenant_id
      AND contract.id=halt.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' lifted=',(SELECT COUNT(*) FROM t_option_trading_halt halt
    JOIN t_option_contract contract ON contract.tenant_id=halt.tenant_id
      AND contract.id=halt.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'
      AND halt.status=2),
  ' cancel_total=',(SELECT COALESCE(SUM(halt.cancel_total),0) FROM t_option_trading_halt halt
    JOIN t_option_contract contract ON contract.tenant_id=halt.tenant_id
      AND contract.id=halt.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
  ' cancel_success=',(SELECT COALESCE(SUM(halt.cancel_success),0) FROM t_option_trading_halt halt
    JOIN t_option_contract contract ON contract.tenant_id=halt.tenant_id
      AND contract.id=halt.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'),
	  ' cancel_failed=',(SELECT COALESCE(SUM(halt.cancel_failed),0) FROM t_option_trading_halt halt
    JOIN t_option_contract contract ON contract.tenant_id=halt.tenant_id
      AND contract.id=halt.contract_id
	    WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P2-HALT-ORDER-RACE-%'));
SELECT CONCAT('contract_series_capacity=',
  ' series=',(SELECT COUNT(*) FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' status=',(SELECT status FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' launch=',(SELECT launch_status FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' expected=',(SELECT expected_contract_count FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' generated=',(SELECT generated_contract_count FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' expiries=',(SELECT COUNT(*) FROM t_option_contract_series_expiry expiry
    JOIN t_option_contract_series series ON series.tenant_id=expiry.tenant_id AND series.id=expiry.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST'),
  ' bands=',(SELECT COUNT(*) FROM t_option_contract_series_strike_band band
    JOIN t_option_contract_series series ON series.tenant_id=band.tenant_id AND series.id=band.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST'),
  ' details=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST'),
  ' contracts=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    JOIN t_option_contract contract ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST'),
  ' pending=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    JOIN t_option_contract contract ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST' AND contract.status=1),
  ' trading=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    JOIN t_option_contract contract ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST' AND contract.status=2),
  ' calls=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST' AND detail.option_type=1),
  ' puts=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST' AND detail.option_type=2),
  ' pairs=',(SELECT COUNT(DISTINCT CONCAT(detail.expiry_id,':',CAST(detail.strike_price AS CHAR)))
    FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-500-REQUEST'),
  ' created_by=',(SELECT created_by FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' reviewed_by=',(SELECT reviewed_by FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' launch_reviewed_by=',(SELECT launch_reviewed_by FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-500-REQUEST'),
  ' collision_status=',(SELECT status FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-COLLISION-REQUEST'),
  ' collision_generated=',(SELECT generated_contract_count FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-COLLISION-REQUEST'),
  ' collision_details=',(SELECT COUNT(*) FROM t_option_contract_series_detail detail
    JOIN t_option_contract_series series ON series.tenant_id=detail.tenant_id AND series.id=detail.series_id
    WHERE series.tenant_id=996031 AND series.request_key='P2-SERIES-COLLISION-REQUEST'),
  ' collision_codes=',(SELECT COUNT(*) FROM t_option_contract
    WHERE tenant_id=996031 AND contract_code LIKE 'P2COLLIDE-V1-%'),
  ' oversize=',(SELECT COUNT(*) FROM t_option_contract_series
    WHERE tenant_id=996031 AND request_key='P2-SERIES-OVERSIZE-REQUEST'));
SELECT CONCAT('public_market_capacity=',
  ' capacity_contracts=',(SELECT COUNT(*) FROM t_option_contract
    WHERE tenant_id=996031 AND underlying_symbol='P2CHAIN500'
      AND expire_time=(SELECT expire_time FROM t_option_contract
        WHERE tenant_id=996031 AND contract_code='P2-PUB-500-001-C')),
  ' paired_strikes=',(SELECT COUNT(*) FROM (
    SELECT strike_price
    FROM t_option_contract
    WHERE tenant_id=996031 AND underlying_symbol='P2CHAIN500'
      AND expire_time=(SELECT expire_time FROM t_option_contract
        WHERE tenant_id=996031 AND contract_code='P2-PUB-500-001-C')
      AND status=2 AND option_type IN (1,2)
    GROUP BY strike_price
    HAVING COUNT(*)=2 AND COUNT(DISTINCT option_type)=2
  ) paired),
  ' overflow_single_leg=',(SELECT COUNT(*) FROM t_option_contract
    WHERE tenant_id=996031 AND contract_code='P2-PUB-501-251-C'),
  ' bid_levels=',(SELECT COUNT(*) FROM (
    SELECT orders.price FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P2-PUB-500-001-C'
      AND orders.side=1 AND orders.status IN (1,2) AND orders.order_type IN (1,3)
      AND orders.combo_order_id=0 AND orders.price>0 AND orders.unfilled_qty>0
    GROUP BY orders.price
  ) bids),
  ' ask_levels=',(SELECT COUNT(*) FROM (
    SELECT orders.price FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P2-PUB-500-001-C'
      AND orders.side=2 AND orders.status IN (1,2) AND orders.order_type IN (1,3)
      AND orders.combo_order_id=0 AND orders.price>0 AND orders.unfilled_qty>0
    GROUP BY orders.price
  ) asks),
  ' resting_orders=',(SELECT COUNT(*) FROM t_option_order orders
    JOIN t_option_contract contract ON contract.tenant_id=orders.tenant_id
      AND contract.id=orders.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P2-PUB-500-001-C'
      AND orders.status IN (1,2) AND orders.order_type IN (1,3)
      AND orders.combo_order_id=0 AND orders.price>0 AND orders.unfilled_qty>0),
  ' cross_tenant_contracts=',(SELECT COUNT(*) FROM t_option_contract
    WHERE tenant_id=996032 AND underlying_symbol='P2CHAIN500'),
  ' sparse_oi_imbalanced=',(SELECT COUNT(*) FROM (
    SELECT position.contract_id,
      SUM(CASE WHEN position.side=1 THEN position.position_qty ELSE 0 END) long_qty,
      SUM(CASE WHEN position.side=2 THEN position.position_qty ELSE 0 END) short_qty
    FROM t_option_position position
    JOIN t_option_contract contract ON contract.tenant_id=position.tenant_id
      AND contract.id=position.contract_id
    WHERE position.tenant_id=996031 AND contract.underlying_symbol='P2CHAINSPARSE'
      AND position.status=1 AND position.position_qty>0
    GROUP BY position.contract_id
    HAVING long_qty<>short_qty
  ) imbalance));
SELECT CONCAT('combo_acceptance=',
  ' idempotent_parents=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND user_id=5961 AND client_combo_id='P2-COMBO-IDEMPOTENT-50'),
  ' idempotent_legs=',(SELECT COUNT(*) FROM t_option_combo_order_leg leg
    JOIN t_option_combo_order parent ON parent.tenant_id=leg.tenant_id AND parent.id=leg.combo_order_id
    WHERE parent.tenant_id=996031 AND parent.user_id=5961
      AND parent.client_combo_id='P2-COMBO-IDEMPOTENT-50'),
  ' idempotent_children=',(SELECT COUNT(*) FROM t_option_order child
    JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
    WHERE parent.tenant_id=996031 AND parent.user_id=5961
      AND parent.client_combo_id='P2-COMBO-IDEMPOTENT-50'),
  ' idempotent_freezes=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order child ON child.tenant_id=instruction.tenant_id AND child.id=instruction.order_id
    JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
    WHERE parent.tenant_id=996031 AND parent.user_id=5961
      AND parent.client_combo_id='P2-COMBO-IDEMPOTENT-50' AND instruction.action=1),
  ' freeze_loss_retried=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order child ON child.tenant_id=instruction.tenant_id AND child.id=instruction.order_id
    JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
    WHERE parent.tenant_id=996031 AND parent.client_combo_id='P2-COMBO-FREEZE-LOSS'
      AND instruction.action=1 AND instruction.status=3 AND instruction.retry_count>0),
  ' kill_canceled=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id='P2-COMBO-KILL-AFTER-PLACE'
      AND status=6 AND cancel_reason='COMBO_USER_KILL_SWITCH_AFTER_FUNDING'),
  ' margin_canceled=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id='P2-COMBO-MARGIN-DRIFT'
      AND status=6 AND cancel_reason='COMBO_SELL_MARGIN_INSUFFICIENT_AFTER_FUNDING'),
  ' fok_zero=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id='P2-COMBO-FOK-INSUFFICIENT'
      AND status=6 AND filled_qty=0 AND cancel_reason='FOK_NOT_FILLED'),
  ' stp_zero=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id='P2-COMBO-STP'
      AND status=6 AND filled_qty=0 AND cancel_reason='SELF_TRADE_PREVENTED'),
  ' admin_canceled=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id='P2-COMBO-ADMIN-CANCEL'
      AND status=6 AND cancel_reason='ADMIN_FORCE_CANCEL:COMBO_009_ACCEPTANCE'),
  ' matched_parents=',(SELECT COUNT(*) FROM t_option_combo_order
    WHERE tenant_id=996031 AND client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER')
      AND status=4 AND filled_qty=qty AND unfilled_qty=0),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade trade
    WHERE trade.tenant_id=996031 AND trade.combo_match_no<>'' AND
      (trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))
       OR trade.sell_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER')))),
  ' match_groups=',(SELECT COUNT(DISTINCT trade.combo_match_no) FROM t_option_trade trade
    WHERE trade.tenant_id=996031 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' trade_legs=',(SELECT COUNT(DISTINCT trade.combo_leg_no) FROM t_option_trade trade
    WHERE trade.tenant_id=996031 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' trade_qty=',(SELECT COALESCE(SUM(trade.qty),0) FROM t_option_trade trade
    WHERE trade.tenant_id=996031 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' settled_outbox=',(SELECT COUNT(*) FROM t_option_outbox outbox
    JOIN t_option_trade trade ON trade.tenant_id=outbox.tenant_id AND trade.id=outbox.trade_id
    WHERE outbox.tenant_id=996031 AND trade.combo_match_no<>'' AND outbox.status=3 AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' positions=',(SELECT COUNT(*) FROM t_option_position position
    JOIN t_option_contract contract ON contract.tenant_id=position.tenant_id AND contract.id=position.contract_id
    WHERE position.tenant_id=996031 AND position.user_id IN (5965,5967)
      AND contract.contract_code IN ('P2-COMBO-100-C','P2-COMBO-110-C')),
  ' trade_instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_trade trade ON trade.tenant_id=instruction.tenant_id AND trade.id=instruction.trade_id
    WHERE instruction.tenant_id=996031 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' trade_instruction_success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_trade trade ON trade.tenant_id=instruction.tenant_id AND trade.id=instruction.trade_id
    WHERE instruction.tenant_id=996031 AND instruction.status=3 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))),
  ' distinct_trade_instructions=',(SELECT COUNT(DISTINCT instruction.instruction_no)
    FROM t_option_asset_instruction instruction
    JOIN t_option_trade trade ON trade.tenant_id=instruction.tenant_id AND trade.id=instruction.trade_id
    WHERE instruction.tenant_id=996031 AND trade.combo_match_no<>'' AND
      trade.buy_order_id IN (SELECT child.id FROM t_option_order child
        JOIN t_option_combo_order parent ON parent.tenant_id=child.tenant_id AND parent.id=child.combo_order_id
        WHERE parent.tenant_id=996031 AND parent.client_combo_id IN ('P2-COMBO-MAKER','P2-COMBO-TAKER'))));
SELECT CONCAT('corporate_action_capacity=',
  ' action_status=',(SELECT status FROM t_option_corporate_action
    WHERE tenant_id=996031 AND event_no='P2-CORPORATE-ACTION-5001'),
  ' mapping_status=',(SELECT mapping.status FROM t_option_corporate_action_contract mapping
    JOIN t_option_corporate_action action ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' total=',(SELECT mapping.position_total FROM t_option_corporate_action_contract mapping
    JOIN t_option_corporate_action action ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' completed=',(SELECT mapping.position_completed FROM t_option_corporate_action_contract mapping
    JOIN t_option_corporate_action action ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' failed=',(SELECT mapping.position_failed FROM t_option_corporate_action_contract mapping
    JOIN t_option_corporate_action action ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' retry=',(SELECT mapping.retry_count FROM t_option_corporate_action_contract mapping
    JOIN t_option_corporate_action action ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' audits=',(SELECT COUNT(*) FROM t_option_corporate_action_position position_audit
    JOIN t_option_corporate_action_contract mapping
      ON mapping.tenant_id=position_audit.tenant_id AND mapping.id=position_audit.action_contract_id
    JOIN t_option_corporate_action action
      ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' unique_sources=',(SELECT COUNT(DISTINCT position_audit.source_position_id)
    FROM t_option_corporate_action_position position_audit
    JOIN t_option_corporate_action_contract mapping
      ON mapping.tenant_id=position_audit.tenant_id AND mapping.id=position_audit.action_contract_id
    JOIN t_option_corporate_action action
      ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' margin_audits=',(SELECT COUNT(*) FROM t_option_corporate_action_margin_lot margin_audit
    JOIN t_option_corporate_action_position position_audit
      ON position_audit.tenant_id=margin_audit.tenant_id AND position_audit.id=margin_audit.action_position_id
    JOIN t_option_corporate_action_contract mapping
      ON mapping.tenant_id=position_audit.tenant_id AND mapping.id=position_audit.action_contract_id
    JOIN t_option_corporate_action action
      ON action.tenant_id=mapping.tenant_id AND action.id=mapping.action_id
    WHERE action.tenant_id=996031 AND action.event_no='P2-CORPORATE-ACTION-5001'),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction
    WHERE tenant_id=996031 AND user_id=96701),
  ' freezes=',(SELECT COUNT(*) FROM t_asset_freeze
    WHERE tenant_id=996031 AND user_id=96701
      AND biz_no IN ('P2-CORPORATE-ACTION-FREEZE-0001','P2-CORPORATE-ACTION-FREEZE-0101')),
  ' frozen=',(SELECT COALESCE(SUM(remain_amount),0) FROM t_asset_freeze
    WHERE tenant_id=996031 AND user_id=96701
      AND biz_no IN ('P2-CORPORATE-ACTION-FREEZE-0001','P2-CORPORATE-ACTION-FREEZE-0101')),
  ' flows=',(SELECT COUNT(*) FROM t_asset_flow
    WHERE tenant_id=996031 AND user_id=96701
      AND biz_no IN ('P2-CORPORATE-ACTION-FREEZE-0001','P2-CORPORATE-ACTION-FREEZE-0101')),
  ' manual=',(SELECT COUNT(*) FROM t_option_corporate_action
    WHERE tenant_id=996031 AND event_no='P2-CORPORATE-ACTION-MANUAL-ONLY' AND status=6));
SELECT CONCAT('instructions=',COUNT(*),' success=',SUM(status=3),' canceled=',SUM(status=6),
  ' reconciled=',SUM(reconciliation_status=2),
  ' weighted_terminal=',SUM(status=3)+2*SUM(status=6))
FROM t_option_asset_instruction WHERE tenant_id=996031;
SELECT CONCAT('freeze_flows=',SUM(scene_type='place_order'),' release_flows=',SUM(scene_type='cancel_order'))
FROM t_asset_flow WHERE tenant_id=996031 AND user_id=104 AND biz_type='option';
SELECT CONCAT('margin_release_instructions=',COUNT(*),' success=',SUM(status=3),
  ' reconciled=',SUM(reconciliation_status=2),' coins=',GROUP_CONCAT(coin ORDER BY coin))
FROM t_option_asset_instruction
WHERE tenant_id=996031 AND instruction_no IN (
  'P0-MARGIN-CALL-CONTROL-RELEASE','P0-MARGIN-PUT-CONTROL-RELEASE'
);
SELECT CONCAT('margin_freeze_flows=',SUM(scene_type='place_order'),
  ' margin_release_flows=',SUM(scene_type='cancel_order'),
  ' coins=',GROUP_CONCAT(DISTINCT coin ORDER BY coin))
FROM t_asset_flow
WHERE tenant_id=996031 AND user_id IN (105,106) AND biz_type='option';
SELECT CONCAT('physical_order_coin=',IF(contract.option_type=1,'CALL','PUT'),
  ' orders=',COUNT(DISTINCT orders.id),
  ' order_types=',COUNT(DISTINCT orders.order_type),
  ' canceled=',COUNT(DISTINCT IF(orders.status=4,orders.id,NULL)),
  ' expired=',COUNT(DISTINCT IF(orders.status=6,orders.id,NULL)),
  ' user_cancel=',COUNT(DISTINCT IF(orders.cancel_reason='USER_CANCEL',orders.id,NULL)),
  ' liquidation_cancel=',COUNT(DISTINCT IF(orders.cancel_reason='LIQUIDATION',orders.id,NULL)),
  ' admin_cancel=',COUNT(DISTINCT IF(orders.cancel_reason LIKE 'P0_PHYSICAL_%_ADMIN_CANCEL',orders.id,NULL)),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' coins=',GROUP_CONCAT(DISTINCT instruction.coin ORDER BY instruction.coin),
  ' wrong_coin=',COUNT(DISTINCT IF(instruction.coin<>IF(contract.option_type=1,'BTC','USDT'),instruction.id,NULL)),
  ' freeze=',CAST(SUM(IF(instruction.action=1,instruction.amount,0)) AS CHAR),
  ' release=',CAST(SUM(IF(instruction.action=3,instruction.amount,0)) AS CHAR),
  ' flows=',COUNT(DISTINCT flow.id),
  ' admin_events=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    JOIN t_option_contract event_contract
      ON event_contract.tenant_id=event.tenant_id AND event_contract.id=event.contract_id
    WHERE event_contract.tenant_id=996031
      AND event_contract.option_type=contract.option_type
      AND event_contract.contract_code IN ('P0-PHYSICAL-COIN-CALL','P0-PHYSICAL-COIN-PUT')
      AND event.event_type='ADMIN_FORCE_CANCEL_ORDER'),
  ' wallet_total=',(SELECT CAST(SUM(wallet.total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.wallet_type=5
      AND wallet.coin=IF(contract.option_type=1,'BTC','USDT')
      AND ((contract.option_type=1 AND
              (wallet.user_id BETWEEN 891001 AND 891011 OR wallet.user_id=893001))
        OR (contract.option_type=2 AND
              (wallet.user_id BETWEEN 892001 AND 892011 OR wallet.user_id=894001)))),
  ' wallet_frozen=',(SELECT CAST(SUM(wallet.frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.wallet_type=5
      AND wallet.coin=IF(contract.option_type=1,'BTC','USDT')
      AND ((contract.option_type=1 AND
              (wallet.user_id BETWEEN 891001 AND 891011 OR wallet.user_id=893001))
        OR (contract.option_type=2 AND
              (wallet.user_id BETWEEN 892001 AND 892011 OR wallet.user_id=894001)))))
FROM t_option_contract contract
JOIN t_option_order orders
  ON orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=orders.tenant_id AND instruction.order_id=orders.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id
  AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
WHERE contract.tenant_id=996031 AND (
  contract.contract_code LIKE 'P0-PHYSICAL-COIN-CALL%' OR
  contract.contract_code LIKE 'P0-PHYSICAL-COIN-PUT%'
)
GROUP BY contract.option_type
ORDER BY contract.option_type;
SELECT CONCAT('cash_settlement=',s.status,' batch=',b.status,
  ' instructions=',b.instruction_count,' success=',b.success_count,
  ' credit=',CAST(b.total_credit AS CHAR),' debit=',CAST(b.total_debit AS CHAR),
  ' contract=',c.status,' delivery_price=',CAST(s.delivery_price AS CHAR))
FROM t_option_settlement s
JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
JOIN t_option_contract c ON c.id=s.contract_id
WHERE s.tenant_id=996031 AND s.contract_id=996301;
SELECT CONCAT('settlement_wallet_total=',CAST(SUM(total_amount) AS CHAR),
  ' available=',CAST(SUM(available_amount) AS CHAR),
  ' frozen=',CAST(SUM(frozen_amount) AS CHAR))
FROM t_user_asset
WHERE tenant_id=996031 AND wallet_type=5 AND coin='USDT' AND user_id IN (107,108);
SELECT CONCAT('missing_price_settlements=',COUNT(*))
FROM t_option_settlement WHERE tenant_id=996031 AND contract_id=996302;
SELECT CONCAT('cash_expiry_capacity=501',
  ' positions=',(SELECT COUNT(*) FROM t_option_position p WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' accounts=',(SELECT COUNT(DISTINCT p.account_id) FROM t_option_position p WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' details=',(SELECT COUNT(*) FROM t_option_settlement_detail d WHERE d.tenant_id=contract.tenant_id AND d.contract_id=contract.id),
  ' partial_qty=',(SELECT CAST(MIN(d.quantity) AS CHAR) FROM t_option_settlement_detail d WHERE d.tenant_id=contract.tenant_id AND d.contract_id=contract.id AND d.side=1),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' long_credit=',CAST(SUM(IF(instruction.action=4 AND instruction.user_id BETWEEN 21000 AND 21500,instruction.amount,0)) AS CHAR),
  ' fee=',CAST(SUM(IF(instruction.action=4 AND instruction.user_id=29999,instruction.amount,0)) AS CHAR),
  ' debit=',CAST(SUM(IF(instruction.action=2 AND instruction.user_id=22000,instruction.amount,0)) AS CHAR),
  ' release=',CAST(SUM(IF(instruction.action=3 AND instruction.user_id=22000,instruction.amount,0)) AS CHAR),
  ' flows=',COUNT(DISTINCT flow.id),
  ' wallet_total=',(SELECT CAST(SUM(wallet.total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND ((wallet.user_id BETWEEN 21000 AND 21500) OR wallet.user_id IN (22000,29999))),
  ' wallet_frozen=',(SELECT CAST(SUM(wallet.frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND ((wallet.user_id BETWEEN 21000 AND 21500) OR wallet.user_id IN (22000,29999))))
FROM t_option_contract contract
JOIN t_option_settlement settlement
  ON settlement.tenant_id=contract.tenant_id AND settlement.contract_id=contract.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=settlement.tenant_id AND instruction.biz_no=settlement.settlement_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-CASH-EXPIRY-CAPACITY-501'
GROUP BY contract.id,contract.tenant_id;
SELECT CONCAT('physical_delivery_contract=',contract.contract_code,
  ' settlement=',settlement.status,' batch=',batch.status,' contract_status=',contract.status,
  ' units=',COUNT(DISTINCT unit.id),
  ' completed=',COUNT(DISTINCT IF(unit.status=5,unit.id,NULL)),
  ' manual_retries=',MAX(unit.manual_retry_count),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id))
FROM t_option_contract contract
JOIN t_option_settlement settlement
  ON settlement.tenant_id=contract.tenant_id AND settlement.contract_id=contract.id
JOIN t_option_settlement_batch batch
  ON batch.tenant_id=settlement.tenant_id AND batch.batch_no=settlement.settlement_no
JOIN t_option_physical_delivery_unit unit
  ON unit.tenant_id=contract.tenant_id AND unit.contract_id=contract.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=unit.tenant_id AND instruction.delivery_unit_id=unit.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE contract.tenant_id=996031 AND contract.id IN (997001,997002,997003,997004,999901,999902)
GROUP BY contract.id,contract.contract_code,contract.status,settlement.status,batch.status
ORDER BY contract.id;
SELECT CONCAT('physical_delivery_flow_cardinality_anomalies=',COUNT(*))
FROM (
  SELECT instruction.id
  FROM t_option_asset_instruction instruction
  JOIN t_option_physical_delivery_unit unit
    ON unit.tenant_id=instruction.tenant_id AND unit.id=instruction.delivery_unit_id
  LEFT JOIN t_asset_flow flow
    ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
  WHERE unit.tenant_id=996031 AND unit.contract_id IN (997001,997002,997003,997004,999901,999902)
  GROUP BY instruction.id
  HAVING COUNT(flow.id)<>1
) physical_flow_anomaly;
SELECT CONCAT('physical_delivery_manual_retry_events=',COUNT(*),
  ' reason=',COALESCE(MAX(reason),''),
  ' operator=',COALESCE(MAX(operator_id),0))
FROM t_option_trading_control_event
WHERE tenant_id=996031 AND contract_id=997003
  AND event_type='PHYSICAL_DELIVERY_MANUAL_RETRY';
SELECT CONCAT('physical_delivery_process_kill=',unit.status,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id),
  ' completion_events=',(
    SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=unit.tenant_id AND event.contract_id=unit.contract_id
      AND event.event_type='PHYSICAL_DELIVERY_COMPLETED'
  ))
FROM t_option_physical_delivery_unit unit
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=unit.tenant_id AND instruction.delivery_unit_id=unit.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE unit.tenant_id=996031 AND unit.contract_id=999901
GROUP BY unit.id,unit.status;
SELECT CONCAT('physical_delivery_capacity=',COUNT(DISTINCT unit.id),
  ' completed=',COUNT(DISTINCT IF(unit.status=5,unit.id,NULL)),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' completion_events=',(
    SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=996031 AND event.contract_id=999902
      AND event.event_type='PHYSICAL_DELIVERY_COMPLETED'
  ))
FROM t_option_physical_delivery_unit unit
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=unit.tenant_id AND instruction.delivery_unit_id=unit.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE unit.tenant_id=996031 AND unit.contract_id=999902;
SELECT CONCAT('settlement_failure_recovery_contract=',s.contract_id,
  ' settlement=',s.status,' batch=',b.status,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',SUM(instruction.status=3),
  ' reconciled=',SUM(instruction.reconciliation_status=2),
  ' max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id))
FROM t_option_settlement s
JOIN t_option_settlement_batch b
  ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=s.tenant_id AND instruction.biz_no=s.settlement_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE s.tenant_id=996031 AND s.contract_id IN (996303,996304)
GROUP BY s.contract_id,s.status,b.status
ORDER BY s.contract_id;
SELECT CONCAT('failure_recovery_wallet_pair=',
  wallet_pair,
  ' total=',CAST(SUM(total_amount) AS CHAR),
  ' available=',CAST(SUM(available_amount) AS CHAR),
  ' frozen=',CAST(SUM(frozen_amount) AS CHAR))
FROM (
  SELECT user_id,total_amount,available_amount,frozen_amount,
    CASE WHEN user_id IN (109,110) THEN 'before-debit' ELSE 'response-loss' END AS wallet_pair
  FROM t_user_asset
  WHERE tenant_id=996031 AND wallet_type=5 AND coin='USDT' AND user_id IN (109,110,111,112)
) recovery_wallet
GROUP BY wallet_pair
ORDER BY MIN(user_id);
SELECT CONCAT('stale_processing_recovery_contract=',s.contract_id,
  ' settlement=',s.status,' batch=',b.status,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',SUM(instruction.status=3),
  ' reconciled=',SUM(instruction.reconciliation_status=2),
  ' max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id))
FROM t_option_settlement s
JOIN t_option_settlement_batch b
  ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=s.tenant_id AND instruction.biz_no=s.settlement_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE s.tenant_id=996031 AND s.contract_id=996305
GROUP BY s.contract_id,s.status,b.status;
SELECT CONCAT('stale_processing_wallet_total=',CAST(SUM(total_amount) AS CHAR),
  ' available=',CAST(SUM(available_amount) AS CHAR),
  ' frozen=',CAST(SUM(frozen_amount) AS CHAR))
FROM t_user_asset
WHERE tenant_id=996031 AND wallet_type=5 AND coin='USDT' AND user_id IN (113,114);
SELECT CONCAT('insufficient_balance_recovery_contract=',s.contract_id,
  ' settlement=',s.status,' batch=',b.status,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',SUM(instruction.status=3),
  ' reconciled=',SUM(instruction.reconciliation_status=2),
  ' final_max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id))
FROM t_option_settlement s
JOIN t_option_settlement_batch b
  ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=s.tenant_id AND instruction.biz_no=s.settlement_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE s.tenant_id=996031 AND s.contract_id=996306
GROUP BY s.contract_id,s.status,b.status;
SELECT CONCAT('insufficient_balance_wallet_total=',CAST(SUM(total_amount) AS CHAR),
  ' available=',CAST(SUM(available_amount) AS CHAR),
  ' frozen=',CAST(SUM(frozen_amount) AS CHAR))
FROM t_user_asset
WHERE tenant_id=996031 AND wallet_type=5 AND coin='USDT' AND user_id IN (115,116);
SELECT CONCAT('insufficient_balance_manual_retry_events=',COUNT(*),
  ' reason=',COALESCE(MAX(reason),''),
  ' operator=',COALESCE(MAX(operator_id),0),
  ' retry20_evidence=',SUM(detail LIKE '%fromRetryCount=20%'))
FROM t_option_trading_control_event
WHERE tenant_id=996031 AND contract_id=996306
  AND event_type='ASSET_INSTRUCTION_MANUAL_RETRY';
SELECT CONCAT('insufficient_balance_topup_flows=',COUNT(*),
  ' amount=',COALESCE(CAST(SUM(change_amount) AS CHAR),'0'))
FROM t_asset_flow
WHERE tenant_id=996031 AND user_id=116
  AND biz_no='P0-SETTLE-INSUFFICIENT-TOPUP-20';
SELECT CONCAT('american_exercise=',e.status,
  ' assignments=',COUNT(DISTINCT a.id),
  ' assignment_done=',COUNT(DISTINCT IF(a.status=2,a.id,NULL)),
  ' instructions=',COUNT(DISTINCT i.id),
  ' success=',COUNT(DISTINCT IF(i.status=3,i.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(i.reconciliation_status=2,i.id,NULL)),
  ' flows=',COUNT(DISTINCT f.id))
FROM t_option_exercise e
JOIN t_option_exercise_assignment a
  ON a.tenant_id=e.tenant_id AND a.exercise_id=e.id
JOIN t_option_asset_instruction i
  ON i.tenant_id=e.tenant_id AND i.biz_no=e.exercise_no
LEFT JOIN t_asset_flow f
  ON f.tenant_id=i.tenant_id AND f.biz_no=i.instruction_no
WHERE e.tenant_id=996031 AND e.client_exercise_id='P0-AMERICAN-EXERCISE-CONCURRENT'
GROUP BY e.id,e.status;
SELECT CONCAT('american_exercise_close_race=',COUNT(*),
  ' done=',SUM(exercise.status=2),
  ' assignments=',(SELECT COUNT(*) FROM t_option_exercise_assignment assignment
    JOIN t_option_exercise race_exercise
      ON race_exercise.tenant_id=assignment.tenant_id AND race_exercise.id=assignment.exercise_id
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=race_exercise.tenant_id AND race_contract.id=race_exercise.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%'),
  ' assignment_done=',(SELECT COUNT(*) FROM t_option_exercise_assignment assignment
    JOIN t_option_exercise race_exercise
      ON race_exercise.tenant_id=assignment.tenant_id AND race_exercise.id=assignment.exercise_id
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=race_exercise.tenant_id AND race_contract.id=race_exercise.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%' AND assignment.status=2),
  ' close_orders=',(SELECT COUNT(*) FROM t_option_order close_order
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=close_order.tenant_id AND race_contract.id=close_order.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%'),
  ' assignment_canceled=',(SELECT COUNT(*) FROM t_option_order close_order
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=close_order.tenant_id AND race_contract.id=close_order.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%'
      AND close_order.status=4 AND close_order.cancel_reason='AMERICAN_EXERCISE_ASSIGNMENT'),
  ' active_close=',(SELECT COUNT(*) FROM t_option_order close_order
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=close_order.tenant_id AND race_contract.id=close_order.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%'
      AND close_order.position_effect=2 AND close_order.status IN (7,1,2)),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_exercise race_exercise
      ON race_exercise.tenant_id=instruction.tenant_id AND race_exercise.exercise_no=instruction.biz_no
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=race_exercise.tenant_id AND race_contract.id=race_exercise.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%'),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_exercise race_exercise
      ON race_exercise.tenant_id=instruction.tenant_id AND race_exercise.exercise_no=instruction.biz_no
    JOIN t_option_contract race_contract
      ON race_contract.tenant_id=race_exercise.tenant_id AND race_contract.id=race_exercise.contract_id
    WHERE race_contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%' AND instruction.status=3))
FROM t_option_exercise exercise
JOIN t_option_contract contract
  ON contract.tenant_id=exercise.tenant_id AND contract.id=exercise.contract_id
WHERE contract.tenant_id=996031 AND contract.contract_code LIKE 'P0-AMERICAN-CLOSE-RACE-%';
SELECT CONCAT('exercise_cutoff_race=',
  ' cutoff_exercises=',(SELECT COUNT(*) FROM t_option_exercise exercise
    JOIN t_option_contract contract ON contract.tenant_id=exercise.tenant_id AND contract.id=exercise.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-EXERCISE-CUTOFF-CROSS'),
  ' cutoff_position_status=',(SELECT position.status FROM t_option_position position
    JOIN t_option_contract contract ON contract.tenant_id=position.tenant_id AND contract.id=position.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-EXERCISE-CUTOFF-CROSS' LIMIT 1),
  ' cutoff_available=',(SELECT CAST(position.available_qty AS CHAR) FROM t_option_position position
    JOIN t_option_contract contract ON contract.tenant_id=position.tenant_id AND contract.id=position.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-EXERCISE-CUTOFF-CROSS' LIMIT 1),
  ' cutoff_frozen=',(SELECT CAST(position.frozen_qty AS CHAR) FROM t_option_position position
    JOIN t_option_contract contract ON contract.tenant_id=position.tenant_id AND contract.id=position.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-EXERCISE-CUTOFF-CROSS' LIMIT 1),
  ' state_instructions=',(SELECT COUNT(*) FROM t_option_exercise_instruction instruction
    JOIN t_option_contract contract ON contract.tenant_id=instruction.tenant_id AND contract.id=instruction.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-INSTRUCTION-STATE-CROSS'),
  ' state_contract_status=',(SELECT contract.status FROM t_option_contract contract
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-INSTRUCTION-STATE-CROSS'),
  ' paused_instructions=',(SELECT COUNT(*) FROM t_option_exercise_instruction instruction
    JOIN t_option_contract contract ON contract.tenant_id=instruction.tenant_id AND contract.id=instruction.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-INSTRUCTION-PAUSED-ACCEPT'),
  ' paused_active=',(SELECT COUNT(*) FROM t_option_exercise_instruction instruction
    JOIN t_option_contract contract ON contract.tenant_id=instruction.tenant_id AND contract.id=instruction.contract_id
    WHERE contract.tenant_id=996031 AND contract.contract_code='P0-INSTRUCTION-PAUSED-ACCEPT'
      AND instruction.status=1 AND instruction.cutoff_time=contract.exercise_cutoff_time));
SELECT CONCAT('last_trade_boundary=',
  ' contract_status=',contract.status,
  ' order_status=',orders.status,
  ' cancel_reason=',orders.cancel_reason,
  ' distinct_times=',(
    contract.list_time < contract.last_trade_time
    AND contract.last_trade_time < contract.exercise_cutoff_time
    AND contract.exercise_cutoff_time < contract.expire_time
    AND contract.expire_time < contract.deliver_time),
  ' premature_exercises=',(SELECT COUNT(*) FROM t_option_exercise exercise
    WHERE exercise.tenant_id=contract.tenant_id AND exercise.contract_id=contract.id),
  ' premature_settlements=',(SELECT COUNT(*) FROM t_option_settlement settlement
    WHERE settlement.tenant_id=contract.tenant_id AND settlement.contract_id=contract.id))
FROM t_option_contract contract
JOIN t_option_order orders
  ON orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
WHERE contract.tenant_id=996031
  AND contract.contract_code='P0-LAST-TRADE-INDEPENDENT'
  AND orders.order_no='P0-LAST-TRADE-PENDING-ORDER';
SELECT CONCAT('exercise_instruction_replacement_race=',
  ' rows=',COUNT(*),
  ' versions=',COUNT(DISTINCT instruction.version),
  ' clients=',COUNT(DISTINCT instruction.client_instruction_id),
  ' active=',SUM(instruction.status=1),
  ' superseded=',SUM(instruction.status=2),
  ' chain_links=',SUM(instruction.version=1 AND instruction.supersedes_id=0)
    +(SELECT COUNT(*) FROM t_option_exercise_instruction current_instruction
      JOIN t_option_exercise_instruction previous_instruction
        ON previous_instruction.tenant_id=current_instruction.tenant_id
       AND previous_instruction.position_id=current_instruction.position_id
       AND previous_instruction.version=current_instruction.version-1
       AND previous_instruction.id=current_instruction.supersedes_id
      WHERE current_instruction.tenant_id=996031
        AND current_instruction.contract_id=contract.id
        AND current_instruction.version>1),
  ' cutoff_matches=',SUM(instruction.cutoff_time=contract.exercise_cutoff_time))
FROM t_option_contract contract
JOIN t_option_exercise_instruction instruction
  ON instruction.tenant_id=contract.tenant_id AND instruction.contract_id=contract.id
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-INSTRUCTION-REPLACE-RACE'
GROUP BY contract.id;
SELECT CONCAT('american_exercise_process_kill=',exercise.status,
  ' assignments=',COUNT(DISTINCT assignment.id),
  ' assignment_done=',COUNT(DISTINCT IF(assignment.status=2,assignment.id,NULL)),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' max_retry=',MAX(instruction.retry_count),
  ' flows=',COUNT(DISTINCT flow.id),
  ' target_flows=',COUNT(DISTINCT IF(instruction.user_id=93151,flow.id,NULL)),
  ' wallet_total=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=996031 AND user_id IN (93151,93152,93153) AND coin='USDT' AND wallet_type=5),
  ' wallet_frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=996031 AND user_id IN (93151,93152,93153) AND coin='USDT' AND wallet_type=5))
FROM t_option_exercise exercise
JOIN t_option_contract contract
  ON contract.tenant_id=exercise.tenant_id AND contract.id=exercise.contract_id
JOIN t_option_exercise_assignment assignment
  ON assignment.tenant_id=exercise.tenant_id AND assignment.exercise_id=exercise.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=exercise.tenant_id AND instruction.biz_no=exercise.exercise_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-AMERICAN-EXERCISE-PROCESS-KILL'
GROUP BY exercise.id,exercise.status;
SELECT CONCAT('american_assignment_capacity=',SUBSTRING_INDEX(c.contract_code,'-',-1),
  ' exercise=',e.status,
  ' assignments=',(SELECT COUNT(*) FROM t_option_exercise_assignment assignment
    WHERE assignment.tenant_id=e.tenant_id AND assignment.exercise_id=e.id),
  ' assignment_done=',(SELECT COUNT(*) FROM t_option_exercise_assignment assignment
    WHERE assignment.tenant_id=e.tenant_id AND assignment.exercise_id=e.id AND assignment.status=2),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=e.tenant_id AND instruction.biz_no=e.exercise_no),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=e.tenant_id AND instruction.biz_no=e.exercise_no AND instruction.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=e.tenant_id AND instruction.biz_no=e.exercise_no
      AND instruction.reconciliation_status=2),
  ' flows=',(SELECT COUNT(*) FROM t_asset_flow flow
    JOIN t_option_asset_instruction instruction
      ON instruction.tenant_id=flow.tenant_id AND instruction.instruction_no=flow.biz_no
    WHERE instruction.tenant_id=e.tenant_id AND instruction.biz_no=e.exercise_no))
FROM t_option_contract c
JOIN t_option_exercise e
  ON e.tenant_id=c.tenant_id AND e.contract_id=c.id
WHERE c.tenant_id=996031 AND c.contract_code IN (
  'P0-AMERICAN-CAPACITY-501','P0-AMERICAN-CAPACITY-5000'
)
ORDER BY CAST(SUBSTRING_INDEX(c.contract_code,'-',-1) AS UNSIGNED);
SELECT CONCAT('expiry_auto_dne=',s.status,' batch=',b.status,' contract=',c.status,
  ' exercises=',COUNT(DISTINCT e.id),
  ' instructions=',COUNT(DISTINCT i.id),
  ' success=',COUNT(DISTINCT IF(i.status=3,i.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(i.reconciliation_status=2,i.id,NULL)),
  ' flows=',COUNT(DISTINCT f.id),
  ' credit=',CAST(b.total_credit AS CHAR),' debit=',CAST(b.total_debit AS CHAR))
FROM t_option_contract c
JOIN t_option_settlement s ON s.tenant_id=c.tenant_id AND s.contract_id=c.id
JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
LEFT JOIN t_option_exercise e ON e.tenant_id=c.tenant_id AND e.contract_id=c.id
JOIN t_option_asset_instruction i ON i.tenant_id=s.tenant_id AND i.biz_no=s.settlement_no
LEFT JOIN t_asset_flow f ON f.tenant_id=i.tenant_id AND f.biz_no=i.instruction_no
WHERE c.tenant_id=996031 AND c.contract_code='P0-EXPIRY-AUTO-DNE-CALL'
GROUP BY c.id,c.status,s.status,b.id,b.status;
SELECT CONCAT('partial_close_trade=',trade.id,
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=trade.tenant_id AND i.trade_id=trade.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=trade.tenant_id AND i.trade_id=trade.id AND i.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=trade.tenant_id AND i.trade_id=trade.id AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(*) FROM t_asset_flow f
    JOIN t_option_asset_instruction i
      ON i.tenant_id=f.tenant_id AND i.instruction_no=f.biz_no
    WHERE i.tenant_id=trade.tenant_id AND i.trade_id=trade.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=trade.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (131,132,133,134)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=trade.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (131,132,133,134)),
  ' remaining_qty=',CAST(position.position_qty AS CHAR),
  ' unrealized=',CAST(position.unrealized_pnl AS CHAR),
  ' trade_realized=',CAST(position.trade_realized_pnl AS CHAR),
  ' fee=',CAST(position.fee_paid AS CHAR),
  ' total_return=',CAST(position.total_return AS CHAR))
FROM t_option_contract contract
JOIN t_option_trade trade
  ON trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id
JOIN t_option_position position
  ON position.tenant_id=trade.tenant_id AND position.contract_id=trade.contract_id
 AND position.user_id=131 AND position.account_id=7030 AND position.side=1
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-PARTIAL-CLOSE-CALL';
SELECT CONCAT('order_admission=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' filled=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=3),
  ' client_keys=',(SELECT COUNT(*) FROM t_option_client_order_key k
    JOIN t_option_order o ON o.tenant_id=k.tenant_id AND o.id=k.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' turnover=',(SELECT CAST(SUM(turnover) AS CHAR) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' buy_fee=',(SELECT CAST(SUM(buy_fee) AS CHAR) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' sell_fee=',(SELECT CAST(SUM(sell_fee) AS CHAR) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.status=3 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.reconciliation_status=2 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' outbox=',(SELECT COUNT(*) FROM t_option_outbox o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=3),
  ' inbox=',(SELECT COUNT(*) FROM t_option_inbox i
    WHERE i.tenant_id=contract.tenant_id AND i.contract_id=contract.id AND i.status=2),
  ' positions=',(SELECT COUNT(*) FROM t_option_position p
    WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' margin_lots=',(SELECT COUNT(*) FROM t_option_margin_lot lot
    WHERE lot.tenant_id=contract.tenant_id AND lot.contract_id=contract.id
      AND lot.initial_margin=50 AND lot.remaining_margin=50),
  ' risk_accounts=',(SELECT COUNT(*) FROM t_option_risk_account risk
    WHERE risk.tenant_id=contract.tenant_id AND risk.user_id IN (161,162)
      AND risk.account_id=0 AND risk.settle_coin='USDT'),
  ' normal=',(SELECT COUNT(*) FROM t_option_risk_account risk
    WHERE risk.tenant_id=contract.tenant_id AND risk.user_id IN (161,162)
      AND risk.account_id=0 AND risk.settle_coin='USDT' AND risk.status=1),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (161,162,163)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (161,162,163)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031
  AND contract.contract_code='P0-FULL-ORDER-ADMISSION-CALL';
SELECT CONCAT('wallet_scope_restriction=',
  ' risk_accounts=',(SELECT COUNT(*) FROM t_option_risk_account risk
    WHERE risk.tenant_id=contract.tenant_id AND risk.user_id=880001
      AND risk.account_id=0 AND risk.settle_coin='USDT' AND risk.status=5),
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' client_keys=',(SELECT COUNT(*) FROM t_option_client_order_key k
    WHERE k.tenant_id=contract.tenant_id AND k.user_id=880001),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.user_id=880001),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=880001
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' available=',(SELECT CAST(available_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=880001
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=880001
      AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_contract contract
WHERE contract.tenant_id=996031
  AND contract.contract_code='P0-WALLET-RESTRICTION-CALL';
SELECT CONCAT('cross_account_stp=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' accounts=',(SELECT COUNT(DISTINCT account_id) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' stp_taker=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND o.cancel_reason='SELF_TRADE_PREVENTED'),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade trade
    WHERE trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id),
  ' events=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='STP_PREVENTED' AND event.reason='SELF_TRADE_PREVENTED'),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
      AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' positions=',(SELECT COUNT(*) FROM t_option_position position
    WHERE position.tenant_id=contract.tenant_id AND position.contract_id=contract.id),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=880002
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=880002
      AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_contract contract
WHERE contract.tenant_id=996031
  AND contract.contract_code='P0-WALLET-SCOPE-STP-CALL';
SELECT CONCAT('portfolio_cross_account_concurrency=',
  ' orders=',COUNT(DISTINCT o.id),
  ' canceled=',COUNT(DISTINCT IF(o.status=4,o.id,NULL)),
  ' accounts=',COUNT(DISTINCT o.account_id),
  ' configs=',COUNT(DISTINCT CONCAT(o.portfolio_risk_config_id,':',o.portfolio_risk_config_version)),
  ' aggregate_margin=',CAST(SUM(IF(i.action=1,i.amount,0)) AS CHAR),
  ' risk_accounts=',(SELECT COUNT(*) FROM t_option_risk_account risk
    WHERE risk.tenant_id=996031 AND risk.user_id=880005
      AND risk.account_id=0 AND risk.settle_coin='USDT'),
  ' instructions=',COUNT(DISTINCT i.id),
  ' success=',COUNT(DISTINCT IF(i.status=3,i.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(i.reconciliation_status=2,i.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' freeze=',CAST(SUM(IF(i.action=1,i.amount,0)) AS CHAR),
  ' release=',CAST(SUM(IF(i.action=3,i.amount,0)) AS CHAR),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=880005
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' available=',(SELECT CAST(available_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=880005
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=880005
      AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_order o
JOIN t_option_contract c ON c.tenant_id=o.tenant_id AND c.id=o.contract_id
LEFT JOIN t_option_asset_instruction i ON i.tenant_id=o.tenant_id AND i.order_id=o.id
LEFT JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
  AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
WHERE c.tenant_id=996031
  AND c.contract_code IN ('P0-PORTFOLIO-CROSS-ACCOUNT-A','P0-PORTFOLIO-CROSS-ACCOUNT-B');
SELECT CONCAT('concurrent_trading_limit=',contract.contract_code,
  ' orders=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND orders.status=4),
  ' clients=',(SELECT COUNT(*) FROM t_option_client_order_key client_key
    JOIN t_option_order orders ON orders.tenant_id=client_key.tenant_id AND orders.id=client_key.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' evaluated=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='ORDER_CONTROL_EVALUATED'),
  ' rejected=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='ORDER_REJECTED'),
  ' reasons=',(SELECT GROUP_CONCAT(DISTINCT event.reason) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='ORDER_REJECTED'),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND instruction.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND instruction.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
      AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' wallet_total=',(SELECT CAST(SUM(wallet.total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND ((contract.contract_code='P0-CONCURRENT-USER-LONG-LIMIT' AND wallet.user_id=95101)
        OR (contract.contract_code='P0-CONCURRENT-OI-LIMIT' AND wallet.user_id BETWEEN 95201 AND 95220))),
  ' wallet_frozen=',(SELECT CAST(SUM(wallet.frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND ((contract.contract_code='P0-CONCURRENT-USER-LONG-LIMIT' AND wallet.user_id=95101)
        OR (contract.contract_code='P0-CONCURRENT-OI-LIMIT' AND wallet.user_id BETWEEN 95201 AND 95220))))
FROM t_option_contract contract
WHERE contract.tenant_id=996031
  AND contract.contract_code IN ('P0-CONCURRENT-USER-LONG-LIMIT','P0-CONCURRENT-OI-LIMIT')
ORDER BY contract.contract_code;
SELECT CONCAT('kill_switch_release_barrier=',contract.contract_code,
  ' orders=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND orders.status=4),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade trade
    WHERE trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id),
  ' activated=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.user_id=96001
      AND event.event_type='KILL_SWITCH_ACTIVATED'),
  ' released=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.user_id=96001
      AND event.event_type='KILL_SWITCH_RELEASED'),
  ' rejected=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.user_id=96001
      AND event.event_type='ORDER_REJECTED' AND event.reason='USER_KILL_SWITCH'),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND instruction.status=3),
  ' canceled_instruction=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND instruction.status=6),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND instruction.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
      AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
  ' wallet_total=',(SELECT CAST(SUM(wallet.total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND wallet.user_id IN (96000,96001,96002)),
  ' wallet_frozen=',(SELECT CAST(SUM(wallet.frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND wallet.user_id IN (96000,96001,96002)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-KILL-SWITCH-RELEASE-BARRIER';
SELECT CONCAT('kill_switch_match_race=',contract.contract_code,
  ' orders=',COUNT(DISTINCT orders.id),
  ' canceled=',COUNT(DISTINCT IF(orders.status=4,orders.id,NULL)),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade trade
    WHERE trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' wallet_total=',(SELECT CAST(SUM(wallet.total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND wallet.user_id IN (96101,96102)),
  ' wallet_frozen=',(SELECT CAST(SUM(wallet.frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5 AND wallet.coin='USDT'
      AND wallet.user_id IN (96101,96102)))
FROM t_option_contract contract
JOIN t_option_order orders ON orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=orders.tenant_id AND instruction.order_id=orders.id
LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
  AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-KILL-SWITCH-MATCH-RACE'
GROUP BY contract.id,contract.contract_code,contract.tenant_id;
SELECT CONCAT('circuit_breaker_batch=',contract.contract_code,
  ' contract_status=',contract.status,
  ' orders=',COUNT(DISTINCT orders.id),
  ' canceled=',COUNT(DISTINCT IF(orders.status=4,orders.id,NULL)),
  ' halt=',halt.cancel_total,'/',halt.cancel_success,'/',halt.cancel_failed,
  ' last_error=',halt.last_error_msg,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade trade
    WHERE trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id),
  ' positions=',(SELECT COUNT(*) FROM t_option_position position
    WHERE position.tenant_id=contract.tenant_id AND position.contract_id=contract.id),
  ' wallet_total=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=96201
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' wallet_frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.user_id=96201
      AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_contract contract
JOIN t_option_trading_halt halt
  ON halt.tenant_id=contract.tenant_id AND halt.contract_id=contract.id AND halt.status=1
JOIN t_option_order orders ON orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=orders.tenant_id AND instruction.order_id=orders.id
LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
  AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-CIRCUIT-BREAKER-BATCH-101'
GROUP BY contract.id,contract.contract_code,contract.tenant_id,contract.status,
  halt.cancel_total,halt.cancel_success,halt.cancel_failed,halt.last_error_msg;
SELECT CONCAT('mmp=',contract.contract_code,
  ' configs=',(SELECT COUNT(*) FROM t_option_mmp_config config
    WHERE config.tenant_id=contract.tenant_id AND config.contract_id=contract.id),
  ' active=',(SELECT COUNT(*) FROM t_option_mmp_config config
    WHERE config.tenant_id=contract.tenant_id AND config.contract_id=contract.id AND config.status=1),
  ' primary=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'),
  ' filled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a' AND orders.status=3),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a' AND orders.status=4),
  ' other_canceled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-b' AND orders.status=4),
  ' ordinary_canceled=',(SELECT COUNT(*) FROM t_option_order orders
    WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp<>1 AND orders.status=4),
  ' releases=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE instruction.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'
      AND instruction.instruction_no LIKE '%-CONTROL-RELEASE'),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE instruction.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'
      AND instruction.instruction_no LIKE '%-CONTROL-RELEASE' AND instruction.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE instruction.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'
      AND instruction.instruction_no LIKE '%-CONTROL-RELEASE' AND instruction.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
    WHERE instruction.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'
      AND instruction.instruction_no LIKE '%-CONTROL-RELEASE'),
  ' max_retry=',(SELECT COALESCE(MAX(instruction.retry_count),0)
    FROM t_option_asset_instruction instruction
    JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
    WHERE instruction.tenant_id=contract.tenant_id AND orders.contract_id=contract.id
      AND orders.mmp=1 AND orders.mmp_group='desk-a'
      AND instruction.instruction_no LIKE '%-CONTROL-RELEASE'),
  ' events=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='MMP_TRIGGERED'),'/',
    (SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='MMP_ORDER_CANCELED'),'/',
    (SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='MMP_MANUAL_RESET'))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P1-MMP-REAL-ASSET-BATCH-102';
SELECT CONCAT('trade_correction=',contract.contract_code,
  ' case_status=',correction.status,
  ' requester=',correction.requested_by,
  ' reviewer=',correction.reviewed_by,
  ' original_trade=',trade.trade_no,'/',CAST(trade.price AS CHAR),'/',CAST(trade.qty AS CHAR),
  ' legs=',COUNT(DISTINCT leg.id),
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' net=',(SELECT CAST(COALESCE(SUM(CASE
      WHEN net_flow.op_type=1 THEN net_flow.change_amount
      WHEN net_flow.op_type=2 THEN -net_flow.change_amount ELSE 0 END),0) AS CHAR)
    FROM t_option_asset_instruction net_instruction
    JOIN t_asset_flow net_flow
      ON net_flow.tenant_id=net_instruction.tenant_id
      AND net_flow.biz_no=net_instruction.instruction_no
    WHERE net_instruction.tenant_id=correction.tenant_id
      AND net_instruction.biz_no=correction.case_no),
  ' events=',
    (SELECT COUNT(*) FROM t_option_trading_control_event event
      WHERE event.tenant_id=correction.tenant_id AND event.contract_id=correction.contract_id
        AND event.event_type='TRADE_CORRECTION_CREATED' AND event.detail LIKE CONCAT('%case_no=',correction.case_no,'%')),'/',
    (SELECT COUNT(*) FROM t_option_trading_control_event event
      WHERE event.tenant_id=correction.tenant_id AND event.contract_id=correction.contract_id
        AND event.event_type='TRADE_CORRECTION_APPROVED' AND event.detail LIKE CONCAT('%case_no=',correction.case_no,'%')),'/',
    (SELECT COUNT(*) FROM t_option_trading_control_event event
      WHERE event.tenant_id=correction.tenant_id AND event.contract_id=correction.contract_id
        AND event.event_type='TRADE_CORRECTION_COMPLETED' AND event.detail LIKE CONCAT('%case_no=',correction.case_no,'%')))
FROM t_option_trade_correction correction
JOIN t_option_contract contract
  ON contract.tenant_id=correction.tenant_id AND contract.id=correction.contract_id
JOIN t_option_trade trade
  ON trade.tenant_id=correction.tenant_id AND trade.id=correction.trade_id
JOIN t_option_trade_correction_leg leg
  ON leg.tenant_id=correction.tenant_id AND leg.correction_id=correction.id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=correction.tenant_id AND instruction.biz_no=correction.case_no
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
WHERE correction.tenant_id=996031 AND contract.contract_code='P1-TRADE-CORRECTION-CALL'
GROUP BY contract.contract_code,correction.id,correction.status,correction.requested_by,
  correction.reviewed_by,correction.tenant_id,correction.contract_id,correction.case_no,
  trade.trade_no,trade.price,trade.qty;
SELECT CONCAT('user_cancel=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled_orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' canceled_before_freeze=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=6),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (171,172)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (171,172)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-USER-CANCEL-CALL';
SELECT CONCAT('ioc_partial=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.status=3 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.reconciliation_status=2 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' release=',(SELECT CAST(amount AS CHAR) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND o.order_type=4 AND i.instruction_no=CONCAT(o.order_no,'-IMMEDIATE-RELEASE')),
  ' outbox=',(SELECT COUNT(*) FROM t_option_outbox o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=3),
  ' inbox=',(SELECT COUNT(*) FROM t_option_inbox i
    WHERE i.tenant_id=contract.tenant_id AND i.contract_id=contract.id AND i.status=2),
  ' positions=',(SELECT COUNT(*) FROM t_option_position p
    WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' margin_lots=',(SELECT COUNT(*) FROM t_option_margin_lot lot
    WHERE lot.tenant_id=contract.tenant_id AND lot.contract_id=contract.id
      AND lot.initial_margin=50 AND lot.remaining_margin=50),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (173,174,175)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (173,174,175)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-IOC-PARTIAL-CALL';
SELECT CONCAT('fok_insufficient=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled_orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' buyer_filled=',(SELECT CAST(filled_qty AS CHAR) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.order_type=5),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' positions=',(SELECT COUNT(*) FROM t_option_position p
    WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (176,177)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (176,177)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-FOK-INSUFFICIENT-CALL';
SELECT CONCAT('market_protection=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' filled=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=3),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.status=3 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=contract.tenant_id AND i.reconciliation_status=2 AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE i.tenant_id=contract.tenant_id AND
      (i.order_id IN (SELECT id FROM t_option_order o
        WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id)
       OR i.trade_id IN (SELECT id FROM t_option_trade t
        WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id))),
  ' release=',(SELECT CAST(i.amount AS CHAR) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND o.order_type=2 AND i.instruction_no=CONCAT(o.order_no,'-RELEASE-REMAINDER')),
  ' outbox=',(SELECT COUNT(*) FROM t_option_outbox o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=3),
  ' inbox=',(SELECT COUNT(*) FROM t_option_inbox i
    WHERE i.tenant_id=contract.tenant_id AND i.contract_id=contract.id AND i.status=2),
  ' positions=',(SELECT COUNT(*) FROM t_option_position p
    WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' margin_lots=',(SELECT COUNT(*) FROM t_option_margin_lot lot
    WHERE lot.tenant_id=contract.tenant_id AND lot.contract_id=contract.id
      AND lot.initial_margin=50 AND lot.remaining_margin=50),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (181,182,183)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (181,182,183)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-MARKET-PROTECTION-CALL';
SELECT CONCAT('post_only=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' would_take=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND o.cancel_reason='POST_ONLY_WOULD_TAKE' AND o.filled_qty=0),
  ' rested_then_user_canceled=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND o.order_type=3 AND o.price=8 AND o.cancel_reason='USER_CANCEL'),
  ' trades=',(SELECT COUNT(*) FROM t_option_trade t
    WHERE t.tenant_id=contract.tenant_id AND t.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' positions=',(SELECT COUNT(*) FROM t_option_position p
    WHERE p.tenant_id=contract.tenant_id AND p.contract_id=contract.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (184,185,186)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (184,185,186)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-POST-ONLY-CALL';
SELECT CONCAT('admin_force_cancel=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled_orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' canceled_before_freeze=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=6),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' audit_events=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='ADMIN_FORCE_CANCEL_ORDER'),
  ' operator_9003=',(SELECT COUNT(*) FROM t_option_trading_control_event event
    WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id
      AND event.event_type='ADMIN_FORCE_CANCEL_ORDER' AND event.operator_id=9003
      AND event.reason='P0_ADMIN_FORCE_CANCEL'),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (191,192)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id IN (191,192)))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-ADMIN-FORCE-CANCEL-CALL';
SELECT CONCAT('cancel_funding_race=',
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' canceled_orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND o.status=4),
  ' client_keys=',(SELECT COUNT(*) FROM t_option_client_order_key key_item
    JOIN t_option_order o ON o.tenant_id=key_item.tenant_id AND o.id=key_item.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=3),
  ' canceled_before_freeze=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id AND i.status=6),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id
      AND i.reconciliation_status=2),
  ' flows=',(SELECT COUNT(DISTINCT flow.id) FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
     AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' terminal_paths=',(SELECT SUM(i.status=6)+SUM(i.status=3)/2
    FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    WHERE o.tenant_id=contract.tenant_id AND o.contract_id=contract.id),
  ' duplicate_instruction_nos=',(SELECT COUNT(*) FROM (
    SELECT i.instruction_no FROM t_option_asset_instruction i
    JOIN t_option_order o ON o.tenant_id=i.tenant_id AND o.id=i.order_id
    JOIN t_option_contract c ON c.tenant_id=o.tenant_id AND c.id=o.contract_id
    WHERE c.tenant_id=996031 AND c.contract_code='P0-CANCEL-FUNDING-RACE-CALL'
    GROUP BY i.instruction_no HAVING COUNT(*)>1
  ) duplicate_instruction),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id=193),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=contract.tenant_id AND wallet.wallet_type=5
      AND wallet.coin='USDT' AND wallet.user_id=193))
FROM t_option_contract contract
WHERE contract.tenant_id=996031 AND contract.contract_code='P0-CANCEL-FUNDING-RACE-CALL';
SELECT CONCAT('multi_instance_kill=',
  ' status=',instruction.status,
  ' retry=',instruction.retry_count,
  ' reconciled=',instruction.reconciliation_status,
  ' flows=',(SELECT COUNT(*) FROM t_asset_flow flow
    WHERE flow.tenant_id=instruction.tenant_id AND flow.user_id=instruction.user_id
      AND flow.biz_no=instruction.target_biz_no),
  ' freezes=',(SELECT COUNT(*) FROM t_asset_freeze freeze_item
    WHERE freeze_item.tenant_id=instruction.tenant_id AND freeze_item.user_id=instruction.user_id
      AND freeze_item.biz_no=instruction.target_biz_no),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=instruction.tenant_id AND wallet.user_id=instruction.user_id
      AND wallet.wallet_type=5 AND wallet.coin=instruction.coin),
  ' available=',(SELECT CAST(available_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=instruction.tenant_id AND wallet.user_id=instruction.user_id
      AND wallet.wallet_type=5 AND wallet.coin=instruction.coin),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=instruction.tenant_id AND wallet.user_id=instruction.user_id
      AND wallet.wallet_type=5 AND wallet.coin=instruction.coin))
FROM t_option_asset_instruction instruction
WHERE instruction.tenant_id=996031
  AND instruction.instruction_no='P0-MULTI-INSTANCE-KILL-FREEZE';
SELECT CONCAT('isolated_liquidation=',liquidation.status,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
  ' reconciled=',COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
  ' flows=',COUNT(DISTINCT flow.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (141,142,143,144)),
  ' frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (141,142,143,144)),
  ' collateral=',CAST(liquidation.collateral_amount AS CHAR),
  ' fee=',CAST(liquidation.liquidation_fee AS CHAR),
  ' source_return=',CAST(source.total_return AS CHAR),
  ' takeover_margin=',CAST(takeover.margin_amount AS CHAR),
  ' takeover_maintenance=',CAST(takeover.maintenance_margin AS CHAR))
FROM t_option_liquidation liquidation
JOIN t_option_position source
  ON source.tenant_id=liquidation.tenant_id AND source.id=liquidation.position_id
JOIN t_option_position takeover
  ON takeover.tenant_id=liquidation.tenant_id AND takeover.id=liquidation.takeover_position_id
JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=liquidation.tenant_id AND instruction.liquidation_id=liquidation.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id
 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
WHERE liquidation.tenant_id=996031 AND liquidation.liquidation_no='P0-ISOLATED-SHORT-LIQUIDATION'
GROUP BY liquidation.id,liquidation.status,source.id,takeover.id;
SELECT CONCAT('insurance_inventory_exit=',exit_item.status,
  ' active_key=',exit_item.active_key,
  ' orders=',(SELECT COUNT(*) FROM t_option_order o
    WHERE o.tenant_id=exit_item.tenant_id AND o.id=exit_item.order_id
      AND o.source=4 AND o.order_type=4 AND o.reduce_only=1),
  ' filled=',CAST(exit_order.filled_qty AS CHAR),
  ' unfilled=',CAST(exit_order.unfilled_qty AS CHAR),
  ' remaining=',CAST(exit_position.position_qty AS CHAR),
  ' margin=',CAST(exit_position.margin_amount AS CHAR),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=exit_item.tenant_id
      AND (i.order_id=exit_item.order_id OR i.margin_lot_id IN (
        SELECT l.id FROM t_option_margin_lot l
        WHERE l.tenant_id=exit_item.tenant_id AND l.position_id=exit_item.position_id
      ))),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=exit_item.tenant_id AND i.status=3
      AND (i.order_id=exit_item.order_id OR i.margin_lot_id IN (
        SELECT l.id FROM t_option_margin_lot l
        WHERE l.tenant_id=exit_item.tenant_id AND l.position_id=exit_item.position_id
      ))),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=exit_item.tenant_id AND i.reconciliation_status=2
      AND (i.order_id=exit_item.order_id OR i.margin_lot_id IN (
        SELECT l.id FROM t_option_margin_lot l
        WHERE l.tenant_id=exit_item.tenant_id AND l.position_id=exit_item.position_id
      ))),
  ' flows=',(SELECT COUNT(DISTINCT flow.id)
    FROM t_option_asset_instruction i
    JOIN t_asset_flow flow ON flow.tenant_id=i.tenant_id
      AND flow.biz_no=CASE WHEN i.action=1 THEN i.target_biz_no ELSE i.instruction_no END
    WHERE i.tenant_id=exit_item.tenant_id
      AND (i.order_id=exit_item.order_id OR i.margin_lot_id IN (
        SELECT l.id FROM t_option_margin_lot l
        WHERE l.tenant_id=exit_item.tenant_id AND l.position_id=exit_item.position_id
      ))),
  ' events=',(SELECT COUNT(*) FROM t_option_trading_control_event e
    WHERE e.tenant_id=exit_item.tenant_id AND e.event_type LIKE 'INSURANCE_EXIT_%'),
  ' wallet=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=exit_item.tenant_id AND wallet.user_id=exit_item.insurance_user_id
      AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=exit_item.tenant_id AND wallet.user_id=exit_item.insurance_user_id
      AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_insurance_inventory_exit exit_item
JOIN t_option_order exit_order
  ON exit_order.tenant_id=exit_item.tenant_id AND exit_order.id=exit_item.order_id
JOIN t_option_position exit_position
  ON exit_position.tenant_id=exit_item.tenant_id AND exit_position.id=exit_item.position_id
WHERE exit_item.tenant_id=996031;
SELECT CONCAT('partial_liquidation=',liquidation.status,
  ' retry=',liquidation.retry_count,
  ' instructions=',COUNT(DISTINCT instruction.id),
  ' success=',COUNT(DISTINCT CASE WHEN instruction.status=3 THEN instruction.id END),
  ' reconciled=',COUNT(DISTINCT CASE WHEN instruction.reconciliation_status=2 THEN instruction.id END),
  ' flows=',COUNT(DISTINCT flow.id),
  ' source_qty=',CAST(position.position_qty AS CHAR),
  ' source_available=',CAST(position.available_qty AS CHAR),
  ' source_frozen=',CAST(position.frozen_qty AS CHAR),
  ' source_margin=',CAST(position.margin_amount AS CHAR),
  ' source_maintenance=',CAST(position.maintenance_margin AS CHAR),
  ' source_return=',CAST(position.total_return AS CHAR),
  ' lot_qty=',CAST(source_lot.remaining_quantity AS CHAR),
  ' lot_margin=',CAST(source_lot.remaining_margin AS CHAR),
  ' lot_pending=',CAST(source_lot.pending_margin AS CHAR),
  ' takeover_qty=',CAST(takeover.position_qty AS CHAR),
  ' takeover_margin=',CAST(takeover.margin_amount AS CHAR),
  ' wallet_total=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (145,156,157)),
  ' wallet_frozen=',(SELECT CAST(SUM(frozen_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (145,156,157)))
FROM t_option_liquidation liquidation
JOIN t_option_contract contract
  ON contract.tenant_id=liquidation.tenant_id AND contract.id=liquidation.contract_id
JOIN t_option_position position
  ON position.tenant_id=liquidation.tenant_id AND position.id=liquidation.position_id
JOIN t_option_position takeover
  ON takeover.tenant_id=liquidation.tenant_id AND takeover.id=liquidation.takeover_position_id
JOIN t_option_margin_lot source_lot
  ON source_lot.tenant_id=liquidation.tenant_id AND source_lot.position_id=position.id
LEFT JOIN t_option_asset_instruction instruction
  ON instruction.tenant_id=liquidation.tenant_id AND instruction.liquidation_id=liquidation.id
LEFT JOIN t_asset_flow flow
  ON flow.tenant_id=instruction.tenant_id
 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
WHERE liquidation.tenant_id=996031
  AND contract.contract_code='P0-ISOLATED-PARTIAL-LIQUIDATION-CALL'
GROUP BY liquidation.id,liquidation.status,liquidation.retry_count,position.id,takeover.id,source_lot.id;
SELECT CONCAT('deficit_liquidation=',liquidation.status,
  ' retry=',liquidation.retry_count,
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=liquidation.tenant_id AND i.liquidation_id=liquidation.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=liquidation.tenant_id AND i.liquidation_id=liquidation.id AND i.status=3),
  ' reconciled=',(SELECT COUNT(*) FROM t_option_asset_instruction i
    WHERE i.tenant_id=liquidation.tenant_id AND i.liquidation_id=liquidation.id AND i.reconciliation_status=2),
  ' max_retry=',(SELECT MAX(retry_count) FROM t_option_asset_instruction i
    WHERE i.tenant_id=liquidation.tenant_id AND i.liquidation_id=liquidation.id),
  ' collateral=',CAST(liquidation.collateral_amount AS CHAR),
  ' deficit=',CAST(liquidation.deficit_amount AS CHAR),
  ' insurance=',CAST(liquidation.insurance_fund_amount AS CHAR),
  ' backstop=',CAST(liquidation.backstop_amount AS CHAR),
  ' insurance_balance=',CAST(15-(SELECT COALESCE(SUM(amount),0) FROM t_asset_platform_flow
    WHERE tenant_id=liquidation.tenant_id AND biz_id=liquidation.id
      AND account_type='INSURANCE_FUND' AND op_type=2) AS CHAR),
  ' backstop_balance=',CAST(0-(SELECT COALESCE(SUM(amount),0) FROM t_asset_platform_flow
    WHERE tenant_id=liquidation.tenant_id AND biz_id=liquidation.id
      AND account_type='OPTION_BACKSTOP' AND op_type=2) AS CHAR),
  ' platform_flows=',(SELECT COUNT(*) FROM t_asset_platform_flow
    WHERE tenant_id=liquidation.tenant_id AND biz_id=liquidation.id),
  ' wallets=',(SELECT CAST(SUM(total_amount) AS CHAR) FROM t_user_asset
    WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (146,147,148)),
  ' conserved=',CAST(
    (SELECT SUM(total_amount) FROM t_user_asset
      WHERE tenant_id=liquidation.tenant_id AND wallet_type=5 AND coin='USDT' AND user_id IN (146,147,148))
    + 15 - (SELECT COALESCE(SUM(amount),0) FROM t_asset_platform_flow
      WHERE tenant_id=liquidation.tenant_id AND biz_id=liquidation.id AND op_type=2)
    AS CHAR))
FROM t_option_liquidation liquidation
WHERE liquidation.tenant_id=996031
  AND liquidation.liquidation_no='P0-LIQUIDATION-DEFICIT-RECOVERY';
SELECT CONCAT('insurance_ledger_semantics=',
  ' raw_inflow=',COALESCE(SUM(CASE WHEN flow_type IN (1,3) THEN ABS(amount) ELSE 0 END),0),
  ' raw_outflow=',COALESCE(SUM(CASE WHEN flow_type IN (2,4) THEN ABS(amount) ELSE 0 END),0),
  ' signed_net=',COALESCE(SUM(CASE WHEN flow_type IN (2,4) THEN -ABS(amount) ELSE ABS(amount) END),0),
  ' nonpositive_rows=',COALESCE(SUM(amount<=0),0))
FROM t_option_insurance_fund_flow
WHERE tenant_id=996031;
SELECT CONCAT('portfolio_partial_liquidation=',liquidation.status,
  ' quantity=',CAST(liquidation.quantity AS CHAR),
  ' collateral=',CAST(liquidation.collateral_amount AS CHAR),
  ' fee=',CAST(liquidation.liquidation_fee AS CHAR),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=liquidation.tenant_id AND instruction.liquidation_id=liquidation.id),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=liquidation.tenant_id AND instruction.liquidation_id=liquidation.id AND instruction.status=3),
  ' source_qty=',CAST(source.position_qty AS CHAR),
  ' source_margin=',CAST(source.margin_amount AS CHAR),
  ' source_return=',CAST(source.total_return AS CHAR),
  ' takeover_qty=',CAST(takeover.position_qty AS CHAR),
  ' takeover_margin=',CAST(takeover.margin_amount AS CHAR),
  ' wallet_total=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=158 AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' wallet_available=',(SELECT CAST(available_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=158 AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' wallet_frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=158 AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_liquidation liquidation
JOIN t_option_position source
  ON source.tenant_id=liquidation.tenant_id AND source.id=liquidation.position_id
JOIN t_option_position takeover
  ON takeover.tenant_id=liquidation.tenant_id AND takeover.id=liquidation.takeover_position_id
WHERE liquidation.tenant_id=996031 AND liquidation.user_id=158
  AND liquidation.liquidation_scope=2;
SELECT CONCAT('portfolio_liquidation_sequential=',COUNT(*),
  ' done=',SUM(liquidation.status=3),
  ' wallet_scope=',SUM(liquidation.liquidation_scope=2 AND liquidation.account_id=0),
  ' risk_reduced=',SUM(liquidation.portfolio_maintenance_before>liquidation.portfolio_maintenance_after),
  ' residual_protected=',SUM(liquidation.portfolio_collateral_after>=liquidation.portfolio_initial_after),
  ' config_versions=',GROUP_CONCAT(DISTINCT liquidation.portfolio_risk_config_version),
  ' source_closed=',(SELECT COUNT(*) FROM t_option_position position
    WHERE position.tenant_id=996031 AND position.user_id=151 AND position.status=2),
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_liquidation evidence
      ON evidence.tenant_id=instruction.tenant_id AND evidence.id=instruction.liquidation_id
    WHERE evidence.tenant_id=996031 AND evidence.user_id=151 AND evidence.liquidation_scope=2),
  ' success=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    JOIN t_option_liquidation evidence
      ON evidence.tenant_id=instruction.tenant_id AND evidence.id=instruction.liquidation_id
    WHERE evidence.tenant_id=996031 AND evidence.user_id=151 AND evidence.liquidation_scope=2
      AND instruction.status=3),
  ' wallet_total=',(SELECT CAST(total_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=151 AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' available=',(SELECT CAST(available_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=151 AND wallet.wallet_type=5 AND wallet.coin='USDT'),
  ' frozen=',(SELECT CAST(frozen_amount AS CHAR) FROM t_user_asset wallet
    WHERE wallet.tenant_id=996031 AND wallet.user_id=151 AND wallet.wallet_type=5 AND wallet.coin='USDT'))
FROM t_option_liquidation liquidation
WHERE liquidation.tenant_id=996031 AND liquidation.user_id=151
  AND liquidation.liquidation_scope=2;
SELECT CONCAT('portfolio_liquidation_recovery_cancel=',liquidation.status,
  ' retry=',liquidation.retry_count,
  ' instructions=',(SELECT COUNT(*) FROM t_option_asset_instruction instruction
    WHERE instruction.tenant_id=liquidation.tenant_id AND instruction.liquidation_id=liquidation.id),
  ' error=',liquidation.last_error_msg)
FROM t_option_liquidation liquidation
WHERE liquidation.tenant_id=996031 AND liquidation.user_id=155
  AND liquidation.liquidation_scope=2;
SELECT CONCAT('duplicate_option_freeze_keys=',COUNT(*),' oldest=',COALESCE(MIN(oldest),0))
FROM (
  SELECT MIN(create_times) oldest
  FROM t_asset_freeze
  WHERE tenant_id=996032 AND biz_type='option' AND TRIM(biz_no)<>''
  GROUP BY tenant_id,biz_type,scene_type,biz_no
  HAVING COUNT(*)>1
) duplicate_keys;
SELECT CONCAT('unique_legacy_idempotency=',
  SUM(tenant_id=996030 AND biz_no='P0-MIGRATION-UNIQUE'),
  ' duplicate_legacy_idempotency=',
  SUM(tenant_id=996032 AND biz_no='P0-MIGRATION-DUP'))
FROM t_asset_idempotent WHERE biz_type='option' AND scene_type='place_order';
"
FAILED=0
echo "P0 Option/Asset RPC acceptance passed"
