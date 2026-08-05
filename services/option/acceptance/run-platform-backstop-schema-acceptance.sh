#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
DEPLOY_ROOT="$REPO_ROOT/deploy"
# shellcheck disable=SC1091
. "$DEPLOY_ROOT/common/scripts/dev-compose.sh"
COMPOSE_FILE=${OPTION_BACKSTOP_SCHEMA_COMPOSE_FILE:-}
DATABASE=${OPTION_BACKSTOP_SCHEMA_DATABASE:-wklive_option_backstop_schema_acceptance}
MYSQL_PASSWORD=${OPTION_BACKSTOP_SCHEMA_MYSQL_PASSWORD:-123456}

case "$DATABASE" in
  wklive_option_backstop_schema_acceptance|wklive_option_backstop_schema_acceptance_[A-Za-z0-9_]*) ;;
  *)
    echo "refusing unsafe acceptance database name: $DATABASE" >&2
    exit 1
    ;;
esac

if [ -n "$COMPOSE_FILE" ]; then compose_mysql=$(docker compose -f "$COMPOSE_FILE" ps -q mysql); else compose_mysql=$(dev_compose ps -q mysql); fi
MYSQL_CONTAINER=${OPTION_BACKSTOP_SCHEMA_MYSQL_CONTAINER:-$compose_mysql}
if [ -z "$MYSQL_CONTAINER" ]; then
  echo "mysql from the development Compose environment must already be running" >&2
  exit 1
fi

mysql_cli() {
  docker exec "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$@"
}

cleanup() {
  mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`;" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

echo "creating isolated database $DATABASE"
mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`; CREATE DATABASE \`$DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/asset/asset.sql"

for pass in 1 2; do
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < \
    "$REPO_ROOT/services/asset/migrations/20260802_asset_backstop_policy_limits.sql"
  echo "migration_pass=$pass"
done

schema_facts=$(mysql_cli -N "$DATABASE" -e "
SELECT CONCAT(
  (SELECT COUNT(*) FROM information_schema.tables
    WHERE table_schema=DATABASE() AND table_name IN
      ('t_asset_backstop_policy','t_asset_backstop_usage_daily','t_asset_backstop_cover')),
  ':',
  (SELECT COUNT(*) FROM information_schema.triggers
    WHERE trigger_schema=DATABASE() AND trigger_name LIKE 'trg_asset_backstop_%'),
  ':',
  (SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_asset_backstop_cover'
      AND column_name IN ('policy_id','policy_version','policy_mode','daily_used_before',
        'daily_used_after','balance_floor','balance_before','balance_after'))
);")
if [ "$schema_facts" != "3:9:8" ]; then
  echo "unexpected schema facts tables:triggers:cover_columns=$schema_facts" >&2
  exit 1
fi
echo "schema_facts=$schema_facts"

expect_rejected() {
  case_name=$1
  sql=$2
  if mysql_cli "$DATABASE" -e "$sql" >/dev/null 2>&1; then
    echo "$case_name=BYPASS" >&2
    exit 1
  fi
  echo "$case_name=REJECTED"
}

expect_rejected direct_approved_insert "
SET @now_ms=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED);
INSERT INTO t_asset_backstop_policy
(tenant_id,coin,request_no,version,mode,per_request_limit,daily_limit,balance_floor,
 effective_from,effective_until,status,reason,evidence_ref,created_by,reviewed_by,review_reason,create_times,update_times)
VALUES (900101,'USDT','DIRECT-APPROVED',1,1,0,0,0,@now_ms+60000,@now_ms+86400000,2,
 'must fail','test://direct',11,22,'must fail',@now_ms,@now_ms);"

mysql_cli "$DATABASE" -e "
SET @now_ms=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED);
INSERT INTO t_asset_backstop_policy
(tenant_id,coin,request_no,version,mode,per_request_limit,daily_limit,balance_floor,
 effective_from,effective_until,status,reason,evidence_ref,created_by,reviewed_by,review_reason,create_times,update_times)
VALUES (900101,'USDT','VALID-DRAFT',1,3,10,100,-50,@now_ms+60000,@now_ms+86400000,1,
 'isolated migration verification','test://valid-draft',11,0,'',@now_ms,@now_ms);"

expect_rejected self_review "
UPDATE t_asset_backstop_policy SET status=2,reviewed_by=11,review_reason='self',
 update_times=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
WHERE tenant_id=900101 AND request_no='VALID-DRAFT';"

mysql_cli "$DATABASE" -e "
UPDATE t_asset_backstop_policy SET status=2,reviewed_by=22,review_reason='four eyes',
 update_times=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)
WHERE tenant_id=900101 AND request_no='VALID-DRAFT';"
echo "four_eyes_review=ACCEPTED"

expect_rejected approved_policy_rewrite "
UPDATE t_asset_backstop_policy SET review_reason='rewrite',
 update_times=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)+1
WHERE tenant_id=900101 AND request_no='VALID-DRAFT';"
expect_rejected approved_policy_delete "
DELETE FROM t_asset_backstop_policy WHERE tenant_id=900101 AND request_no='VALID-DRAFT';"

expect_rejected fake_usage_policy "
INSERT INTO t_asset_backstop_usage_daily
(tenant_id,coin,usage_day,covered_amount,last_policy_id,create_times,update_times)
VALUES (900102,'BTC','20260802',1,999999,1,1);"

expect_rejected fake_cover_policy "
INSERT INTO t_asset_backstop_cover
(tenant_id,platform_account_id,coin,liquidation_id,liquidation_no,policy_id,policy_version,
 policy_mode,covered_amount,daily_used_before,daily_used_after,balance_floor,balance_before,
 balance_after,status,create_times,update_times)
VALUES (900102,999999,'BTC',1,'FAKE-COVER-POLICY',999999,1,2,1,0,1,0,1,0,1,1,1);"

echo "platform_backstop_schema_acceptance=PASS"
