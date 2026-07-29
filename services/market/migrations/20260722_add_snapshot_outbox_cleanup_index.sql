-- Run this migration before enabling the snapshot outbox cleanup worker.
-- On a large production table, use the database's approved online-DDL process
-- and monitor replication lag and free disk space while the index is built.
ALTER TABLE `t_market_snapshot_outbox`
  ADD INDEX `idx_snapshot_outbox_cleanup` (`status`, `update_times`, `id`);
