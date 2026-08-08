ALTER TABLE `t_pay_outbox`
  ADD COLUMN `claimed_by` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '当前处理实例的唯一领取标识' AFTER `status`,
  ADD COLUMN `claimed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '领取时间；超过租约可由其他实例恢复' AFTER `claimed_by`,
  ADD KEY `idx_status_claim` (`status`, `claimed_at`);
