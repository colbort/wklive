#!/bin/sh
set -u

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
DOCKERFILE="$SCRIPT_DIR/Dockerfile.beanstalkd"
RESTART_SMOKE="$SCRIPT_DIR/beanstalk-restart-smoke.sh"
RESILIENCE_SMOKE="$SCRIPT_DIR/beanstalk-resilience-smoke.sh"
failures=0
repository_only=false

if [ "${1:-}" = "--repository-only" ]; then
  repository_only=true
elif [ "$#" -gt 0 ]; then
  echo "usage: $0 [--repository-only]" >&2
  exit 2
fi

pass() {
  printf 'PASS  %s\n' "$1"
}

fail() {
  printf 'FAIL  %s\n' "$1"
  failures=$((failures + 1))
}

require_contains() {
  file_name="$1"
  pattern="$2"
  description="$3"
  if grep -Eq "$pattern" "$file_name"; then
    pass "$description"
  else
    fail "$description"
  fi
}

require_absent() {
  file_name="$1"
  pattern="$2"
  description="$3"
  if grep -Eq "$pattern" "$file_name"; then
    fail "$description"
  else
    pass "$description"
  fi
}

require_contains "$DOCKERFILE" \
  '^ARG ALPINE_IMAGE=alpine:3[.]20@sha256:[0-9a-f]{64}$' \
  "Beanstalkd base uses a pinned multi-architecture image digest"
require_contains "$DOCKERFILE" \
  'apk add --no-cache beanstalkd=1[.]13-r0' \
  "Beanstalkd package version is pinned"
require_contains "$DOCKERFILE" \
  '"-b", "/var/lib/beanstalkd", "-f", "0"' \
  "Beanstalkd enables WAL and fsyncs every accepted job"
require_contains "$COMPOSE_FILE" \
  'dockerfile: Dockerfile[.]beanstalkd' \
  "Compose builds the repository-owned Beanstalkd image"
require_contains "$COMPOSE_FILE" \
  "printf 'stats\\\\r\\\\nquit\\\\r\\\\n'.*grep -q '\^OK '" \
  "Compose checks the Beanstalkd protocol rather than only the TCP port"
require_absent "$COMPOSE_FILE" \
  'schickling/beanstalkd|platform:[[:space:]]*linux/(amd64|arm64)' \
  "Compose does not force an emulated or host-specific Beanstalkd platform"
require_contains "$COMPOSE_FILE" \
  'beanstalk-primary-data:/var/lib/beanstalkd' \
  "Primary Beanstalkd uses a durable dedicated volume"
require_contains "$COMPOSE_FILE" \
  'beanstalk-secondary-data:/var/lib/beanstalkd' \
  "Secondary Beanstalkd uses a durable dedicated volume"
require_contains "$RESTART_SMOKE" \
  'docker rm -f "[$]container_name"' \
  "Beanstalkd acceptance removes and recreates an isolated container"
require_contains "$RESTART_SMOKE" \
  'recovered job %s from WAL' \
  "Beanstalkd acceptance verifies WAL recovery"
if [ -x "$RESILIENCE_SMOKE" ]; then
  pass "Beanstalkd isolated resilience gate is executable"
else
  fail "Beanstalkd isolated resilience gate is executable"
fi
require_contains "$RESILIENCE_SMOKE" \
  'docker kill --signal KILL "[$]container_name"' \
  "Beanstalkd resilience gate injects a hard process failure"
require_contains "$RESILIENCE_SMOKE" \
  'current-jobs-ready' \
  "Beanstalkd resilience gate verifies the full WAL backlog"
require_contains "$RESILIENCE_SMOKE" \
  'BEANSTALK_RESILIENCE_JOBS' \
  "Beanstalkd resilience gate exposes an explicit capacity workload"
require_contains "$RESILIENCE_SMOKE" \
  'new connection consumed all' \
  "Beanstalkd resilience gate proves post-recovery reconnection and consumption"

if [ "$repository_only" = true ]; then
  if [ "$failures" -eq 0 ]; then
    printf '\nREADY: repository-owned Beanstalkd artifacts passed.\n'
    exit 0
  fi
  printf '\nNOT READY: %d repository check(s) failed.\n' "$failures" >&2
  exit 1
fi

case "$(uname -m)" in
  arm64|aarch64)
    host_arch=arm64
    ;;
  x86_64|amd64)
    host_arch=amd64
    ;;
  *)
    fail "Host architecture is supported ($(uname -m))"
    host_arch=unknown
    ;;
esac

for service_name in beanstalk-primary beanstalk-secondary; do
  container_id=$(docker compose -f "$COMPOSE_FILE" ps -q "$service_name" 2>/dev/null)
  if [ -z "$container_id" ]; then
    fail "$service_name is running"
    continue
  fi

  health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$container_id" 2>/dev/null)
  if [ "$health" = "healthy" ]; then
    pass "$service_name protocol health check is healthy"
  else
    fail "$service_name protocol health check is healthy (status=${health:-unknown})"
  fi

  image_id=$(docker inspect --format '{{.Image}}' "$container_id" 2>/dev/null)
  image_arch=$(docker image inspect --format '{{.Architecture}}' "$image_id" 2>/dev/null)
  if [ "$host_arch" != unknown ] && [ "$image_arch" = "$host_arch" ]; then
    pass "$service_name image architecture matches host ($host_arch)"
  else
    fail "$service_name image architecture matches host (host=$host_arch image=${image_arch:-unknown})"
  fi

  wal_mount=$(docker inspect --format '{{range .Mounts}}{{if eq .Destination "/var/lib/beanstalkd"}}{{.Type}}{{end}}{{end}}' "$container_id" 2>/dev/null)
  if [ "$wal_mount" = "volume" ]; then
    pass "$service_name WAL directory uses a Docker volume"
  else
    fail "$service_name WAL directory uses a Docker volume (type=${wal_mount:-missing})"
  fi
done

if [ "$failures" -eq 0 ]; then
  printf '\nREADY: native Beanstalkd runtime passed.\n'
  exit 0
fi
printf '\nNOT READY: %d check(s) failed.\n' "$failures" >&2
exit 1
