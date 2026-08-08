#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
REPO_ROOT=$(CDPATH= cd -- "$DEPLOY_ROOT/.." && pwd)
BASE_COMPOSE="$DEPLOY_ROOT/common/compose.base.yml"
PREPROD_COMPOSE="$SCRIPT_DIR/compose.yml"
ENV_FILE="${1:-${PREPROD_ENV_FILE:-$SCRIPT_DIR/.env}}"
failures=0

pass() {
  printf 'PASS  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  failures=$((failures + 1))
}

require_contains() {
  file="$1"
  pattern="$2"
  description="$3"
  if grep -Eq "$pattern" "$file"; then
    pass "$description"
  else
    fail "$description"
  fi
}

repository_checks() {
  require_contains "$PREPROD_COMPOSE" '^name:.*wklive-preprod' "pre-production uses an isolated Compose project"
  require_contains "$PREPROD_COMPOSE" 'ports: !override \[\]' "infrastructure host ports are removed"
  require_contains "$PREPROD_COMPOSE" 'SERVICE_MODE: pre' "services render in go-zero pre-release mode"
  require_contains "$PREPROD_COMPOSE" 'ADMIN_REQUEST_ENCRYPTION_MODE.*REQUIRED' "admin request encryption is fail-closed"
  require_contains "$PREPROD_COMPOSE" 'read_only: true' "application containers use a read-only root filesystem"
  require_contains "$PREPROD_COMPOSE" 'no-new-privileges:true' "application containers prohibit privilege escalation"
  require_contains "$PREPROD_COMPOSE" 'RELEASE_TAG.*required' "application images require an immutable release tag"
  require_contains "$PREPROD_COMPOSE" 'volumes: !override \[\]' "db-init rejects inherited source mounts"
  if ! grep -qE -- '- \.\.:/workspace' "$PREPROD_COMPOSE"; then pass "pre-production does not mount the repository workspace"; else fail "pre-production does not mount the repository workspace"; fi
  require_contains "$DEPLOY_ROOT/common/docker/Dockerfile.db-init" 'COPY --from=database-release /release /release' "database SQL is bundled into the release image"
  require_contains "$PREPROD_COMPOSE" 'DB_INIT_PROFILE: preprod' "pre-production business seed profile is explicit"
  require_contains "$REPO_ROOT/init.sql" 'INSERT INTO (`)?sys_menu(`)?' "management menus are part of initialization data"
  require_contains "$DEPLOY_ROOT/common/dbinit/main.go" '"super_admin", "超级管理员"' "system super administrator is seeded from release credentials"
  require_contains "$REPO_ROOT/services/system/system.sql" 'CREATE TABLE sys_job' "scheduled jobs are part of the System schema"
  require_contains "$DEPLOY_ROOT/common/dbinit/main.go" '20260730_add_option_jobs.sql' "Option scheduled jobs are replayed during initialization"
  require_contains "$DEPLOY_ROOT/common/dbinit/main.go" '20260803_add_staking_jobs.sql' "Staking scheduled jobs are replayed during initialization"
  require_contains "$SCRIPT_DIR/data/bootstrap.sql" "'BTCUSDT'.*'ETHUSDT'|'BTCUSDT'" "BTCUSDT baseline is versioned"
  require_contains "$SCRIPT_DIR/data/bootstrap.sql" "'ETHUSDT'" "ETHUSDT baseline is versioned"
  if ! grep -Eq "'([A-Z0-9]+USDT)'" "$SCRIPT_DIR/data/bootstrap.sql" ||
     ! grep -Eq "'BTCUSDT'|'ETHUSDT'" "$SCRIPT_DIR/data/bootstrap.sql"; then
    fail "pre-production baseline symbols are declared"
  else
    unexpected_symbols=$(grep -Eo "'[A-Z0-9]+USDT'" "$SCRIPT_DIR/data/bootstrap.sql" | sort -u | grep -Ev "^'(BTCUSDT|ETHUSDT)'$" || true)
    if [ -z "$unexpected_symbols" ]; then pass "pre-production baseline is limited to BTCUSDT and ETHUSDT"; else fail "pre-production baseline is limited to BTCUSDT and ETHUSDT"; fi
  fi
  require_contains "$SCRIPT_DIR/config/common.yaml" 'Encoding: json' "pre-production logs use structured JSON"
  require_contains "$SCRIPT_DIR/config/common.yaml" '__MYSQL_APP_USER__' "runtime services use a non-root database identity"
  require_contains "$SCRIPT_DIR/config/common.yaml" 'Pass: __REDIS_PASSWORD__' "runtime Redis connections require authentication"
  if grep -A1 '^AutomaticLiquidation:' "$REPO_ROOT/services/trade/etc/trade.yaml" | grep -q 'Enabled: false'; then pass "trade automatic liquidation defaults to disabled"; else fail "trade automatic liquidation defaults to disabled"; fi
  if grep -A1 '^CrossMarginTrading:' "$REPO_ROOT/services/trade/etc/trade.yaml" | grep -q 'Enabled: false'; then pass "trade cross-margin defaults to disabled"; else fail "trade cross-margin defaults to disabled"; fi
  option_scope_false=$(grep -A7 '^ProductScope:' "$REPO_ROOT/services/option/etc/option.yaml" | grep -c 'Enabled: false' | tr -d ' ')
  if [ "$option_scope_false" -eq 7 ]; then pass "all Option optional product capabilities default to disabled"; else fail "all Option optional product capabilities default to disabled"; fi

  for lock_file in \
    "$REPO_ROOT/app-packages/package-lock.json" \
    "$REPO_ROOT/admin-ui/package-lock.json" \
    "$REPO_ROOT/app-web/package-lock.json" \
    "$REPO_ROOT/chat-admin-ui/package-lock.json" \
    "$REPO_ROOT/liquidity-admin-ui/package-lock.json"; do
    if [ -s "$lock_file" ]; then
      pass "$(basename "$(dirname "$lock_file")") has a reproducible npm lock file"
    else
      fail "$(basename "$(dirname "$lock_file")") has a reproducible npm lock file"
    fi
  done
}

repository_checks

if [ "${1:-}" = "--repository-only" ]; then
  if [ "$failures" -eq 0 ]; then
    printf '\nREADY: pre-production repository artifacts passed.\n'
    exit 0
  fi
  printf '\nNOT READY: %d repository check(s) failed.\n' "$failures" >&2
  exit 1
fi

if [ ! -r "$ENV_FILE" ]; then
  fail "pre-production environment file is readable"
  printf '\nNOT READY: %d runtime check(s) failed.\n' "$failures" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a
PREPROD_PROJECT_NAME="${PREPROD_PROJECT_NAME:-wklive-preprod}"

compose() {
  docker compose \
    --project-directory "$DEPLOY_ROOT" \
    --env-file "$ENV_FILE" \
    -p "$PREPROD_PROJECT_NAME" \
    -f "$BASE_COMPOSE" \
    -f "$PREPROD_COMPOSE" \
    "$@"
}

long_running_services="
etcd mysql redis mongo kafka beanstalk-primary beanstalk-secondary
system-rpc user-rpc market-rpc payment-rpc asset-rpc option-rpc staking-rpc trade-rpc chat-rpc liquidity-rpc
payment-api app-api chat-admin-api chat-api admin-api liquidity-admin-api
admin-ui app-web chat-admin-ui liquidity-admin-ui
"

for service_name in $long_running_services; do
  container_id=$(compose ps -q "$service_name" 2>/dev/null)
  if [ -z "$container_id" ]; then
    fail "$service_name is running"
    continue
  fi
  state=$(docker inspect --format '{{.State.Status}}' "$container_id" 2>/dev/null)
  health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null)
  if [ "$state" = "running" ] && [ "$health" = "healthy" ]; then
    pass "$service_name is running and healthy"
  else
    fail "$service_name is running and healthy (state=${state:-unknown} health=${health:-unknown})"
  fi
done

for service_name in etcd mysql redis mongo kafka beanstalk-primary beanstalk-secondary; do
  container_id=$(compose ps -q "$service_name" 2>/dev/null)
  [ -n "$container_id" ] || continue
  if [ -z "$(docker port "$container_id" 2>/dev/null)" ]; then
    pass "$service_name has no host-published port"
  else
    fail "$service_name has no host-published port"
  fi
done

for one_shot in db-init kafka-init config-seed; do
  container_id=$(compose ps -a -q "$one_shot" 2>/dev/null)
  exit_code=$(docker inspect --format '{{.State.ExitCode}}' "$container_id" 2>/dev/null || true)
  if [ -n "$container_id" ] && [ "$exit_code" = "0" ]; then
    pass "$one_shot completed successfully"
  else
    fail "$one_shot completed successfully"
  fi
done

baseline_counts=$(compose exec -T -e BASELINE_TENANT_ID="${PREPROD_TENANT_ID:-0}" mysql sh -lc '
mysql -N -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "
SET @tenant_id=CAST($BASELINE_TENANT_ID AS UNSIGNED);
SELECT CONCAT(
  (SELECT COUNT(*) FROM sys_tenant WHERE id=@tenant_id AND enabled=1), CHAR(124),
  (SELECT COUNT(*) FROM t_itick_tenant_category WHERE tenant_id=@tenant_id AND enabled=1), CHAR(124),
  (SELECT COUNT(*) FROM t_itick_tenant_product tenant_product JOIN t_itick_product product ON product.id=tenant_product.product_id WHERE tenant_product.tenant_id=@tenant_id AND product.symbol IN (0x42544355534454,0x45544855534454)), CHAR(124),
  (SELECT COUNT(*) FROM t_trade_symbol WHERE tenant_id=@tenant_id AND symbol IN (0x42544355534454,0x45544855534454)), CHAR(124),
  (SELECT COUNT(*) FROM t_option_contract WHERE tenant_id=@tenant_id AND underlying_symbol IN (0x42544355534454,0x45544855534454)), CHAR(124),
  (SELECT COUNT(*) FROM t_stake_product WHERE tenant_id=@tenant_id AND coin_symbol IN (0x425443,0x455448)), CHAR(124),
  (SELECT COUNT(*) FROM t_itick_price_formula WHERE symbol IN (0x42544355534454,0x45544855534454)), CHAR(124),
  (SELECT COUNT(*) FROM t_itick_tenant_product tenant_product JOIN t_itick_product product ON product.id=tenant_product.product_id WHERE tenant_product.tenant_id=@tenant_id AND product.symbol NOT IN (0x42544355534454,0x45544855534454))
);" 2>/dev/null
' 2>/dev/null || true)
if [ "$baseline_counts" = "1|1|2|8|4|2|8|0" ]; then
  pass "pre-production tenant, Market, Trade, Option and Staking baseline data is complete"
else
  fail "pre-production tenant, Market, Trade, Option and Staking baseline data is complete (counts=${baseline_counts:-unavailable})"
fi

system_seed_status=$(compose exec -T mysql sh -lc '
mysql -N -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" -e "
SELECT CONCAT(
  IF((SELECT COUNT(*)
      FROM sys_user user_account
      JOIN sys_user_role user_role ON user_role.tenant_id=user_account.tenant_id AND user_role.user_id=user_account.id
      JOIN sys_role role_record ON role_record.tenant_id=user_role.tenant_id AND role_record.id=user_role.role_id
      WHERE user_account.tenant_id=0 AND user_account.app_scope=1 AND user_account.enabled=1
        AND role_record.code=0x73757065725F61646D696E AND role_record.enabled=1)=1,1,0), CHAR(124),
  IF((SELECT COUNT(*) FROM sys_menu WHERE app_scope=1 AND enabled=1)>0,1,0), CHAR(124),
  IF((SELECT COUNT(*)
      FROM sys_menu menu_record
      WHERE menu_record.app_scope=1 AND menu_record.enabled=1
        AND NOT EXISTS (
          SELECT 1 FROM sys_role role_record
          JOIN sys_role_menu role_menu ON role_menu.tenant_id=role_record.tenant_id AND role_menu.role_id=role_record.id
          WHERE role_record.tenant_id=0 AND role_record.app_scope=1
            AND role_record.code=0x73757065725F61646D696E AND role_menu.menu_id=menu_record.id
        ))=0,1,0), CHAR(124),
  IF((SELECT COUNT(DISTINCT perms) FROM sys_menu
      WHERE perms IN (0x7379733A6A6F623A6C697374,0x7379733A6A6F623A6C6F673A6C697374))=2,1,0), CHAR(124),
  IF((SELECT COUNT(*) FROM sys_job WHERE invoke_target IN (
        0x74726164652E50726F636573734F726465724D61746368696E67,
        0x74726164652E50726F63657373506F736974696F6E73,
        0x74726164652E50726F63657373436F6E7472616374536574746C656D656E7473,
        0x74726164652E50726F6365737354726164654576656E7473,
        0x74726164652E50726F636573735365636F6E6473536574746C656D656E7473))=5,1,0), CHAR(124),
  IF((SELECT COUNT(*) FROM sys_job WHERE invoke_target IN (
        0x6F7074696F6E2E50726F636573734173736574496E737472756374696F6E73,
        0x6F7074696F6E2E50726F6365737354726164654576656E7473,
        0x6F7074696F6E2E50726F636573735269736B4163636F756E7473,
        0x6F7074696F6E2E50726F636573734C69717569646174696F6E73,
        0x6F7074696F6E2E50726F63657373457865726369736573,
        0x6F7074696F6E2E50726F63657373436F6E74726163744C6966656379636C65,
        0x6F7074696F6E2E50726F636573734461696C795265636F6E63696C696174696F6E,
        0x6F7074696F6E2E436C65616E4D61726B6574536E617073686F7473))=8,1,0), CHAR(124),
  IF((SELECT COUNT(*) FROM sys_job WHERE invoke_target IN (
        0x7374616B696E672E50726F6365737352657761726473416E64536574746C654F7264657273,
        0x7374616B696E672E5265636F6E63696C655374616B696E67))=2,1,0)
);" 2>/dev/null
' 2>/dev/null || true)
if [ "$system_seed_status" = "1|1|1|1|1|1|1" ]; then
  pass "management menus, system super administrator and scheduled jobs are complete"
else
  fail "management menus, system super administrator and scheduled jobs are complete (status=${system_seed_status:-unavailable})"
fi

trade_config=$(compose exec -T etcd etcdctl --endpoints=http://127.0.0.1:2379 get /wklive/trade-rpc/config --print-value-only 2>/dev/null || true)
option_config=$(compose exec -T etcd etcdctl --endpoints=http://127.0.0.1:2379 get /wklive/option-rpc/config --print-value-only 2>/dev/null || true)
admin_config=$(compose exec -T etcd etcdctl --endpoints=http://127.0.0.1:2379 get /wklive/admin-api/config --print-value-only 2>/dev/null || true)
common_config=$(compose exec -T etcd etcdctl --endpoints=http://127.0.0.1:2379 get /wklive/common/config --print-value-only 2>/dev/null || true)

if printf '%s\n' "$trade_config" | grep -q '^Mode: pre$'; then pass "Trade runs in pre-release mode"; else fail "Trade runs in pre-release mode"; fi
if printf '%s\n' "$option_config" | grep -q '^Mode: pre$'; then pass "Option runs in pre-release mode"; else fail "Option runs in pre-release mode"; fi
if printf '%s\n' "$trade_config" | grep -A1 '^AutomaticLiquidation:' | grep -q 'Enabled: false'; then pass "automatic liquidation remains disabled"; else fail "automatic liquidation remains disabled"; fi
if printf '%s\n' "$trade_config" | grep -A1 '^CrossMarginTrading:' | grep -q 'Enabled: false'; then pass "cross-margin trading remains disabled"; else fail "cross-margin trading remains disabled"; fi
if printf '%s\n' "$option_config" | grep -A7 '^ProductScope:' | grep -q 'SellerTradingEnabled: false'; then pass "Option seller trading remains disabled"; else fail "Option seller trading remains disabled"; fi
if printf '%s\n' "$admin_config" | grep -A3 '^RequestEncryption:' | grep -q 'Mode: REQUIRED'; then pass "Admin request encryption is required"; else fail "Admin request encryption is required"; fi
if printf '%s\n' "$common_config" | grep -q "DataSource: ${MYSQL_APP_USER}:" && ! printf '%s\n' "$common_config" | grep -q 'DataSource: root:'; then pass "runtime database identity is non-root"; else fail "runtime database identity is non-root"; fi

if [ "$failures" -eq 0 ]; then
  printf '\nREADY: pre-production runtime passed all checks.\n'
  exit 0
fi
printf '\nNOT READY: %d check(s) failed.\n' "$failures" >&2
exit 1
