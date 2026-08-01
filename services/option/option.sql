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
  `underlying_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '实物交割标的币种，如 BTC',
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
  `exercise_cutoff_time` BIGINT NOT NULL DEFAULT 0 COMMENT '主动行权及到期指令截止时间',
  `auto_exercise_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '每单位内在价值自动行权阈值',
  `max_user_long_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '单用户跨账户多头持仓及开仓委托上限，0表示未配置',
  `max_user_short_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '单用户跨账户空头持仓及开仓委托上限，0表示未配置',
  `max_open_interest` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '合约单边持仓及开仓委托上限，0表示未配置',
  `order_price_band_ratio` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '相对新鲜标记价的下单价格带比例，0表示未配置',
  `circuit_breaker_ratio` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '标记价相对前值跳变熔断比例，0表示未配置',
  `greeks_max_age_seconds` BIGINT NOT NULL DEFAULT 0 COMMENT 'IV与Greeks最大允许陈旧秒数，0表示未审批/未配置',
  `settlement_price_source` VARCHAR(32) NOT NULL DEFAULT 'authoritative-market' COMMENT '最终结算价来源',
  `settlement_price_method` VARCHAR(16) NOT NULL DEFAULT 'MEDIAN' COMMENT '最终结算价算法',
  `settlement_window_seconds` INT NOT NULL DEFAULT 60 COMMENT '到期前取价窗口秒数',
  `settlement_min_samples` INT NOT NULL DEFAULT 3 COMMENT '最终结算价最小样本数',
  `is_auto_exercise` TINYINT NOT NULL DEFAULT 2 COMMENT '是否自动行权：1是 2否',
  `maker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Maker成交手续费率',
  `taker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT 'Taker成交手续费率',
  `exercise_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '行权手续费率',
  `fee_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '平台手续费归集用户ID',
  `fee_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '平台手续费归集Option账户ID',
  `seller_margin_mode` TINYINT NOT NULL DEFAULT 1 COMMENT '卖方保证金模式：1关闭 2逐仓 3组合 4实物全额担保',
  `initial_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方初始保证金率',
  `maintenance_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方维持保证金率',
  `min_margin_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '卖方最低保证金率',
  `liquidation_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '强平手续费率',
  `insurance_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金用户ID',
  `insurance_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保险基金Option账户ID',
  `liquidation_deficit_policy` TINYINT NOT NULL DEFAULT 1 COMMENT '保险不足策略：1人工 2平台兜底',
  `physical_delivery_policy` TINYINT NOT NULL DEFAULT 0 COMMENT '实物交割策略：0不适用 1严格全额交收',
  `physical_delivery_cure_seconds` BIGINT NOT NULL DEFAULT 0 COMMENT '实物交割补资期限秒数，现金合约为0',
  `trading_calendar_code` VARCHAR(64) NOT NULL DEFAULT 'CONTINUOUS_24_7' COMMENT '不可变交易日历代码，运行时解析已批准版本',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1待上市 2可交易 3暂停交易 4已到期 5已结算 6已下线',
  `sort` INT NOT NULL DEFAULT 0 COMMENT '排序值',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `is_deleted` TINYINT NOT NULL DEFAULT 2 COMMENT '是否删除：1是 2否',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract_code` (`tenant_id`, `contract_code`),
  KEY `idx_tenant_underlying_symbol` (`tenant_id`, `underlying_symbol`),
  KEY `idx_option_public_chain` (`tenant_id`, `underlying_symbol`, `expire_time`, `status`, `is_deleted`, `strike_price`, `option_type`, `id`),
  KEY `idx_tenant_expire_time` (`tenant_id`, `expire_time`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_option_contract_monitor` (`status`, `update_times`, `tenant_id`, `id`),
  KEY `idx_option_contract_lifecycle_monitor` (`status`, `expire_time`, `tenant_id`, `id`),
  KEY `idx_option_public_chain_monitor` (`status`, `is_deleted`, `tenant_id`, `underlying_symbol`, `expire_time`, `strike_price`, `option_type`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权合约表';

-- =========================================================
-- 1.1 版本化交易日历、会话、例外和临时休市
-- =========================================================
DROP TABLE IF EXISTS `t_option_trading_halt`;
DROP TABLE IF EXISTS `t_option_trading_calendar_exception`;
DROP TABLE IF EXISTS `t_option_trading_calendar_session`;
DROP TABLE IF EXISTS `t_option_trading_calendar`;

CREATE TABLE `t_option_trading_calendar` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `calendar_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '稳定日历代码',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '同代码递增版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1草案 2已批准 3已拒绝 4已替代',
  `timezone` VARCHAR(64) NOT NULL DEFAULT 'UTC' COMMENT 'IANA时区',
  `effective_from` BIGINT NOT NULL DEFAULT 0 COMMENT 'UTC生效秒',
  `effective_until` BIGINT NOT NULL DEFAULT 0 COMMENT 'UTC失效秒，0表示无上限',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '替代的上一版本ID',
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '变更原因',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '节假日来源/证据引用',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '复核管理员',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '复核意见',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '复核时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_calendar_version` (`tenant_id`,`calendar_code`,`version`),
  KEY `idx_trading_calendar_effective` (`tenant_id`,`calendar_code`,`status`,`effective_from`,`effective_until`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可覆盖的期权交易日历版本';

CREATE TABLE `t_option_trading_calendar_session` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `calendar_id` BIGINT NOT NULL DEFAULT 0 COMMENT '日历版本ID',
  `weekday` TINYINT NOT NULL DEFAULT 0 COMMENT '0周日到6周六',
  `open_second` INT NOT NULL DEFAULT 0 COMMENT '本地会话日起始秒',
  `close_second` INT NOT NULL DEFAULT 0 COMMENT '本地会话日结束秒，可超过86400表示跨午夜',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_calendar_session` (`tenant_id`,`calendar_id`,`weekday`,`open_second`,`close_second`),
  KEY `idx_trading_calendar_session` (`tenant_id`,`calendar_id`,`weekday`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易日历周会话';

CREATE TABLE `t_option_trading_calendar_exception` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `calendar_id` BIGINT NOT NULL DEFAULT 0 COMMENT '日历版本ID',
  `exception_type` TINYINT NOT NULL DEFAULT 0 COMMENT '1休市 2特别开市',
  `start_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'UTC开始秒，包含',
  `end_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'UTC结束秒，不包含',
  `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '原因',
  `announcement_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '公告引用',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_trading_calendar_exception` (`tenant_id`,`calendar_id`,`start_time`,`end_time`,`exception_type`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易日历节假日与特别开市窗口';

CREATE TABLE `t_option_trading_halt` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `halt_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '临时休市幂等编号',
  `active_key` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '活动时为CONTRACT:id，解除后为HALT:halt_no',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `source` TINYINT NOT NULL DEFAULT 1 COMMENT '1人工 2熔断 3系统',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1活动 2已解除',
  `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '休市原因',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '证据引用',
  `started_at` BIGINT NOT NULL DEFAULT 0 COMMENT '生效时间',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `cancel_total` BIGINT NOT NULL DEFAULT 0 COMMENT '发现活动订单',
  `cancel_success` BIGINT NOT NULL DEFAULT 0 COMMENT '已进入撤销/已撤订单',
  `cancel_failed` BIGINT NOT NULL DEFAULT 0 COMMENT '撤销失败订单',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `lifted_at` BIGINT NOT NULL DEFAULT 0 COMMENT '解除时间',
  `lifted_by` BIGINT NOT NULL DEFAULT 0 COMMENT '解除管理员',
  `lift_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '解除原因',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_halt_no` (`tenant_id`,`halt_no`),
  UNIQUE KEY `uk_trading_halt_active` (`tenant_id`,`active_key`),
  KEY `idx_trading_halt_contract` (`tenant_id`,`contract_id`,`status`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权临时休市与恢复审计';

CREATE TABLE `t_option_corporate_action` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `event_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '内部事件编号',
  `external_event_ref` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '外部事件唯一引用',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '同一外部事件版本',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '受影响标的',
  `action_type` TINYINT NOT NULL DEFAULT 0 COMMENT '事件类型',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1草案 2批准 3拒绝 4执行中 5完成 6人工处理 7失败',
  `announcement_time` BIGINT NOT NULL DEFAULT 0 COMMENT '公告时间',
  `ex_time` BIGINT NOT NULL DEFAULT 0 COMMENT '除权/除息时间',
  `record_time` BIGINT NOT NULL DEFAULT 0 COMMENT '登记时间',
  `effective_time` BIGINT NOT NULL DEFAULT 0 COMMENT '调整生效时间',
  `pay_time` BIGINT NOT NULL DEFAULT 0 COMMENT '支付时间',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '官方证据链接/文号',
  `evidence_hash` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '证据内容哈希',
  `description` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '事件与适用规则说明',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '复核管理员',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '复核说明',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '复核时间',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_event_no` (`tenant_id`,`event_no`),
  UNIQUE KEY `uk_corporate_action_external_version` (`tenant_id`,`external_event_ref`,`version`),
  KEY `idx_corporate_action_due` (`tenant_id`,`status`,`effective_time`,`id`),
  KEY `idx_corporate_action_underlying` (`tenant_id`,`underlying_symbol`,`id`),
  KEY `idx_corporate_action_monitor` (`status`,`effective_time`,`tenant_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可覆盖的期权公司行动事件版本';

CREATE TABLE `t_option_corporate_action_contract` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `action_id` BIGINT NOT NULL DEFAULT 0 COMMENT '公司行动ID',
  `source_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被替代合约',
  `successor_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '后继合约；人工事件可为0',
  `execution_mode` TINYINT NOT NULL DEFAULT 0 COMMENT '1自动单一现金后继 2仅人工',
  `quantity_numerator` DECIMAL(32,0) NOT NULL DEFAULT 1 COMMENT '数量换算分子',
  `quantity_denominator` DECIMAL(32,0) NOT NULL DEFAULT 1 COMMENT '数量换算分母',
  `halt_id` BIGINT NOT NULL DEFAULT 0 COMMENT '创建事件时产生的停牌记录',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1已停牌 2就绪 3执行中 4完成 5人工处理 6失败',
  `position_total` BIGINT NOT NULL DEFAULT 0 COMMENT '待迁移持仓总数',
  `position_completed` BIGINT NOT NULL DEFAULT 0 COMMENT '已迁移持仓数',
  `position_failed` BIGINT NOT NULL DEFAULT 0 COMMENT '失败持仓数',
  `last_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '批处理游标',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_source` (`tenant_id`,`action_id`,`source_contract_id`),
  KEY `idx_corporate_action_contract_status` (`tenant_id`,`action_id`,`status`,`id`),
  KEY `idx_corporate_action_contract_monitor` (`status`,`tenant_id`,`action_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动受影响合约及后继映射';

CREATE TABLE `t_option_corporate_action_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `action_id` BIGINT NOT NULL DEFAULT 0 COMMENT '公司行动ID',
  `action_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '公司行动合约映射ID',
  `source_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原持仓ID',
  `successor_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '后继持仓ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '账户ID',
  `side` TINYINT NOT NULL DEFAULT 0 COMMENT '方向',
  `source_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '原数量',
  `successor_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '后继数量',
  `source_available_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '原可用数量',
  `successor_available_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '后继可用数量',
  `source_open_avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '原成本价',
  `successor_open_avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '后继成本价',
  `source_effective_multiplier` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '原有效乘数',
  `successor_effective_multiplier` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '后继有效乘数',
  `cost_basis_before` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '迁移前成本基数',
  `cost_basis_after` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '迁移后成本基数',
  `cash_difference` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '显式现金差额；V1自动路径必须为0',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1待执行 2完成 3失败',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后错误',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_position` (`tenant_id`,`action_contract_id`,`source_position_id`),
  KEY `idx_corporate_action_position_status` (`tenant_id`,`action_contract_id`,`status`,`source_position_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动逐持仓不可覆盖迁移审计';

CREATE TABLE `t_option_corporate_action_margin_lot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `action_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '逐持仓迁移审计ID',
  `margin_lot_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被重新指向的保证金批次ID',
  `source_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原合约ID',
  `successor_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '后继合约ID',
  `source_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原持仓ID',
  `successor_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '后继持仓ID',
  `source_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '批次原数量',
  `successor_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '批次后继数量',
  `source_remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '批次原剩余数量',
  `successor_remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '批次后继剩余数量',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_margin_lot` (`tenant_id`,`action_position_id`,`margin_lot_id`),
  KEY `idx_corporate_action_margin_lot` (`tenant_id`,`margin_lot_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动保证金批次换算审计';


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
  `snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '兼容字段：行情记录最后快照时间',
  `underlying_snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '标的价格快照时间',
  `mark_snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '期权标记价格快照时间',
  `greeks_snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT 'IV与Greeks快照时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_tenant_snapshot_time` (`tenant_id`, `snapshot_time`),
  KEY `idx_tenant_underlying_time` (`tenant_id`, `underlying_snapshot_time`),
  KEY `idx_tenant_mark_time` (`tenant_id`, `mark_snapshot_time`)
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
  `source_type` TINYINT NOT NULL DEFAULT 0 COMMENT '来源：0未知 1权威标的行情 2管理行情 3结算审计',
  `source_snapshot_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '来源快照唯一标识',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_contract_snapshot_time` (`tenant_id`, `contract_id`, `snapshot_time`),
  KEY `idx_contract_source_time` (`tenant_id`, `contract_id`, `source_type`, `snapshot_time`),
  KEY `idx_option_settlement_snapshot_evidence`
    (`tenant_id`, `contract_id`, `source_type`, `source_snapshot_id`, `snapshot_time`, `id`)
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
DROP TABLE IF EXISTS `t_option_combo_order_leg`;
DROP TABLE IF EXISTS `t_option_combo_order`;
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
  `margin_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '订单冻结资产币种',
  `portfolio_risk_config_id` BIGINT NOT NULL DEFAULT 0 COMMENT '组合保证金准入采用的参数ID；非组合保证金卖单或迁移前历史单为0',
  `portfolio_risk_config_version` BIGINT NOT NULL DEFAULT 0 COMMENT '组合保证金准入采用的参数版本；与参数ID成对保存',
  `source` TINYINT NOT NULL DEFAULT 0 COMMENT '订单来源：1APP 2WEB 3API 4ADMIN',
  `client_order_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端订单号',
  `reduce_only` TINYINT NOT NULL DEFAULT 2 COMMENT '是否只减仓：1是 2否',
  `mmp` TINYINT NOT NULL DEFAULT 2 COMMENT '是否做市商保护单：1是 2否',
  `mmp_group` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'MMP报价组，mmp=1时必填',
  `combo_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '组合父单ID；0为普通单',
  `combo_leg_no` BIGINT NOT NULL DEFAULT 0 COMMENT '组合腿序号；普通单为0',
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
  KEY `idx_option_public_book` (`tenant_id`, `contract_id`, `status`, `side`, `price`, `id`),
  KEY `idx_option_combo_child` (`tenant_id`, `combo_order_id`, `combo_leg_no`, `id`),
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
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权客户端订单幂等键';

CREATE TABLE `t_option_combo_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `combo_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '组合父单号',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '业务归属账户ID',
  `client_combo_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端组合幂等号',
  `strategy_key` CHAR(64) NOT NULL DEFAULT '' COMMENT '规范化实际方向策略SHA-256',
  `inverse_strategy_key` CHAR(64) NOT NULL DEFAULT '' COMMENT '完全反向策略SHA-256',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '共同标的',
  `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '共同到期时间',
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '共同结算币',
  `quote_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '共同报价币',
  `order_type` TINYINT NOT NULL DEFAULT 0 COMMENT '1 LIMIT 2 FOK',
  `net_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '买腿减卖腿的每份组合净价，可为负',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '父单组合份数',
  `filled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '已成交组合份数',
  `unfilled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未成交组合份数',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '1资金中 2活动 3部分成交 4完成 5撤销中 6已撤销 7拒绝 8人工',
  `payload_hash` CHAR(64) NOT NULL DEFAULT '' COMMENT '规范化请求SHA-256',
  `cancel_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '撤销/拒绝原因',
  `cancel_time` BIGINT NOT NULL DEFAULT 0 COMMENT '撤销时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_combo_no` (`tenant_id`,`combo_no`),
  UNIQUE KEY `uk_option_combo_client` (`tenant_id`,`user_id`,`client_combo_id`),
  KEY `idx_option_combo_match` (`tenant_id`,`strategy_key`,`status`,`net_price`,`id`),
  KEY `idx_option_combo_user` (`tenant_id`,`user_id`,`account_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权组合父单';

CREATE TABLE `t_option_combo_order_leg` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `combo_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '组合父单ID',
  `leg_no` BIGINT NOT NULL DEFAULT 0 COMMENT '稳定腿序号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `side` TINYINT NOT NULL DEFAULT 0 COMMENT '1买 2卖',
  `position_effect` TINYINT NOT NULL DEFAULT 1 COMMENT '首版仅1开仓',
  `ratio` BIGINT NOT NULL DEFAULT 0 COMMENT '已约分整数比例1至8',
  `price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '腿限价',
  `qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '父单数量乘比例',
  `filled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '腿已成交量',
  `unfilled_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '腿未成交量',
  `child_order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '影子子单ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_combo_leg_no` (`tenant_id`,`combo_order_id`,`leg_no`),
  UNIQUE KEY `uk_option_combo_leg_contract` (`tenant_id`,`combo_order_id`,`contract_id`),
  UNIQUE KEY `uk_option_combo_leg_child` (`tenant_id`,`child_order_id`),
  KEY `idx_option_combo_leg_contract` (`tenant_id`,`contract_id`,`combo_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权组合不可变腿及影子子单';


-- =========================================================
-- 5. 期权成交表
-- =========================================================
DROP TABLE IF EXISTS `t_option_trade`;
CREATE TABLE `t_option_trade` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `trade_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '成交号',
  `combo_match_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '组合成交组号；普通成交为空',
  `combo_leg_no` BIGINT NOT NULL DEFAULT 0 COMMENT '组合成交腿序号；普通成交为0',
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
  KEY `idx_option_combo_match_trade` (`tenant_id`,`combo_match_no`,`combo_leg_no`,`id`),
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
  UNIQUE KEY `uk_tenant_contract` (`tenant_id`, `contract_id`)
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
  KEY `idx_outbox_retry` (`status`, `next_retry_at`, `id`)
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
  KEY `idx_inbox_contract_sequence` (`tenant_id`, `contract_id`, `match_sequence`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权事件消费幂等箱';

CREATE TABLE `t_option_margin_lot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方Option账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头持仓ID，事件入账后回填',
  `origin_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '首次建批次时的合约ID，公司行动后保持不变',
  `origin_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '首次入账时的持仓ID，公司行动后保持不变',
  `corporate_action_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '最近一次公司行动逐持仓审计ID',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖出开仓订单ID',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '成交ID',
  `freeze_biz_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT 'Asset冻结业务号',
  `collateral_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '该批次实际冻结的担保资产币种',
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
  KEY `idx_margin_lot_order` (`tenant_id`, `order_id`, `status`, `id`)
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
  KEY `idx_application_margin_lot` (`tenant_id`, `margin_lot_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='保证金批次资产指令应用幂等记录';

CREATE TABLE `t_option_risk_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Option账户ID',
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '结算币种',
  `equity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '风险权益',
  `net_option_value` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '多头市值减空头市值',
  `position_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '卖方持仓保证金',
  `maintenance_margin` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '维持保证金',
  `unrealized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '未实现盈亏',
  `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '维持保证金/权益',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2追保 3强平中 4破产 5限制',
  `portfolio_risk_method` TINYINT NOT NULL DEFAULT 0 COMMENT '组合风险算法：0无 1到期损益情景V1',
  `portfolio_risk_config_id` BIGINT NOT NULL DEFAULT 0 COMMENT '本次计算使用的组合风险参数版本ID',
  `portfolio_risk_config_version` BIGINT NOT NULL DEFAULT 0 COMMENT '本次计算使用的组合风险参数版本号',
  `portfolio_scenario_loss` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '组合价格情景最大损失',
  `portfolio_short_floor` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '组合裸空头最低保证金',
  `portfolio_concentration_addon` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '组合集中度附加保证金',
  `portfolio_liquidity_addon` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '组合流动性附加保证金',
  `last_calc_time` BIGINT NOT NULL DEFAULT 0 COMMENT '最后计算时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_account_coin` (`tenant_id`, `user_id`, `account_id`, `settle_coin`),
  KEY `idx_risk_account_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_risk_account_portfolio_monitor` (`portfolio_risk_method`, `last_calc_time`, `tenant_id`, `settle_coin`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权卖方风险账户';

CREATE TABLE `t_option_portfolio_risk_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '结算币种',
  `version` BIGINT NOT NULL DEFAULT 0 COMMENT '租户及结算币内递增版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待审批 2已批准 3已拒绝 4已替代',
  `model_method` TINYINT NOT NULL DEFAULT 1 COMMENT '组合风险算法：1到期损益情景V1',
  `initial_shock_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '初始保证金最低价格冲击率',
  `maintenance_shock_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '维持保证金最低价格冲击率',
  `scenario_shocks` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '逗号分隔的相对价格冲击集合',
  `concentration_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '同到期组合净空头风险名义金额阈值',
  `concentration_addon_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '超过集中度阈值部分的附加率',
  `liquidity_addon_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT '净空头风险名义金额流动性附加率',
  `effective_from` BIGINT NOT NULL DEFAULT 0 COMMENT '生效时间（秒）',
  `effective_until` BIGINT NOT NULL DEFAULT 0 COMMENT '失效时间（秒），0为未安排失效',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '批准后替代的上一版本ID',
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '变更或回滚原因',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '回测、验证或审批证据引用',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员ID',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '审批管理员ID',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '审批意见',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '审批时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_portfolio_risk_config_version` (`tenant_id`, `settle_coin`, `version`),
  KEY `idx_portfolio_risk_config_active` (`tenant_id`, `settle_coin`, `status`, `effective_from`, `effective_until`),
  KEY `idx_option_portfolio_config_monitor` (`status`, `effective_from`, `effective_until`, `tenant_id`, `settle_coin`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组合保证金风险模型参数版本';

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
  `backstop_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '平台兜底负债金额',
  `deficit_resolution` TINYINT NOT NULL DEFAULT 1 COMMENT '缺口处置：1无 2保险 3平台兜底 4保险加兜底 5人工',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_liquidation_no` (`tenant_id`, `liquidation_no`),
  KEY `idx_liquidation_status` (`tenant_id`, `status`, `id`),
  KEY `idx_liquidation_position` (`tenant_id`, `position_id`, `id`)
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
  KEY `idx_insurance_liquidation` (`tenant_id`, `liquidation_id`)
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
  `trade_realized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '平仓产生的权利金交易毛盈亏',
  `settlement_realized_pnl` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权/到期产生的结算毛盈亏',
  `fee_paid` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '归属于持仓的累计交易/行权/强平费用',
  `total_return` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '累计已实现总收益=交易+结算-费用',
  `exerciseable_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '可行权数量',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0未知 1持仓中 2已平仓 3已行权 4已到期 5已结算 6公司行动迁出',
  `last_calc_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上次风控计算时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_uid_account_contract_side` (`tenant_id`, `user_id`, `account_id`, `contract_id`, `side`),
  KEY `idx_tenant_contract_id` (`tenant_id`, `contract_id`),
  KEY `idx_option_position_assignment_fifo` (`tenant_id`, `contract_id`, `side`, `status`, `create_times`, `id`),
  KEY `idx_tenant_uid_account` (`tenant_id`, `user_id`, `account_id`),
  KEY `idx_option_position_monitor` (`status`, `tenant_id`, `contract_id`, `user_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权持仓表';

-- =========================================================
-- 7. 期权行权表
-- =========================================================
DROP TABLE IF EXISTS `t_option_exercise`;
CREATE TABLE `t_option_exercise` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `exercise_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '行权单号',
  `client_exercise_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端行权幂等号；用户主动行权禁止为空',
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
  UNIQUE KEY `uk_tenant_user_client_exercise` (`tenant_id`, `user_id`, `client_exercise_id`),
  KEY `idx_tenant_uid_account_contract_id` (`tenant_id`, `user_id`, `account_id`, `contract_id`),
  KEY `idx_tenant_position_id` (`tenant_id`, `position_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_option_exercise_monitor` (`status`, `tenant_id`, `contract_id`, `create_times`, `id`)
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
  KEY `idx_assignment_exercise` (`tenant_id`, `exercise_id`, `status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权主动行权空头指派';

DROP TABLE IF EXISTS `t_option_exercise_instruction`;
CREATE TABLE `t_option_exercise_instruction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头持仓ID',
  `client_instruction_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端幂等号',
  `instruction_type` TINYINT NOT NULL DEFAULT 1 COMMENT '指令：1自动 2放弃 3相反行权',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '持仓指令版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1生效 2已被新版本替代',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被本版本替代的指令ID',
  `cutoff_time` BIGINT NOT NULL DEFAULT 0 COMMENT '提交时适用的截止时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exercise_instruction_client` (`tenant_id`, `user_id`, `client_instruction_id`),
  UNIQUE KEY `uk_exercise_instruction_version` (`tenant_id`, `position_id`, `version`),
  KEY `idx_exercise_instruction_active` (`tenant_id`, `contract_id`, `position_id`, `status`, `version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权到期行权指令版本';

-- 行权指令经济字段不可变、状态单向迁移及禁止删除触发器由
-- migrations/20260730_zd_option_exercise_governance.sql 安装；
-- 新建数据库后同样必须执行该迁移。

-- =========================================================
-- 8.3 用户交易控制及审计事件
-- =========================================================
DROP TABLE IF EXISTS `t_option_user_trading_control`;
CREATE TABLE `t_option_user_trading_control` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `kill_switch` TINYINT NOT NULL DEFAULT 2 COMMENT 'kill switch：1开启 2关闭',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近启用/解除原因',
  `activated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近启用时间',
  `released_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近解除时间',
  `activated_by` BIGINT NOT NULL DEFAULT 0 COMMENT '启用操作人，用户本人等于user_id',
  `released_by` BIGINT NOT NULL DEFAULT 0 COMMENT '解除管理员ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_user_trading_control` (`tenant_id`, `user_id`),
  KEY `idx_option_user_control_monitor` (`kill_switch`, `activated_at`, `tenant_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权用户交易控制';

DROP TABLE IF EXISTS `t_option_trading_control_event`;
CREATE TABLE `t_option_trading_control_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID，合约事件可为0',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID，用户级事件可为0',
  `order_id` BIGINT NOT NULL DEFAULT 0 COMMENT '相关订单ID',
  `event_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '事件类型',
  `reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '机器可读原因',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '参数及观测值',
  `operator_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户或管理员操作人',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_option_control_event_contract` (`tenant_id`, `contract_id`, `id`),
  KEY `idx_option_control_event_user` (`tenant_id`, `user_id`, `id`),
  KEY `idx_option_control_event_reason` (`tenant_id`, `event_type`, `reason`, `id`),
  KEY `idx_option_control_event_monitor` (`event_type`, `reason`, `create_times`, `tenant_id`, `user_id`, `contract_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易控制审计事件';

-- TRADING 参数门禁和审计事件不可修改/删除触发器由
-- migrations/20260730_ze_option_trading_controls.sql 安装；
-- 新建数据库后同样必须执行该迁移。


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
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '钱包镜像固定为0，业务账户仅保留在订单/持仓/流水',
  `margin_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '保证金币种',
  `balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包总额镜像',
  `available_balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包可用额镜像',
  `frozen_balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包冻结额镜像',
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
  `change_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动金额，正负都有可能',
  `balance_before` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动前余额',
  `balance_after` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动后余额',
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
  `delivery_unit_id` BIGINT NOT NULL DEFAULT 0 COMMENT '实物交割配对单元ID',
  `execution_group` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '步骤屏障执行域；空值回退biz_no',
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
  KEY `idx_instruction_delivery_unit` (`tenant_id`, `delivery_unit_id`, `step_no`, `status`, `id`),
  KEY `idx_option_asset_instruction_control_monitor`
    (`action`, `status`, `tenant_id`, `user_id`, `create_times`, `id`)
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
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待复核 2已确认 3已拒绝 4已被新版本替代',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被本版本替代的结算价ID',
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '计算、拒绝或人工更正原因',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建人，0为系统计算',
  `confirmed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '确认人，0为系统',
  `confirmed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '确认时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_price_version` (`tenant_id`, `contract_id`, `version`),
  KEY `idx_settlement_price_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_settlement_price_monitor` (`status`, `tenant_id`, `contract_id`, `confirmed_at`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权不可变结算价快照';

-- 已确认结算价的经济字段不可变触发器由
-- migrations/20260730_z_option_settlement_price_approval.sql 安装。
-- 新建数据库不能仅导入本文件，仍须按顺序执行 migrations。

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
  KEY `idx_settlement_batch_status` (`tenant_id`, `status`, `id`)
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
  `delivery_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '实物交割标的币种',
  `delivery_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '实物交割标的数量',
  `payment_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '行权款币种',
  `payment_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权款金额',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_detail_position` (`tenant_id`, `batch_id`, `position_id`),
  KEY `idx_settlement_detail_user` (`tenant_id`, `user_id`, `account_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权逐持仓结算明细';

-- =========================================================
-- 14. 实物交割配对单元
-- =========================================================
DROP TABLE IF EXISTS `t_option_physical_delivery_unit`;
CREATE TABLE `t_option_physical_delivery_unit` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `delivery_unit_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '稳定交割单元号',
  `batch_id` BIGINT NOT NULL DEFAULT 0 COMMENT '结算批次ID',
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '结算批次号',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '期权合约ID',
  `long_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头持仓ID',
  `long_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头用户ID',
  `long_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头Option账户ID',
  `short_position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头持仓ID',
  `short_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头用户ID',
  `short_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '空头Option账户ID',
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '配对合约数量',
  `delivery_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '标的交割币',
  `delivery_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '标的交割数量',
  `payment_coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '行权款币',
  `payment_amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权款金额',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1已创建 2资产处理中 3补资 4人工 5完成 6违约',
  `cure_deadline` BIGINT NOT NULL DEFAULT 0 COMMENT '补资截止时间',
  `failed_instruction_id` BIGINT NOT NULL DEFAULT 0 COMMENT '最近失败资产指令ID',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成时间',
  `manual_retry_count` BIGINT NOT NULL DEFAULT 0 COMMENT '逾期后人工重试代次',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_physical_delivery_unit_no` (`tenant_id`, `delivery_unit_no`),
  KEY `idx_physical_delivery_batch` (`tenant_id`, `batch_id`, `id`),
  KEY `idx_physical_delivery_status` (`tenant_id`, `status`, `cure_deadline`, `id`),
  KEY `idx_option_physical_delivery_monitor` (`status`, `cure_deadline`, `tenant_id`, `id`),
  KEY `idx_physical_delivery_long` (`tenant_id`, `long_user_id`, `long_position_id`, `id`),
  KEY `idx_physical_delivery_short` (`tenant_id`, `short_user_id`, `short_position_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='实物交割多空配对执行单元';


-- =========================================================
-- 15. Option 与 Asset 对账差异
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
  KEY `idx_option_reconciliation_instruction` (`tenant_id`, `instruction_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option与Asset对账差异';

DROP TABLE IF EXISTS `t_option_reconciliation_run_detail`;
DROP TABLE IF EXISTS `t_option_reconciliation_run`;
CREATE TABLE `t_option_reconciliation_run` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `business_date` CHAR(10) NOT NULL DEFAULT '' COMMENT 'UTC业务日期YYYY-MM-DD',
  `scope` TINYINT NOT NULL DEFAULT 1 COMMENT '范围：1钱包镜像 2完整资金守恒',
  `attempt_no` INT NOT NULL DEFAULT 1 COMMENT '当日同范围执行序号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1成功 2有差异 3执行失败',
  `snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '一致性快照Unix秒',
  `snapshot_ref` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '快照或证据引用',
  `coin_count` BIGINT NOT NULL DEFAULT 0 COMMENT '检查币种数',
  `account_count` BIGINT NOT NULL DEFAULT 0 COMMENT '检查用户钱包数',
  `mismatch_count` BIGINT NOT NULL DEFAULT 0 COMMENT '差异用户钱包数',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '执行摘要或失败原因',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成Unix秒',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_run_attempt` (`tenant_id`,`business_date`,`scope`,`attempt_no`),
  UNIQUE KEY `uk_option_reconciliation_run_identity` (`id`,`tenant_id`,`business_date`,`scope`),
  KEY `idx_option_reconciliation_run_latest` (`tenant_id`,`scope`,`completed_at`,`status`,`id`),
  KEY `idx_option_reconciliation_run_monitor` (`scope`,`status`,`completed_at`,`tenant_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option日终对账运行与成功心跳';

CREATE TABLE `t_option_reconciliation_run_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `run_id` BIGINT NOT NULL COMMENT '不可变对账运行ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID，冗余用于租户隔离查询',
  `business_date` CHAR(10) NOT NULL COMMENT 'UTC业务日期YYYY-MM-DD',
  `scope` TINYINT NOT NULL DEFAULT 2 COMMENT '范围，明细当前仅允许完整资金守恒scope=2',
  `dimension_type` TINYINT NOT NULL COMMENT '维度：1用户钱包逐币 2平台账户逐币 3Option子账逐币',
  `dimension_key` VARCHAR(96) NOT NULL COMMENT 'coin或account_type:coin稳定键',
  `opening_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '业务日期期初总额',
  `external_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '充值提现及跨钱包划转净额',
  `option_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Option业务Asset流水净额',
  `manual_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '人工增减净额',
  `expected_closing` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '期初加三类净额',
  `actual_closing` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '由一致性截止点反推的实际日终',
  `difference_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '实际日终减预期日终',
  `flow_count` BIGINT NOT NULL DEFAULT 0 COMMENT '纳入计算的权威流水数',
  `mismatch_count` BIGINT NOT NULL DEFAULT 0 COMMENT '余额链、字段恒等式或子账关联异常数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1一致 2有差异 3数据不完整',
  `evidence_ref` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '游标、查询或快照证据引用',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '分类和完整性摘要',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_detail_dimension` (`run_id`,`dimension_type`,`dimension_key`),
  KEY `idx_option_reconciliation_detail_lookup` (`tenant_id`,`business_date`,`scope`,`status`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option完整资金守恒逐维度不可变证据';

-- =========================================================
-- 15. 异常成交现金更正案件与不可变资金分录
-- =========================================================
DROP TABLE IF EXISTS `t_option_trade_correction_leg`;
DROP TABLE IF EXISTS `t_option_trade_correction`;
CREATE TABLE `t_option_trade_correction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `case_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '案件号及资金业务号',
  `trade_id` BIGINT NOT NULL DEFAULT 0 COMMENT '原始成交ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `action` TINYINT NOT NULL DEFAULT 1 COMMENT '处置：1现金更正',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待复核 2拒绝 3执行中 4完成 5人工处理',
  `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '异常成交判定及更正依据',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '不可变证据包引用',
  `requested_by` BIGINT NOT NULL DEFAULT 0 COMMENT '申请管理员ID',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '复核管理员ID',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '复核意见',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '复核时间',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '资金更正完成时间',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后执行错误',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_trade_correction_case` (`tenant_id`, `case_no`),
  KEY `idx_option_trade_correction_trade` (`tenant_id`, `trade_id`, `id`),
  KEY `idx_option_trade_correction_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_trade_correction_monitor` (`status`, `update_times`, `tenant_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异常成交现金更正案件';

CREATE TABLE `t_option_trade_correction_leg` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `correction_id` BIGINT NOT NULL DEFAULT 0 COMMENT '更正案件ID',
  `leg_no` BIGINT NOT NULL DEFAULT 0 COMMENT '案件内分录序号',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '资金用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '业务账户ID',
  `coin` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '资金币种',
  `direction` TINYINT NOT NULL DEFAULT 0 COMMENT '方向：1扣可用 2加可用',
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '正数金额',
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '审批后使用的幂等资金指令号',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_trade_correction_leg` (`tenant_id`, `correction_id`, `leg_no`),
  UNIQUE KEY `uk_option_trade_correction_instruction` (`tenant_id`, `instruction_no`),
  KEY `idx_option_trade_correction_leg_user` (`tenant_id`, `user_id`, `account_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异常成交不可变现金更正分录';

-- 案件经济字段、分录及合法状态迁移触发器由
-- migrations/20260730_zf_option_trade_correction.sql 安装。

-- =========================================================
-- 16. 做市商保护（MMP）配置及滚动窗口状态
-- =========================================================
DROP TABLE IF EXISTS `t_option_mmp_config`;
CREATE TABLE `t_option_mmp_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '做市用户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `group_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '报价组',
  `enabled` TINYINT NOT NULL DEFAULT 2 COMMENT '1启用 2禁用',
  `qty_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '窗口内maker成交量阈值，0禁用',
  `trade_count_threshold` BIGINT NOT NULL DEFAULT 0 COMMENT '窗口内maker成交笔数阈值，0禁用',
  `loss_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '窗口内相对标记价不利成交及手续费阈值，0禁用',
  `window_seconds` BIGINT NOT NULL DEFAULT 0 COMMENT '滚动窗口秒数',
  `cooldown_seconds` BIGINT NOT NULL DEFAULT 0 COMMENT '触发后冷静期秒数',
  `status` TINYINT NOT NULL DEFAULT 3 COMMENT '状态：1活动 2已触发 3禁用',
  `window_start` BIGINT NOT NULL DEFAULT 0 COMMENT '当前窗口开始',
  `accumulated_qty` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '窗口累计maker成交量',
  `trade_count` BIGINT NOT NULL DEFAULT 0 COMMENT '窗口累计成交笔数',
  `accumulated_loss` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '窗口累计不利成交损失',
  `triggered_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近触发时间',
  `cooldown_until` BIGINT NOT NULL DEFAULT 0 COMMENT '冷静期截止时间',
  `trigger_reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '机器触发原因',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '自动撤单最后错误',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `updated_by` BIGINT NOT NULL DEFAULT 0 COMMENT '最近更新/恢复管理员',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_mmp_config` (`tenant_id`, `user_id`, `contract_id`, `group_code`),
  KEY `idx_option_mmp_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_mmp_contract` (`tenant_id`, `contract_id`, `id`),
  KEY `idx_option_mmp_monitor` (`status`, `triggered_at`, `update_times`, `tenant_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权做市商保护配置与实时窗口';

-- MMP 配置门禁触发器由 migrations/20260730_zg_option_mmp.sql 安装。

-- =========================================================
-- 17. 合约到期/行权价系列、审批与生成谱系
-- =========================================================
DROP TABLE IF EXISTS `t_option_contract_series_detail`;
DROP TABLE IF EXISTS `t_option_contract_series_strike_band`;
DROP TABLE IF EXISTS `t_option_contract_series_expiry`;
DROP TABLE IF EXISTS `t_option_contract_series`;
CREATE TABLE `t_option_contract_series` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `request_key` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '客户端幂等键',
  `series_code` VARCHAR(24) NOT NULL DEFAULT '' COMMENT '系列代码',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '追加版本号',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被替代系列版本ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待复核 2已生成 3已拒绝',
  `template_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保留字段；V1使用内嵌参数快照',
  `template_snapshot` JSON NOT NULL COMMENT '创建时的完整模板经济参数快照',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板标的',
  `reference_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '审批参考价快照',
  `reference_source` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '权威参考价来源',
  `reference_time` BIGINT NOT NULL DEFAULT 0 COMMENT '参考价时间',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '不可变审批证据引用',
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '创建/修订原因',
  `payload_hash` CHAR(64) NOT NULL DEFAULT '' COMMENT '规范化输入SHA-256',
  `expected_contract_count` BIGINT NOT NULL DEFAULT 0 COMMENT '预计生成合约数',
  `generated_contract_count` BIGINT NOT NULL DEFAULT 0 COMMENT '实际生成合约数',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '复核管理员',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '复核意见',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '复核时间',
  `generated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '原子生成完成时间',
  `launch_status` TINYINT NOT NULL DEFAULT 0 COMMENT '上市复核：0不适用 1待复核 2批准 3拒绝',
  `launch_reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '上市复核管理员',
  `launch_review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '上市复核意见',
  `launch_reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '上市复核时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_request` (`tenant_id`, `request_key`),
  UNIQUE KEY `uk_option_contract_series_version` (`tenant_id`, `series_code`, `version`),
  KEY `idx_option_contract_series_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_contract_series_monitor` (`status`, `create_times`, `tenant_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可变合约系列版本及审批';

CREATE TABLE `t_option_contract_series_expiry` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `sequence_no` BIGINT NOT NULL DEFAULT 0 COMMENT '稳定到期序号',
  `cycle_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '运营确认的周期标签',
  `list_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上市时间',
  `exercise_cutoff_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行权指令截止时间',
  `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '到期/最后交易时间',
  `deliver_time` BIGINT NOT NULL DEFAULT 0 COMMENT '交割时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_expiry_seq` (`tenant_id`, `series_id`, `sequence_no`),
  UNIQUE KEY `uk_option_contract_series_expiry_time` (`tenant_id`, `series_id`, `expire_time`),
  KEY `idx_option_contract_series_expiry` (`tenant_id`, `series_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列显式到期规格';

CREATE TABLE `t_option_contract_series_strike_band` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `sequence_no` BIGINT NOT NULL DEFAULT 0 COMMENT '稳定梯度序号',
  `lower_strike` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '闭区间下界',
  `upper_strike` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '闭区间上界',
  `strike_step` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '精确步长',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_band_seq` (`tenant_id`, `series_id`, `sequence_no`),
  KEY `idx_option_contract_series_band` (`tenant_id`, `series_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列绝对行权价梯度';

CREATE TABLE `t_option_contract_series_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `expiry_id` BIGINT NOT NULL DEFAULT 0 COMMENT '到期规格ID',
  `option_type` TINYINT NOT NULL DEFAULT 0 COMMENT '1 Call 2 Put',
  `strike_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权价',
  `contract_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '确定性合约代码',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '生成的PENDING合约ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_detail_leg` (`tenant_id`, `series_id`, `expiry_id`, `option_type`, `strike_price`),
  UNIQUE KEY `uk_option_contract_series_detail_code` (`tenant_id`, `contract_code`),
  UNIQUE KEY `uk_option_contract_series_detail_contract` (`tenant_id`, `contract_id`),
  KEY `idx_option_contract_series_detail` (`tenant_id`, `series_id`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列生成合约不可变谱系';

-- 系列经济字段、子规格、谱系和合法状态迁移触发器由
-- migrations/20260731_zn_option_contract_series.sql 安装。

SET FOREIGN_KEY_CHECKS = 1;
