#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
BUILD_SERVICES="
db-init
config-seed
market-rpc
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

docker_free_kb() {
  running_container=$(
    docker compose -f "$COMPOSE_FILE" ps -q 2>/dev/null |
      sed -n '1p'
  )
  if [ -z "$running_container" ]; then
    return 1
  fi
  docker exec "$running_container" df -Pk / 2>/dev/null |
    awk 'NR == 2 {print $4; exit}'
}

check_docker_disk() {
  minimum_free_kb="${DOCKER_MIN_FREE_KB:-4194304}"
  case "$minimum_free_kb" in
    ''|*[!0-9]*|0)
      echo "invalid DOCKER_MIN_FREE_KB: expected a positive integer" >&2
      return 1
      ;;
  esac

  available_free_kb=$(docker_free_kb || true)
  case "$available_free_kb" in
    ''|*[!0-9]*)
      echo "Docker disk preflight skipped: no running Compose container is available"
      return 0
      ;;
  esac

  echo "Docker disk preflight: available=${available_free_kb}KiB required=${minimum_free_kb}KiB"
  if [ "$available_free_kb" -lt "$minimum_free_kb" ]; then
    echo "Docker disk preflight failed: free space is below the protected reserve" >&2
    echo "Run 'docker image prune' after reviewing unused images; do not delete database volumes." >&2
    return 1
  fi
}

prune_project_dangling_images() {
  docker image prune --force \
    --filter "label=com.docker.compose.project=wklive" >/dev/null ||
    echo "warning: unable to prune dangling wklive images" >&2
}

start_services() {
  compose_status=0
  docker compose -f "$COMPOSE_FILE" up -d "$@" || compose_status=$?
  prune_project_dangling_images
  return "$compose_status"
}

build_images() {
  for build_service in "$@"; do
    check_docker_disk
    docker compose -f "$COMPOSE_FILE" build "$build_service"
  done
}

usage() {
  echo "usage: $0 {up|start|down|restart|build|logs|ps|disk-check|config|seed|data|merge-data|database|db-upgrade|compose-config|db-init|kafka-init|contract-readiness} [service ...]"
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
    start_services "$@"
    ;;
  start)
    start_services "$@"
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
  disk-check)
    check_docker_disk
    ;;
  compose-config)
    docker compose -f "$COMPOSE_FILE" config
    ;;
  config|seed)
    check_docker_disk
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps config-seed
    ;;
  data|merge-data)
    check_docker_disk
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps \
      -e DB_INIT_MODE=data \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  database|db-upgrade)
    check_docker_disk
    docker compose -f "$COMPOSE_FILE" run --build --rm --no-deps \
      -e DB_INIT_MODE=full \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  db-init)
    check_docker_disk
    docker compose -f "$COMPOSE_FILE" run --build --rm db-init
    ;;
  kafka-init)
    docker compose -f "$COMPOSE_FILE" run --build --rm kafka-init
    ;;
  contract-readiness)
    check_docker_disk
    "$SCRIPT_DIR/contract-readiness.sh" "$@"
    ;;
  *)
    usage
    exit 1
    ;;
esac
