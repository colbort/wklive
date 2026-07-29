-- 覆盖 Snapshot Outbox 状态计数及最老未完成任务查询，避免读取大体积 payload。
ALTER TABLE `t_market_snapshot_outbox`
  ADD INDEX `idx_snapshot_outbox_health` (`status`, `create_times`);
