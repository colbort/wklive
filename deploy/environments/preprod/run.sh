#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
BASE_COMPOSE="$DEPLOY_ROOT/common/compose.base.yml"
PREPROD_COMPOSE="$SCRIPT_DIR/compose.yml"
ENV_FILE="${PREPROD_ENV_FILE:-$SCRIPT_DIR/.env}"
if [ ! -r "$ENV_FILE" ] && [ -r "$DEPLOY_ROOT/.env.preprod" ]; then
  ENV_FILE="$DEPLOY_ROOT/.env.preprod"
fi

if [ ! -r "$ENV_FILE" ]; then
  echo "pre-production environment file is missing or unreadable: $ENV_FILE" >&2
  echo "copy $SCRIPT_DIR/.env.example to $SCRIPT_DIR/.env and replace every placeholder" >&2
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

fail() {
  echo "pre-production validation failed: $1" >&2
  exit 1
}

require_value() {
  name="$1"
  value="$2"
  [ -n "$value" ] || fail "$name is required"
  case "$value" in
    replace-*|*replace-with*|*example.com*)
      fail "$name still contains an example or replacement value"
      ;;
  esac
}

require_safe_secret() {
  name="$1"
  value="$2"
  minimum="$3"
  require_value "$name" "$value"
  case "$value" in
    *[!A-Za-z0-9._-]*) fail "$name contains characters that are unsafe for config rendering" ;;
  esac
  length=$(LC_ALL=C printf '%s' "$value" | wc -c | tr -d ' ')
  [ "$length" -ge "$minimum" ] || fail "$name must contain at least $minimum bytes"
}

require_private_file() {
  name="$1"
  path="$2"
  require_value "$name" "$path"
  case "$path" in
    /*) ;;
    *) fail "$name must be an absolute path" ;;
  esac
  [ -r "$path" ] || fail "$name is missing or unreadable: $path"
  mode=$(stat -f '%Lp' "$path" 2>/dev/null || stat -c '%a' "$path" 2>/dev/null || true)
  case "$mode" in
    400|600) ;;
    *) fail "$name must use mode 400 or 600, found ${mode:-unknown}" ;;
  esac
}

validate_release_tag() {
  require_value RELEASE_TAG "${RELEASE_TAG:-}"
  case "$RELEASE_TAG" in
    latest|*[!A-Za-z0-9._-]*) fail "RELEASE_TAG must be an immutable tag and cannot be latest" ;;
  esac
}

validate_environment() {
  env_mode=$(stat -f '%Lp' "$ENV_FILE" 2>/dev/null || stat -c '%a' "$ENV_FILE" 2>/dev/null || true)
  case "$env_mode" in
    400|600) ;;
    *) fail "$ENV_FILE must use mode 400 or 600, found ${env_mode:-unknown}" ;;
  esac

  validate_release_tag
  require_safe_secret MYSQL_ROOT_PASSWORD "${MYSQL_ROOT_PASSWORD:-}" 20
  require_safe_secret MYSQL_APP_USER "${MYSQL_APP_USER:-}" 3
  require_safe_secret MYSQL_APP_PASSWORD "${MYSQL_APP_PASSWORD:-}" 20
  require_safe_secret MONGO_ROOT_PASSWORD "${MONGO_ROOT_PASSWORD:-}" 20
  require_safe_secret REDIS_PASSWORD "${REDIS_PASSWORD:-}" 20
  require_safe_secret JWT_ACCESS_SECRET "${JWT_ACCESS_SECRET:-}" 32
  require_safe_secret ADMIN_PASSWORD "${ADMIN_PASSWORD:-}" 20
  require_safe_secret LIQUIDITY_ADMIN_PASSWORD "${LIQUIDITY_ADMIN_PASSWORD:-}" 20

  case "${PREPROD_TENANT_ID:-}" in
    ''|*[!0-9]*|0) fail "PREPROD_TENANT_ID must be a positive integer" ;;
  esac
  require_safe_secret PREPROD_TENANT_CODE "${PREPROD_TENANT_CODE:-}" 1
  require_value PREPROD_TENANT_NAME "${PREPROD_TENANT_NAME:-}"

  [ "$MYSQL_ROOT_PASSWORD" != "$MYSQL_APP_PASSWORD" ] || fail "MySQL root and application passwords must differ"
  [ "$ADMIN_PASSWORD" != "$LIQUIDITY_ADMIN_PASSWORD" ] || fail "the two administrator passwords must differ"
  [ "$REDIS_PASSWORD" != "$MYSQL_APP_PASSWORD" ] || fail "Redis and MySQL application passwords must differ"

  case "${ADMIN_REQUEST_ENCRYPTION_MODE:-REQUIRED}" in
    REQUIRED) ;;
    *) fail "ADMIN_REQUEST_ENCRYPTION_MODE must be REQUIRED in pre-production" ;;
  esac
  require_safe_secret ADMIN_SESSION_WRAP_KEY "${ADMIN_SESSION_WRAP_KEY:-}" 32
  wrap_length=$(LC_ALL=C printf '%s' "$ADMIN_SESSION_WRAP_KEY" | wc -c | tr -d ' ')
  [ "$wrap_length" -eq 32 ] || fail "ADMIN_SESSION_WRAP_KEY must be exactly 32 bytes"

  require_value ADMIN_ALLOWED_ORIGIN "${ADMIN_ALLOWED_ORIGIN:-}"
  case "$ADMIN_ALLOWED_ORIGIN" in
    *[\&\#[:space:]]*) fail "ADMIN_ALLOWED_ORIGIN contains characters that are unsafe for config rendering" ;;
  esac
  case "$ADMIN_ALLOWED_ORIGIN" in
    https://*) ;;
    http://127.0.0.1:*|http://localhost:*) ;;
    *) fail "ADMIN_ALLOWED_ORIGIN must use HTTPS unless it is a loopback URL" ;;
  esac

  require_private_file PREPROD_ADMIN_RSA_PRIVATE_KEY_FILE "${PREPROD_ADMIN_RSA_PRIVATE_KEY_FILE:-}"
  grep -Eq '^-----BEGIN (RSA )?PRIVATE KEY-----$' "$PREPROD_ADMIN_RSA_PRIVATE_KEY_FILE" ||
    fail "PREPROD_ADMIN_RSA_PRIVATE_KEY_FILE is not a PEM private key"
  require_private_file PREPROD_ITICK_TOKEN_FILE "${PREPROD_ITICK_TOKEN_FILE:-}"
  require_private_file PREPROD_OPERATORS_ENV_FILE "${PREPROD_OPERATORS_ENV_FILE:-}"

  case "${PREPROD_BIND_ADDRESS:-127.0.0.1}" in
    127.0.0.1|::1) ;;
    0.0.0.0)
      [ "${PREPROD_ALLOW_PUBLIC_BIND:-false}" = "true" ] ||
        fail "0.0.0.0 binding requires PREPROD_ALLOW_PUBLIC_BIND=true"
      ;;
    *) fail "PREPROD_BIND_ADDRESS must be loopback or explicitly approved 0.0.0.0" ;;
  esac

  case "${KAFKA_PARTITIONS:-3}" in
    ''|*[!0-9]*|0) fail "KAFKA_PARTITIONS must be a positive integer" ;;
  esac

  compose config -q
  echo "pre-production environment validation passed"
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
liquidity-rpc
admin-api
app-api
chat-admin-api
chat-api
payment-api
liquidity-admin-api
admin-ui
app-web
chat-admin-ui
liquidity-admin-ui
"

build_images() {
  for service_name in $BUILD_SERVICES; do
    compose build "$service_name"
  done
}

usage() {
  echo "usage: $0 {validate|config|build|pull|up|start|restart|stop|down|ps|logs|db-init|kafka-init|readiness|rollback} [service|release-tag ...]" >&2
}

command="${1:-}"
[ -n "$command" ] || {
  usage
  exit 1
}
shift

case "$command" in
  validate)
    validate_environment
    ;;
  config)
    validate_environment
    compose config
    ;;
  build)
    validate_environment
    if [ "$#" -gt 0 ]; then
      for service_name in "$@"; do
        compose build "$service_name"
      done
    else
      build_images
    fi
    ;;
  pull)
    validate_environment
    compose pull "$@"
    ;;
  up)
    validate_environment
    if [ "${PREPROD_BUILD_LOCAL:-true}" = "true" ]; then
      build_images
    else
      compose pull
    fi
    compose up -d --no-build --wait --wait-timeout "${PREPROD_START_TIMEOUT_SECONDS:-300}"
    "$SCRIPT_DIR/readiness.sh" "$ENV_FILE"
    ;;
  start)
    validate_environment
    compose up -d --no-build --wait --wait-timeout "${PREPROD_START_TIMEOUT_SECONDS:-300}" "$@"
    ;;
  restart)
    validate_environment
    compose restart "$@"
    ;;
  stop)
    compose stop "$@"
    ;;
  down)
    compose down "$@"
    ;;
  ps)
    compose ps "$@"
    ;;
  logs)
    compose logs -f --tail=200 "$@"
    ;;
  db-init)
    validate_environment
    compose run --rm db-init
    ;;
  kafka-init)
    validate_environment
    compose run --rm kafka-init
    ;;
  readiness)
    validate_environment
    "$SCRIPT_DIR/readiness.sh" "$ENV_FILE"
    ;;
  rollback)
    [ "$#" -eq 1 ] || fail "rollback requires exactly one previous immutable release tag"
    rollback_tag="$1"
    case "$rollback_tag" in
      latest|replace-*|*[!A-Za-z0-9._-]*) fail "invalid rollback release tag" ;;
    esac
    validate_environment
    RELEASE_TAG="$rollback_tag" compose pull
    RELEASE_TAG="$rollback_tag" compose up -d --no-build --wait --wait-timeout "${PREPROD_START_TIMEOUT_SECONDS:-300}"
    RELEASE_TAG="$rollback_tag" "$SCRIPT_DIR/readiness.sh" "$ENV_FILE"
    ;;
  *)
    usage
    exit 1
    ;;
esac
