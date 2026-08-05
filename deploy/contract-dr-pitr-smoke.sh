#!/bin/sh
set -eu

DEPLOY_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
exec "$DEPLOY_ROOT/operations/dr/pitr-smoke.sh" "$@"
