SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- =========================================================
-- 1. 期权合约表
-- =========================================================
DROP TABLE IF EXISTS `t_option_contract`;
CREATE TABLE `t_option_contract` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '合约编码，如 BTC-20260630-50000-C',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '标的资产，如 BTCUSDT',
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '结算币种，如 USDT',
  `quote_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '报价币种，如 USDT',
  `option_type` TINYINT NOT NULL DEFAULT 0 COMMENT '期权类型：1看涨 2看跌',
  `exercise_style` TINYINT NOT NULL DEFAULT 0 COMMENT '行权方式：1欧式 2美式',
  `settlement_type` TINYINT NOT NULL DEFAULT 0 COMMENT '结算方式：1现金结算 2实物交割',
  `strike_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权价',
  `contract_unit` DECIMAL(32,16) NOT NULL DEFAULT 1 COMMENT '每张合约对应标的数量',
  `min_order_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最小下单数量',
  `max_order_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最大下单数量',
  `price_tick` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最小价格变动单位',
  `qty_step` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最小数量变动单位',
  `multiplier` DECIMAL(32,16) NOT NULL DEFAULT 1 COMMENT '合约乘数',
  `list_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上市时间',
  `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '到期时间',
  `deliver_time` BIGINT NOT NULL DEFAULT 0 COMMENT '交割/结算时间',
  `is_auto_exercise` TINYINT NOT NULL DEFAULT 2 COMMENT '是否自动行权：1是 2否',
  `maker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Maker成交手续费率',
  `taker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Taker成交手续费率',
  `exercise_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '行权手续费率',
  `fee_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '平台手续费归集用户ID',
  `fee_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '平台手续费归集Option账户ID',
  `seller_margin_mode` TINYINT NOT NULL DEFAULT 1 COMMENT '卖方保证金模式：1关闭 2逐仓 3组合',
  `initial_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方初始保证金率',
  `maintenance_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方维持保证金率',
  `min_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方最低保证金率',
  `liquidation_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '强平手续费率',
  `insurance_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金用户ID',
  `insurance_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金Option账户ID',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1待上市 2可交易 3暂停交易 4已到期 5已结算 6已下线',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `is_deleted` TINYINT NOT NULL DEFAULT 2 COMMENT '是否删除：1是 2否',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract_code` (`tenant_id`, `contract_code`),
  KEY `idx_tenant_underlying_symbol` (`tenant_id`, `underlying_symbol`),
  KEY `idx_tenant_expire_time` (`tenant_id`, `expire_time`),
  KEY `idx_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权合约表';

-- =========================================================
-- 2. 期权当前行情表
-- =========================================================
DROP TABLE IF EXISTS `t_option_market`;
CREATE TABLE `t_option_market` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `underlying_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标的价格',
  `mark_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标记价格/参考权利金',
  `last_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最新成交价',
  `bid_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '买一价',
  `ask_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖一价',
  `theoretical_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '理论价/风险定价',
  `intrinsic_value` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '内在价值',
  `time_value` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '时间价值',
  `iv` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '隐含波动率',
  `delta` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Delta',
  `gamma` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Gamma',
  `theta` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Theta',
  `vega` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Vega',
  `rho` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Rho',
  `risk_free_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '无风险利率',
  `pricing_model` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '定价模型，如 Black-Scholes',
  `snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行情快照时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_tenant_snapshot_time` (`tenant_id`, `snapshot_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权当前行情表';

-- =========================================================
-- 3. 期权行情快照表
-- =========================================================
DROP TABLE IF EXISTS `t_option_market_snapshot`;
CREATE TABLE `t_option_market_snapshot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `underlying_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标的价格',
  `mark_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标记价格',
  `last_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最新成交价',
  `bid_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '买一价',
  `ask_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖一价',
  `theoretical_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '理论价',
  `iv` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '隐含波动率',
  `delta` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Delta',
  `gamma` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Gamma',
  `theta` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Theta',
  `vega` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Vega',
  `rho` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Rho',
  `snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '快照时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_contract_snapshot_time` (`tenant_id`, `contract_id`, `snapshot_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权行情快照表';

DROP TABLE IF EXISTS `t_option_market_snapshot_inbox`;
CREATE TABLE `t_option_market_snapshot_inbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `snapshot_id` VARCHAR(64) NOT NULL COMMENT '权威行情快照ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_snapshot_contract` (`snapshot_id`, `contract_id`),
  KEY `idx_option_snapshot_inbox_created` (`create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权权威行情逐合约消费幂等表';

-- =========================================================
-- 4. 期权委托表
-- =========================================================
DROP TABLE IF EXISTS `t_option_order`;
CREATE TABLE `t_option_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '标的',
  `side` TINYINT NOT NULL DEFAULT 0 COMMENT '买卖方向：1买 2卖',
  `position_effect` TINYINT NOT NULL DEFAULT 0 COMMENT '开平方向：1开仓 2平仓',
  `order_type` TINYINT NOT NULL DEFAULT 0 COMMENT '订单类型：1限价 2市价 3只做maker 4IOC 5FOK',
  `price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '委托价格/权利金',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '委托数量',
  `filled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已成交数量',
  `unfilled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未成交数量',
  `avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交均价',
  `turnover` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交额',
  `fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '手续费',
  `fee_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `margin_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '冻结保证金',
  `source` TINYINT NOT NULL DEFAULT 0 COMMENT '订单来源：1APP 2WEB 3API 4ADMIN',
  `client_order_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端订单号',
  `reduce_only` TINYINT NOT NULL DEFAULT 2 COMMENT '是否只减仓：1是 2否',
  `mmp` TINYINT NOT NULL DEFAULT 2 COMMENT '是否做市商保护单：1是 2否',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1待撮合 2部分成交 3完全成交 4已撤单 5拒单 6已过期',
  `cancel_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '撤单/拒单原因',
  `match_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后成交时间',
  `cancel_time` BIGINT NOT NULL DEFAULT 0 COMMENT '撤单时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid_client_order_id` (`tenant_id`, `user_id`, `client_order_id`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `user_id`, `account_id`),
  KEY `idx_tenant_uid_contract_id` (`tenant_id`, `user_id`, `contract_id`),
  KEY `idx_tenant_contract_id_status` (`tenant_id`, `contract_id`, `status`),
  KEY `idx_tenant_create_times` (`tenant_id`, `create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权委托表';

CREATE TABLE `t_option_client_order_key` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `client_order_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端幂等订单号，禁止为空',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Option订单号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_client_order` (`tenant_id`, `user_id`, `client_order_id`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  CONSTRAINT `chk_option_client_order_key` CHECK (`client_order_id` <> '' AND `order_id` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权客户端订单幂等键';

-- =========================================================
-- 5. 期权成交表
-- =========================================================
DROP TABLE IF EXISTS `t_option_trade`;
CREATE TABLE `t_option_trade` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `trade_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '成交号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '标的',
  `buy_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '买单ID',
  `buy_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '买单订单号',
  `buy_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '买方用户ID',
  `buy_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '买方账户ID',
  `sell_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖单ID',
  `sell_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '卖单订单号',
  `sell_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方用户ID',
  `sell_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方账户ID',
  `price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交价格/权利金',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交数量',
  `turnover` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交额',
  `buy_fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '买方手续费',
  `sell_fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖方手续费',
  `fee_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `maker_side` TINYINT NOT NULL DEFAULT 0 COMMENT 'maker方向：1买 2卖',
  `match_sequence` BIGINT NOT NULL DEFAULT 0 COMMENT '合约内严格递增撮合序号',
  `trade_time` BIGINT NOT NULL DEFAULT 0 COMMENT '成交时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_trade_no` (`tenant_id`, `trade_no`),
  UNIQUE KEY `uk_tenant_contract_match_sequence` (`tenant_id`, `contract_id`, `match_sequence`),
  KEY `idx_tenant_buy_user_id` (`tenant_id`, `buy_user_id`),
  KEY `idx_tenant_sell_user_id` (`tenant_id`, `sell_user_id`),
  KEY `idx_tenant_contract_trade_time` (`tenant_id`, `contract_id`, `trade_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权成交表';

CREATE TABLE `t_option_match_sequence` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL COMMENT '期权合约ID',
  `next_sequence` BIGINT NOT NULL DEFAULT 1 COMMENT '下一个撮合序号',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract` (`tenant_id`, `contract_id`),
  CONSTRAINT `chk_option_match_sequence` CHECK (`next_sequence` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权合约撮合序列';

CREATE TABLE `t_option_outbox` (
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

CREATE TABLE `t_option_inbox` (
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

CREATE TABLE `t_option_margin_lot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方Option账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头持仓ID，事件入账后回填',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖出开仓订单ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `freeze_biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT 'Asset冻结业务号',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '批次数量',
  `remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '尚未平仓的批次数量',
  `initial_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '分配的初始保证金',
  `remaining_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '尚未释放或消费的保证金',
  `pending_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已生成但尚未完成资产指令的保证金',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1有效 2释放中 3消费中 4已释放 5已消费',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_trade` (`tenant_id`, `trade_id`),
  KEY `idx_margin_lot_position` (`tenant_id`, `position_id`, `status`, `id`),
  KEY `idx_margin_lot_order` (`tenant_id`, `order_id`, `status`, `id`),
  CONSTRAINT `chk_option_margin_lot` CHECK (
    `quantity` > 0 AND `remaining_quantity` >= 0 AND `remaining_quantity` <= `quantity`
    AND `initial_margin` > 0 AND `remaining_margin` >= 0 AND `pending_margin` >= 0
    AND `pending_margin` <= `remaining_margin`
    AND `status` IN (1,2,3,4,5,6)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权卖方保证金批次';

CREATE TABLE `t_option_margin_lot_application` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `instruction_id` BIGINT NOT NULL DEFAULT 0 COMMENT '资产指令ID',
  `margin_lot_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保证金批次ID',
  `action` TINYINT NOT NULL DEFAULT 0 COMMENT '资产指令动作',
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已应用金额',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_instruction` (`tenant_id`, `instruction_id`),
  KEY `idx_application_margin_lot` (`tenant_id`, `margin_lot_id`),
  CONSTRAINT `chk_margin_lot_application` CHECK (`action` IN (2,3) AND `amount` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='保证金批次资产指令应用幂等记录';

CREATE TABLE `t_option_risk_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '结算币种',
  `equity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '风险权益',
  `position_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖方持仓保证金',
  `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '维持保证金',
  `unrealized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未实现盈亏',
  `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '维持保证金/权益',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2追保 3强平中 4破产 5限制',
  `last_calc_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后计算时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_account_coin` (`tenant_id`, `user_id`, `account_id`, `settle_coin`),
  KEY `idx_risk_account_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_risk_account` CHECK (
    `position_margin` >= 0 AND `maintenance_margin` >= 0
    AND `risk_rate` >= 0 AND `status` IN (1,2,3,4,5)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权卖方风险账户';

CREATE TABLE `t_option_liquidation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `liquidation_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '强平单号',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头持仓ID',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '强平数量',
  `mark_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '触发标记价',
  `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '触发维持保证金',
  `equity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '触发权益',
  `deficit_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '穿仓缺口',
  `liquidation_fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '强平手续费',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2执行中 3完成 4失败 5破产 6人工',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `collateral_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '本次消费的用户保证金',
  `insurance_fund_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '保险基金补足金额',
  `remaining_deficit` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '保险后剩余穿仓',
  `takeover_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险账户接管持仓ID',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `insurance_attempt` INT NOT NULL DEFAULT 0 COMMENT '保险基金人工重试代次',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_liquidation_no` (`tenant_id`, `liquidation_no`),
  KEY `idx_liquidation_status` (`tenant_id`, `status`, `id`),
  KEY `idx_liquidation_position` (`tenant_id`, `position_id`, `id`),
  CONSTRAINT `chk_option_liquidation` CHECK (
    `quantity` > 0 AND `deficit_amount` >= 0 AND `liquidation_fee` >= 0
    AND `status` IN (1,2,3,4,5,6) AND `retry_count` >= 0 AND `insurance_attempt` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权强平记录';

CREATE TABLE `t_option_insurance_fund_flow` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `flow_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '保险基金流水号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `liquidation_id` BIGINT NOT NULL DEFAULT 0 COMMENT '强平记录ID',
  `flow_type` TINYINT NOT NULL DEFAULT 0 COMMENT '类型：1强平费 2缺口赔付 3人工注资 4人工提取',
  `coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '币种',
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '正数入金，负数出金',
  `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Asset实际流水号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_flow_no` (`tenant_id`, `flow_no`),
  KEY `idx_insurance_liquidation` (`tenant_id`, `liquidation_id`),
  CONSTRAINT `chk_option_insurance_fund_flow` CHECK (`flow_type` IN (1,2,3,4) AND `amount` <> 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权保险基金流水';

-- =========================================================
-- 6. 期权持仓表
-- =========================================================
DROP TABLE IF EXISTS `t_option_position`;
CREATE TABLE `t_option_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '标的',
  `side` TINYINT NOT NULL DEFAULT 0 COMMENT '持仓方向：1多头 2空头',
  `position_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '持仓数量',
  `available_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '可用数量',
  `frozen_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '冻结数量',
  `open_avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '开仓均价/平均权利金',
  `mark_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标记价格',
  `position_value` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '持仓价值',
  `margin_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '占用保证金',
  `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '维持保证金',
  `unrealized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未实现盈亏',
  `realized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已实现盈亏',
  `exerciseable_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '可行权数量',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1持仓中 2已平仓 3已行权 4已到期 5已结算',
  `last_calc_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上次风控计算时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_uid_account_contract_side` (`tenant_id`, `user_id`, `account_id`, `contract_id`, `side`),
  KEY `idx_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `user_id`, `account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权持仓表';

-- =========================================================
-- 7. 期权行权表
-- =========================================================
DROP TABLE IF EXISTS `t_option_exercise`;
CREATE TABLE `t_option_exercise` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `exercise_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '行权单号',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '持仓ID',
  `exercise_type` TINYINT NOT NULL DEFAULT 0 COMMENT '行权类型：1用户主动 2系统自动',
  `exercise_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权数量',
  `strike_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权价',
  `settlement_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '结算价',
  `exercise_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权金额',
  `profit_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权收益',
  `fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权手续费',
  `fee_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1待处理 2已执行 3已拒绝 4已取消',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `exercise_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行权时间',
  `finish_time` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_exercise_no` (`tenant_id`, `exercise_no`),
  KEY `idx_tenant_uid_account_contract_id` (`tenant_id`, `user_id`, `account_id`, `contract_id`),
  KEY `idx_tenant_position_id` (`tenant_id`, `position_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权行权表';

CREATE TABLE `t_option_exercise_assignment` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `exercise_id` BIGINT NOT NULL DEFAULT 0 COMMENT '行权记录ID',
  `exercise_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '行权单号',
  `long_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头持仓ID',
  `short_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被指派空头持仓ID',
  `short_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头用户ID',
  `short_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头账户ID',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '指派数量',
  `payoff` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '空头应付金额',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2完成 3失败 4人工',
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '首个空头资产指令号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exercise_short_position` (`tenant_id`, `exercise_id`, `short_position_id`),
  KEY `idx_assignment_exercise` (`tenant_id`, `exercise_id`, `status`, `id`),
  CONSTRAINT `chk_option_exercise_assignment` CHECK (
    `quantity` > 0 AND `payoff` > 0 AND `status` IN (1,2,3,4)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权主动行权空头指派';

-- =========================================================
-- 8. 期权到期结算表
-- =========================================================
DROP TABLE IF EXISTS `t_option_settlement`;
CREATE TABLE `t_option_settlement` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `settlement_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '结算单号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '标的',
  `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '到期时间',
  `settlement_time` BIGINT NOT NULL DEFAULT 0 COMMENT '结算时间',
  `delivery_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '交割结算价',
  `theoretical_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '结算理论价',
  `iv` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '结算时IV',
  `is_itm` TINYINT NOT NULL DEFAULT 2 COMMENT '是否实值：1是 2否',
  `exercise_result` TINYINT NOT NULL DEFAULT 0 COMMENT '行权结果：1未执行 2自动行权 3自动放弃',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1待结算 2结算中 3已完成 4失败',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_settlement_no` (`tenant_id`, `settlement_no`),
  UNIQUE KEY `uk_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_tenant_expire_time` (`tenant_id`, `expire_time`),
  KEY `idx_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权到期结算表';

-- =========================================================
-- 9. 期权账户资产表
-- =========================================================
DROP TABLE IF EXISTS `t_option_account`;
CREATE TABLE `t_option_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `margin_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '保证金币种',
  `balance` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '账户余额',
  `available_balance` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '可用余额',
  `frozen_balance` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '冻结余额',
  `position_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '持仓保证金',
  `order_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '委托保证金',
  `unrealized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未实现盈亏',
  `realized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已实现盈亏',
  `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '风险率',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2冻结 3限制交易',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_uid_account_margin_coin` (`tenant_id`, `user_id`, `account_id`, `margin_coin`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `user_id`, `account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权账户资产表';

-- =========================================================
-- 10. 期权资金流水表
-- =========================================================
DROP TABLE IF EXISTS `t_option_bill`;
CREATE TABLE `t_option_bill` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `biz_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '业务流水号',
  `ref_type` TINYINT NOT NULL DEFAULT 0 COMMENT '关联类型：1下单 2成交 3撤单 4行权 5结算 6手续费',
  `ref_id` BIGINT NOT NULL DEFAULT 0 COMMENT '关联ID',
  `coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '币种',
  `change_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '变动金额，正负都有可能',
  `balance_before` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '变动前余额',
  `balance_after` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '变动后余额',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_biz_no` (`tenant_id`, `biz_no`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `user_id`, `account_id`),
  KEY `idx_tenant_ref_type_ref_id` (`tenant_id`, `ref_type`, `ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权资金流水表';

-- =========================================================
-- 11. Option 资产指令表
-- =========================================================
DROP TABLE IF EXISTS `t_option_asset_instruction`;
CREATE TABLE `t_option_asset_instruction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '资产指令号/幂等键',
  `biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '业务单号',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '订单ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '持仓ID',
  `margin_lot_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方保证金批次ID',
  `liquidation_id` BIGINT NOT NULL DEFAULT 0 COMMENT '强平记录ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `action` TINYINT NOT NULL DEFAULT 0 COMMENT '动作：1冻结 2扣冻结 3释放冻结 4可用入账 5可用扣减',
  `target_biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '扣减或释放关联的原冻结业务号',
  `coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '币种',
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '金额，必须为正数',
  `step_no` INT NOT NULL DEFAULT 1 COMMENT '同一业务内执行顺序',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待执行 2执行中 3成功 4失败 5人工处理 6未执行取消',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT 0 COMMENT '下次重试时间',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `asset_flow_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Asset实际流水号',
  `reconciliation_status` TINYINT NOT NULL DEFAULT 1 COMMENT '对账状态：1待对账 2一致 3不一致',
  `reconciled_at` BIGINT NOT NULL DEFAULT 0 COMMENT '对账完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_instruction_no` (`tenant_id`, `instruction_no`),
  KEY `idx_instruction_retry` (`tenant_id`, `status`, `next_retry_at`, `id`),
  KEY `idx_instruction_biz_step` (`tenant_id`, `biz_no`, `step_no`, `status`, `id`),
  KEY `idx_instruction_order` (`tenant_id`, `order_id`),
  KEY `idx_instruction_trade` (`tenant_id`, `trade_id`),
  KEY `idx_instruction_margin_lot` (`tenant_id`, `margin_lot_id`),
  KEY `idx_instruction_liquidation` (`tenant_id`, `liquidation_id`, `status`, `id`),
  CONSTRAINT `chk_option_asset_instruction` CHECK (
    `action` IN (1,2,3,4,5)
    AND `amount` > 0
    AND `step_no` > 0
    AND `status` IN (1,2,3,4,5,6)
    AND `retry_count` >= 0
    AND `reconciliation_status` IN (1,2,3)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option发送给Asset的幂等资金指令';

-- =========================================================
-- 12. 期权不可变结算价快照
-- =========================================================
DROP TABLE IF EXISTS `t_option_settlement_price`;
CREATE TABLE `t_option_settlement_price` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `price_source` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '权威价格来源',
  `window_start` BIGINT NOT NULL DEFAULT 0 COMMENT '取价窗口开始时间',
  `window_end` BIGINT NOT NULL DEFAULT 0 COMMENT '取价窗口结束时间',
  `sample_count` BIGINT NOT NULL DEFAULT 0 COMMENT '有效样本数',
  `calculation_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '计算方法',
  `delivery_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '最终结算价',
  `source_snapshot_ids` TEXT NOT NULL COMMENT '原始快照依据',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '结算价版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1等待价格 2已确认 3已拒绝',
  `confirmed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '确认人，0为系统',
  `confirmed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '确认时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_price_contract` (`tenant_id`, `contract_id`),
  KEY `idx_settlement_price_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_settlement_price` CHECK (
    `status` IN (1,2,3)
    AND `version` > 0
    AND (`status` <> 2 OR (`delivery_price` > 0 AND `sample_count` > 0 AND `confirmed_at` > 0))
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权不可变结算价快照';

-- =========================================================
-- 13. 期权结算批次及逐持仓明细
-- =========================================================
DROP TABLE IF EXISTS `t_option_settlement_batch`;
CREATE TABLE `t_option_settlement_batch` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '结算批次号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `settlement_price_id` BIGINT NOT NULL DEFAULT 0 COMMENT '不可变结算价ID',
  `total_credit` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '多头应收合计',
  `total_debit` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '空头应付合计',
  `instruction_count` BIGINT NOT NULL DEFAULT 0 COMMENT '资产指令数',
  `success_count` BIGINT NOT NULL DEFAULT 0 COMMENT '已成功指令数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '批次状态：1等待价格 2价格锁定 3计算中 4指令已创建 5资产处理中 6对账中 7完成 8失败',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_batch_no` (`tenant_id`, `batch_no`),
  UNIQUE KEY `uk_settlement_batch_contract` (`tenant_id`, `contract_id`),
  KEY `idx_settlement_batch_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_settlement_batch` CHECK (
    `status` IN (1,2,3,4,5,6,7,8)
    AND `total_credit` >= 0
    AND `total_debit` >= 0
    AND `instruction_count` >= 0
    AND `success_count` >= 0
    AND `success_count` <= `instruction_count`
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权结算批次';

DROP TABLE IF EXISTS `t_option_settlement_detail`;
CREATE TABLE `t_option_settlement_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `batch_id` BIGINT NOT NULL DEFAULT 0 COMMENT '结算批次ID',
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '结算批次号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '持仓ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `side` TINYINT NOT NULL DEFAULT 0 COMMENT '持仓方向：1多头 2空头',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '结算数量',
  `payoff` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '绝对结算金额',
  `direction` TINYINT NOT NULL DEFAULT 0 COMMENT '方向：1应收 2应付 3放弃',
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '关联资产指令号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_detail_position` (`tenant_id`, `batch_id`, `position_id`),
  KEY `idx_settlement_detail_user` (`tenant_id`, `user_id`, `account_id`, `id`),
  CONSTRAINT `chk_option_settlement_detail` CHECK (
    `side` IN (1,2)
    AND `quantity` >= 0
    AND `payoff` >= 0
    AND `direction` IN (1,2,3)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权逐持仓结算明细';

-- =========================================================
-- 14. Option 与 Asset 对账差异
-- =========================================================
DROP TABLE IF EXISTS `t_option_reconciliation_issue`;
CREATE TABLE `t_option_reconciliation_issue` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `issue_key` VARCHAR(160) NOT NULL DEFAULT '' COMMENT '稳定差异键',
  `check_type` TINYINT NOT NULL DEFAULT 0 COMMENT '检查类型：1资产流水 2余额镜像 3结算守恒',
  `biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '业务单号',
  `instruction_id` BIGINT NOT NULL DEFAULT 0 COMMENT '资产指令ID',
  `expected_value` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '预期摘要',
  `actual_value` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '实际摘要',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '差异详情',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2已恢复 3人工忽略',
  `occurrence_count` BIGINT NOT NULL DEFAULT 1 COMMENT '累计发现次数',
  `resolved_at` BIGINT NOT NULL DEFAULT 0 COMMENT '恢复时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_issue` (`tenant_id`, `issue_key`),
  KEY `idx_option_reconciliation_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_reconciliation_instruction` (`tenant_id`, `instruction_id`),
  CONSTRAINT `chk_option_reconciliation_issue` CHECK (
    `check_type` IN (1,2,3)
    AND `status` IN (1,2,3)
    AND `occurrence_count` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option与Asset对账差异';

SET FOREIGN_KEY_CHECKS = 1;
