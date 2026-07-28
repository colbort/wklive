-- 对账扫描可以高频累计 occurrence_count，但同一未变化差异不应每轮刷告警。
ALTER TABLE `t_contract_reconciliation_issue`
  ADD COLUMN `last_alert_at` BIGINT NOT NULL DEFAULT 0
    COMMENT '最近一次输出或发送告警的时间' AFTER `last_seen_at`;
