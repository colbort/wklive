#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
VERIFY="$SCRIPT_DIR/option-external-contract-intake-verify.sh"
TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/option-contract-intake.XXXXXX")
trap 'rm -rf "$TMP_DIR"' EXIT HUP INT TERM

header='contract_code,underlying,call_put,exercise_style,settlement_type,quote_coin,settle_coin,underlying_coin,strike,contract_unit,multiplier,price_tick,qty_step,min_qty,max_qty,list_time_ms,last_trade_time_ms,exercise_cutoff_time_ms,expire_time_ms,deliver_time_ms,iana_timezone,series_id'
row_one='BTC-20261225-50000-C,BTCUSD,CALL,EUROPEAN,CASH,USDT,USDT,BTC,50000,0.01,1,0.1,0.001,0.001,10,1798000000000,1798100000000,1798100000000,1798200000000,1798200000000,Asia/Hong_Kong,SERIES-20261225'
row_two='BTC-20261225-45000-P,BTCUSD,PUT,AMERICAN,PHYSICAL,USDT,USDT,BTC,45000,0.01,1,0.1,0.001,0.001,10,1798000000000,1798100000000,1798100000000,1798200000000,1798300000000,UTC,NOT_APPLICABLE'

printf '%s\n' "$header" "$row_one" "$row_two" > "$TMP_DIR/valid.csv"

expect_pass() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.csv" >/dev/null 2>&1; then
    printf 'PASS  %s\n' "$case_name"
  else
    printf 'FAIL  %s should pass\n' "$case_name" >&2
    exit 1
  fi
}

expect_fail() {
  case_name=$1
  if "$VERIFY" "$TMP_DIR/$case_name.csv" >/dev/null 2>&1; then
    printf 'FAIL  %s should be rejected\n' "$case_name" >&2
    exit 1
  else
    printf 'PASS  %s rejected\n' "$case_name"
  fi
}

expect_pass valid

printf '%s\n' 'contract_code,underlying' "$row_one" > "$TMP_DIR/header.csv"
expect_fail header

printf '%s\n' "$header" "$row_one" "$row_one" > "$TMP_DIR/duplicate.csv"
expect_fail duplicate

printf '%s\n' "$header" "$(printf '%s\n' "$row_one" | sed 's/,BTCUSD,/,输入待填写,/')" > "$TMP_DIR/placeholder.csv"
expect_fail placeholder

printf '%s\n' "$header" "$(printf '%s\n' "$row_one" | sed 's/,CALL,/,BUY,/')" > "$TMP_DIR/enum.csv"
expect_fail enum

printf '%s\n' "$header" "$(printf '%s\n' "$row_one" | sed 's/,0.001,10,/,20,10,/')" > "$TMP_DIR/quantity.csv"
expect_fail quantity

printf '%s\n' "$header" "$(printf '%s\n' "$row_one" | sed 's/1798000000000,1798100000000/1798100000000,1798000000000/')" > "$TMP_DIR/time.csv"
expect_fail time

printf '%s\n' "$header" "$(printf '%s\n' "$row_one" | sed 's/Asia\/Hong_Kong/UnknownZone/')" > "$TMP_DIR/timezone.csv"
expect_fail timezone

printf '%s\n' "$header" > "$TMP_DIR/empty.csv"
expect_fail empty

printf '\nOption external contract-intake verifier self-test passed.\n'

