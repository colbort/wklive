#!/usr/bin/env bash

set -uo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

TARGET_DIRS=(
  "admin-api"
  "app-api"
  "payment-api"
  "chat-api"
  "chat-admin-api"
  "liquidity-admin-api"
  "services"
  "common"
  "proto"
)

TEMP_FILE="$(mktemp)"
trap 'rm -f "$TEMP_FILE"' EXIT

echo "工作区目录：$ROOT_DIR"
echo
echo "正在查找 go.mod..."
echo

for target in "${TARGET_DIRS[@]}"; do
  target_path="$ROOT_DIR/$target"

  if [[ ! -d "$target_path" ]]; then
    echo "跳过不存在的目录：$target"
    continue
  fi

  find "$target_path" \
    -type f \
    -name "go.mod" \
    -not -path "*/vendor/*" \
    -not -path "*/.git/*" \
    >> "$TEMP_FILE"
done

sort -u "$TEMP_FILE" -o "$TEMP_FILE"

if [[ ! -s "$TEMP_FILE" ]]; then
  echo "没有找到任何 go.mod"
  exit 1
fi

MODULE_COUNT="$(wc -l < "$TEMP_FILE" | tr -d ' ')"

echo "共找到 $MODULE_COUNT 个 Go Module："
echo

while IFS= read -r go_mod; do
  echo "  - ${go_mod#$ROOT_DIR/}"
done < "$TEMP_FILE"

echo
echo "开始升级依赖..."
echo

SUCCESS_COUNT=0
FAILED_COUNT=0
FAILED_MODULES=()

run_with_timeout() {
  if command -v gtimeout >/dev/null 2>&1; then
    gtimeout 10m "$@"
  elif command -v timeout >/dev/null 2>&1; then
    timeout 10m "$@"
  else
    "$@"
  fi
}

process_module() {
  local module_dir="$1"

  cd "$module_dir" || return 1

  echo "Go 版本：$(go version)"
  echo "GOPROXY：$(go env GOPROXY)"
  echo

  echo "[1/3] 升级当前模块实际使用的依赖..."
  run_with_timeout go get -u ./...
  local code=$?

  if [[ "$code" -ne 0 ]]; then
    return "$code"
  fi

  echo
  echo "[2/3] 整理 go.mod 和 go.sum..."
  run_with_timeout go mod tidy
  code=$?

  if [[ "$code" -ne 0 ]]; then
    return "$code"
  fi

  echo
  echo "[3/3] 编译检查..."
  run_with_timeout go test -run='^$' ./...
  code=$?

  if [[ "$code" -ne 0 ]]; then
    return "$code"
  fi

  return 0
}

while IFS= read -r go_mod; do
  module_dir="$(dirname "$go_mod")"
  module_name="${module_dir#$ROOT_DIR/}"

  echo "============================================================"
  echo "处理模块：$module_name"
  echo "============================================================"

  process_module "$module_dir"
  exit_code=$?

  if [[ "$exit_code" -eq 0 ]]; then
    echo
    echo "成功：$module_name"
    SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
  else
    echo

    if [[ "$exit_code" -eq 124 ]]; then
      echo "超时：$module_name"
    else
      echo "失败：$module_name，退出码：$exit_code"
    fi

    FAILED_COUNT=$((FAILED_COUNT + 1))
    FAILED_MODULES+=("$module_name")
  fi

  echo
done < "$TEMP_FILE"

echo "============================================================"
echo "依赖升级完成"
echo "============================================================"
echo "成功：$SUCCESS_COUNT"
echo "失败：$FAILED_COUNT"

if [[ "$FAILED_COUNT" -gt 0 ]]; then
  echo
  echo "失败或超时的模块："

  for module_name in "${FAILED_MODULES[@]}"; do
    echo "  - $module_name"
  done

  exit 1
fi

echo
echo "检查修改："
echo "  git status"
echo "  git diff -- '**/go.mod' '**/go.sum'"
