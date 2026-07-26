CREATE TABLE IF NOT EXISTS `t_option_market_snapshot_inbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `snapshot_id` VARCHAR(64) NOT NULL COMMENT '权威行情快照ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_snapshot_contract` (`snapshot_id`, `contract_id`),
  KEY `idx_option_snapshot_inbox_created` (`create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权权威行情逐合约消费幂等表';
