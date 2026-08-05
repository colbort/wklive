#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEPLOY_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
# shellcheck disable=SC1091
. "$DEPLOY_ROOT/common/scripts/dev-compose.sh"
MODE="${1:-}"
ENV_FILE="${DR_BACKUP_ENV_FILE:-$DEPLOY_ROOT/.env}"
PROBE_DB="wklive_dr_backup_probe"
RESTORE_DB="wklive_dr_backup_restore"
WORK_DIR=""
SMOKE_REMOTE_DIR=""
RESTORE_CONTAINER=""
created_probe=0
created_restore=0
created_restore_container=0
PITR_PROBE_TABLE=""
created_pitr_probe=0

usage() {
  echo "usage: $0 {smoke|local-verify|local-restore-verify|local-pitr-restore-verify|production}" >&2
}

case "$MODE" in
  smoke|local-verify|local-restore-verify|local-pitr-restore-verify|production)
    ;;
  *)
    usage
    exit 1
    ;;
esac

if [ ! -f "$ENV_FILE" ]; then
  echo "DR backup environment file does not exist: $ENV_FILE" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$ENV_FILE"
set +a

MYSQL_HOST="${DR_BACKUP_MYSQL_HOST:-127.0.0.1}"
MYSQL_PORT="${DR_BACKUP_MYSQL_PORT:-3306}"
MYSQL_USER="${DR_BACKUP_MYSQL_USER:-root}"
SOURCE_DB="${DR_BACKUP_SOURCE_DB:-wklive}"

if [ -z "${MYSQL_ROOT_PASSWORD:-}" ]; then
  echo "MYSQL_ROOT_PASSWORD is required in the DR backup environment file" >&2
  exit 1
fi

valid_database_name() {
  case "$1" in
    ''|*[!a-zA-Z0-9_]*)
      return 1
      ;;
  esac
}

if ! valid_database_name "$SOURCE_DB"; then
  echo "invalid DR backup source database name" >&2
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
OPENSSL_BIN=$(find_client openssl "${OPENSSL_CLIENT_BIN:-}") || {
  echo "OpenSSL was not found; set OPENSSL_CLIENT_BIN" >&2
  exit 1
}
MYSQLBINLOG_BIN=""
if [ "$MODE" = "local-pitr-restore-verify" ]; then
  MYSQLBINLOG_BIN=$(find_client mysqlbinlog "${MYSQLBINLOG_CLIENT_BIN:-}") || {
    echo "mysqlbinlog client was not found; set MYSQLBINLOG_CLIENT_BIN" >&2
    exit 1
  }
fi
use_compose_client="${DR_BACKUP_USE_COMPOSE_CLIENT:-}"
if [ -z "$use_compose_client" ] && [ "$MODE" = "smoke" ]; then
  use_compose_client=true
fi
case "$use_compose_client" in
  true)
    compose_mysql_id=$(
      dev_compose ps -q mysql 2>/dev/null || true
    )
    if [ -z "$compose_mysql_id" ]; then
      echo "Compose MySQL client requested but the Compose runtime is unavailable" >&2
      exit 1
    fi
    ;;
  false|'')
    MYSQLDUMP_BIN=$(find_client mysqldump "${MYSQLDUMP_CLIENT_BIN:-}") || {
      echo "mysqldump client was not found; set MYSQLDUMP_CLIENT_BIN" >&2
      exit 1
    }
    ;;
  *)
    echo "DR_BACKUP_USE_COMPOSE_CLIENT must be true or false" >&2
    exit 1
    ;;
esac

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"

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

load_system_object_storage() {
  use_system_storage="${DR_BACKUP_USE_SYSTEM_OBJECT_STORAGE:-true}"
  case "$use_system_storage" in
    true)
      ;;
    false)
      return 0
      ;;
    *)
      echo "DR_BACKUP_USE_SYSTEM_OBJECT_STORAGE must be true or false" >&2
      exit 1
      ;;
  esac

  storage_bucket="${DR_BACKUP_BUCKET_NAME:-}"
  case "$storage_bucket" in
    ''|*[!a-z0-9.-]*)
      echo "DR_BACKUP_BUCKET_NAME must be a lowercase DNS-safe dedicated bucket" >&2
      exit 1
      ;;
  esac

  object_prefix="${DR_BACKUP_OBJECT_PREFIX:-wklive/mysql}"
  object_prefix=${object_prefix#/}
  object_prefix=${object_prefix%/}
  case "$object_prefix" in
    ''|*[[:space:]]*)
      echo "DR_BACKUP_OBJECT_PREFIX must be non-empty and contain no whitespace" >&2
      exit 1
      ;;
  esac

  DR_BACKUP_DESTINATION_URI="s3://$storage_bucket/$object_prefix"
  export DR_BACKUP_DESTINATION_URI
  echo "DR object storage uses system config with dedicated bucket=$storage_bucket"
}

cleanup() {
  cleanup_status=$?
  trap - 0 HUP INT TERM
  if [ "$created_restore_container" -eq 1 ] && [ -n "$RESTORE_CONTAINER" ]; then
    docker rm -f "$RESTORE_CONTAINER" >/dev/null 2>&1 || true
  fi
  if [ "$created_pitr_probe" -eq 1 ] && [ -n "$PITR_PROBE_TABLE" ]; then
    mysql_exec -e \
      "DROP TABLE IF EXISTS \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\`;" \
      >/dev/null 2>&1 || true
  fi
  if [ "$created_probe" -eq 1 ] || [ "$created_restore" -eq 1 ]; then
    mysql_exec -e \
      "DROP DATABASE IF EXISTS \`$PROBE_DB\`; DROP DATABASE IF EXISTS \`$RESTORE_DB\`;" \
      >/dev/null 2>&1 || true
  fi
  if [ -n "$WORK_DIR" ] && [ -d "$WORK_DIR" ]; then
    rm -rf "$WORK_DIR"
  fi
  if [ -n "$SMOKE_REMOTE_DIR" ] && [ -d "$SMOKE_REMOTE_DIR" ]; then
    rm -rf "$SMOKE_REMOTE_DIR"
  fi
  unset MYSQL_PWD
  exit "$cleanup_status"
}
trap cleanup 0 HUP INT TERM

tmp_root="${TMPDIR:-/private/tmp}"
WORK_DIR=$(mktemp -d "$tmp_root/wklive-dr-backup.XXXXXX")
chmod 700 "$WORK_DIR"

check_free_space() {
  required_kb="$1"
  case "$required_kb" in
    ''|*[!0-9]*|0)
      echo "DR_BACKUP_MIN_FREE_KB must be a positive integer" >&2
      exit 1
      ;;
  esac
  available_kb=$(df -Pk "$WORK_DIR" | awk 'NR == 2 {print $4; exit}')
  case "$available_kb" in
    ''|*[!0-9]*)
      echo "unable to inspect DR backup temporary disk capacity" >&2
      exit 1
      ;;
  esac
  if [ "$available_kb" -lt "$required_kb" ]; then
    echo "DR backup temporary disk preflight failed: available=${available_kb}KiB required=${required_kb}KiB" >&2
    exit 1
  fi
}

backup_started_epoch=$(date +%s)
backup_started_utc=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
object_stamp=$(date -u '+%Y%m%dT%H%M%SZ')
recipient_cert=""
recipient_key=""
recipient_key_passphrase_file=""
key_id=""
operator_account=""
reviewer_account=""
retention_days=""
destination_uri=""
aws_bin=""
use_system_storage=false
storage_bucket=""
object_prefix=""
upload_performed=false

if [ "$MODE" = "smoke" ]; then
  check_free_space 262144
  SOURCE_DB="$PROBE_DB"
  recipient_cert="$WORK_DIR/recipient-cert.pem"
  recipient_key="$WORK_DIR/recipient-key.pem"
  key_id="ephemeral-smoke-key"
  operator_account="dr_operator"
  reviewer_account="production_reviewer"
  retention_days="0"
  SMOKE_REMOTE_DIR=$(mktemp -d "$tmp_root/wklive-dr-offsite-smoke.XXXXXX")
  chmod 700 "$SMOKE_REMOTE_DIR"
  destination_uri="file://$SMOKE_REMOTE_DIR"

  "$OPENSSL_BIN" req \
    -x509 \
    -newkey rsa:3072 \
    -nodes \
    -subj "/CN=wklive-dr-backup-smoke" \
    -days 1 \
    -keyout "$recipient_key" \
    -out "$recipient_cert" \
    >/dev/null 2>&1
  chmod 600 "$recipient_key"

  existing_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.schemata
WHERE schema_name IN ('$PROBE_DB','$RESTORE_DB');")
  if [ "$existing_count" -ne 0 ]; then
    echo "DR backup probe database already exists; refusing to overwrite it" >&2
    exit 1
  fi
  created_probe=1
  created_restore=1
  mysql_exec -e "
CREATE DATABASE \`$PROBE_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE DATABASE \`$RESTORE_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE TABLE \`$PROBE_DB\`.backup_fact (
  id BIGINT NOT NULL,
  payload VARCHAR(64) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB;
INSERT INTO \`$PROBE_DB\`.backup_fact(id,payload)
VALUES(1,'encrypted-backup-before'),(2,'encrypted-backup-after');"
else
  if [ "$MODE" = "local-restore-verify" ] ||
     [ "$MODE" = "local-pitr-restore-verify" ]; then
    check_free_space "${DR_BACKUP_RESTORE_MIN_FREE_KB:-12582912}"
  else
    check_free_space "${DR_BACKUP_MIN_FREE_KB:-8388608}"
  fi
  recipient_cert="${DR_BACKUP_RECIPIENT_CERT:-}"
  key_id="${DR_BACKUP_KEY_ID:-}"
  operator_account="${DR_BACKUP_OPERATOR_ACCOUNT:-}"
  reviewer_account="${DR_BACKUP_REVIEWER_ACCOUNT:-}"
  retention_days="${DR_BACKUP_RETENTION_DAYS:-}"
  if [ ! -f "$recipient_cert" ]; then
    echo "DR_BACKUP_RECIPIENT_CERT must reference an existing public certificate" >&2
    exit 1
  fi
  if ! "$OPENSSL_BIN" x509 -in "$recipient_cert" -noout -checkend 86400 >/dev/null 2>&1; then
    echo "DR_BACKUP_RECIPIENT_CERT is invalid or expires within 24 hours" >&2
    exit 1
  fi
  if [ -z "$key_id" ] || [ -z "$operator_account" ] || [ -z "$reviewer_account" ]; then
    echo "DR key ID, operator account, and reviewer account are required" >&2
    exit 1
  fi
  case "$key_id" in
    *[!a-zA-Z0-9_./:@-]*)
      echo "DR_BACKUP_KEY_ID contains unsupported characters" >&2
      exit 1
      ;;
  esac
  for account_name in "$operator_account" "$reviewer_account"; do
    case "$account_name" in
      ''|*[!a-zA-Z0-9_.-]*)
        echo "DR operator and reviewer accounts must use safe system-account identifiers" >&2
        exit 1
        ;;
    esac
  done
  case "$retention_days" in
    ''|*[!0-9]*|0)
      echo "DR_BACKUP_RETENTION_DAYS must be a positive integer" >&2
      exit 1
      ;;
  esac

  if [ "$MODE" = "local-verify" ] ||
     [ "$MODE" = "local-restore-verify" ] ||
     [ "$MODE" = "local-pitr-restore-verify" ]; then
    destination_uri="NOT_UPLOADED"
    recipient_key="${DR_BACKUP_RECIPIENT_KEY:-}"
    recipient_key_passphrase_file="${DR_BACKUP_RECIPIENT_KEY_PASSPHRASE_FILE:-}"
    if [ ! -f "$recipient_key" ]; then
      echo "DR_BACKUP_RECIPIENT_KEY must reference the local verification private key" >&2
      exit 1
    fi
    if [ ! -f "$recipient_key_passphrase_file" ]; then
      echo "DR_BACKUP_RECIPIENT_KEY_PASSPHRASE_FILE must reference an existing protected file" >&2
      exit 1
    fi
    cert_public_key_sha=$(
      "$OPENSSL_BIN" x509 -in "$recipient_cert" -pubkey -noout |
        "$OPENSSL_BIN" pkey -pubin -outform DER 2>/dev/null |
        sha256sum |
        awk '{print $1}'
    )
    private_public_key_sha=$(
      "$OPENSSL_BIN" pkey \
        -in "$recipient_key" \
        -passin "file:$recipient_key_passphrase_file" \
        -pubout \
        -outform DER 2>/dev/null |
        sha256sum |
        awk '{print $1}'
    )
    if [ -z "$cert_public_key_sha" ] ||
       [ "$cert_public_key_sha" != "$private_public_key_sha" ]; then
      echo "DR backup recipient certificate and local verification key do not match" >&2
      exit 1
    fi
  else
    load_system_object_storage
    destination_uri="${DR_BACKUP_DESTINATION_URI:-}"
    case "$destination_uri" in
      *[[:space:]]*)
        echo "production DR backup destination must not contain whitespace" >&2
        exit 1
        ;;
    esac
    case "$destination_uri" in
      s3://*)
        ;;
      file://*)
        echo "production DR backup refuses local file destinations; use an offsite s3:// URI" >&2
        exit 1
        ;;
      *)
        echo "production DR backup requires DR_BACKUP_DESTINATION_URI=s3://bucket/prefix" >&2
        exit 1
        ;;
    esac
    if [ "$use_system_storage" = "true" ]; then
      dev_compose run --rm --no-deps \
        --entrypoint /usr/local/bin/object-storage \
        -e OBJECT_STORAGE_BUCKET="$storage_bucket" \
        -e OBJECT_STORAGE_REQUIRE_PRIVATE_VERSIONED=true \
        db-init inspect
    else
      aws_bin=$(find_client aws "${AWS_CLIENT_BIN:-}") || {
        echo "AWS CLI was not found; set AWS_CLIENT_BIN for S3-compatible offsite storage" >&2
        exit 1
      }
    fi
    upload_performed=true
  fi
fi

decrypt_backup() {
  encrypted_input="$1"
  decrypted_output="$2"
  if [ -n "$recipient_key_passphrase_file" ]; then
    "$OPENSSL_BIN" cms \
      -decrypt \
      -binary \
      -inform DER \
      -recip "$recipient_cert" \
      -inkey "$recipient_key" \
      -passin "file:$recipient_key_passphrase_file" \
      -in "$encrypted_input" \
      -out "$decrypted_output"
  else
    "$OPENSSL_BIN" cms \
      -decrypt \
      -binary \
      -inform DER \
      -recip "$recipient_cert" \
      -inkey "$recipient_key" \
      -in "$encrypted_input" \
      -out "$decrypted_output"
  fi
}

raw_dump_file="$WORK_DIR/$SOURCE_DB.sql"
dump_file="$WORK_DIR/$SOURCE_DB.sql.gz"
cipher_file="$WORK_DIR/$SOURCE_DB-$object_stamp.sql.gz.cms"
readback_file="$WORK_DIR/readback.sql.gz.cms"
restore_dump="$WORK_DIR/restored.sql.gz"
manifest_file="$WORK_DIR/$SOURCE_DB-$object_stamp.manifest"
object_name=$(basename "$cipher_file")
manifest_name=$(basename "$manifest_file")
if [ "$MODE" = "local-verify" ] ||
   [ "$MODE" = "local-restore-verify" ] ||
   [ "$MODE" = "local-pitr-restore-verify" ]; then
  remote_object="NOT_UPLOADED"
  remote_manifest="NOT_UPLOADED"
else
  remote_object="${destination_uri%/}/$object_name"
  remote_manifest="${destination_uri%/}/$manifest_name"
fi

pitr_snapshot_file=""
pitr_snapshot_position=0
pitr_stop_file=""
pitr_stop_position=0
pitr_end_file=""
pitr_end_position=0
pitr_binlog_file_count=0
pitr_tail_bytes=0
pitr_tail_sha256=""
pitr_baseline_verified=false
pitr_recovery_point_verified=false
pitr_boundary_verified=false
pitr_source_cleanup_verified=false
pitr_business_table_count=0
pitr_expected_restore_table_count=0
pitr_tail_file="$WORK_DIR/pitr-tail.sql"
pitr_selected_logs_file="$WORK_DIR/pitr-selected-logs.txt"

if [ "$MODE" = "local-pitr-restore-verify" ]; then
  PITR_PROBE_TABLE="wklive_dr_pitr_${object_stamp}_$$"
  pitr_run_token="$object_stamp-$$"
  case "$PITR_PROBE_TABLE" in
    *[!a-zA-Z0-9_]*)
      echo "generated PITR probe table name is unsafe" >&2
      exit 1
      ;;
  esac
  pitr_probe_exists=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema='$SOURCE_DB'
  AND table_name='$PITR_PROBE_TABLE';")
  if [ "$pitr_probe_exists" -ne 0 ]; then
    echo "generated PITR probe table already exists; refusing to overwrite it" >&2
    exit 1
  fi
  created_pitr_probe=1
  mysql_exec -e "
CREATE TABLE \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\` (
  id TINYINT UNSIGNED NOT NULL,
  run_token VARCHAR(64) NOT NULL,
  payload VARCHAR(64) NOT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB;
INSERT INTO \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\`(id,run_token,payload)
VALUES(1,'$pitr_run_token','full-backup-baseline');"
fi

binlog_before=$(mysql_exec -e "SHOW BINARY LOG STATUS;" || true)

dump_database() {
  target_file="$1"
  if [ "$use_compose_client" = "true" ]; then
    dev_compose exec -T mysql \
      sh -lc '
        export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"
        exec mysqldump \
          -uroot \
          --single-transaction \
          --routines \
          --events \
          --triggers \
          --hex-blob \
          --set-gtid-purged=OFF \
          --source-data=2 \
          --no-tablespaces \
          "$1"
      ' sh "$SOURCE_DB" >"$target_file"
    return
  fi
  "$MYSQLDUMP_BIN" \
    --protocol=tcp \
    --host="$MYSQL_HOST" \
    --port="$MYSQL_PORT" \
    --user="$MYSQL_USER" \
    --single-transaction \
    --routines \
    --events \
    --triggers \
    --hex-blob \
    --set-gtid-purged=OFF \
    --source-data=2 \
    --no-tablespaces \
    "$SOURCE_DB" >"$target_file"
}

dump_database "$raw_dump_file"
if [ ! -s "$raw_dump_file" ]; then
  echo "mysqldump produced an empty backup" >&2
  exit 1
fi
raw_dump_bytes=$(wc -c <"$raw_dump_file" | awk '{$1=$1; print}')

if [ "$MODE" = "local-pitr-restore-verify" ]; then
  pitr_snapshot_line=$(
    grep -m 1 '^-- CHANGE REPLICATION ' "$raw_dump_file" || true
  )
  pitr_snapshot_file=$(
    printf '%s\n' "$pitr_snapshot_line" |
      awk -F "'" 'NF >= 3 {print $2; exit}'
  )
  pitr_snapshot_position=$(
    printf '%s\n' "$pitr_snapshot_line" |
      sed -E 's/.*(SOURCE_LOG_POS|MASTER_LOG_POS)=([0-9]+).*/\2/'
  )
  case "$pitr_snapshot_file" in
    ''|*[!a-zA-Z0-9._-]*)
      echo "unable to parse a safe snapshot Binlog file from mysqldump" >&2
      exit 1
      ;;
  esac
  case "$pitr_snapshot_position" in
    ''|*[!0-9]*|0)
      echo "unable to parse the snapshot Binlog position from mysqldump" >&2
      exit 1
      ;;
  esac

  mysql_exec -e "
INSERT INTO \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\`(id,run_token,payload)
VALUES(2,'$pitr_run_token','at-recovery-point');"
  pitr_stop_status=$(mysql_exec -e "SHOW BINARY LOG STATUS;")
  pitr_stop_file=$(printf '%s\n' "$pitr_stop_status" | awk '{print $1; exit}')
  pitr_stop_position=$(printf '%s\n' "$pitr_stop_status" | awk '{print $2; exit}')

  mysql_exec -e "
INSERT INTO \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\`(id,run_token,payload)
VALUES(3,'$pitr_run_token','after-recovery-point');"
  pitr_end_status=$(mysql_exec -e "SHOW BINARY LOG STATUS;")
  pitr_end_file=$(printf '%s\n' "$pitr_end_status" | awk '{print $1; exit}')
  pitr_end_position=$(printf '%s\n' "$pitr_end_status" | awk '{print $2; exit}')

  for pitr_coordinate in \
    "$pitr_stop_file" "$pitr_end_file"; do
    case "$pitr_coordinate" in
      ''|*[!a-zA-Z0-9._-]*)
        echo "PITR Binlog status returned an unsafe file name" >&2
        exit 1
        ;;
    esac
  done
  for pitr_coordinate in \
    "$pitr_stop_position" "$pitr_end_position"; do
    case "$pitr_coordinate" in
      ''|*[!0-9]*|0)
        echo "PITR Binlog status returned an invalid position" >&2
        exit 1
        ;;
    esac
  done
  if [ "$pitr_stop_file" = "$pitr_end_file" ] &&
     [ "$pitr_stop_position" -ge "$pitr_end_position" ]; then
    echo "PITR recovery boundary does not precede the post-boundary transaction" >&2
    exit 1
  fi

  mysql_exec -e "SHOW BINARY LOGS;" |
    awk -v start="$pitr_snapshot_file" -v stop="$pitr_stop_file" '
      $1 == start {capture=1}
      capture {print $1}
      $1 == stop {found_stop=1; exit}
      END {
        if (!capture || !found_stop) {
          exit 1
        }
      }
    ' >"$pitr_selected_logs_file" || {
      echo "snapshot or recovery-point Binlog file is no longer available" >&2
      exit 1
    }
  pitr_binlog_file_count=$(
    wc -l <"$pitr_selected_logs_file" | awk '{$1=$1; print}'
  )
  if [ "$pitr_binlog_file_count" -eq 0 ] ||
     [ "$(sed -n '1p' "$pitr_selected_logs_file")" != "$pitr_snapshot_file" ] ||
     [ "$(sed -n '$p' "$pitr_selected_logs_file")" != "$pitr_stop_file" ]; then
    echo "PITR Binlog range selection is incomplete" >&2
    exit 1
  fi

  stream_remote_binlog() {
    "$MYSQLBINLOG_BIN" \
      --read-from-remote-server \
      --host="$MYSQL_HOST" \
      --port="$MYSQL_PORT" \
      --user="$MYSQL_USER" \
      "$@"
  }

  : >"$pitr_tail_file"
  if [ "$pitr_snapshot_file" = "$pitr_stop_file" ]; then
    stream_remote_binlog \
      "--start-position=$pitr_snapshot_position" \
      "--stop-position=$pitr_stop_position" \
      --disable-log-bin \
      "--database=$RESTORE_DB" \
      "--rewrite-db=$SOURCE_DB->$RESTORE_DB" \
      "$pitr_snapshot_file" >"$pitr_tail_file"
  else
    stream_remote_binlog \
      "--start-position=$pitr_snapshot_position" \
      --disable-log-bin \
      "--database=$RESTORE_DB" \
      "--rewrite-db=$SOURCE_DB->$RESTORE_DB" \
      "$pitr_snapshot_file" >"$pitr_tail_file"
    sed -n '2,$p' "$pitr_selected_logs_file" |
      sed '$d' |
      while IFS= read -r pitr_middle_file; do
        stream_remote_binlog \
          --disable-log-bin \
          "--database=$RESTORE_DB" \
          "--rewrite-db=$SOURCE_DB->$RESTORE_DB" \
          "$pitr_middle_file" >>"$pitr_tail_file"
      done
    stream_remote_binlog \
      "--stop-position=$pitr_stop_position" \
      --disable-log-bin \
      "--database=$RESTORE_DB" \
      "--rewrite-db=$SOURCE_DB->$RESTORE_DB" \
      "$pitr_stop_file" >>"$pitr_tail_file"
  fi
  if [ ! -s "$pitr_tail_file" ]; then
    echo "PITR Binlog extraction produced an empty SQL stream" >&2
    exit 1
  fi
  pitr_tail_bytes=$(wc -c <"$pitr_tail_file" | awk '{$1=$1; print}')
  pitr_tail_sha256=$(sha256sum "$pitr_tail_file" | awk '{print $1}')
  pitr_expected_restore_table_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema='$SOURCE_DB'
  AND table_type='BASE TABLE';")

  mysql_exec -e "DROP TABLE \`$SOURCE_DB\`.\`$PITR_PROBE_TABLE\`;"
  created_pitr_probe=0
  pitr_probe_exists=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema='$SOURCE_DB'
  AND table_name='$PITR_PROBE_TABLE';")
  if [ "$pitr_probe_exists" -ne 0 ]; then
    echo "PITR source probe table cleanup failed" >&2
    exit 1
  fi
  pitr_source_cleanup_verified=true
  pitr_business_table_count=$((pitr_expected_restore_table_count - 1))
fi

gzip -c "$raw_dump_file" >"$dump_file"
rm -f "$raw_dump_file"

gzip -t "$dump_file"
dump_sha256=$(sha256sum "$dump_file" | awk '{print $1}')

"$OPENSSL_BIN" cms \
  -encrypt \
  -binary \
  -stream \
  -outform DER \
  -aes-256-gcm \
  -recip "$recipient_cert" \
  -in "$dump_file" \
  -out "$cipher_file"

if ! "$OPENSSL_BIN" cms \
  -cmsout \
  -inform DER \
  -in "$cipher_file" \
  -print |
  grep -q "algorithm: aes-256-gcm"; then
  echo "encrypted backup is not CMS AES-256-GCM" >&2
  exit 1
fi

cipher_sha256=$(sha256sum "$cipher_file" | awk '{print $1}')
cipher_bytes=$(wc -c <"$cipher_file" | awk '{$1=$1; print}')

aws_s3_copy() {
  source_path="$1"
  target_path="$2"
  if [ -n "${DR_BACKUP_S3_ENDPOINT:-}" ]; then
    "$aws_bin" \
      --endpoint-url "$DR_BACKUP_S3_ENDPOINT" \
      s3 cp "$source_path" "$target_path" --only-show-errors
  else
    "$aws_bin" s3 cp "$source_path" "$target_path" --only-show-errors
  fi
}

system_s3_copy() {
  source_path="$1"
  target_path="$2"
  remote_prefix="s3://$storage_bucket/"
  case "$source_path" in
    "$remote_prefix"*)
      operation="download"
      object_key=${source_path#"$remote_prefix"}
      local_path="$target_path"
      ;;
    *)
      case "$target_path" in
        "$remote_prefix"*)
          operation="upload"
          object_key=${target_path#"$remote_prefix"}
          local_path="$source_path"
          ;;
        *)
          echo "system object storage copy requires one s3:// endpoint" >&2
          exit 1
          ;;
      esac
      ;;
  esac
  case "$local_path" in
    "$WORK_DIR"/*)
      container_path="/backup/${local_path#"$WORK_DIR"/}"
      ;;
    *)
      echo "system object storage local path must be inside the protected work directory" >&2
      exit 1
      ;;
  esac

  dev_compose run --rm --no-deps \
    --entrypoint /usr/local/bin/object-storage \
    -e OBJECT_STORAGE_BUCKET="$storage_bucket" \
    -e OBJECT_STORAGE_OBJECT_KEY="$object_key" \
    -e OBJECT_STORAGE_LOCAL_FILE="$container_path" \
    -v "$WORK_DIR:/backup" \
    db-init "$operation"
}

remote_copy() {
  if [ "$use_system_storage" = "true" ]; then
    system_s3_copy "$1" "$2"
  else
    aws_s3_copy "$1" "$2"
  fi
}

case "$MODE" in
  smoke)
    cp "$cipher_file" "$SMOKE_REMOTE_DIR/$object_name"
    cp "$SMOKE_REMOTE_DIR/$object_name" "$readback_file"
    ;;
  local-verify)
    cp "$cipher_file" "$readback_file"
    ;;
  local-restore-verify)
    cp "$cipher_file" "$readback_file"
    ;;
  local-pitr-restore-verify)
    cp "$cipher_file" "$readback_file"
    ;;
  production)
    remote_copy "$cipher_file" "$remote_object"
    remote_copy "$remote_object" "$readback_file"
    ;;
esac

readback_sha256=$(sha256sum "$readback_file" | awk '{print $1}')
if [ "$readback_sha256" != "$cipher_sha256" ]; then
  echo "encrypted backup readback checksum mismatch" >&2
  exit 1
fi

decrypt_verified=false
tamper_rejected=false
restored_fact_count=0
restored_dump_sha256=""
full_restore_verified=false
restore_cleanup_verified=false
restored_table_count=0
source_table_count=0
restored_check_ok_count=0
restored_total_rows=0
restored_counts_sha256=""
restored_schema_migrations=0
restored_trade_orders=0
restored_trade_fills=0
restored_contract_positions=0
restored_position_history=0
restored_settlement_instructions=0
restored_trade_events=0
restored_trade_event_inbox=0
restored_funding_batches=0
restored_delivery_batches=0
restored_reconciliation_issues=0
restored_authoritative_snapshots=0
restored_snapshot_outbox=0
if [ "$MODE" = "smoke" ] ||
   [ "$MODE" = "local-verify" ] ||
   [ "$MODE" = "local-restore-verify" ] ||
   [ "$MODE" = "local-pitr-restore-verify" ]; then
  decrypt_backup "$readback_file" "$restore_dump"
  gzip -t "$restore_dump"
  restored_dump_sha256=$(sha256sum "$restore_dump" | awk '{print $1}')
  if [ "$restored_dump_sha256" != "$dump_sha256" ]; then
    echo "decrypted backup checksum does not match the compressed SQL backup" >&2
    exit 1
  fi
  decrypt_verified=true

  if [ "$MODE" = "smoke" ]; then
    gzip -dc "$restore_dump" |
      "$MYSQL_BIN" \
        --protocol=tcp \
        --host="$MYSQL_HOST" \
        --port="$MYSQL_PORT" \
        --user="$MYSQL_USER" \
        "$RESTORE_DB"
    restored_fact_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM \`$RESTORE_DB\`.backup_fact
WHERE (id=1 AND payload='encrypted-backup-before')
   OR (id=2 AND payload='encrypted-backup-after');")
    if [ "$restored_fact_count" -ne 2 ]; then
      echo "decrypted DR backup restore fact verification failed" >&2
      exit 1
    fi
  fi

  tampered_file="$WORK_DIR/tampered.sql.gz.cms"
  cp "$readback_file" "$tampered_file"
  tamper_offset=$((cipher_bytes / 2))
  printf '\001' |
    dd of="$tampered_file" bs=1 seek="$tamper_offset" conv=notrunc >/dev/null 2>&1
  if decrypt_backup "$tampered_file" "$WORK_DIR/tampered-output" >/dev/null 2>&1; then
    echo "tampered encrypted backup was unexpectedly accepted" >&2
    exit 1
  fi
  tamper_rejected=true
fi

if [ "$MODE" = "local-restore-verify" ] ||
   [ "$MODE" = "local-pitr-restore-verify" ]; then
  RESTORE_CONTAINER="wklive-dr-full-restore-$object_stamp-$$"
  restore_container_name="$RESTORE_CONTAINER"
  if docker container inspect "$RESTORE_CONTAINER" >/dev/null 2>&1; then
    echo "DR restore verification container already exists; refusing to overwrite it" >&2
    exit 1
  fi

  restore_data_dir="$WORK_DIR/mysql-restore-data"
  mkdir -p "$restore_data_dir"
  chmod 700 "$restore_data_dir"
  restore_password="wklive-local-restore-$object_stamp-$$"

  docker run -d \
    --pull=never \
    --name "$RESTORE_CONTAINER" \
    --label wklive.dr.restore=true \
    --network none \
    --mount "type=bind,src=$restore_data_dir,dst=/var/lib/mysql" \
    -e "MYSQL_ROOT_PASSWORD=$restore_password" \
    mysql:8.4 \
    --skip-log-bin \
    --innodb-buffer-pool-size=536870912 \
    --innodb-flush-log-at-trx-commit=2 \
    --sync-binlog=0 \
    >/dev/null
  created_restore_container=1

  restore_ready=false
  restore_wait_attempt=0
  while [ "$restore_wait_attempt" -lt 90 ]; do
    if docker exec \
      -e "MYSQL_PWD=$restore_password" \
      "$RESTORE_CONTAINER" \
      mysqladmin -uroot ping --silent >/dev/null 2>&1; then
      restore_ready=true
      break
    fi
    if [ "$(docker inspect -f '{{.State.Running}}' "$RESTORE_CONTAINER" 2>/dev/null || true)" != "true" ]; then
      docker logs "$RESTORE_CONTAINER" >&2 || true
      echo "isolated MySQL restore container exited before becoming ready" >&2
      exit 1
    fi
    restore_wait_attempt=$((restore_wait_attempt + 1))
    sleep 2
  done
  if [ "$restore_ready" != "true" ]; then
    docker logs "$RESTORE_CONTAINER" >&2 || true
    echo "isolated MySQL restore container did not become ready" >&2
    exit 1
  fi

  restored_mysql_exec() {
    docker exec \
      -e "MYSQL_PWD=$restore_password" \
      "$RESTORE_CONTAINER" \
      mysql \
      -uroot \
      --batch \
      --skip-column-names \
      "$@"
  }

  restored_mysql_exec -e \
    "CREATE DATABASE \`$RESTORE_DB\` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
  if ! gzip -dc "$restore_dump" |
    docker exec -i \
      -e "MYSQL_PWD=$restore_password" \
      "$RESTORE_CONTAINER" \
      mysql -uroot "$RESTORE_DB"; then
    echo "isolated full database restore failed" >&2
    exit 1
  fi

  if [ "$MODE" = "local-pitr-restore-verify" ]; then
    if ! docker exec -i \
      -e "MYSQL_PWD=$restore_password" \
      "$RESTORE_CONTAINER" \
      mysql -uroot <"$pitr_tail_file"; then
      echo "isolated Binlog point-in-time replay failed" >&2
      exit 1
    fi
    pitr_baseline_count=$(restored_mysql_exec -e "
SELECT COUNT(*)
FROM \`$RESTORE_DB\`.\`$PITR_PROBE_TABLE\`
WHERE id=1
  AND run_token='$pitr_run_token'
  AND payload='full-backup-baseline';")
    pitr_recovery_point_count=$(restored_mysql_exec -e "
SELECT COUNT(*)
FROM \`$RESTORE_DB\`.\`$PITR_PROBE_TABLE\`
WHERE id=2
  AND run_token='$pitr_run_token'
  AND payload='at-recovery-point';")
    pitr_after_boundary_count=$(restored_mysql_exec -e "
SELECT COUNT(*)
FROM \`$RESTORE_DB\`.\`$PITR_PROBE_TABLE\`
WHERE id=3;")
    if [ "$pitr_baseline_count" -ne 1 ] ||
       [ "$pitr_recovery_point_count" -ne 1 ] ||
       [ "$pitr_after_boundary_count" -ne 0 ]; then
      echo "isolated full-database PITR boundary verification failed" >&2
      exit 1
    fi
    pitr_baseline_verified=true
    pitr_recovery_point_verified=true
    pitr_boundary_verified=true
  fi

  restored_tables_file="$WORK_DIR/restored-tables.txt"
  restored_mysql_exec -e "
SELECT table_name
FROM information_schema.tables
WHERE table_schema='$RESTORE_DB'
  AND table_type='BASE TABLE'
ORDER BY table_name;" >"$restored_tables_file"
  restored_table_count=$(wc -l <"$restored_tables_file" | awk '{$1=$1; print}')
  if [ "$MODE" = "local-pitr-restore-verify" ]; then
    source_table_count="$pitr_expected_restore_table_count"
  else
    source_table_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema='$SOURCE_DB'
  AND table_type='BASE TABLE';")
  fi
  if [ "$restored_table_count" -ne "$source_table_count" ]; then
    echo "isolated restore base-table count does not match the source schema" >&2
    exit 1
  fi

  restore_validation_sql="$WORK_DIR/restore-validation.sql"
  while IFS= read -r restored_table; do
    case "$restored_table" in
      ''|*[!a-zA-Z0-9_]*)
        echo "isolated restore contains an unsafe table identifier" >&2
        exit 1
        ;;
    esac
    printf "SELECT 'COUNT','%s',COUNT(*) FROM \`%s\`;\n" \
      "$restored_table" "$restored_table"
    printf "CHECK TABLE \`%s\`;\n" "$restored_table"
  done <"$restored_tables_file" >"$restore_validation_sql"

  restore_validation_output="$WORK_DIR/restore-validation.tsv"
  docker exec -i \
    -e "MYSQL_PWD=$restore_password" \
    "$RESTORE_CONTAINER" \
    mysql \
    -uroot \
    --batch \
    --skip-column-names \
    "$RESTORE_DB" \
    <"$restore_validation_sql" \
    >"$restore_validation_output"

  restored_count_rows=$(
    awk -F '\t' '$1=="COUNT" {count++} END {print count+0}' \
      "$restore_validation_output"
  )
  restored_check_ok_count=$(
    awk -F '\t' '$2=="check" && $3=="status" && $4=="OK" {count++} END {print count+0}' \
      "$restore_validation_output"
  )
  if [ "$restored_count_rows" -ne "$restored_table_count" ] ||
     [ "$restored_check_ok_count" -ne "$restored_table_count" ]; then
    echo "isolated restore table count or CHECK TABLE verification failed" >&2
    exit 1
  fi
  restored_total_rows=$(
    awk -F '\t' '$1=="COUNT" {sum+=$3} END {printf "%.0f\n",sum}' \
      "$restore_validation_output"
  )
  restored_counts_sha256=$(
    awk -F '\t' '$1=="COUNT" {print $2 "\t" $3}' \
      "$restore_validation_output" |
      sha256sum |
      awk '{print $1}'
  )

  restored_count_for() {
    table_name="$1"
    awk -F '\t' -v table_name="$table_name" \
      '$1=="COUNT" && $2==table_name {print $3; found=1; exit}
       END {if (!found) print 0}' \
      "$restore_validation_output"
  }

  restored_schema_migrations=$(restored_count_for schema_migrations)
  restored_trade_orders=$(restored_count_for t_trade_order)
  restored_trade_fills=$(restored_count_for t_trade_fill)
  restored_contract_positions=$(restored_count_for t_contract_position)
  restored_position_history=$(restored_count_for t_contract_position_history)
  restored_settlement_instructions=$(restored_count_for t_trade_settlement_instruction)
  restored_trade_events=$(restored_count_for t_biz_trade_event)
  restored_trade_event_inbox=$(restored_count_for t_trade_event_inbox)
  restored_funding_batches=$(restored_count_for t_contract_funding_batch)
  restored_delivery_batches=$(restored_count_for t_contract_delivery_batch)
  restored_reconciliation_issues=$(restored_count_for t_contract_reconciliation_issue)
  restored_authoritative_snapshots=$(restored_count_for t_itick_authoritative_snapshot)
  restored_snapshot_outbox=$(restored_count_for t_itick_snapshot_outbox)
  full_restore_verified=true

  docker rm -f "$RESTORE_CONTAINER" >/dev/null
  created_restore_container=0
  RESTORE_CONTAINER=""
  rm -rf "$restore_data_dir"
  if docker container inspect "$restore_container_name" >/dev/null 2>&1 ||
     [ -e "$restore_data_dir" ]; then
    echo "isolated restore container or data directory cleanup failed" >&2
    exit 1
  fi
  restore_cleanup_verified=true
fi

binlog_after=$(mysql_exec -e "SHOW BINARY LOG STATUS;" || true)
schema_migration_count=0
schema_migration_table_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema='$SOURCE_DB' AND table_name='schema_migrations';")
if [ "$schema_migration_table_count" -eq 1 ]; then
  schema_migration_count=$(mysql_exec -e "
SELECT COUNT(*)
FROM \`$SOURCE_DB\`.schema_migrations;")
fi
if { [ "$MODE" = "local-restore-verify" ] ||
     [ "$MODE" = "local-pitr-restore-verify" ]; } &&
   [ "$restored_schema_migrations" -ne "$schema_migration_count" ]; then
  echo "isolated restore schema migration count does not match the source" >&2
  exit 1
fi
backup_finished_epoch=$(date +%s)
backup_finished_utc=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
duration_seconds=$((backup_finished_epoch - backup_started_epoch))

{
  printf '%s\n' \
    "DR_BACKUP_FORMAT_VERSION=1" \
    "DR_BACKUP_MODE=$MODE" \
    "DR_BACKUP_SOURCE_DB=$SOURCE_DB" \
    "DR_BACKUP_STARTED_AT=$backup_started_utc" \
    "DR_BACKUP_FINISHED_AT=$backup_finished_utc" \
    "DR_BACKUP_DURATION_SECONDS=$duration_seconds" \
    "DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM" \
    "DR_BACKUP_KEY_ID=$key_id" \
    "DR_BACKUP_OPERATOR_ACCOUNT=$operator_account" \
    "DR_BACKUP_REVIEWER_ACCOUNT=$reviewer_account" \
    "DR_BACKUP_RETENTION_DAYS=$retention_days" \
    "DR_BACKUP_DESTINATION_URI=$destination_uri" \
    "DR_BACKUP_OBJECT=$remote_object" \
    "DR_BACKUP_UPLOAD_PERFORMED=$upload_performed" \
    "DR_BACKUP_RAW_SQL_BYTES=$raw_dump_bytes" \
    "DR_BACKUP_CIPHER_BYTES=$cipher_bytes" \
    "DR_BACKUP_CIPHER_SHA256=$cipher_sha256" \
    "DR_BACKUP_COMPRESSED_SQL_SHA256=$dump_sha256" \
    "DR_BACKUP_READBACK_SHA256=$readback_sha256" \
    "DR_BACKUP_RESTORED_COMPRESSED_SQL_SHA256=$restored_dump_sha256" \
    "DR_BACKUP_DECRYPT_VERIFIED=$decrypt_verified" \
    "DR_BACKUP_TAMPER_REJECTED=$tamper_rejected" \
    "DR_BACKUP_RESTORED_FACT_COUNT=$restored_fact_count" \
    "DR_BACKUP_FULL_RESTORE_VERIFIED=$full_restore_verified" \
    "DR_BACKUP_RESTORE_CLEANUP_VERIFIED=$restore_cleanup_verified" \
    "DR_BACKUP_RESTORED_TABLE_COUNT=$restored_table_count" \
    "DR_BACKUP_SOURCE_TABLE_COUNT=$source_table_count" \
    "DR_BACKUP_RESTORED_CHECK_OK_COUNT=$restored_check_ok_count" \
    "DR_BACKUP_RESTORED_TOTAL_ROWS=$restored_total_rows" \
    "DR_BACKUP_RESTORED_COUNTS_SHA256=$restored_counts_sha256" \
    "DR_BACKUP_RESTORED_SCHEMA_MIGRATIONS=$restored_schema_migrations" \
    "DR_BACKUP_RESTORED_TRADE_ORDERS=$restored_trade_orders" \
    "DR_BACKUP_RESTORED_TRADE_FILLS=$restored_trade_fills" \
    "DR_BACKUP_RESTORED_CONTRACT_POSITIONS=$restored_contract_positions" \
    "DR_BACKUP_RESTORED_POSITION_HISTORY=$restored_position_history" \
    "DR_BACKUP_RESTORED_SETTLEMENT_INSTRUCTIONS=$restored_settlement_instructions" \
    "DR_BACKUP_RESTORED_TRADE_EVENTS=$restored_trade_events" \
    "DR_BACKUP_RESTORED_TRADE_EVENT_INBOX=$restored_trade_event_inbox" \
    "DR_BACKUP_RESTORED_FUNDING_BATCHES=$restored_funding_batches" \
    "DR_BACKUP_RESTORED_DELIVERY_BATCHES=$restored_delivery_batches" \
    "DR_BACKUP_RESTORED_RECONCILIATION_ISSUES=$restored_reconciliation_issues" \
    "DR_BACKUP_RESTORED_AUTHORITATIVE_SNAPSHOTS=$restored_authoritative_snapshots" \
    "DR_BACKUP_RESTORED_SNAPSHOT_OUTBOX=$restored_snapshot_outbox" \
    "DR_BACKUP_PITR_SNAPSHOT_FILE=$pitr_snapshot_file" \
    "DR_BACKUP_PITR_SNAPSHOT_POSITION=$pitr_snapshot_position" \
    "DR_BACKUP_PITR_STOP_FILE=$pitr_stop_file" \
    "DR_BACKUP_PITR_STOP_POSITION=$pitr_stop_position" \
    "DR_BACKUP_PITR_END_FILE=$pitr_end_file" \
    "DR_BACKUP_PITR_END_POSITION=$pitr_end_position" \
    "DR_BACKUP_PITR_BINLOG_FILE_COUNT=$pitr_binlog_file_count" \
    "DR_BACKUP_PITR_TAIL_BYTES=$pitr_tail_bytes" \
    "DR_BACKUP_PITR_TAIL_SHA256=$pitr_tail_sha256" \
    "DR_BACKUP_PITR_BASELINE_VERIFIED=$pitr_baseline_verified" \
    "DR_BACKUP_PITR_RECOVERY_POINT_VERIFIED=$pitr_recovery_point_verified" \
    "DR_BACKUP_PITR_BOUNDARY_VERIFIED=$pitr_boundary_verified" \
    "DR_BACKUP_PITR_SOURCE_CLEANUP_VERIFIED=$pitr_source_cleanup_verified" \
    "DR_BACKUP_PITR_BUSINESS_TABLE_COUNT=$pitr_business_table_count" \
    "DR_BACKUP_SCHEMA_MIGRATION_COUNT=$schema_migration_count" \
    "DR_BACKUP_BINLOG_BEFORE=$binlog_before" \
    "DR_BACKUP_BINLOG_AFTER=$binlog_after"
} >"$manifest_file"

case "$MODE" in
  smoke)
    cp "$manifest_file" "$SMOKE_REMOTE_DIR/$manifest_name"
    ;;
  local-verify)
    ;;
  local-restore-verify)
    ;;
  local-pitr-restore-verify)
    ;;
  production)
    remote_copy "$manifest_file" "$remote_manifest"
    ;;
esac

printf '%s\n' \
  "DR_ENCRYPTED_BACKUP_RESULT=PASS" \
  "DR_BACKUP_MODE=$MODE" \
  "DR_BACKUP_ENCRYPTION=CMS-AES-256-GCM" \
  "DR_BACKUP_KEY_ID=$key_id" \
  "DR_BACKUP_DESTINATION_URI=$destination_uri" \
  "DR_BACKUP_OBJECT=$remote_object" \
  "DR_BACKUP_MANIFEST=$remote_manifest" \
  "DR_BACKUP_UPLOAD_PERFORMED=$upload_performed" \
  "DR_BACKUP_RAW_SQL_BYTES=$raw_dump_bytes" \
  "DR_BACKUP_CIPHER_BYTES=$cipher_bytes" \
  "DR_BACKUP_CIPHER_SHA256=$cipher_sha256" \
  "DR_BACKUP_COMPRESSED_SQL_SHA256=$dump_sha256" \
  "DR_BACKUP_READBACK_SHA256=$readback_sha256" \
  "DR_BACKUP_RESTORED_COMPRESSED_SQL_SHA256=$restored_dump_sha256" \
  "DR_BACKUP_DECRYPT_VERIFIED=$decrypt_verified" \
  "DR_BACKUP_TAMPER_REJECTED=$tamper_rejected" \
  "DR_BACKUP_RESTORED_FACT_COUNT=$restored_fact_count" \
  "DR_BACKUP_FULL_RESTORE_VERIFIED=$full_restore_verified" \
  "DR_BACKUP_RESTORE_CLEANUP_VERIFIED=$restore_cleanup_verified" \
  "DR_BACKUP_RESTORED_TABLE_COUNT=$restored_table_count" \
  "DR_BACKUP_SOURCE_TABLE_COUNT=$source_table_count" \
  "DR_BACKUP_RESTORED_CHECK_OK_COUNT=$restored_check_ok_count" \
  "DR_BACKUP_RESTORED_TOTAL_ROWS=$restored_total_rows" \
  "DR_BACKUP_RESTORED_COUNTS_SHA256=$restored_counts_sha256" \
  "DR_BACKUP_RESTORED_SCHEMA_MIGRATIONS=$restored_schema_migrations" \
  "DR_BACKUP_RESTORED_TRADE_ORDERS=$restored_trade_orders" \
  "DR_BACKUP_RESTORED_TRADE_FILLS=$restored_trade_fills" \
  "DR_BACKUP_RESTORED_CONTRACT_POSITIONS=$restored_contract_positions" \
  "DR_BACKUP_RESTORED_POSITION_HISTORY=$restored_position_history" \
  "DR_BACKUP_RESTORED_SETTLEMENT_INSTRUCTIONS=$restored_settlement_instructions" \
  "DR_BACKUP_RESTORED_TRADE_EVENTS=$restored_trade_events" \
  "DR_BACKUP_RESTORED_TRADE_EVENT_INBOX=$restored_trade_event_inbox" \
  "DR_BACKUP_RESTORED_FUNDING_BATCHES=$restored_funding_batches" \
  "DR_BACKUP_RESTORED_DELIVERY_BATCHES=$restored_delivery_batches" \
  "DR_BACKUP_RESTORED_RECONCILIATION_ISSUES=$restored_reconciliation_issues" \
  "DR_BACKUP_RESTORED_AUTHORITATIVE_SNAPSHOTS=$restored_authoritative_snapshots" \
  "DR_BACKUP_RESTORED_SNAPSHOT_OUTBOX=$restored_snapshot_outbox" \
  "DR_BACKUP_PITR_SNAPSHOT_FILE=$pitr_snapshot_file" \
  "DR_BACKUP_PITR_SNAPSHOT_POSITION=$pitr_snapshot_position" \
  "DR_BACKUP_PITR_STOP_FILE=$pitr_stop_file" \
  "DR_BACKUP_PITR_STOP_POSITION=$pitr_stop_position" \
  "DR_BACKUP_PITR_END_FILE=$pitr_end_file" \
  "DR_BACKUP_PITR_END_POSITION=$pitr_end_position" \
  "DR_BACKUP_PITR_BINLOG_FILE_COUNT=$pitr_binlog_file_count" \
  "DR_BACKUP_PITR_TAIL_BYTES=$pitr_tail_bytes" \
  "DR_BACKUP_PITR_TAIL_SHA256=$pitr_tail_sha256" \
  "DR_BACKUP_PITR_BASELINE_VERIFIED=$pitr_baseline_verified" \
  "DR_BACKUP_PITR_RECOVERY_POINT_VERIFIED=$pitr_recovery_point_verified" \
  "DR_BACKUP_PITR_BOUNDARY_VERIFIED=$pitr_boundary_verified" \
  "DR_BACKUP_PITR_SOURCE_CLEANUP_VERIFIED=$pitr_source_cleanup_verified" \
  "DR_BACKUP_PITR_BUSINESS_TABLE_COUNT=$pitr_business_table_count" \
  "DR_BACKUP_DURATION_SECONDS=$duration_seconds"
