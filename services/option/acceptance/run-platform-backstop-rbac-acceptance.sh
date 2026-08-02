#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
COMPOSE_FILE=${OPTION_BACKSTOP_RBAC_COMPOSE_FILE:-$REPO_ROOT/deploy/docker-compose.yml}
DATABASE=${OPTION_BACKSTOP_RBAC_DATABASE:-wklive_option_backstop_rbac_acceptance}
MYSQL_PASSWORD=${OPTION_BACKSTOP_RBAC_MYSQL_PASSWORD:-123456}

case "$DATABASE" in
  wklive_option_backstop_rbac_acceptance|wklive_option_backstop_rbac_acceptance_[A-Za-z0-9_]*) ;;
  *)
    echo "refusing unsafe acceptance database name: $DATABASE" >&2
    exit 1
    ;;
esac

MYSQL_CONTAINER=${OPTION_BACKSTOP_RBAC_MYSQL_CONTAINER:-$(docker compose -f "$COMPOSE_FILE" ps -q mysql)}
if [ -z "$MYSQL_CONTAINER" ]; then
  echo "mysql from deploy/docker-compose.yml must already be running" >&2
  exit 1
fi
mysql_cli() {
  docker exec "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$@"
}
cleanup() {
  mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`;" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

echo "creating isolated RBAC database $DATABASE"
mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`; CREATE DATABASE \`$DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/system/system.sql"
mysql_cli "$DATABASE" -e "
INSERT INTO sys_menu
  (id,parent_id,app_scope,name,menu_type,method,path,perms,component,icon,sort)
VALUES
  (500,0,1,'资产管理',1,'','','','','',500),
  (710,0,1,'期权管理',1,'','','','','',710);"

for pass in 1 2; do
  for migration in \
    "$REPO_ROOT/services/system/migrations/20260731_zl_option_trading_calendar_permissions.sql" \
    "$REPO_ROOT/services/system/migrations/20260802_asset_platform_backstop_policy_permissions.sql" \
    "$REPO_ROOT/services/system/migrations/20260802_option_insurance_inventory_exit_permissions.sql"
  do
    docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$migration"
  done
  echo "rbac_migration_pass=$pass"
done

facts=$(mysql_cli -N "$DATABASE" -e "
SELECT CONCAT(
  (SELECT COUNT(*) FROM sys_menu WHERE id BETWEEN 733 AND 758), ':',
  (SELECT COUNT(DISTINCT id) FROM sys_menu WHERE id BETWEEN 733 AND 758), ':',
  (SELECT COUNT(DISTINCT perms) FROM sys_menu WHERE id BETWEEN 733 AND 758), ':',
  (SELECT COUNT(*) FROM sys_role WHERE code IN
    ('platform_backstop_policy_creator','platform_backstop_policy_reviewer')), ':',
  (SELECT COUNT(*) FROM sys_role_menu rm JOIN sys_role r ON r.id=rm.role_id
    JOIN sys_menu m ON m.id=rm.menu_id
    WHERE r.code='platform_backstop_policy_creator'
      AND m.perms='asset:platform-backstop-policy:create'), ':',
  (SELECT COUNT(*) FROM sys_role_menu rm JOIN sys_role r ON r.id=rm.role_id
    JOIN sys_menu m ON m.id=rm.menu_id
    WHERE r.code='platform_backstop_policy_creator'
      AND m.perms='asset:platform-backstop-policy:review'), ':',
  (SELECT COUNT(*) FROM sys_role_menu rm JOIN sys_role r ON r.id=rm.role_id
    JOIN sys_menu m ON m.id=rm.menu_id
    WHERE r.code='platform_backstop_policy_reviewer'
      AND m.perms='asset:platform-backstop-policy:review'), ':',
  (SELECT COUNT(*) FROM sys_role_menu rm JOIN sys_role r ON r.id=rm.role_id
    JOIN sys_menu m ON m.id=rm.menu_id
    WHERE r.code='platform_backstop_policy_reviewer'
      AND m.perms='asset:platform-backstop-policy:create')
);")
if [ "$facts" != "14:14:14:2:1:0:1:0" ]; then
  echo "unexpected RBAC facts menus:ids:perms:roles:creator_create:creator_review:reviewer_review:reviewer_create=$facts" >&2
  exit 1
fi
echo "rbac_facts=$facts"
echo "platform_backstop_rbac_acceptance=PASS"
