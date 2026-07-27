CREATE TABLE `t_liquidity_provider_provision` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `provider_code` VARCHAR(64) NOT NULL,
  `request_hash` CHAR(64) NOT NULL,
  `trade_user_id` BIGINT NOT NULL DEFAULT 0,
  `step` TINYINT NOT NULL DEFAULT 1 COMMENT '1待处理 2账户已创建 3资金已初始化 4已完成 5失败待重试',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_provider_code` (`provider_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='内部做市提供方开通状态';
