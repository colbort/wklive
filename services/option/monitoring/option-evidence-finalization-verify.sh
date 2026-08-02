#!/bin/sh
set -u

if [ "$#" -eq 0 ]; then
  printf 'usage: %s EVIDENCE_FILE [EVIDENCE_FILE ...]\n' "$0" >&2
  exit 2
fi

failures=0

for evidence_file in "$@"; do
  if [ ! -s "$evidence_file" ]; then
    printf 'FAIL  evidence file is missing or empty: %s\n' "$evidence_file" >&2
    failures=$((failures + 1))
    continue
  fi

  unresolved=$(
    awk '
      function report(reason) {
        printf "%d:%s: %s\n", NR, reason, $0
        bad=1
      }

      /^[[:space:]]*[A-Z0-9_]+_STATUS:[[:space:]]*(DRAFT|OPEN|REJECTED)[[:space:]]*$/ {
        report("non-final document status")
        next
      }

      /^[[:space:]]*-[[:space:]]*\[[[:space:]]\]/ {
        report("unchecked acceptance item")
        next
      }

      /REPLACE_BEFORE_DEPLOY/ {
        report("deployment placeholder")
        next
      }

      /[:：][[:space:]]*待填写([。；;[:space:]]*)$/ {
        report("unresolved prose field")
        next
      }

      /____/ {
        report("blank field")
        next
      }

      {
        content=$0
        gsub(/\[[^]]*\]\([^)]*\)/, "", content)
        if (content ~ /\[[[:upper:]][[:upper:][:digit:]_.\/-]*\]/ ||
            content ~ /\[(填写|待填写|待补|待批准)\]/) {
          report("bracketed placeholder")
          next
        }
      }

      /^[[:space:]]*\|/ {
        if ($0 ~ /待填写|待补|待批准|TBD|TODO/) {
          report("unresolved table placeholder")
          next
        }
        if ($0 ~ /通过\/拒绝|通过\/不适用\/拒绝|接受\/拒绝\/有条件接受|OPEN\/CLOSED|DRAFT\/APPROVED/) {
          report("unresolved table decision")
          next
        }
        if ($0 ~ /\[[[:space:]]\]/) {
          report("unchecked table item")
          next
        }
        if ($0 ~ /\|[[:space:]]*\|/) {
          report("empty table cell")
          next
        }
        if ($0 ~ /\|[[:space:]]*(OPEN|REJECTED|DRAFT)[[:space:]]*\|/) {
          report("non-final table status")
          next
        }
      }

      END { exit bad ? 1 : 0 }
    ' "$evidence_file"
  )
  result=$?

  if [ "$result" -eq 0 ]; then
    printf 'PASS  evidence has no unresolved finalization fields: %s\n' "$evidence_file"
  else
    printf 'FAIL  evidence contains unresolved finalization fields: %s\n' "$evidence_file" >&2
    printf '%s\n' "$unresolved" | sed 's/^/      /' >&2
    failures=$((failures + 1))
  fi
done

if [ "$failures" -ne 0 ]; then
  exit 1
fi
