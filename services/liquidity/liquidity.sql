-- Liquidity service schema (MySQL 8.0+)
--
-- Design conventions:
-- 1. All timestamps are Unix milliseconds (BIGINT).
-- 2. Price, quantity and notional columns use natural trading units, not asset minor units.
-- 3. Cross-service identifiers (trade symbol/order/user) are not protected by foreign keys.
-- 4. API secrets are never stored here; credential_ref points to an external secret manager.
-- 5. Every write path must use tenant_id and an idempotent business number.

SET NAMES utf8mb4;

DROP TABLE IF EXISTS `t_liquidity_event_outbox`;
DROP TABLE IF EXISTS `t_liquidity_event_inbox`;
DROP TABLE IF EXISTS `t_liquidity_reconcile_detail`;
DROP TABLE IF EXISTS `t_liquidity_reconcile_batch`;
DROP TABLE IF EXISTS `t_liquidity_risk_event`;
DROP TABLE IF EXISTS `t_liquidity_inventory_snapshot`;
DROP TABLE IF EXISTS `t_liquidity_hedge_task`;
DROP TABLE IF EXISTS `t_liquidity_external_fill`;
DROP TABLE IF EXISTS `t_liquidity_external_order`;
DROP TABLE IF EXISTS `t_liquidity_quote_order`;
DROP TABLE IF EXISTS `t_liquidity_quote_cycle`;
DROP TABLE IF EXISTS `t_liquidity_strategy_level`;
DROP TABLE IF EXISTS `t_liquidity_symbol_config`;
DROP TABLE IF EXISTS `t_liquidity_provider`;

CREATE TABLE `t_liquidity_provider` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `provider_code` VARCHAR(64) NOT NULL COMMENT '租户内唯一提供方编码',
  `provider_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '提供方名称',
  `provider_type` TINYINT NOT NULL COMMENT '类型：1平台内部做市 2外部流动性',
  `trade_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '内部做市用户ID；外部提供方为0',
  `venue_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '外部场所编码，如BINANCE；内部为空',
  `environment` TINYINT NOT NULL DEFAULT 1 COMMENT '环境：1生产 2沙箱',
  `credential_ref` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密钥管理系统引用，禁止保存明文密钥',
  `account_ref` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '外部账户或子账户标识',
  `rate_limit_per_second` INT NOT NULL DEFAULT 10 COMMENT '外部接口每秒请求上限',
  `status` TINYINT NOT NULL DEFAULT 2 COMMENT '状态：1启用 2停用',
  `last_health_status` TINYINT NOT NULL DEFAULT 0 COMMENT '健康状态：0未知 1正常 2异常',
  `last_health_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近健康检查时间',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_provider_code` (`tenant_id`, `provider_code`),
  KEY `idx_tenant_type_status` (`tenant_id`, `provider_type`, `status`),
  KEY `idx_trade_user` (`tenant_id`, `trade_user_id`),
  CONSTRAINT `chk_liquidity_provider_type` CHECK (`provider_type` IN (1, 2)),
  CONSTRAINT `chk_liquidity_provider_environment` CHECK (`environment` IN (1, 2)),
  CONSTRAINT `chk_liquidity_provider_status` CHECK (`status` IN (1, 2)),
  CONSTRAINT `chk_liquidity_provider_health` CHECK (`last_health_status` IN (0, 1, 2)),
  CONSTRAINT `chk_liquidity_provider_target` CHECK (
    (`provider_type` = 1 AND `trade_user_id` > 0 AND `venue_code` = '' AND `credential_ref` = '')
    OR
    (`provider_type` = 2 AND `trade_user_id` = 0 AND `venue_code` <> '' AND `credential_ref` <> '')
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流动性提供方及账户';

CREATE TABLE `t_liquidity_symbol_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `symbol_id` BIGINT NOT NULL COMMENT 'trade.t_trade_symbol.id',
  `symbol` VARCHAR(64) NOT NULL COMMENT '内部交易标的代码快照',
  `product_type` TINYINT NOT NULL COMMENT '产品：1现货 2衍生品',
  `contract_type` TINYINT NOT NULL DEFAULT 0 COMMENT '合约类型：0不适用 1永续 2交割',
  `liquidity_mode` TINYINT NOT NULL COMMENT '模式：1内部做市 2外部路由 3内部做市并外部对冲',
  `internal_provider_id` BIGINT NOT NULL DEFAULT 0 COMMENT '内部做市提供方ID',
  `external_provider_id` BIGINT NOT NULL DEFAULT 0 COMMENT '外部流动性或对冲提供方ID',
  `external_symbol` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '外部场所交易对代码',
  `reference_price_source` VARCHAR(255) NOT NULL COMMENT '权威行情源，可配置多个逗号分隔源',
  `reference_price_kind` VARCHAR(32) NOT NULL DEFAULT 'FINAL_QUOTE' COMMENT '参考价类型',
  `quote_validity_ms` INT NOT NULL DEFAULT 3000 COMMENT '行情最大有效期',
  `refresh_interval_ms` INT NOT NULL DEFAULT 1000 COMMENT '最小重新报价间隔',
  `quote_ttl_ms` INT NOT NULL DEFAULT 5000 COMMENT '内部报价最大存活时间',
  `reprice_threshold_bps` DECIMAL(20,8) NOT NULL DEFAULT 1 COMMENT '重新报价阈值，基点',
  `base_spread_bps` DECIMAL(20,8) NOT NULL DEFAULT 10 COMMENT '基础单边点差，基点',
  `max_spread_bps` DECIMAL(20,8) NOT NULL DEFAULT 100 COMMENT '允许的最大单边点差',
  `max_price_deviation_bps` DECIMAL(20,8) NOT NULL DEFAULT 100 COMMENT '内部报价相对参考价最大偏离',
  `price_tick` DECIMAL(36,18) NOT NULL COMMENT '价格步长快照',
  `qty_step` DECIMAL(36,18) NOT NULL COMMENT '数量步长快照',
  `min_quote_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单档最小报价数量',
  `max_quote_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单档最大报价数量，0表示不限',
  `max_quote_notional` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单档最大名义金额，0表示不限',
  `target_base_inventory` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产目标库存',
  `min_base_inventory` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产最低库存',
  `max_base_inventory` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产最高库存，0表示不限',
  `max_net_exposure` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '最大净敞口，0表示不限',
  `max_daily_notional` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '每日最大成交名义金额，0表示不限',
  `inventory_skew_bps` DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT '库存偏移达到上限时的最大报价倾斜',
  `hedge_threshold` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发外部对冲的净敞口阈值',
  `hedge_ratio` DECIMAL(20,10) NOT NULL DEFAULT 1 COMMENT '目标对冲比例，0到1',
  `self_trade_prevention` TINYINT NOT NULL DEFAULT 1 COMMENT '自成交保护：1启用 2停用',
  `status` TINYINT NOT NULL DEFAULT 2 COMMENT '状态：1运行 2停用 3暂停 4熔断',
  `pause_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '暂停或熔断原因',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_symbol_product` (`tenant_id`, `symbol_id`, `product_type`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_internal_provider` (`tenant_id`, `internal_provider_id`, `status`),
  KEY `idx_external_provider` (`tenant_id`, `external_provider_id`, `status`),
  CONSTRAINT `chk_liquidity_symbol_product` CHECK (`product_type` IN (1, 2)),
  CONSTRAINT `chk_liquidity_contract_type` CHECK (
    (`product_type` = 1 AND `contract_type` = 0)
    OR
    (`product_type` = 2 AND `contract_type` IN (1, 2))
  ),
  CONSTRAINT `chk_liquidity_mode` CHECK (`liquidity_mode` IN (1, 2, 3)),
  CONSTRAINT `chk_liquidity_mode_provider` CHECK (
    (`liquidity_mode` = 1 AND `internal_provider_id` > 0)
    OR
    (`liquidity_mode` = 2 AND `external_provider_id` > 0)
    OR
    (`liquidity_mode` = 3 AND `internal_provider_id` > 0 AND `external_provider_id` > 0)
  ),
  CONSTRAINT `chk_liquidity_symbol_status` CHECK (`status` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_symbol_positive` CHECK (
    `quote_validity_ms` > 0 AND `refresh_interval_ms` > 0 AND `quote_ttl_ms` > 0
    AND `price_tick` > 0 AND `qty_step` > 0
    AND `base_spread_bps` >= 0 AND `max_spread_bps` >= `base_spread_bps`
    AND `max_price_deviation_bps` > 0
    AND `min_quote_qty` >= 0
    AND (`max_quote_qty` = 0 OR `max_quote_qty` >= `min_quote_qty`)
    AND `target_base_inventory` >= 0 AND `min_base_inventory` >= 0
    AND (`max_base_inventory` = 0 OR `max_base_inventory` >= `min_base_inventory`)
    AND `hedge_ratio` >= 0 AND `hedge_ratio` <= 1
  ),
  CONSTRAINT `chk_liquidity_symbol_stp` CHECK (`self_trade_prevention` IN (1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易对流动性与风控配置';

CREATE TABLE `t_liquidity_strategy_level` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `level_no` INT NOT NULL COMMENT '深度档位，从1开始',
  `bid_spread_bps` DECIMAL(20,8) NOT NULL COMMENT '买单相对参考价向下偏移基点',
  `ask_spread_bps` DECIMAL(20,8) NOT NULL COMMENT '卖单相对参考价向上偏移基点',
  `bid_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '买单基础资产数量',
  `ask_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '卖单基础资产数量',
  `enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '启用：1是 2否',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_config_level` (`tenant_id`, `config_id`, `level_no`),
  KEY `idx_config_enabled` (`tenant_id`, `config_id`, `enabled`),
  CONSTRAINT `chk_liquidity_level_no` CHECK (`level_no` > 0),
  CONSTRAINT `chk_liquidity_level_spread` CHECK (`bid_spread_bps` >= 0 AND `ask_spread_bps` >= 0),
  CONSTRAINT `chk_liquidity_level_qty` CHECK (`bid_qty` > 0 AND `ask_qty` > 0),
  CONSTRAINT `chk_liquidity_level_enabled` CHECK (`enabled` IN (1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='多档做市报价策略';

CREATE TABLE `t_liquidity_quote_cycle` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `cycle_no` VARCHAR(64) NOT NULL COMMENT '报价周期业务号',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `symbol_id` BIGINT NOT NULL COMMENT '内部交易标的ID',
  `reference_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '本轮参考价',
  `reference_source` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '实际选中的行情源',
  `reference_snapshot_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '权威行情快照ID',
  `reference_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行情源时间',
  `target_bid_count` INT NOT NULL DEFAULT 0 COMMENT '目标买单数',
  `target_ask_count` INT NOT NULL DEFAULT 0 COMMENT '目标卖单数',
  `placed_bid_count` INT NOT NULL DEFAULT 0 COMMENT '成功买单数',
  `placed_ask_count` INT NOT NULL DEFAULT 0 COMMENT '成功卖单数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1计算中 2执行中 3成功 4部分成功 5失败 6已过期',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_cycle_no` (`tenant_id`, `cycle_no`),
  KEY `idx_config_status_time` (`tenant_id`, `config_id`, `status`, `create_times`),
  KEY `idx_symbol_time` (`tenant_id`, `symbol_id`, `create_times`),
  CONSTRAINT `chk_liquidity_cycle_status` CHECK (`status` IN (1, 2, 3, 4, 5, 6))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='做市报价执行周期';

CREATE TABLE `t_liquidity_quote_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `quote_no` VARCHAR(64) NOT NULL COMMENT '报价订单业务号',
  `cycle_id` BIGINT NOT NULL COMMENT '报价周期ID',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `provider_id` BIGINT NOT NULL COMMENT '流动性提供方ID',
  `symbol_id` BIGINT NOT NULL COMMENT '内部交易标的ID',
  `side` TINYINT NOT NULL COMMENT '方向：1买 2卖',
  `level_no` INT NOT NULL COMMENT '档位',
  `price` DECIMAL(36,18) NOT NULL COMMENT '报价价格',
  `qty` DECIMAL(36,18) NOT NULL COMMENT '报价数量',
  `filled_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '累计成交数量',
  `internal_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'trade订单ID',
  `internal_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'trade订单号',
  `client_order_id` VARCHAR(64) NOT NULL COMMENT '发送给trade的幂等客户订单号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待提交 2已挂单 3部分成交 4已成交 5撤单中 6已撤单 7失败 8未知',
  `expire_at` BIGINT NOT NULL DEFAULT 0 COMMENT '报价到期时间',
  `cancel_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '撤单原因',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_quote_no` (`tenant_id`, `quote_no`),
  UNIQUE KEY `uk_tenant_client_order_id` (`tenant_id`, `client_order_id`),
  KEY `idx_config_status_expire` (`tenant_id`, `config_id`, `status`, `expire_at`),
  KEY `idx_internal_order` (`tenant_id`, `internal_order_id`),
  KEY `idx_cycle_side_level` (`tenant_id`, `cycle_id`, `side`, `level_no`),
  CONSTRAINT `chk_liquidity_quote_side` CHECK (`side` IN (1, 2)),
  CONSTRAINT `chk_liquidity_quote_value` CHECK (`level_no` > 0 AND `price` > 0 AND `qty` > 0 AND `filled_qty` >= 0 AND `filled_qty` <= `qty`),
  CONSTRAINT `chk_liquidity_quote_status` CHECK (`status` IN (1, 2, 3, 4, 5, 6, 7, 8))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='内部做市报价订单映射';

CREATE TABLE `t_liquidity_external_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `order_no` VARCHAR(64) NOT NULL COMMENT '内部外部订单业务号',
  `request_no` VARCHAR(64) NOT NULL COMMENT '外部下单幂等请求号',
  `provider_id` BIGINT NOT NULL COMMENT '外部流动性提供方ID',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `symbol_id` BIGINT NOT NULL COMMENT '内部交易标的ID',
  `external_symbol` VARCHAR(64) NOT NULL COMMENT '外部交易对',
  `purpose` TINYINT NOT NULL COMMENT '用途：1用户订单路由 2净敞口对冲 3外部做市',
  `reference_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '关联类型：TRADE_ORDER/HEDGE_TASK/QUOTE',
  `reference_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联记录ID',
  `side` TINYINT NOT NULL COMMENT '方向：1买 2卖',
  `order_type` TINYINT NOT NULL COMMENT '类型：1限价 2市价',
  `time_in_force` TINYINT NOT NULL DEFAULT 2 COMMENT '有效方式：1GTC 2IOC 3FOK 4POST_ONLY',
  `price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '委托价格',
  `qty` DECIMAL(36,18) NOT NULL COMMENT '委托数量',
  `filled_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '累计成交数量',
  `avg_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '成交均价',
  `fee_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '累计手续费',
  `fee_asset` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `external_order_id` VARCHAR(128) NULL DEFAULT NULL COMMENT '外部订单ID；提交成功前为空',
  `external_client_order_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '外部客户订单号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待提交 2已提交 3部分成交 4已成交 5撤单中 6已撤单 7拒绝 8失败 9未知',
  `submitted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '提交时间',
  `finished_at` BIGINT NOT NULL DEFAULT 0 COMMENT '终态时间',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最近外部错误码',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `raw_response` JSON NULL COMMENT '最近一次外部响应',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_order_no` (`tenant_id`, `order_no`),
  UNIQUE KEY `uk_tenant_request_no` (`tenant_id`, `request_no`),
  UNIQUE KEY `uk_provider_external_order` (`provider_id`, `external_order_id`),
  KEY `idx_provider_status_retry` (`tenant_id`, `provider_id`, `status`, `next_retry_at`),
  KEY `idx_reference` (`tenant_id`, `reference_type`, `reference_id`),
  KEY `idx_symbol_time` (`tenant_id`, `symbol_id`, `create_times`),
  CONSTRAINT `chk_liquidity_external_purpose` CHECK (`purpose` IN (1, 2, 3)),
  CONSTRAINT `chk_liquidity_external_side` CHECK (`side` IN (1, 2)),
  CONSTRAINT `chk_liquidity_external_order_type` CHECK (`order_type` IN (1, 2)),
  CONSTRAINT `chk_liquidity_external_tif` CHECK (`time_in_force` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_external_value` CHECK (
    `qty` > 0 AND `filled_qty` >= 0 AND `filled_qty` <= `qty`
    AND (`order_type` = 2 OR `price` > 0)
  ),
  CONSTRAINT `chk_liquidity_external_status` CHECK (`status` IN (1, 2, 3, 4, 5, 6, 7, 8, 9))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部订单路由与对冲订单';

CREATE TABLE `t_liquidity_external_fill` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `provider_id` BIGINT NOT NULL COMMENT '外部流动性提供方ID',
  `external_order_id` BIGINT NOT NULL COMMENT 't_liquidity_external_order.id',
  `fill_no` VARCHAR(128) NOT NULL COMMENT '内部成交幂等号',
  `external_trade_id` VARCHAR(128) NOT NULL COMMENT '外部成交ID',
  `side` TINYINT NOT NULL COMMENT '方向：1买 2卖',
  `price` DECIMAL(36,18) NOT NULL COMMENT '成交价格',
  `qty` DECIMAL(36,18) NOT NULL COMMENT '成交数量',
  `amount` DECIMAL(36,18) NOT NULL COMMENT '成交名义金额',
  `fee_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '手续费',
  `fee_asset` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `liquidity_type` TINYINT NOT NULL DEFAULT 0 COMMENT '流动性：0未知 1Maker 2Taker',
  `trade_time` BIGINT NOT NULL COMMENT '外部成交时间',
  `settlement_status` TINYINT NOT NULL DEFAULT 1 COMMENT '内部入账：1待处理 2处理中 3成功 4失败',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `raw_payload` JSON NULL COMMENT '外部成交原文',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_provider_external_trade` (`provider_id`, `external_trade_id`),
  UNIQUE KEY `uk_tenant_fill_no` (`tenant_id`, `fill_no`),
  KEY `idx_order` (`tenant_id`, `external_order_id`),
  KEY `idx_settlement_retry` (`tenant_id`, `settlement_status`, `next_retry_at`),
  CONSTRAINT `chk_liquidity_fill_side` CHECK (`side` IN (1, 2)),
  CONSTRAINT `chk_liquidity_fill_value` CHECK (`price` > 0 AND `qty` > 0 AND `amount` > 0 AND `fee_amount` >= 0),
  CONSTRAINT `chk_liquidity_fill_type` CHECK (`liquidity_type` IN (0, 1, 2)),
  CONSTRAINT `chk_liquidity_fill_settlement` CHECK (`settlement_status` IN (1, 2, 3, 4))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部成交回报';

CREATE TABLE `t_liquidity_hedge_task` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `hedge_no` VARCHAR(64) NOT NULL COMMENT '对冲业务号',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `provider_id` BIGINT NOT NULL COMMENT '外部对冲提供方ID',
  `symbol_id` BIGINT NOT NULL COMMENT '内部交易标的ID',
  `trigger_type` TINYINT NOT NULL COMMENT '触发：1成交事件 2敞口阈值 3人工 4恢复任务',
  `exposure_before` DECIMAL(36,18) NOT NULL COMMENT '对冲前净敞口；正数多头，负数空头',
  `target_exposure` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '目标净敞口',
  `side` TINYINT NOT NULL COMMENT '外部对冲方向：1买 2卖',
  `target_qty` DECIMAL(36,18) NOT NULL COMMENT '目标对冲数量',
  `executed_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '已执行数量',
  `avg_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '对冲均价',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待执行 2执行中 3部分完成 4完成 5失败 6取消',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '乐观锁版本',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_hedge_no` (`tenant_id`, `hedge_no`),
  KEY `idx_config_status_retry` (`tenant_id`, `config_id`, `status`, `next_retry_at`),
  KEY `idx_provider_status` (`tenant_id`, `provider_id`, `status`),
  CONSTRAINT `chk_liquidity_hedge_trigger` CHECK (`trigger_type` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_hedge_side` CHECK (`side` IN (1, 2)),
  CONSTRAINT `chk_liquidity_hedge_qty` CHECK (`target_qty` > 0 AND `executed_qty` >= 0 AND `executed_qty` <= `target_qty`),
  CONSTRAINT `chk_liquidity_hedge_status` CHECK (`status` IN (1, 2, 3, 4, 5, 6))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='净敞口外部对冲任务';

CREATE TABLE `t_liquidity_inventory_snapshot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `snapshot_no` VARCHAR(64) NOT NULL COMMENT '库存快照业务号',
  `config_id` BIGINT NOT NULL COMMENT '交易对流动性配置ID',
  `provider_id` BIGINT NOT NULL COMMENT '流动性提供方ID',
  `symbol_id` BIGINT NOT NULL COMMENT '内部交易标的ID',
  `base_asset` VARCHAR(32) NOT NULL COMMENT '基础资产',
  `quote_asset` VARCHAR(32) NOT NULL COMMENT '计价资产',
  `base_total` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产总额',
  `base_available` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产可用',
  `base_frozen` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '基础资产冻结',
  `quote_total` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '计价/保证金资产总额',
  `quote_available` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '计价/保证金资产可用',
  `quote_frozen` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '计价/保证金资产冻结',
  `position_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '合约净持仓；现货为0',
  `pending_buy_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '未成交买单数量',
  `pending_sell_qty` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '未成交卖单数量',
  `net_exposure` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '归一化净敞口',
  `reference_price` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '快照参考价',
  `exposure_notional` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '净敞口名义金额',
  `source` TINYINT NOT NULL COMMENT '来源：1内部资产持仓 2外部账户 3合并视图',
  `snapshot_time` BIGINT NOT NULL COMMENT '快照时间',
  `raw_payload` JSON NULL COMMENT '外部或计算原文',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_snapshot_no` (`tenant_id`, `snapshot_no`),
  KEY `idx_config_snapshot_time` (`tenant_id`, `config_id`, `snapshot_time`),
  KEY `idx_provider_snapshot_time` (`tenant_id`, `provider_id`, `snapshot_time`),
  CONSTRAINT `chk_liquidity_inventory_source` CHECK (`source` IN (1, 2, 3)),
  CONSTRAINT `chk_liquidity_inventory_amount` CHECK (
    `base_total` >= 0 AND `base_available` >= 0 AND `base_frozen` >= 0
    AND `quote_total` >= 0 AND `quote_available` >= 0 AND `quote_frozen` >= 0
    AND `pending_buy_qty` >= 0 AND `pending_sell_qty` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='做市库存与敞口快照';

CREATE TABLE `t_liquidity_risk_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `event_no` VARCHAR(64) NOT NULL COMMENT '风控事件号',
  `config_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易对流动性配置ID',
  `provider_id` BIGINT NOT NULL DEFAULT 0 COMMENT '流动性提供方ID',
  `symbol_id` BIGINT NOT NULL DEFAULT 0 COMMENT '内部交易标的ID',
  `risk_type` VARCHAR(64) NOT NULL COMMENT '风险类型',
  `risk_level` TINYINT NOT NULL COMMENT '级别：1提示 2警告 3严重 4致命',
  `metric_value` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '触发指标值',
  `threshold_value` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '配置阈值',
  `action_type` TINYINT NOT NULL DEFAULT 1 COMMENT '动作：1仅记录 2暂停报价 3撤销报价 4熔断 5人工介入',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2处理中 3已恢复 4已关闭',
  `message` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '事件描述',
  `context_json` JSON NULL COMMENT '触发上下文',
  `triggered_at` BIGINT NOT NULL COMMENT '触发时间',
  `recovered_at` BIGINT NOT NULL DEFAULT 0 COMMENT '恢复时间',
  `closed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '关闭时间',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '处理人ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_event_no` (`tenant_id`, `event_no`),
  KEY `idx_tenant_status_level` (`tenant_id`, `status`, `risk_level`, `triggered_at`),
  KEY `idx_config_type_time` (`tenant_id`, `config_id`, `risk_type`, `triggered_at`),
  CONSTRAINT `chk_liquidity_risk_level` CHECK (`risk_level` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_risk_action` CHECK (`action_type` IN (1, 2, 3, 4, 5)),
  CONSTRAINT `chk_liquidity_risk_status` CHECK (`status` IN (1, 2, 3, 4))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流动性风险事件与熔断记录';

CREATE TABLE `t_liquidity_reconcile_batch` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `batch_no` VARCHAR(64) NOT NULL COMMENT '对账批次号',
  `provider_id` BIGINT NOT NULL COMMENT '外部流动性提供方ID',
  `reconcile_type` TINYINT NOT NULL COMMENT '类型：1订单 2成交 3余额 4持仓',
  `window_start` BIGINT NOT NULL COMMENT '对账窗口开始',
  `window_end` BIGINT NOT NULL COMMENT '对账窗口结束',
  `local_count` BIGINT NOT NULL DEFAULT 0 COMMENT '本地记录数',
  `external_count` BIGINT NOT NULL DEFAULT 0 COMMENT '外部记录数',
  `matched_count` BIGINT NOT NULL DEFAULT 0 COMMENT '一致记录数',
  `difference_count` BIGINT NOT NULL DEFAULT 0 COMMENT '差异记录数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待执行 2执行中 3一致 4存在差异 5失败',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` BIGINT NOT NULL DEFAULT 0 COMMENT '结束时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_batch_no` (`tenant_id`, `batch_no`),
  KEY `idx_provider_type_window` (`tenant_id`, `provider_id`, `reconcile_type`, `window_start`),
  KEY `idx_status_time` (`tenant_id`, `status`, `create_times`),
  CONSTRAINT `chk_liquidity_reconcile_type` CHECK (`reconcile_type` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_reconcile_window` CHECK (`window_end` > `window_start`),
  CONSTRAINT `chk_liquidity_reconcile_status` CHECK (`status` IN (1, 2, 3, 4, 5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部流动性对账批次';

CREATE TABLE `t_liquidity_reconcile_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `batch_id` BIGINT NOT NULL COMMENT '对账批次ID',
  `difference_no` VARCHAR(64) NOT NULL COMMENT '差异业务号',
  `difference_type` TINYINT NOT NULL COMMENT '差异：1本地缺失 2外部缺失 3状态不一致 4金额不一致 5其他',
  `business_type` VARCHAR(32) NOT NULL COMMENT '业务类型',
  `local_reference` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '本地记录标识',
  `external_reference` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '外部记录标识',
  `local_value` JSON NULL COMMENT '本地值',
  `external_value` JSON NULL COMMENT '外部值',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2自动修复中 3已修复 4忽略 5需人工',
  `resolution` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '处理结果',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '处理人ID',
  `resolved_at` BIGINT NOT NULL DEFAULT 0 COMMENT '处理时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_difference_no` (`tenant_id`, `difference_no`),
  KEY `idx_batch_status` (`tenant_id`, `batch_id`, `status`),
  KEY `idx_business_reference` (`tenant_id`, `business_type`, `local_reference`, `external_reference`),
  CONSTRAINT `chk_liquidity_difference_type` CHECK (`difference_type` IN (1, 2, 3, 4, 5)),
  CONSTRAINT `chk_liquidity_difference_status` CHECK (`status` IN (1, 2, 3, 4, 5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外部流动性对账差异';

CREATE TABLE `t_liquidity_event_inbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `consumer` VARCHAR(64) NOT NULL COMMENT '消费者标识',
  `event_no` VARCHAR(128) NOT NULL COMMENT '上游事件唯一号',
  `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型',
  `aggregate_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '聚合类型',
  `aggregate_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '聚合标识',
  `payload` JSON NULL COMMENT '事件负载',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1处理中 2成功 3失败',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `processed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '处理完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_consumer_event` (`consumer`, `event_no`),
  KEY `idx_status_retry` (`status`, `next_retry_at`),
  KEY `idx_aggregate` (`tenant_id`, `aggregate_type`, `aggregate_id`),
  CONSTRAINT `chk_liquidity_inbox_status` CHECK (`status` IN (1, 2, 3))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='可靠消费与幂等收件箱';

CREATE TABLE `t_liquidity_event_outbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `event_no` VARCHAR(128) NOT NULL COMMENT '事件唯一号',
  `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型',
  `topic` VARCHAR(128) NOT NULL COMMENT '目标主题',
  `message_key` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '消息分区键',
  `aggregate_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '聚合类型',
  `aggregate_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '聚合标识',
  `payload` JSON NOT NULL COMMENT '事件负载',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待发送 2发送中 3成功 4失败',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `max_retry_count` INT NOT NULL DEFAULT 20 COMMENT '最大重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `sent_at` BIGINT NOT NULL DEFAULT 0 COMMENT '发送成功时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_event_no` (`tenant_id`, `event_no`),
  KEY `idx_status_retry` (`status`, `next_retry_at`),
  KEY `idx_aggregate` (`tenant_id`, `aggregate_type`, `aggregate_id`),
  CONSTRAINT `chk_liquidity_outbox_status` CHECK (`status` IN (1, 2, 3, 4)),
  CONSTRAINT `chk_liquidity_outbox_retry` CHECK (`retry_count` >= 0 AND `max_retry_count` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='可靠发布发件箱';
