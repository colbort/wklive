#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OPTION_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
VERIFIER="$SCRIPT_DIR/option-product-scope-verify.sh"
fixture_dir=$(mktemp -d "${TMPDIR:-/tmp}/option-product-scope.XXXXXX") || exit 1
trap 'rm -rf "$fixture_dir"' EXIT HUP INT TERM

write_valid_declaration() {
  target="$1"
  printf '%s\n' \
    'OPTION_SELLER_TRADING_ENABLED=false' \
    'OPTION_PORTFOLIO_MARGIN_ENABLED=false' \
    'OPTION_PHYSICAL_DELIVERY_ENABLED=false' \
    'OPTION_COMPLEX_ORDERS_ENABLED=false' \
    'OPTION_PUBLIC_MARKET_ENABLED=false' \
    'OPTION_MMP_ENABLED=false' \
    'OPTION_AMERICAN_EXERCISE_ENABLED=false' \
    'OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=false' >"$target"
}

expect_rejected() {
  name="$1"
  declaration="$2"
  runtime="$3"
  if "$VERIFIER" "$declaration" "$runtime" >/dev/null 2>&1; then
    printf 'FAIL  %s was accepted\n' "$name" >&2
    exit 1
  fi
  printf 'PASS  %s rejected\n' "$name"
}

valid_env="$fixture_dir/valid.env"
valid_yaml="$fixture_dir/valid.yaml"
write_valid_declaration "$valid_env"
cp "$OPTION_DIR/etc/option.yaml" "$valid_yaml"
"$VERIFIER" "$valid_env" "$valid_yaml" >/dev/null || exit 1
printf 'PASS  valid declaration and runtime config\n'

mismatch_yaml="$fixture_dir/mismatch.yaml"
sed 's/SellerTradingEnabled: false/SellerTradingEnabled: true/' "$valid_yaml" >"$mismatch_yaml"
expect_rejected "declaration/runtime mismatch" "$valid_env" "$mismatch_yaml"

duplicate_yaml="$fixture_dir/duplicate.yaml"
awk '{ print; if ($0 ~ /SellerTradingEnabled: false/) print $0 }' "$valid_yaml" >"$duplicate_yaml"
expect_rejected "duplicate runtime key" "$valid_env" "$duplicate_yaml"

missing_yaml="$fixture_dir/missing.yaml"
sed '/MMPEnabled: false/d' "$valid_yaml" >"$missing_yaml"
expect_rejected "missing runtime key" "$valid_env" "$missing_yaml"

duplicate_env="$fixture_dir/duplicate.env"
cp "$valid_env" "$duplicate_env"
printf '%s\n' 'OPTION_PUBLIC_MARKET_ENABLED=false' >>"$duplicate_env"
expect_rejected "duplicate release declaration" "$duplicate_env" "$valid_yaml"

greeks_env="$fixture_dir/greeks.env"
sed 's/OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=false/OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED=true/' \
  "$valid_env" >"$greeks_env"
expect_rejected "unsupported Greeks-dependent scope" "$greeks_env" "$valid_yaml"

for dependency_mapping in \
  'OPTION_PORTFOLIO_MARGIN_ENABLED:PortfolioMarginEnabled' \
  'OPTION_PHYSICAL_DELIVERY_ENABLED:PhysicalDeliveryEnabled' \
  'OPTION_COMPLEX_ORDERS_ENABLED:ComplexOrdersEnabled' \
  'OPTION_MMP_ENABLED:MMPEnabled' \
  'OPTION_AMERICAN_EXERCISE_ENABLED:AmericanExerciseEnabled'; do
  dependency_flag=${dependency_mapping%%:*}
  dependency_key=${dependency_mapping#*:}
  dependency_env="$fixture_dir/$dependency_flag.env"
  dependency_yaml="$fixture_dir/$dependency_key.yaml"
  sed "s/$dependency_flag=false/$dependency_flag=true/" "$valid_env" >"$dependency_env"
  sed "s/$dependency_key: false/$dependency_key: true/" "$valid_yaml" >"$dependency_yaml"
  expect_rejected "$dependency_flag without seller trading" "$dependency_env" "$dependency_yaml"
done

printf '\nOption ProductScope verifier self-test passed.\n'
