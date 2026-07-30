CREATE TABLE IF NOT EXISTS `t_option_outbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `event_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '事件幂等号',
  `event_type` TINYINT NOT NULL DEFAULT 0 COMMENT '事件类型：1成交持仓',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `match_sequence` BIGINT NOT NULL DEFAULT 0 COMMENT '合约内撮合序号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2处理中 3成功 4失败 5人工处理',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_event_no` (`tenant_id`, `event_no`),
  UNIQUE KEY `uk_tenant_contract_sequence_type` (`tenant_id`, `contract_id`, `match_sequence`, `event_type`),
  KEY `idx_outbox_retry` (`status`, `next_retry_at`, `id`),
  CONSTRAINT `chk_option_outbox` CHECK (
    `event_type` IN (1) AND `match_sequence` > 0
    AND `status` IN (1,2,3,4,5) AND `retry_count` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权事务事件发件箱';

CREATE TABLE IF NOT EXISTS `t_option_inbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `event_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '事件幂等号',
  `event_type` TINYINT NOT NULL DEFAULT 0 COMMENT '事件类型：1成交持仓',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `match_sequence` BIGINT NOT NULL DEFAULT 0 COMMENT '合约内撮合序号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1处理中 2成功 3失败',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_event_no` (`tenant_id`, `event_no`),
  KEY `idx_inbox_contract_sequence` (`tenant_id`, `contract_id`, `match_sequence`),
  CONSTRAINT `chk_option_inbox` CHECK (
    `event_type` IN (1) AND `match_sequence` > 0 AND `status` IN (1,2,3)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权事件消费幂等箱';
