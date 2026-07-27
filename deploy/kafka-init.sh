#!/bin/bash
set -euo pipefail

KAFKA_BOOTSTRAP_SERVER="${KAFKA_BOOTSTRAP_SERVER:-kafka:9092}"
KAFKA_PARTITIONS="${KAFKA_PARTITIONS:-1}"
KAFKA_REPLICATION_FACTOR="${KAFKA_REPLICATION_FACTOR:-1}"
KAFKA_TOPICS_BIN="${KAFKA_TOPICS_BIN:-/opt/kafka/bin/kafka-topics.sh}"

topics=(
  system.scheduled-tasks
  admin.notifications
  trade.domain-events
  user.business-events
  market.authoritative-snapshot.v1
  wklive.chat.app.events
  wklive.chat.admin.events
  system.scheduled-tasks.dlq
  trade.domain-events.dlq
  user.business-events.dlq
  market.authoritative-snapshot.v1.dlq
  wklive.chat.app.events.dlq
  wklive.chat.admin.events.dlq
)

for topic in "${topics[@]}"; do
  echo "creating Kafka topic: ${topic}"
  "${KAFKA_TOPICS_BIN}" \
    --bootstrap-server "${KAFKA_BOOTSTRAP_SERVER}" \
    --create \
    --if-not-exists \
    --topic "${topic}" \
    --partitions "${KAFKA_PARTITIONS}" \
    --replication-factor "${KAFKA_REPLICATION_FACTOR}"
done

echo "Kafka topics are ready"
