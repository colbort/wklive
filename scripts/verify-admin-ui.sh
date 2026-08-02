#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "$0")/.." && pwd)"
node_major="$(node -p 'Number(process.versions.node.split(".")[0])')"

if ((node_major < 20)); then
  echo "error: admin UI verification requires Node.js 20 or newer (current: $(node --version))" >&2
  exit 1
fi

for app in admin-ui chat-admin-ui liquidity-admin-ui; do
  echo "==> building $app"
  (
    cd "$root_dir/$app"
	if node -e 'const p=require("./package.json"); process.exit(p.scripts?.test ? 0 : 1)'; then
	  npm test
	fi
    npm run build
  )
done

echo "All admin UI type checks and production builds passed."
