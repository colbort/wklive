#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
DEPLOY_ROOT="$REPO_ROOT/deploy"
# shellcheck disable=SC1091
. "$DEPLOY_ROOT/common/scripts/dev-compose.sh"
COMPOSE_FILE=${OPTION_BACKSTOP_RPC_COMPOSE_FILE:-}
DATABASE=${OPTION_BACKSTOP_RPC_DATABASE:-wklive_option_backstop_rpc_acceptance}
MYSQL_PASSWORD=${OPTION_BACKSTOP_RPC_MYSQL_PASSWORD:-123456}
ASSET_PORT=${OPTION_BACKSTOP_RPC_ASSET_PORT:-18085}
KEEP=${OPTION_BACKSTOP_RPC_KEEP:-0}

case "$DATABASE" in
  wklive_option_backstop_rpc_acceptance|wklive_option_backstop_rpc_acceptance_[A-Za-z0-9_]*) ;;
  *)
    echo "refusing unsafe acceptance database name: $DATABASE" >&2
    exit 1
    ;;
esac

compose_ps() {
  if [ -n "$COMPOSE_FILE" ]; then docker compose -f "$COMPOSE_FILE" ps -q "$1"; else dev_compose ps -q "$1"; fi
}
MYSQL_CONTAINER=${OPTION_BACKSTOP_RPC_MYSQL_CONTAINER:-$(compose_ps mysql)}
ETCD_CONTAINER=${OPTION_BACKSTOP_RPC_ETCD_CONTAINER:-$(compose_ps etcd)}
if [ -z "$MYSQL_CONTAINER" ] || [ -z "$ETCD_CONTAINER" ]; then
  echo "mysql and etcd from the development Compose environment must already be running" >&2
  exit 1
fi
if nc -z 127.0.0.1 "$ASSET_PORT" >/dev/null 2>&1; then
  echo "Asset acceptance port is already in use: $ASSET_PORT" >&2
  exit 1
fi

WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/wklive-option-backstop-rpc.XXXXXX")
REDIS_CONTAINER="wklive-option-backstop-rpc-redis-$$"
COMMON_KEY="/wklive/acceptance/option-backstop-rpc-$$/common"
ASSET_KEY="/wklive/acceptance/option-backstop-rpc-$$/asset"
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
    "${TMPDIR:-/tmp}"/wklive-option-backstop-rpc.*) rm -rf -- "$WORK_DIR" ;;
  esac
}
trap cleanup EXIT INT TERM

echo "creating isolated database $DATABASE"
mysql_cli -e "DROP DATABASE IF EXISTS \`$DATABASE\`; CREATE DATABASE \`$DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/asset/asset.sql"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/option/option.sql"
docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < "$REPO_ROOT/services/option/schema/90_constraints.sql"
for pass in 1 2; do
  docker exec -i "$MYSQL_CONTAINER" mysql -uroot "-p$MYSQL_PASSWORD" "$DATABASE" < \
    "$REPO_ROOT/services/asset/migrations/20260802_asset_backstop_policy_limits.sql"
  echo "migration_pass=$pass"
done

echo "starting isolated Redis"
docker run -d --rm --name "$REDIS_CONTAINER" --label wklive.acceptance=option-backstop-rpc \
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
  Key: option-backstop-rpc
  Host: $REDIS_ADDR
  Type: node
EOF
cat > "$WORK_DIR/asset.yaml" <<EOF
Name: asset.option.backstop.acceptance.rpc
ListenOn: 127.0.0.1:$ASSET_PORT
Mode: test
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: asset.option.backstop.acceptance.rpc
EOF
docker exec -i "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 put "$COMMON_KEY" < "$WORK_DIR/common.yaml" >/dev/null
docker exec -i "$ETCD_CONTAINER" etcdctl --endpoints=http://127.0.0.1:2379 put "$ASSET_KEY" < "$WORK_DIR/asset.yaml" >/dev/null

echo "building and starting repository Asset RPC"
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

echo "running platform-backstop policy and atomic-limit RPC acceptance"
(
  cd "$REPO_ROOT/services/asset"
  ASSET_BACKSTOP_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  ASSET_BACKSTOP_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/asset -run '^TestPlatformBackstopRPCLimitsMySQL$' \
      -count=1 -timeout=3m -v
)

echo "running Option liquidation deficit to governed Asset backstop acceptance"
(
  cd "$REPO_ROOT/services/option"
  OPTION_P0_ASSET_E2E_DSN="root:$MYSQL_PASSWORD@tcp(127.0.0.1:3306)/$DATABASE?charset=utf8mb4&parseTime=true&loc=Local" \
  OPTION_P0_ASSET_E2E_RPC_ADDR="127.0.0.1:$ASSET_PORT" \
  OPTION_P0_ASSET_E2E_REDIS_ADDR="$REDIS_ADDR" \
  GOCACHE="$WORK_DIR/go-build-cache" \
    go test ./internal/logic/task \
      -run '^TestP0AssetRPCEndToEnd/liquidation_deficit_backstop_failure_recovery$' \
      -count=1 -timeout=5m -v
)

mysql_cli -N "$DATABASE" -e "
SELECT CONCAT('policies=',COUNT(*),' approved=',COALESCE(SUM(status=2),0),
 ' rejected=',COALESCE(SUM(status=3),0)) FROM t_asset_backstop_policy;
SELECT CONCAT('covers=',COUNT(*),' amount=',COALESCE(SUM(covered_amount),0),
 ' policies_used=',COUNT(DISTINCT policy_id)) FROM t_asset_backstop_cover;
SELECT CONCAT('daily_rows=',COUNT(*),' used=',COALESCE(SUM(covered_amount),0))
 FROM t_asset_backstop_usage_daily;
SELECT CONCAT('flows=',COUNT(*),' amount=',COALESCE(SUM(amount),0))
 FROM t_asset_platform_flow
 WHERE account_type='OPTION_BACKSTOP' AND scene_type='platform_backstop_cover';
"

FAILED=0
echo "platform_backstop_rpc_acceptance=PASS database_removed=$([ "$KEEP" = "1" ] && echo no || echo yes)"
