#!/bin/sh
set -eu

IMAGE_NAME="${BEANSTALK_SMOKE_IMAGE:-wklive/beanstalkd:1.13-alpine3.20}"
smoke_suffix="$$"
container_name="wklive-beanstalk-restart-smoke-$smoke_suffix"
volume_name="wklive-beanstalk-restart-smoke-$smoke_suffix"
job_body="wklive-beanstalk-wal-$smoke_suffix"

cleanup() {
  docker rm -f "$container_name" >/dev/null 2>&1 || true
  docker volume rm "$volume_name" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

wait_for_protocol() {
  host_port="$1"
  attempts=0
  while [ "$attempts" -lt 30 ]; do
    if printf 'stats\r\nquit\r\n' | nc -w 2 127.0.0.1 "$host_port" 2>/dev/null |
       grep -q '^OK '; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  return 1
}

start_smoke_container() {
  docker run -d --name "$container_name" \
    -v "$volume_name:/var/lib/beanstalkd" \
    -p 127.0.0.1::11300 \
    "$IMAGE_NAME" >/dev/null
  docker port "$container_name" 11300/tcp | awk -F: 'NR==1 {print $NF}'
}

docker image inspect "$IMAGE_NAME" >/dev/null
docker volume create "$volume_name" >/dev/null

first_port=$(start_smoke_container)
if ! wait_for_protocol "$first_port"; then
  echo "FAIL  first Beanstalkd instance did not become ready" >&2
  exit 1
fi

body_bytes=$(printf '%s' "$job_body" | wc -c | tr -d ' ')
put_output=$(
  printf 'use wklive-readiness\r\nput 0 0 60 %s\r\n%s\r\nquit\r\n' \
    "$body_bytes" "$job_body" |
    nc -w 3 127.0.0.1 "$first_port"
)
job_id=$(printf '%s\n' "$put_output" | awk '/^INSERTED / {gsub("\\r", "", $2); print $2; exit}')
case "$job_id" in
  ''|*[!0-9]*)
    echo "FAIL  durable smoke job was not accepted: $put_output" >&2
    exit 1
    ;;
esac
printf 'PASS  durable smoke job %s accepted\n' "$job_id"

# Remove the entire container, not merely restart the process. Only the named
# WAL volume survives, which proves Compose-style recreation durability.
docker rm -f "$container_name" >/dev/null

second_port=$(start_smoke_container)
if ! wait_for_protocol "$second_port"; then
  echo "FAIL  recreated Beanstalkd instance did not become ready" >&2
  exit 1
fi

reserve_output=$(
  printf 'watch wklive-readiness\r\nignore default\r\nreserve-with-timeout 3\r\nquit\r\n' |
    nc -w 5 127.0.0.1 "$second_port"
)
reserved_id=$(printf '%s\n' "$reserve_output" | awk '/^RESERVED / {gsub("\\r", "", $2); print $2; exit}')
reserved_body=$(printf '%s\n' "$reserve_output" | tr -d '\r' | awk -v body="$job_body" '$0==body {print; exit}')
if [ "$reserved_id" != "$job_id" ] || [ "$reserved_body" != "$job_body" ]; then
  echo "FAIL  recreated instance did not recover the durable job: $reserve_output" >&2
  exit 1
fi

printf 'delete %s\r\nquit\r\n' "$reserved_id" |
  nc -w 3 127.0.0.1 "$second_port" >/dev/null
printf 'PASS  recreated instance recovered job %s from WAL\n' "$reserved_id"
printf '\nREADY: Beanstalkd container-recreation persistence smoke passed.\n'
