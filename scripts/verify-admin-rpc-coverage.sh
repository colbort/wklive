#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "$0")/.." && pwd)"
services=(asset chat liquidity market option payment staking system trade user)
total_rpc=0
total_server=0
total_admin_api=0
failed=0

printf '%-12s %8s %10s %12s\n' SERVICE RPC SERVER ADMIN_API

for service in "${services[@]}"; do
  proto_file="$root_dir/proto/$service/$service.proto"
  server_dir="$root_dir/services/$service/internal/server/admin"
  case "$service" in
    chat)
      admin_dir="$root_dir/chat-admin-api"
      client_pattern='ChatAdminCli'
      ;;
    liquidity)
      admin_dir="$root_dir/liquidity-admin-api"
      client_pattern='LiquidityCli'
      ;;
    *)
      admin_dir="$root_dir/admin-api"
      client_pattern="${service^}Cli"
      ;;
  esac

  mapfile -t methods < <(
    sed -n '/service Admin[[:space:]]*{/,/^[[:space:]]*}/s/^[[:space:]]*rpc[[:space:]]\([A-Za-z0-9_]*\).*/\1/p' "$proto_file"
  )
  rpc_count=${#methods[@]}
  server_count=0
  api_count=0
  missing_server=()
  missing_api=()

  for method in "${methods[@]}"; do
    if rg -q "func[[:space:]]*\\([^)]*\\*AdminServer\\)[[:space:]]+$method\\(" "$server_dir"; then
      ((server_count += 1))
    else
      missing_server+=("$method")
    fi

    api_pattern="${client_pattern}\\.${method}([^A-Za-z0-9_]|$)"
    if [[ "$service" == "system" ]]; then
	  api_pattern="(SystemCli|client)\\.${method}([^A-Za-z0-9_]|$)"
    fi
    if rg -q "$api_pattern" "$admin_dir/internal" -g '*.go'; then
      ((api_count += 1))
    else
      missing_api+=("$method")
    fi
  done

  printf '%-12s %8d %10d %12d\n' "$service" "$rpc_count" "$server_count" "$api_count"
  if ((${#missing_server[@]})); then
    printf '  missing server: %s\n' "${missing_server[*]}"
    failed=1
  fi
  if ((${#missing_api[@]})); then
    printf '  missing admin API: %s\n' "${missing_api[*]}"
    failed=1
  fi
  total_rpc=$((total_rpc + rpc_count))
  total_server=$((total_server + server_count))
  total_admin_api=$((total_admin_api + api_count))
done

printf '%-12s %8d %10d %12d\n' TOTAL "$total_rpc" "$total_server" "$total_admin_api"

if rg -n 'todo: add your logic' \
  "$root_dir/admin-api/internal/logic" \
  "$root_dir/chat-admin-api/internal/logic" \
  "$root_dir/liquidity-admin-api/internal/logic"; then
  echo 'error: generated logic scaffolds still contain empty implementations' >&2
  failed=1
fi

if ((failed)); then
  exit 1
fi

echo 'Admin RPC coverage verification passed.'
