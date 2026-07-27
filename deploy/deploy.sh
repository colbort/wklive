#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

usage() {
  echo "usage: $0 {up|down|restart|build|logs|ps|config|seed|compose-config|db-init|kafka-init} [service ...]"
}

command="${1:-}"
if [ -z "$command" ]; then
  usage
  exit 1
fi
shift

case "$command" in
  up)
    docker compose -f "$COMPOSE_FILE" up -d --build "$@"
    ;;
  down)
    docker compose -f "$COMPOSE_FILE" down "$@"
    ;;
  restart)
    docker compose -f "$COMPOSE_FILE" restart "$@"
    ;;
  build)
    docker compose -f "$COMPOSE_FILE" build "$@"
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
    docker compose -f "$COMPOSE_FILE" run --rm --no-deps config-seed
    ;;
  db-init)
    docker compose -f "$COMPOSE_FILE" run --rm db-init
    ;;
  kafka-init)
    docker compose -f "$COMPOSE_FILE" run --rm kafka-init
    ;;
  *)
    usage
    exit 1
    ;;
esac
