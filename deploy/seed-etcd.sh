#!/bin/sh
set -eu

ETCD_ENDPOINT="${ETCD_ENDPOINT:-http://etcd:2379}"
WORKSPACE="${WORKSPACE:-/workspace}"
COMMON_CONFIG="${COMMON_CONFIG:-/deploy-config/common.yaml}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-123456}"
MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-openIM123}"
JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-change-this-secret-before-production}"

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

wait_for_etcd() {
  attempts=0
  until ETCDCTL_API=3 etcdctl --endpoints="$ETCD_ENDPOINT" endpoint health >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ "$attempts" -ge 60 ]; then
      echo "etcd did not become healthy" >&2
      exit 1
    fi
    sleep 1
  done
}

render_config() {
  sed \
    -e 's#127\.0\.0\.1:2379#etcd:2379#g' \
    -e 's#127\.0\.0\.1:8088#chat-rpc:8088#g' \
    -e 's#127\.0\.0\.1:27017#mongo:27017#g' \
    -e 's#127\.0\.0\.1:11300#beanstalk-primary:11300#g' \
    -e 's#127\.0\.0\.1:11301#beanstalk-secondary:11300#g' \
    -e "s#123456#${MYSQL_ROOT_PASSWORD}#g" \
    -e "s#openIM123#${MONGO_ROOT_PASSWORD}#g" \
    -e "s#change-this-secret-before-production#${JWT_ACCESS_SECRET}#g" \
    "$1"
}

put_file() {
  key="$1"
  file="$2"
  echo "seeding $key from $file"
  render_config "$file" | ETCDCTL_API=3 etcdctl --endpoints="$ETCD_ENDPOINT" put "$key"
}

validate_secret MYSQL_ROOT_PASSWORD "$MYSQL_ROOT_PASSWORD"
validate_secret MONGO_ROOT_PASSWORD "$MONGO_ROOT_PASSWORD"
validate_secret JWT_ACCESS_SECRET "$JWT_ACCESS_SECRET"
wait_for_etcd

put_file /wklive/common/config "$COMMON_CONFIG"
put_file /wklive/admin-api/config "$WORKSPACE/admin-api/etc/admin.yaml"
put_file /wklive/app-api/config "$WORKSPACE/app-api/etc/app.yaml"
put_file /wklive/chat-admin-api/config "$WORKSPACE/chat-admin-api/etc/chatadmin.yaml"
put_file /wklive/chat-api/config "$WORKSPACE/chat-api/etc/chat.yaml"
put_file /wklive/liquidity-admin-api/config "$WORKSPACE/liquidity-admin-api/etc/liquidityadmin.yaml"
put_file /wklive/payment-api/config "$WORKSPACE/payment-api/etc/payment-api.yaml"
put_file /wklive/asset-rpc/config "$WORKSPACE/services/asset/etc/asset.yaml"
put_file /wklive/chat-rpc/config "$WORKSPACE/services/chat/etc/chat.yaml"
put_file /wklive/itick-rpc/config "$WORKSPACE/services/itick/etc/itick.yaml"
put_file /wklive/liquidity-rpc/config "$WORKSPACE/services/liquidity/etc/liquidity.yaml"
put_file /wklive/option-rpc/config "$WORKSPACE/services/option/etc/option.yaml"
put_file /wklive/payment-rpc/config "$WORKSPACE/services/payment/etc/payment.yaml"
put_file /wklive/staking-rpc/config "$WORKSPACE/services/staking/etc/staking.yaml"
put_file /wklive/system-rpc/config "$WORKSPACE/services/system/etc/system.yaml"
put_file /wklive/trade-rpc/config "$WORKSPACE/services/trade/etc/trade.yaml"
put_file /wklive/user-rpc/config "$WORKSPACE/services/user/etc/user.yaml"

echo "all etcd configurations were seeded"
