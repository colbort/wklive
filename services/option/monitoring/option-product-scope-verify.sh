#!/bin/sh
set -u

if [ "$#" -ne 2 ]; then
  printf 'usage: %s READINESS_ENV RENDERED_OPTION_YAML\n' "$0" >&2
  exit 2
fi

readiness_file="$1"
runtime_file="$2"

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  exit 1
}

[ -s "$readiness_file" ] || fail "readiness declaration is missing or empty"
[ -s "$runtime_file" ] || fail "rendered Option YAML is missing or empty"

read_unique_declaration() {
  setting_name="$1"
  awk -v setting_name="$setting_name" '
    $0 ~ "^[[:space:]]*" setting_name "[[:space:]]*=" {
      count++
      value=$0
      sub("^[[:space:]]*" setting_name "[[:space:]]*=[[:space:]]*", "", value)
      sub(/[[:space:]]+$/, "", value)
    }
    END {
      if (count != 1) exit 2
      print value
    }
  ' "$readiness_file"
}

read_unique_product_scope_value() {
  setting_name="$1"
  awk -v setting_name="$setting_name" '
    /^ProductScope:[[:space:]]*$/ { sections++; section=1; next }
    section && /^[^[:space:]#]/ { section=0 }
    section {
      line=$0
      sub(/^[[:space:]]+/, "", line)
      if (line ~ ("^" setting_name ":[[:space:]]*")) {
        count++
        sub("^" setting_name ":[[:space:]]*", "", line)
        sub(/[[:space:]]+$/, "", line)
        value=line
      }
    }
    END {
      if (sections != 1 || count != 1) exit 2
      print value
    }
  ' "$runtime_file"
}

verify_mapping() {
  declaration_name="$1"
  runtime_name="$2"

  declared=$(read_unique_declaration "$declaration_name") ||
    fail "$declaration_name must be declared exactly once"
  case "$declared" in
    true|false) ;;
    *) fail "$declaration_name must be the unquoted boolean true or false" ;;
  esac

  runtime_value=$(read_unique_product_scope_value "$runtime_name") ||
    fail "ProductScope.$runtime_name and the ProductScope section must each occur exactly once"
  case "$runtime_value" in
    true|false) ;;
    *) fail "ProductScope.$runtime_name must be the unquoted boolean true or false" ;;
  esac

  [ "$runtime_value" = "$declared" ] ||
    fail "ProductScope.$runtime_name=$runtime_value does not match $declaration_name=$declared"
  printf 'PASS  %s matches ProductScope.%s\n' "$declaration_name" "$runtime_name"
}

verify_mapping OPTION_SELLER_TRADING_ENABLED SellerTradingEnabled
verify_mapping OPTION_PORTFOLIO_MARGIN_ENABLED PortfolioMarginEnabled
verify_mapping OPTION_PHYSICAL_DELIVERY_ENABLED PhysicalDeliveryEnabled
verify_mapping OPTION_COMPLEX_ORDERS_ENABLED ComplexOrdersEnabled
verify_mapping OPTION_PUBLIC_MARKET_ENABLED PublicMarketEnabled
verify_mapping OPTION_MMP_ENABLED MMPEnabled
verify_mapping OPTION_AMERICAN_EXERCISE_ENABLED AmericanExerciseEnabled

seller_scope=$(read_unique_declaration OPTION_SELLER_TRADING_ENABLED) ||
  fail "OPTION_SELLER_TRADING_ENABLED must be declared exactly once"
for dependent_scope_name in OPTION_PORTFOLIO_MARGIN_ENABLED OPTION_PHYSICAL_DELIVERY_ENABLED \
  OPTION_COMPLEX_ORDERS_ENABLED OPTION_MMP_ENABLED OPTION_AMERICAN_EXERCISE_ENABLED; do
  dependent_scope=$(read_unique_declaration "$dependent_scope_name") ||
    fail "$dependent_scope_name must be declared exactly once"
  if [ "$dependent_scope" = "true" ] && [ "$seller_scope" != "true" ]; then
    fail "$dependent_scope_name requires OPTION_SELLER_TRADING_ENABLED=true"
  fi
done
printf 'PASS  seller-dependent product scopes are internally consistent\n'

greeks_scope=$(read_unique_declaration OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED) ||
  fail "OPTION_GREEKS_DEPENDENT_FEATURES_ENABLED must be declared exactly once"
[ "$greeks_scope" = "false" ] ||
  fail "Greeks-dependent optional features have no runtime entry and must remain false"
printf 'PASS  Greeks-dependent optional features remain unavailable\n'

printf '\nOption ProductScope declaration matches the rendered runtime config.\n'
