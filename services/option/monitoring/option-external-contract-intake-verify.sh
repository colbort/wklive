#!/bin/sh
set -u

if [ "$#" -ne 1 ]; then
  printf 'usage: %s COMPLETE_CONTRACT_INTAKE.csv\n' "$0" >&2
  exit 2
fi

intake_file=$1
[ -s "$intake_file" ] || {
  printf 'FAIL  contract intake CSV is missing or empty\n' >&2
  exit 1
}

expected_header='contract_code,underlying,call_put,exercise_style,settlement_type,quote_coin,settle_coin,underlying_coin,strike,contract_unit,multiplier,price_tick,qty_step,min_qty,max_qty,list_time_ms,last_trade_time_ms,exercise_cutoff_time_ms,expire_time_ms,deliver_time_ms,iana_timezone,series_id'

actual_header=$(sed -n '1{s/\r$//;p;}' "$intake_file")
[ "$actual_header" = "$expected_header" ] || {
  printf 'FAIL  contract intake CSV header does not match the governed 22-column schema\n' >&2
  exit 1
}

awk -F ',' '
  function fail(reason) {
    printf "FAIL  row %d: %s\n", NR, reason > "/dev/stderr"
    bad=1
  }
  function positive_decimal(value) {
    return value ~ /^([0-9]+([.][0-9]+)?|[.][0-9]+)$/ && (value + 0) > 0
  }
  function positive_integer(value) {
    return value ~ /^[1-9][0-9]*$/
  }
  function coin(value) {
    return value ~ /^[A-Z][A-Z0-9]{1,15}$/
  }
  function forbidden(value) {
    return value ~ /待填写|待补|TBD|TODO|同上|默认/
  }
  NR == 1 {next}
  {
    sub(/\r$/, "", $NF)
    rows++
    if (NF != 22) {
      fail("expected exactly 22 comma-separated fields")
      next
    }
    for (i=1; i<=22; i++) {
      if ($i == "") fail("field " i " is empty")
      if ($i ~ /^[[:space:]]|[[:space:]]$/) fail("field " i " has surrounding whitespace")
      if (forbidden($i)) fail("field " i " contains a placeholder")
    }
    if ($1 !~ /^[A-Za-z0-9][A-Za-z0-9_.-]{1,63}$/) fail("contract_code format is invalid")
    if (seen_code[$1]++) fail("contract_code is duplicated")
    if ($2 !~ /^[A-Z][A-Z0-9_.-]{1,31}$/) fail("underlying format is invalid")
    if ($3 != "CALL" && $3 != "PUT") fail("call_put must be CALL or PUT")
    if ($4 != "EUROPEAN" && $4 != "AMERICAN") fail("exercise_style must be EUROPEAN or AMERICAN")
    if ($5 != "CASH" && $5 != "PHYSICAL") fail("settlement_type must be CASH or PHYSICAL")
    if (!coin($6) || !coin($7) || !coin($8)) fail("coin codes must be uppercase system identifiers")
    for (i=9; i<=15; i++) {
      if (!positive_decimal($i)) fail("economic field " i " must be a positive decimal")
    }
    if (($14 + 0) > ($15 + 0)) fail("min_qty must not exceed max_qty")
    for (i=16; i<=20; i++) {
      if (!positive_integer($i)) fail("time field " i " must be a positive Unix millisecond integer")
    }
    if (!(($16 + 0) < ($17 + 0) && ($17 + 0) <= ($18 + 0) &&
          ($18 + 0) <= ($19 + 0) && ($19 + 0) <= ($20 + 0))) {
      fail("times must satisfy list < last_trade <= exercise_cutoff <= expire <= deliver")
    }
    if ($21 != "UTC" && $21 !~ /^[A-Za-z_+-]+\/[A-Za-z0-9_+.-]+(\/[A-Za-z0-9_+.-]+)?$/) {
      fail("iana_timezone must be UTC or an IANA zone name")
    }
    if ($22 != "NOT_APPLICABLE" && $22 !~ /^[A-Za-z0-9][A-Za-z0-9_.-]{1,63}$/) {
      fail("series_id format is invalid")
    }
  }
  END {
    if (rows < 1) {
      print "FAIL  contract intake CSV must contain at least one contract" > "/dev/stderr"
      bad=1
    }
    if (bad) exit 1
    printf "PASS  complete contract intake contains %d unique contract(s)\n", rows
  }
' "$intake_file"

