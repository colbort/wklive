CREATE TABLE IF NOT EXISTS `t_contract_adl_execution` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `execution_no` VARCHAR(96) NOT NULL,
  `liquidation_id` BIGINT NOT NULL,
  `liquidation_no` VARCHAR(64) NOT NULL,
  `position_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `position_version` BIGINT NOT NULL,
  `qty` DECIMAL(36,18) NOT NULL,
  `position_margin_release` DECIMAL(36,18) NOT NULL,
  `isolated_margin_release` DECIMAL(36,18) NOT NULL,
  `realized_pnl` DECIMAL(36,18) NOT NULL,
  `asset_credit` DECIMAL(36,18) NOT NULL,
  `asset` VARCHAR(32) NOT NULL,
  `bankruptcy_price` DECIMAL(36,18) NOT NULL,
  `relief_amount` DECIMAL(36,18) NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 PREPARED 2 ASSET_DONE 3 POSITION_DONE 4 FAILED 5 MANUAL',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_execution` (`tenant_id`,`execution_no`),
  UNIQUE KEY `uk_liquidation_position` (`tenant_id`,`liquidation_id`,`position_id`),
  KEY `idx_adl_recovery` (`tenant_id`,`status`,`update_times`),
  CONSTRAINT `chk_adl_execution` CHECK (`qty` > 0 AND `position_margin_release` >= 0 AND `isolated_margin_release` >= 0 AND `asset_credit` >= 0 AND `relief_amount` > 0 AND `status` IN (1,2,3,4,5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE `t_trade_settlement_instruction`
  ADD KEY `idx_biz_id_status` (`tenant_id`,`biz_type`,`biz_id`,`status`,`step_no`);

-- Recover pre-Saga delivery rows. Biz numbers intentionally match the legacy
-- direct RPC suffixes, so Asset idempotency also covers partially executed rows.
INSERT IGNORE INTO `t_trade_settlement_instruction`
(`tenant_id`,`instruction_no`,`biz_type`,`biz_id`,`batch_no`,`fill_id`,`order_id`,`position_id`,`reservation_no`,`user_id`,`action`,`asset`,`amount`,`step_no`,`status`,`retry_count`,`next_retry_at`,`last_error_msg`,`create_times`,`update_times`)
SELECT s.tenant_id,CONCAT(s.settlement_no,'-',x.suffix),'delivery',s.settlement_no,s.batch_no,0,0,s.position_id,'',s.user_id,x.action,s.settle_asset,
       CASE x.suffix WHEN 'LOSS' THEN GREATEST(-s.realized_pnl,0) WHEN 'FEE' THEN s.delivery_fee WHEN 'MARGIN' THEN p.position_margin+p.isolated_margin ELSE GREATEST(s.realized_pnl,0) END,
       x.step_no,1,0,s.next_retry_at,'',s.create_times,s.update_times
FROM t_contract_delivery_settlement s
JOIN t_contract_position p ON p.id=s.position_id AND p.tenant_id=s.tenant_id
CROSS JOIN (
  SELECT 'LOSS' suffix,8 action,1 step_no UNION ALL SELECT 'FEE',8,1 UNION ALL
  SELECT 'MARGIN',3,2 UNION ALL SELECT 'PROFIT',3,2
) x
WHERE s.status IN (1,3)
AND (CASE x.suffix WHEN 'LOSS' THEN GREATEST(-s.realized_pnl,0) WHEN 'FEE' THEN s.delivery_fee WHEN 'MARGIN' THEN p.position_margin+p.isolated_margin ELSE GREATEST(s.realized_pnl,0) END)>0;
