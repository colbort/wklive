#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
REPO_ROOT=$(CDPATH= cd -- "$DEPLOY_ROOT/.." && pwd)
BASE_COMPOSE="$DEPLOY_ROOT/common/compose.base.yml"
DEV_COMPOSE="$SCRIPT_DIR/compose.yml"
ENV_FILE="${DEV_ENV_FILE:-$SCRIPT_DIR/.env}"
if [ ! -r "$ENV_FILE" ] && [ -r "$DEPLOY_ROOT/.env" ]; then
  ENV_FILE="$DEPLOY_ROOT/.env"
fi
if [ ! -r "$ENV_FILE" ]; then
  echo "development environment file is missing: $ENV_FILE" >&2
  echo "copy $SCRIPT_DIR/.env.example to $SCRIPT_DIR/.env" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

compose() {
  docker compose \
    --project-directory "$DEPLOY_ROOT" \
    --env-file "$ENV_FILE" \
    -f "$BASE_COMPOSE" \
    -f "$DEV_COMPOSE" \
    "$@"
}
BUILD_SERVICES="
beanstalk-primary
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
admin-ui
app-web
app-mobile
chat-admin-ui
chat-ui
liquidity-admin-ui
"

docker_free_kb() {
  running_container=$(
    compose ps -q 2>/dev/null |
      sed -n '1p'
  )
  if [ -z "$running_container" ]; then
    return 1
  fi
  docker exec "$running_container" df -Pk / 2>/dev/null |
    awk 'NR == 2 {print $4; exit}'
}

check_docker_disk() {
  minimum_free_kb="${1:-${DOCKER_MIN_FREE_KB:-4194304}}"
  minimum_name="${2:-DOCKER_MIN_FREE_KB}"
  case "$minimum_free_kb" in
    ''|*[!0-9]*|0)
      echo "invalid ${minimum_name}: expected a positive integer" >&2
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
    --filter "label=com.docker.compose.project=${DEV_PROJECT_NAME:-wklive}" >/dev/null ||
    echo "warning: unable to prune dangling wklive images" >&2
}

start_services() {
  compose_status=0
  compose up -d "$@" || compose_status=$?
  prune_project_dangling_images
  return "$compose_status"
}

build_images() {
  for build_service in "$@"; do
    check_docker_disk
    compose build "$build_service"
  done
}

usage() {
  echo "usage: $0 {up|start|down|restart|build|logs|ps|disk-check|config|seed|data|merge-data|database|db-upgrade|compose-config|db-init|kafka-init|beanstalk-readiness|beanstalk-restart-smoke|beanstalk-resilience-smoke|contract-readiness|option-readiness|contract-delivery-preflight|contract-dr-pitr-smoke|contract-dr-storage-check|contract-dr-storage-init|contract-dr-backup-smoke|contract-dr-backup-local-verify|contract-dr-backup-local-restore-verify|contract-dr-backup-local-pitr-restore-verify|contract-dr-backup} [service ...]"
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
    compose down "$@"
    ;;
  restart)
    compose restart "$@"
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
    compose logs -f --tail=200 "$@"
    ;;
  ps)
    compose ps "$@"
    ;;
  disk-check)
    check_docker_disk
    ;;
  compose-config)
    compose config
    ;;
  config|seed)
    check_docker_disk
    compose run --build --rm --no-deps config-seed
    ;;
  data|merge-data)
    check_docker_disk
    compose run --build --rm --no-deps \
      -e DB_INIT_MODE=data \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  database|db-upgrade)
    check_docker_disk
    compose run --build --rm --no-deps \
      -e DB_INIT_MODE=full \
      -e DB_INIT_TARGET=external \
      db-init
    ;;
  db-init)
    check_docker_disk
    compose run --build --rm db-init
    ;;
  kafka-init)
    compose run --build --rm kafka-init
    ;;
  beanstalk-readiness)
    "$DEPLOY_ROOT/operations/beanstalk/readiness.sh" "$@"
    ;;
  beanstalk-restart-smoke)
    "$DEPLOY_ROOT/operations/beanstalk/restart-smoke.sh" "$@"
    ;;
  beanstalk-resilience-smoke)
    "$DEPLOY_ROOT/operations/beanstalk/resilience-smoke.sh" "$@"
    ;;
  contract-readiness)
    check_docker_disk "${DOCKER_READINESS_MIN_FREE_KB:-2097152}" "DOCKER_READINESS_MIN_FREE_KB"
    "$DEPLOY_ROOT/operations/contract/readiness.sh" "$@"
    ;;
  option-readiness)
    "$REPO_ROOT/services/option/monitoring/option-production-readiness.sh" "$@"
    ;;
  contract-delivery-preflight)
    check_docker_disk "${DOCKER_READINESS_MIN_FREE_KB:-2097152}" "DOCKER_READINESS_MIN_FREE_KB"
    "$DEPLOY_ROOT/operations/contract/delivery-preflight.sh" "$@"
    ;;
  contract-dr-pitr-smoke)
    "$DEPLOY_ROOT/operations/dr/pitr-smoke.sh" "$@"
    ;;
  contract-dr-storage-check|contract-dr-storage-init)
    dr_env_file="${DR_BACKUP_ENV_FILE:-$ENV_FILE}"
    if [ ! -f "$dr_env_file" ]; then
      echo "DR backup environment file does not exist: $dr_env_file" >&2
      exit 1
    fi
    set -a
    # shellcheck disable=SC1090
    . "$dr_env_file"
    set +a
    dr_bucket="${DR_BACKUP_BUCKET_NAME:-}"
    if [ -z "$dr_bucket" ]; then
      echo "DR_BACKUP_BUCKET_NAME is required" >&2
      exit 1
    fi
    if [ "$command" = "contract-dr-storage-init" ]; then
      check_docker_disk
      build_images db-init
    else
      check_docker_disk "${DOCKER_READINESS_MIN_FREE_KB:-2097152}" \
        "DOCKER_READINESS_MIN_FREE_KB"
    fi
    if [ "$command" = "contract-dr-storage-init" ]; then
      compose run --rm --no-deps \
        --entrypoint /usr/local/bin/object-storage \
        -e OBJECT_STORAGE_BUCKET="$dr_bucket" \
        -e OBJECT_STORAGE_ALLOW_MUTATION=true \
        db-init ensure-private-versioned
    fi
    compose run --rm --no-deps \
      --entrypoint /usr/local/bin/object-storage \
      -e OBJECT_STORAGE_BUCKET="$dr_bucket" \
      -e OBJECT_STORAGE_REQUIRE_PRIVATE_VERSIONED=true \
      db-init inspect
    ;;
  contract-dr-backup-smoke)
    "$DEPLOY_ROOT/operations/dr/encrypted-backup.sh" smoke "$@"
    ;;
  contract-dr-backup-local-verify)
    "$DEPLOY_ROOT/operations/dr/encrypted-backup.sh" local-verify "$@"
    ;;
  contract-dr-backup-local-restore-verify)
    "$DEPLOY_ROOT/operations/dr/encrypted-backup.sh" local-restore-verify "$@"
    ;;
  contract-dr-backup-local-pitr-restore-verify)
    "$DEPLOY_ROOT/operations/dr/encrypted-backup.sh" local-pitr-restore-verify "$@"
    ;;
  contract-dr-backup)
    "$DEPLOY_ROOT/operations/dr/encrypted-backup.sh" production "$@"
    ;;
  *)
    usage
    exit 1
    ;;
esac
