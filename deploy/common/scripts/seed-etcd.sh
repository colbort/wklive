#!/bin/sh
set -eu

ETCD_ENDPOINT="${ETCD_ENDPOINT:-http://etcd:2379}"
WORKSPACE="${WORKSPACE:-/workspace}"
COMMON_CONFIG="${COMMON_CONFIG:-/deploy-config/common.yaml}"
CHAT_API_CONFIG="${CHAT_API_CONFIG:-$WORKSPACE/chat-api/etc/chat.yaml}"
CHAT_ADMIN_API_CONFIG="${CHAT_ADMIN_API_CONFIG:-$WORKSPACE/chat-admin-api/etc/chatadmin.yaml}"
DEV_BACKEND_SUBNET="${DEV_BACKEND_SUBNET:-172.20.0.0/16}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-123456}"
MYSQL_APP_USER="${MYSQL_APP_USER:-root}"
MYSQL_APP_PASSWORD="${MYSQL_APP_PASSWORD:-$MYSQL_ROOT_PASSWORD}"
MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-openIM123}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-change-this-secret-before-production}"
ITICK_TOKEN_FILE="${ITICK_TOKEN_FILE:-/run/secrets/itick_token}"
SERVICE_MODE="${SERVICE_MODE:-}"
ADMIN_ALLOWED_ORIGIN="${ADMIN_ALLOWED_ORIGIN:-}"
ADMIN_REQUEST_ENCRYPTION_MODE="${ADMIN_REQUEST_ENCRYPTION_MODE:-}"
ADMIN_RSA_PRIVATE_KEY_PATH="${ADMIN_RSA_PRIVATE_KEY_PATH:-}"
ADMIN_SESSION_WRAP_KEY="${ADMIN_SESSION_WRAP_KEY:-}"

validate_secret() {
  name="$1"
  value="$2"
  case "$value" in
    *[!A-Za-z0-9._-]*|'')
      echo "$name may only contain letters, numbers, dot, underscore and dash" >&2
      exit 1
      ;;
  esac
}

check_etcd() {
  echo "checking etcd at $ETCD_ENDPOINT"
  if ! ETCDCTL_DIAL_TIMEOUT=3s ETCDCTL_COMMAND_TIMEOUT=5s \
    etcdctl --endpoints="$ETCD_ENDPOINT" endpoint health >/dev/null 2>&1; then
    echo "etcd at $ETCD_ENDPOINT is unreachable or unhealthy; start it first or set ETCD_ENDPOINT in the selected environment .env" >&2
    exit 1
  fi
}

load_file_secret() {
  name="$1"
  file="$2"
  if [ ! -r "$file" ]; then
    echo "$name secret file is missing or unreadable: $file" >&2
    exit 1
  fi
  value=$(sed -n '1p' "$file")
  validate_secret "$name" "$value"
  printf '%s' "$value"
}

render_config() {
  sed \
    -e 's#127\.0\.0\.1:2379#etcd:2379#g' \
    -e 's#127\.0\.0\.1:8088#chat-rpc:8088#g' \
    -e 's#127\.0\.0\.1:27017#mongo:27017#g' \
    -e 's#127\.0\.0\.1:11300#beanstalk-primary:11300#g' \
    -e 's#127\.0\.0\.1:11301#beanstalk-secondary:11300#g' \
    -e "s#__DEV_BACKEND_SUBNET__#${DEV_BACKEND_SUBNET}#g" \
    -e "s#123456#${MYSQL_ROOT_PASSWORD}#g" \
    -e "s#__MYSQL_APP_USER__#${MYSQL_APP_USER}#g" \
    -e "s#__MYSQL_APP_PASSWORD__#${MYSQL_APP_PASSWORD}#g" \
    -e "s#openIM123#${MONGO_ROOT_PASSWORD}#g" \
    -e "s#__REDIS_PASSWORD__#${REDIS_PASSWORD}#g" \
    -e "s#change-this-secret-before-production#${JWT_ACCESS_SECRET}#g" \
    -e "s#__ITICK_TOKEN__#${ITICK_TOKEN}#g" \
    "$1" |
    if [ -n "$SERVICE_MODE" ]; then
      sed \
        -e "s#^Mode: dev\$#Mode: ${SERVICE_MODE}#" \
        -e "s#^Mode: pro\$#Mode: ${SERVICE_MODE}#"
    else
      cat
    fi
}

render_admin_config() {
  render_config "$1" |
    sed \
      -e "s#http://localhost:5173#${ADMIN_ALLOWED_ORIGIN:-http://localhost:5173}#g" \
      -e "s#http://127\.0\.0\.1:5173#${ADMIN_ALLOWED_ORIGIN:-http://127.0.0.1:5173}#g" \
      -e "s#http://localhost:8888#${ADMIN_ALLOWED_ORIGIN:-http://localhost:8888}#g" \
      -e "s#http://127\.0\.0\.1:8888#${ADMIN_ALLOWED_ORIGIN:-http://127.0.0.1:8888}#g" \
      -e "s#  Mode: DISABLED#  Mode: ${ADMIN_REQUEST_ENCRYPTION_MODE:-DISABLED}#" \
      -e "s#/Users/sky/local/go/src/wklive/secrets/admin-api-private.pem#${ADMIN_RSA_PRIVATE_KEY_PATH:-/Users/sky/local/go/src/wklive/secrets/admin-api-private.pem}#" \
      -e "s#0@29odn4s\*dP@tcMZ=U_sQHppG4Av114#${ADMIN_SESSION_WRAP_KEY:-0@29odn4s*dP@tcMZ=U_sQHppG4Av114}#"
}

put_file() {
  key="$1"
  file="$2"
  echo "seeding $key from $file"
  render_config "$file" | etcdctl --endpoints="$ETCD_ENDPOINT" put "$key"
}

put_admin_file() {
  key="$1"
  file="$2"
  echo "seeding $key from $file"
  render_admin_config "$file" | etcdctl --endpoints="$ETCD_ENDPOINT" put "$key"
}

validate_secret MYSQL_ROOT_PASSWORD "$MYSQL_ROOT_PASSWORD"
validate_secret MYSQL_APP_USER "$MYSQL_APP_USER"
validate_secret MYSQL_APP_PASSWORD "$MYSQL_APP_PASSWORD"
validate_secret MONGO_ROOT_PASSWORD "$MONGO_ROOT_PASSWORD"
if [ -n "$REDIS_PASSWORD" ]; then
  validate_secret REDIS_PASSWORD "$REDIS_PASSWORD"
fi
validate_secret JWT_ACCESS_SECRET "$JWT_ACCESS_SECRET"
if [ -n "$SERVICE_MODE" ]; then
  case "$SERVICE_MODE" in
    dev|test|rt|pre|pro) ;;
    *)
      echo "SERVICE_MODE must be one of dev, test, rt, pre or pro" >&2
      exit 1
      ;;
  esac
fi
if [ -n "$ADMIN_REQUEST_ENCRYPTION_MODE" ]; then
  case "$ADMIN_REQUEST_ENCRYPTION_MODE" in
    DISABLED|OPTIONAL|REQUIRED) ;;
    *)
      echo "ADMIN_REQUEST_ENCRYPTION_MODE must be DISABLED, OPTIONAL or REQUIRED" >&2
      exit 1
      ;;
  esac
fi
if [ "$ADMIN_REQUEST_ENCRYPTION_MODE" != "" ] && [ "$ADMIN_REQUEST_ENCRYPTION_MODE" != "DISABLED" ]; then
  if [ -z "$ADMIN_RSA_PRIVATE_KEY_PATH" ]; then
    echo "ADMIN_RSA_PRIVATE_KEY_PATH is required when request encryption is enabled" >&2
    exit 1
  fi
  if [ "${#ADMIN_SESSION_WRAP_KEY}" -ne 32 ]; then
    echo "ADMIN_SESSION_WRAP_KEY must be exactly 32 bytes" >&2
    exit 1
  fi
fi
ITICK_TOKEN=$(load_file_secret ITICK_TOKEN "$ITICK_TOKEN_FILE")
check_etcd

put_file /wklive/common/config "$COMMON_CONFIG"
put_admin_file /wklive/admin-api/config "$WORKSPACE/admin-api/etc/admin.yaml"
put_file /wklive/app-api/config "$WORKSPACE/app-api/etc/app.yaml"
put_file /wklive/chat-admin-api/config "$CHAT_ADMIN_API_CONFIG"
put_file /wklive/chat-api/config "$CHAT_API_CONFIG"
put_file /wklive/liquidity-admin-api/config "$WORKSPACE/liquidity-admin-api/etc/liquidityadmin.yaml"
put_file /wklive/payment-api/config "$WORKSPACE/payment-api/etc/payment-api.yaml"
put_file /wklive/asset-rpc/config "$WORKSPACE/services/asset/etc/asset.yaml"
put_file /wklive/chat-rpc/config "$WORKSPACE/services/chat/etc/chat.yaml"
put_file /wklive/market-rpc/config "$WORKSPACE/services/market/etc/market.yaml"
put_file /wklive/liquidity-rpc/config "$WORKSPACE/services/liquidity/etc/liquidity.yaml"
put_file /wklive/option-rpc/config "$WORKSPACE/services/option/etc/option.yaml"
put_file /wklive/payment-rpc/config "$WORKSPACE/services/payment/etc/payment.yaml"
put_file /wklive/staking-rpc/config "$WORKSPACE/services/staking/etc/staking.yaml"
put_file /wklive/system-rpc/config "$WORKSPACE/services/system/etc/system.yaml"
put_file /wklive/trade-rpc/config "$WORKSPACE/services/trade/etc/trade.yaml"
put_file /wklive/user-rpc/config "$WORKSPACE/services/user/etc/user.yaml"

echo "all etcd configurations were seeded"
