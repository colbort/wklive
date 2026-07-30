#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ENV_FILE="${DR_PITR_ENV_FILE:-$SCRIPT_DIR/.env}"
SOURCE_DB="${DR_PITR_SOURCE_DB:-wklive_dr_pitr_probe}"
RESTORE_DB="${DR_PITR_RESTORE_DB:-wklive_dr_pitr_restore}"
MYSQL_HOST="${DR_PITR_MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${DR_PITR_MYSQL_PORT:-3306}"
MYSQL_USER="${DR_PITR_MYSQL_USER:-root}"

case "$SOURCE_DB" in
  ''|*[!a-zA-Z0-9_]*)
    echo "invalid PITR source database name" >&2
    exit 1
    ;;
esac
case "$RESTORE_DB" in
  ''|*[!a-zA-Z0-9_]*)
    echo "invalid PITR restore database name" >&2
    exit 1
    ;;
esac
if [ "$SOURCE_DB" = "$RESTORE_DB" ] || [ "$SOURCE_DB" = "wklive" ] || [ "$RESTORE_DB" = "wklive" ]; then
  echo "PITR probe databases must be distinct and must not be wklive" >&2
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "PITR environment file does not exist: $ENV_FILE" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

if [ -z "${MYSQL_ROOT_PASSWORD:-}" ]; then
  echo "MYSQL_ROOT_PASSWORD is required in the PITR environment file" >&2
  exit 1
fi

find_client() {
  client_name="$1"
  configured_path="$2"
  if [ -n "$configured_path" ] && [ -x "$configured_path" ]; then
    printf '%s\n' "$configured_path"
    return 0
  fi
  if command -v "$client_name" >/dev/null 2>&1; then
    command -v "$client_name"
    return 0
  fi
  for client_path in \
    "/opt/homebrew/opt/mysql-client/bin/$client_name" \
    "/usr/local/opt/mysql-client/bin/$client_name"
  do
    if [ -x "$client_path" ]; then
      printf '%s\n' "$client_path"
      return 0
    fi
  done
  return 1
}

MYSQL_BIN=$(find_client mysql "${MYSQL_CLIENT_BIN:-}") || {
  echo "mysql client was not found; set MYSQL_CLIENT_BIN" >&2
  exit 1
}
MYSQLBINLOG_BIN=$(find_client mysqlbinlog "${MYSQLBINLOG_CLIENT_BIN:-}") || {
  echo "mysqlbinlog client was not found; set MYSQLBINLOG_CLIENT_BIN" >&2
  exit 1
}

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"
created_source=0
created_restore=0

mysql_exec() {
  "$MYSQL_BIN" \
    --protocol=tcp \
    --host="$MYSQL_HOST" \
    --port="$MYSQL_PORT" \
    --user="$MYSQL_USER" \
    --batch \
    --skip-column-names \
    "$@"
}

cleanup() {
  cleanup_status=$?
  trap - 0 HUP INT TERM
  if [ "$created_source" -eq 1 ] || [ "$created_restore" -eq 1 ]; then
    mysql_exec -e "DROP DATABASE IF EXISTS \`$SOURCE_DB\`; DROP DATABASE IF EXISTS \`$RESTORE_DB\`;" \
      >/dev/null 2>&1 || true
  fi
  unset MYSQL_PWD
  exit "$cleanup_status"
}
trap cleanup 0 HUP INT TERM

existing_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.schemata
WHERE schema_name IN ('$SOURCE_DB','$RESTORE_DB');")
if [ "$existing_count" -ne 0 ]; then
  echo "PITR probe database already exists; refusing to overwrite it" >&2
  exit 1
fi

started_at=$(date +%s)
created_source=1
created_restore=1
mysql_exec -e "
CREATE DATABASE \`$SOURCE_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE DATABASE \`$RESTORE_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE TABLE \`$SOURCE_DB\`.pitr_fact (
  id BIGINT NOT NULL,
  payload VARCHAR(64) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB;
CREATE TABLE \`$RESTORE_DB\`.pitr_fact LIKE \`$SOURCE_DB\`.pitr_fact;"

start_status=$(mysql_exec -e "SHOW BINARY LOG STATUS;")
set -- $start_status
binlog_file="$1"
start_position="$2"

mysql_exec -e "
INSERT INTO \`$SOURCE_DB\`.pitr_fact(id,payload)
VALUES(1,'before-recovery-point');"

stop_status=$(mysql_exec -e "SHOW BINARY LOG STATUS;")
set -- $stop_status
stop_file="$1"
stop_position="$2"
if [ "$stop_file" != "$binlog_file" ]; then
  echo "binary log rotated during PITR probe; retry the command" >&2
  exit 1
fi

mysql_exec -e "
INSERT INTO \`$SOURCE_DB\`.pitr_fact(id,payload)
VALUES(2,'after-recovery-point');"

end_status=$(mysql_exec -e "SHOW BINARY LOG STATUS;")
set -- $end_status
end_file="$1"
end_position="$2"
if [ "$end_file" != "$binlog_file" ]; then
  echo "binary log rotated during PITR probe; retry the command" >&2
  exit 1
fi

"$MYSQLBINLOG_BIN" \
  --read-from-remote-server \
  --host="$MYSQL_HOST" \
  --port="$MYSQL_PORT" \
  --user="$MYSQL_USER" \
  --start-position="$start_position" \
  --stop-position="$stop_position" \
  --disable-log-bin \
  --database="$RESTORE_DB" \
  --rewrite-db="$SOURCE_DB->$RESTORE_DB" \
  "$binlog_file" |
  "$MYSQL_BIN" \
    --protocol=tcp \
    --host="$MYSQL_HOST" \
    --port="$MYSQL_PORT" \
    --user="$MYSQL_USER"

source_count=$(mysql_exec -e "SELECT COUNT(*) FROM \`$SOURCE_DB\`.pitr_fact;")
restore_count=$(mysql_exec -e "SELECT COUNT(*) FROM \`$RESTORE_DB\`.pitr_fact;")
restore_payload=$(mysql_exec -e "
SELECT payload
FROM \`$RESTORE_DB\`.pitr_fact
WHERE id=1;")
unexpected_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM \`$RESTORE_DB\`.pitr_fact
WHERE id=2;")

if [ "$source_count" -ne 2 ] ||
  [ "$restore_count" -ne 1 ] ||
  [ "$restore_payload" != "before-recovery-point" ] ||
  [ "$unexpected_count" -ne 0 ]; then
  printf '%s\n' \
    "DR_PITR_BINLOG_FILE=$binlog_file" \
    "DR_PITR_START_POSITION=$start_position" \
    "DR_PITR_STOP_POSITION=$stop_position" \
    "DR_PITR_END_POSITION=$end_position" \
    "DR_PITR_SOURCE_COUNT=$source_count" \
    "DR_PITR_RESTORED_COUNT=$restore_count" \
    "DR_PITR_RESTORED_PAYLOAD=$restore_payload" \
    "DR_PITR_UNEXPECTED_COUNT=$unexpected_count" >&2
  echo "PITR verification failed" >&2
  exit 1
fi

finished_at=$(date +%s)
duration_seconds=$((finished_at - started_at))
printf '%s\n' \
  "DR_PITR_SMOKE_RESULT=PASS" \
  "DR_PITR_BINLOG_FILE=$binlog_file" \
  "DR_PITR_START_POSITION=$start_position" \
  "DR_PITR_STOP_POSITION=$stop_position" \
  "DR_PITR_END_POSITION=$end_position" \
  "DR_PITR_SOURCE_COUNT=$source_count" \
  "DR_PITR_RESTORED_COUNT=$restore_count" \
  "DR_PITR_RESTORED_PAYLOAD=$restore_payload" \
  "DR_PITR_DURATION_SECONDS=$duration_seconds"
