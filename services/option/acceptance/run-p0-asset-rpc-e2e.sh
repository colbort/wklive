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
  "$REPO_ROOT/services/option/migrations/20260730_x_settlement_price_rule.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_z_option_settlement_price_approval.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_zb_option_wallet_account_mirror.sql" \
  "$REPO_ROOT/services/option/migrations/20260730_zc_option_risk_net_value.sql" \
  "$REPO_ROOT/services/option/migrations/20260731_zx_option_margin_coin_evidence.sql" \
  "$REPO_ROOT/services/option/migrations/20260731_zy_option_settlement_price_evidence.sql" \
  "$REPO_ROOT/services/option/migrations/20260801_option_portfolio_liquidation.sql"
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
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task -run '^TestP0AssetMultiInstanceKillTakeover$' -count=1 -timeout=90s
)

echo "verifying evidence and cleanup scope"
mysql_cli -N "$DATABASE" -e "
SELECT CONCAT('risk_accounts=',COUNT(*),' account_id_sum=',COALESCE(SUM(account_id),0))
FROM t_option_risk_account WHERE tenant_id=996031;
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
