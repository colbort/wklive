#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
MODE=${1:---scope-only}

case "$MODE" in
  --scope-only|--release-clean) ;;
  *)
    printf 'usage: %s [--scope-only|--release-clean]\n' "$0" >&2
    exit 2
    ;;
esac

if ! git -C "$REPO_ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'SCOPE_FAIL repository metadata is unavailable\n' >&2
  exit 1
fi

total=0
modified=0
untracked=0
failures=0
status_file=$(mktemp "${TMPDIR:-/tmp}/option-release-scope.XXXXXX") || exit 1
trap 'rm -f "$status_file"' EXIT HUP INT TERM
git -C "$REPO_ROOT" status --porcelain -uall >"$status_file" || exit 1

while IFS= read -r line; do
  [ -n "$line" ] || continue
  status=$(printf '%s\n' "$line" | cut -c1-2)
  path=$(printf '%s\n' "$line" | cut -c4-)
  total=$((total + 1))
  if [ "$status" = "??" ]; then
    untracked=$((untracked + 1))
  else
    modified=$((modified + 1))
  fi
  case "$status" in
    *D*|*U*|*R*)
      printf 'SCOPE_FAIL destructive/unmerged status=%s path=%s\n' "$status" "$path" >&2
      failures=$((failures + 1))
      continue
      ;;
  esac
  case "$path" in
    init.sql|proto/option/*|services/option/*|services/system/migrations/20260802_option_insurance_inventory_exit_permissions.sql|admin-api/api/option.api|admin-api/internal/handler/routes.go|admin-api/internal/types/types.go|admin-api/internal/handler/option/*|admin-api/internal/logic/option/*|admin-ui/src/api/option/*|admin-ui/src/i18n/locales/en-US.ts|admin-ui/src/i18n/locales/zh-CN.ts|admin-ui/src/services/index.ts|admin-ui/src/services/option/*|admin-ui/src/views/option/*)
      ;;
    *)
      printf 'SCOPE_FAIL out-of-scope path=%s status=%s\n' "$path" "$status" >&2
      failures=$((failures + 1))
      ;;
  esac
done <"$status_file"

if ! git -C "$REPO_ROOT" diff --check >/dev/null; then
  printf 'SCOPE_FAIL git diff --check failed\n' >&2
  failures=$((failures + 1))
fi

if [ "$MODE" = "--release-clean" ] && [ "$total" -ne 0 ]; then
  printf 'SCOPE_FAIL release worktree is dirty changed=%s\n' "$total" >&2
  failures=$((failures + 1))
fi

if [ "$failures" -ne 0 ]; then
  printf 'SCOPE_REJECTED changed=%s modified=%s untracked=%s failures=%s\n' \
    "$total" "$modified" "$untracked" "$failures" >&2
  exit 1
fi

printf 'SCOPE_OK changed=%s modified=%s untracked=%s mode=%s\n' \
  "$total" "$modified" "$untracked" "$MODE"
