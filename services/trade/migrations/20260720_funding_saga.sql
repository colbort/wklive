CREATE TABLE IF NOT EXISTS `t_contract_funding_difference_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `settle_asset` VARCHAR(32) NOT NULL,
  `fund_user_id` BIGINT NOT NULL,
  `wallet_type` TINYINT NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `version` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_asset` (`tenant_id`,`settle_asset`),
  CONSTRAINT `chk_funding_difference_account` CHECK (`tenant_id` > 0 AND `fund_user_id` > 0 AND `wallet_type` > 0 AND `status` IN (1,2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE `t_trade_settlement_instruction`
  ADD KEY `idx_biz_pending` (`tenant_id`,`biz_type`,`batch_no`,`status`,`next_retry_at`,`step_no`);

-- Backfill batches created by the pre-Saga worker. That worker rejected every
-- non-zero batch difference, so these historical batches need user steps only.
INSERT INTO `t_trade_settlement_instruction`
(`tenant_id`,`instruction_no`,`biz_type`,`biz_id`,`batch_no`,`fill_id`,`order_id`,`position_id`,`reservation_no`,`user_id`,`action`,`asset`,`amount`,`step_no`,`status`,`retry_count`,`next_retry_at`,`last_error_msg`,`create_times`,`update_times`)
SELECT s.`tenant_id`,CONCAT(s.`settlement_no`,'-ASSET'),'funding',s.`settlement_no`,s.`batch_no`,0,0,s.`position_id`,'',s.`user_id`,
       IF(s.`fee_amount` < 0,8,3),s.`fee_asset`,ABS(s.`fee_amount`),IF(s.`fee_amount` < 0,1,2),
       IF(s.`status`=4,5,IF(s.`status`=3,4,1)),s.`retry_count`,s.`next_retry_at`,s.`last_error_msg`,s.`create_times`,s.`update_times`
FROM `t_contract_funding_settlement` s
LEFT JOIN `t_trade_settlement_instruction` i
  ON i.`tenant_id`=s.`tenant_id` AND i.`instruction_no`=CONCAT(s.`settlement_no`,'-ASSET')
WHERE s.`status` IN (1,3,4) AND s.`fee_amount` <> 0 AND i.`id` IS NULL;
