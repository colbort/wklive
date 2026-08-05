#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
# shellcheck disable=SC1091
. "$DEPLOY_ROOT/common/scripts/dev-compose.sh"
READINESS_FILE="${1:-$SCRIPT_DIR/production-readiness.env}"

if [ ! -f "$READINESS_FILE" ]; then
  echo "Production readiness file does not exist: $READINESS_FILE" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$READINESS_FILE"
set +a

case "${DELIVERY_ALGORITHM:-}" in
  WEIGHTED_MEAN)
    delivery_algorithm_number=1
    ;;
  MEDIAN)
    delivery_algorithm_number=2
    ;;
  *)
    echo "DELIVERY_ALGORITHM must be WEIGHTED_MEAN or MEDIAN" >&2
    exit 1
    ;;
esac

source_count=$(
  printf '%s\n' "${PRODUCTION_PRICE_SOURCE_IDS:-}" |
    awk -F, '{
      for (i = 1; i <= NF; i++) {
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", $i)
        if ($i != "") unique[$i] = 1
      }
    } END {
      for (source in unique) count++
      print count + 0
    }'
)

dev_compose run --build --rm --no-deps \
  --entrypoint /usr/local/bin/delivery-preflight \
  -e "DELIVERY_PREFLIGHT_TENANT_ID=${PRODUCTION_TENANT_ID:-}" \
  -e "DELIVERY_PREFLIGHT_SYMBOL=${PRODUCTION_DELIVERY_SYMBOL:-}" \
  -e "DELIVERY_PREFLIGHT_SETTLEMENT_ASSET=${PRODUCTION_SETTLEMENT_COIN:-}" \
  -e "DELIVERY_PREFLIGHT_CATEGORY_CODE=${PRODUCTION_CATEGORY_CODE:-}" \
  -e "DELIVERY_PREFLIGHT_MARKET=${PRODUCTION_MARKET:-}" \
  -e "DELIVERY_PREFLIGHT_PRICE_SYMBOL=${PRODUCTION_PRICE_SYMBOL:-}" \
  -e "DELIVERY_PREFLIGHT_FORMULA_VERSION=${DELIVERY_FORMULA_VERSION:-}" \
  -e "DELIVERY_PREFLIGHT_FORMULA_ALGORITHM=$delivery_algorithm_number" \
  -e "DELIVERY_PREFLIGHT_MAX_LOOKBACK_MS=${DELIVERY_LOCK_WINDOW_MS:-}" \
  -e "DELIVERY_PREFLIGHT_MAX_DEVIATION_BPS=${DELIVERY_MAX_DEVIATION_BPS:-}" \
  -e "DELIVERY_PREFLIGHT_MIN_INPUT_COUNT=$source_count" \
  -e "DELIVERY_PREFLIGHT_INTERVAL_MS=${PRICE_FORMULA_INTERVAL_MS:-}" \
  db-init
