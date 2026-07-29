#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
BUILD_SERVICES="
db-init
config-seed
itick-rpc
system-rpc
user-rpc
asset-rpc
chat-rpc
option-rpc
payment-rpc
trade-rpc
staking-rpc
admin-api
app-api
chat-admin-api
liquidity-rpc
chat-api
payment-api
liquidity-admin-api
"

build_images() {
  for build_service in "$@"; do
    docker compose -f "$COMPOSE_FILE" build "$build_service"
  done
}

usage() {
  echo "usage: $0 {up|start|down|restart|build|logs|ps|config|seed|data|merge-data|database|db-upgrade|compose-config|db-init|kafka-init|contract-readiness} [service ...]"
}

command="${1:-}"
if [ -z "$command" ]; then
  usage
  exit 1
fi
shift

case "$command" in
  up)
    if [ "$#" -eq 0 ]; then
      # Compose/Bake builds independent services concurrently. Building one
      # service at a time keeps first deployment reliable on smaller hosts.
      # shellcheck disable=SC2086
      build_images $BUILD_SERVICES
    else
      build_images "$@"
    fi
    docker compose -f "$COMPOSE_FILE" up -d "$@"
    ;;
  start)
    docker compose -f "$COMPOSE_FILE" up -d "$@"
    ;;
  down)
    docker compose -f "$COMPOSE_FILE" down "$@"
    ;;
  restart)
    docker compose -f "$COMPOSE_FILE" restart "$@"
    ;;
  build)
    if [ "$#" -eq 0 ]; then
      # shellcheck disable=SC2086
      build_images $BUILD_SERVICES
    else
      build_images "$@"
    fi
    ;;
  logs)
    docker compose -f "$COMPOSE_FILE" logs -f --tail=200 "$@"
    ;;
  ps)
    docker compose -f "$COMPOSE_FILE" ps "$@"
    ;;
  compose-config)
    docker compose -f "$COMPOSE_FILE" config
    ;;
  config|seed)
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps config-seed
    ;;
  data|merge-data)
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps \
      -e DB_INIT_MODE=data \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  database|db-upgrade)
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps \
      -e DB_INIT_MODE=full \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  db-init)
    docker compose -f "$COMPOSE_FILE" run --build --rm db-init
    ;;
  kafka-init)
    docker compose -f "$COMPOSE_FILE" run --build --rm kafka-init
    ;;
  contract-readiness)
    "$SCRIPT_DIR/contract-readiness.sh" "$@"
    ;;
  *)
    usage
    exit 1
    ;;
esac
