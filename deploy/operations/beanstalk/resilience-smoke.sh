#!/bin/sh
set -u

image_name=${BEANSTALK_RESILIENCE_IMAGE:-wklive/beanstalkd:1.13-alpine3.20}
job_count=${BEANSTALK_RESILIENCE_JOBS:-1000}
minimum_put_rate=${BEANSTALK_RESILIENCE_MIN_PUT_RATE:-50}
maximum_recovery_seconds=${BEANSTALK_RESILIENCE_MAX_RECOVERY_SECONDS:-15}
maximum_disconnect_seconds=${BEANSTALK_RESILIENCE_MAX_DISCONNECT_SECONDS:-10}
resource_suffix=$$
container_name="wklive-beanstalk-resilience-${resource_suffix}"
volume_name="wklive-beanstalk-resilience-${resource_suffix}"
tube_name="wklive-resilience-${resource_suffix}"
probe_tube="wklive-resilience-disconnect-${resource_suffix}"
job_prefix="wklive-resilience-${resource_suffix}-job"
temporary_dir=$(mktemp -d "${TMPDIR:-/tmp}/wklive-beanstalk-resilience.XXXXXX") || exit 1
put_protocol="$temporary_dir/put.protocol"
put_response="$temporary_dir/put.response"
stats_response="$temporary_dir/stats.response"
reserve_protocol="$temporary_dir/reserve.protocol"
reserve_response="$temporary_dir/reserve.response"
disconnect_response="$temporary_dir/disconnect.response"
disconnect_protocol="$temporary_dir/disconnect.protocol"
disconnect_pid=
disconnect_writer_open=false
host_port=

fail() {
  printf 'FAIL  %s\n' "$1" >&2
  exit 1
}

pass() {
  printf 'PASS  %s\n' "$1"
}

require_positive_integer() {
  value=$1
  name=$2
  case "$value" in
    ''|*[!0-9]*|0)
      fail "$name must be a positive integer"
      ;;
  esac
}

cleanup() {
  if [ "$disconnect_writer_open" = true ]; then
    exec 3>&-
    disconnect_writer_open=false
  fi
  if [ -n "$disconnect_pid" ] && kill -0 "$disconnect_pid" 2>/dev/null; then
    kill "$disconnect_pid" 2>/dev/null || true
    wait "$disconnect_pid" 2>/dev/null || true
  fi
  docker rm -f "$container_name" >/dev/null 2>&1 || true
  docker volume rm -f "$volume_name" >/dev/null 2>&1 || true
  rm -f "$put_protocol" "$put_response" "$stats_response" \
    "$reserve_protocol" "$reserve_response" "$disconnect_response" "$disconnect_protocol"
  rmdir "$temporary_dir" 2>/dev/null || true
}

trap cleanup EXIT INT TERM

require_positive_integer "$job_count" BEANSTALK_RESILIENCE_JOBS
require_positive_integer "$minimum_put_rate" BEANSTALK_RESILIENCE_MIN_PUT_RATE
require_positive_integer "$maximum_recovery_seconds" BEANSTALK_RESILIENCE_MAX_RECOVERY_SECONDS
require_positive_integer "$maximum_disconnect_seconds" BEANSTALK_RESILIENCE_MAX_DISCONNECT_SECONDS
if [ "$job_count" -gt 50000 ]; then
  fail "BEANSTALK_RESILIENCE_JOBS must not exceed 50000"
fi

docker image inspect "$image_name" >/dev/null 2>&1 ||
  fail "Beanstalkd image is unavailable: $image_name"
command -v nc >/dev/null 2>&1 || fail "nc is required"
docker volume create "$volume_name" >/dev/null || fail "cannot create isolated WAL volume"

start_container() {
  docker run -d --name "$container_name" \
    -p 127.0.0.1::11300 \
    -v "$volume_name:/var/lib/beanstalkd" \
    "$image_name" >/dev/null || fail "cannot start isolated Beanstalkd container"
  host_port=$(docker port "$container_name" 11300/tcp 2>/dev/null | awk -F: 'NR == 1 {print $NF}')
  case "$host_port" in
    ''|*[!0-9]*) fail "cannot resolve isolated Beanstalkd host port" ;;
  esac
}

wait_until_ready() {
  attempts=0
  while [ "$attempts" -lt 60 ]; do
    if printf 'stats\r\nquit\r\n' | nc -w 1 127.0.0.1 "$host_port" 2>/dev/null |
      grep -q '^OK '; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  return 1
}

ready_job_count() {
  printf 'stats-tube %s\r\nquit\r\n' "$tube_name" |
    nc -w 5 127.0.0.1 "$host_port" >"$stats_response" 2>/dev/null || return 1
  tr -d '\r' <"$stats_response" |
    awk '$1 == "current-jobs-ready:" {print $2; exit}'
}

printf 'use %s\r\n' "$tube_name" >"$put_protocol"
job_number=1
while [ "$job_number" -le "$job_count" ]; do
  job_body="${job_prefix}-${job_number}"
  body_size=$(printf '%s' "$job_body" | wc -c | tr -d ' ')
  printf 'put 0 0 3600 %s\r\n%s\r\n' "$body_size" "$job_body" >>"$put_protocol"
  job_number=$((job_number + 1))
done
printf 'quit\r\n' >>"$put_protocol"

printf 'watch %s\r\nignore default\r\n' "$tube_name" >"$reserve_protocol"
job_number=1
while [ "$job_number" -le "$job_count" ]; do
  printf 'reserve-with-timeout 0\r\n' >>"$reserve_protocol"
  job_number=$((job_number + 1))
done
printf 'quit\r\n' >>"$reserve_protocol"

start_container
wait_until_ready || fail "isolated Beanstalkd did not become protocol-ready"
pass "isolated Beanstalkd became protocol-ready"

put_started=$(date +%s)
nc -w 60 127.0.0.1 "$host_port" <"$put_protocol" >"$put_response" 2>/dev/null ||
  fail "bulk put connection failed"
put_finished=$(date +%s)
inserted_count=$(tr -d '\r' <"$put_response" |
  awk '$1 == "INSERTED" {count++} END {print count + 0}')
if [ "$inserted_count" -ne "$job_count" ]; then
  fail "inserted $inserted_count of $job_count jobs"
fi
put_elapsed=$((put_finished - put_started))
if [ "$put_elapsed" -lt 1 ]; then
  put_elapsed=1
fi
put_rate=$((job_count / put_elapsed))
if [ "$put_rate" -lt "$minimum_put_rate" ]; then
  fail "put rate ${put_rate} jobs/s is below ${minimum_put_rate} jobs/s"
fi
pass "inserted $job_count jobs at ${put_rate} jobs/s (minimum ${minimum_put_rate})"

ready_before_kill=$(ready_job_count) || fail "cannot read workload tube before kill"
if [ "$ready_before_kill" != "$job_count" ]; then
  fail "workload tube has $ready_before_kill ready jobs before kill; expected $job_count"
fi
pass "workload tube reports current-jobs-ready=$job_count before kill"

mkfifo "$disconnect_protocol" || fail "cannot create disconnect probe pipe"
nc -w 35 127.0.0.1 "$host_port" \
  <"$disconnect_protocol" >"$disconnect_response" 2>/dev/null &
disconnect_pid=$!
exec 3>"$disconnect_protocol"
disconnect_writer_open=true
printf 'watch %s\r\nignore default\r\nreserve-with-timeout 30\r\n' "$probe_tube" >&3
sleep 1
if ! kill -0 "$disconnect_pid" 2>/dev/null; then
  fail "disconnect probe did not hold an active connection"
fi

recovery_started=$(date +%s)
docker kill --signal KILL "$container_name" >/dev/null || fail "cannot SIGKILL isolated Beanstalkd"
disconnect_waited=0
while kill -0 "$disconnect_pid" 2>/dev/null &&
  [ "$disconnect_waited" -lt "$maximum_disconnect_seconds" ]; do
  sleep 1
  disconnect_waited=$((disconnect_waited + 1))
done
if kill -0 "$disconnect_pid" 2>/dev/null; then
  fail "client connection did not observe SIGKILL within ${maximum_disconnect_seconds}s"
fi
exec 3>&-
disconnect_writer_open=false
wait "$disconnect_pid" 2>/dev/null || true
disconnect_pid=
pass "active client connection observed the hard disconnect in ${disconnect_waited}s"

docker rm "$container_name" >/dev/null || fail "cannot remove killed isolated container"
start_container
wait_until_ready || fail "Beanstalkd did not recover from WAL after SIGKILL"
recovery_finished=$(date +%s)
recovery_elapsed=$((recovery_finished - recovery_started))
if [ "$recovery_elapsed" -gt "$maximum_recovery_seconds" ]; then
  fail "WAL recovery took ${recovery_elapsed}s; maximum is ${maximum_recovery_seconds}s"
fi
pass "protocol recovered from WAL after SIGKILL in ${recovery_elapsed}s"

ready_after_restart=$(ready_job_count) || fail "cannot read workload tube after recovery"
if [ "$ready_after_restart" != "$job_count" ]; then
  fail "recovered tube has $ready_after_restart ready jobs; expected $job_count"
fi
pass "recovered workload preserves current-jobs-ready=$job_count"

nc -w 60 127.0.0.1 "$host_port" <"$reserve_protocol" >"$reserve_response" 2>/dev/null ||
  fail "new consumer connection failed after recovery"
reserved_count=$(tr -d '\r' <"$reserve_response" |
  awk '$1 == "RESERVED" {count++} END {print count + 0}')
if [ "$reserved_count" -ne "$job_count" ]; then
  fail "new connection reserved $reserved_count of $job_count recovered jobs"
fi
first_body="${job_prefix}-1"
last_body="${job_prefix}-${job_count}"
tr -d '\r' <"$reserve_response" | grep -Fxq "$first_body" ||
  fail "first recovered job body is missing"
tr -d '\r' <"$reserve_response" | grep -Fxq "$last_body" ||
  fail "last recovered job body is missing"
pass "new connection consumed all $job_count recovered jobs with intact boundary bodies"

printf '\nREADY: isolated Beanstalkd resilience passed '
printf '(jobs=%s put_rate=%s/s recovery=%ss disconnect=%ss).\n' \
  "$job_count" "$put_rate" "$recovery_elapsed" "$disconnect_waited"
