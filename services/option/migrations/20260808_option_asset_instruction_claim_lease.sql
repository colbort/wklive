ALTER TABLE `t_option_asset_instruction`
  ADD COLUMN `claimed_by` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '当前领取实例' AFTER `next_retry_at`,
  ADD COLUMN `claimed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '领取时间（秒）' AFTER `claimed_by`,
  ADD KEY `idx_instruction_claim` (`status`, `claimed_at`, `id`);
