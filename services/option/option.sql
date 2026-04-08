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
  `is_auto_exercise` TINYINT NOT NULL DEFAULT 1 COMMENT '是否自动行权：0否 1是',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：1待上市 2可交易 3暂停交易 4已到期 5已结算 6已下线',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除：0否 1是',
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

-- =========================================================
-- 4. 期权委托表
-- =========================================================
DROP TABLE IF EXISTS `t_option_order`;
CREATE TABLE `t_option_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '订单号',
  `uid` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
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
  `reduce_only` TINYINT NOT NULL DEFAULT 0 COMMENT '是否只减仓：0否 1是',
  `mmp` TINYINT NOT NULL DEFAULT 0 COMMENT '是否做市商保护单：0否 1是',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：1待撮合 2部分成交 3完全成交 4已撤单 5拒单 6已过期',
  `cancel_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '撤单/拒单原因',
  `match_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后成交时间',
  `cancel_time` BIGINT NOT NULL DEFAULT 0 COMMENT '撤单时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_order_no` (`tenant_id`, `order_no`),
  UNIQUE KEY `uk_tenant_uid_client_order_id` (`tenant_id`, `uid`, `client_order_id`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `uid`, `account_id`),
  KEY `idx_tenant_uid_contract_id` (`tenant_id`, `uid`, `contract_id`),
  KEY `idx_tenant_contract_id_status` (`tenant_id`, `contract_id`, `status`),
  KEY `idx_tenant_create_times` (`tenant_id`, `create_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权委托表';

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
  `buy_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '买方用户ID',
  `buy_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '买方账户ID',
  `sell_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖单ID',
  `sell_order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '卖单订单号',
  `sell_uid` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方用户ID',
  `sell_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方账户ID',
  `price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交价格/权利金',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交数量',
  `turnover` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '成交额',
  `buy_fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '买方手续费',
  `sell_fee` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖方手续费',
  `fee_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '手续费币种',
  `maker_side` TINYINT NOT NULL DEFAULT 0 COMMENT 'maker方向：1买 2卖',
  `trade_time` BIGINT NOT NULL DEFAULT 0 COMMENT '成交时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_trade_no` (`tenant_id`, `trade_no`),
  KEY `idx_tenant_buy_uid` (`tenant_id`, `buy_uid`),
  KEY `idx_tenant_sell_uid` (`tenant_id`, `sell_uid`),
  KEY `idx_tenant_contract_trade_time` (`tenant_id`, `contract_id`, `trade_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权成交表';

-- =========================================================
-- 6. 期权持仓表
-- =========================================================
DROP TABLE IF EXISTS `t_option_position`;
CREATE TABLE `t_option_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `uid` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
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
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：1持仓中 2已平仓 3已行权 4已到期 5已结算',
  `last_calc_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上次风控计算时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_uid_account_contract_side` (`tenant_id`, `uid`, `account_id`, `contract_id`, `side`),
  KEY `idx_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `uid`, `account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权持仓表';

-- =========================================================
-- 7. 期权行权表
-- =========================================================
DROP TABLE IF EXISTS `t_option_exercise`;
CREATE TABLE `t_option_exercise` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `exercise_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '行权单号',
  `uid` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
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
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：1待处理 2已执行 3已拒绝 4已取消',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `exercise_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行权时间',
  `finish_time` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_exercise_no` (`tenant_id`, `exercise_no`),
  KEY `idx_tenant_uid_account_contract_id` (`tenant_id`, `uid`, `account_id`, `contract_id`),
  KEY `idx_tenant_position_id` (`tenant_id`, `position_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权行权表';

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
  `is_itm` TINYINT NOT NULL DEFAULT 0 COMMENT '是否实值：0否 1是',
  `exercise_result` TINYINT NOT NULL DEFAULT 0 COMMENT '行权结果：0未执行 1自动行权 2自动放弃',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：1待结算 2结算中 3已完成 4失败',
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
  `uid` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
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
  UNIQUE KEY `uk_tenant_uid_account_margin_coin` (`tenant_id`, `uid`, `account_id`, `margin_coin`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `uid`, `account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权账户资产表';

-- =========================================================
-- 10. 期权资金流水表
-- =========================================================
DROP TABLE IF EXISTS `t_option_bill`;
CREATE TABLE `t_option_bill` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `uid` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
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
  KEY `idx_tenant_uid_account` (`tenant_id`, `uid`, `account_id`),
  KEY `idx_tenant_ref_type_ref_id` (`tenant_id`, `ref_type`, `ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权资金流水表';

SET FOREIGN_KEY_CHECKS = 1;