#!/bin/sh

if [ -z "${DEPLOY_ROOT:-}" ]; then
  helper_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
  DEPLOY_ROOT=$(CDPATH= cd -- "$helper_dir/../.." && pwd)
fi

DEV_BASE_COMPOSE="$DEPLOY_ROOT/common/compose.base.yml"
DEV_ENV_COMPOSE="$DEPLOY_ROOT/environments/dev/compose.yml"
DEV_ENV_FILE="${DEV_ENV_FILE:-$DEPLOY_ROOT/environments/dev/.env}"
if [ ! -r "$DEV_ENV_FILE" ] && [ -r "$DEPLOY_ROOT/.env" ]; then
  DEV_ENV_FILE="$DEPLOY_ROOT/.env"
fi

dev_compose() {
  docker compose \
    --project-directory "$DEPLOY_ROOT" \
    --env-file "$DEV_ENV_FILE" \
    -f "$DEV_BASE_COMPOSE" \
    -f "$DEV_ENV_COMPOSE" \
    "$@"
}
