SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_stake_product
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_product`;
CREATE TABLE `t_stake_product` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `product_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押产品编号',
  `product_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '质押产品名称',
  `product_type` TINYINT NOT NULL DEFAULT '1' COMMENT '产品类型：1活期 2定期',
  `coin_name` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '质押币种名称',
  `coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '质押币种符号',
  `reward_coin_name` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '收益币种名称',
  `reward_coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '收益币种符号',
  `apr` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '年化收益率，例如12.5000表示12.5%',
  `lock_days` INT NOT NULL DEFAULT '0' COMMENT '锁仓天数，0表示活期',
  `min_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '最小质押数量',
  `max_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '最大质押数量，0表示不限制',
  `step_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '递增数量，0表示不限制步长',
  `total_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '产品总可质押数量，0表示不限制',
  `staked_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '当前已质押数量',
  `user_limit_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '单用户最大可质押数量，0表示不限制',
  `interest_mode` TINYINT NOT NULL DEFAULT '1' COMMENT '计息方式：1按天计息 2到期一次性计息',
  `reward_mode` TINYINT NOT NULL DEFAULT '1' COMMENT '发息方式：1每日发放 2到期发放',
  `allow_early_redeem` TINYINT NOT NULL DEFAULT '2' COMMENT '是否允许提前赎回：1是 2否',
  `early_redeem_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '提前赎回手续费率，例如5.0000表示5%',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '产品状态：1禁用 2启用 3下架',
  `sort` INT NOT NULL DEFAULT '0' COMMENT '排序值，越大越靠前',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_product_no` (`tenant_id`, `product_no`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_tenant_coin_symbol` (`tenant_id`, `coin_symbol`),
  KEY `idx_tenant_sort` (`tenant_id`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押产品表';

-- ----------------------------
-- Table structure for t_stake_user_position
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_user_position`;
CREATE TABLE `t_stake_user_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `staked_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '当前在途和生效质押本金',
  `version` BIGINT NOT NULL DEFAULT '1' COMMENT '并发控制版本',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳（毫秒）',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳（毫秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_user_product` (`tenant_id`, `user_id`, `product_id`),
  KEY `idx_tenant_product` (`tenant_id`, `product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押用户产品额度聚合表';

-- ----------------------------
-- Table structure for t_stake_order
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_order`;
CREATE TABLE `t_stake_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `request_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '创建请求幂等号',
  `active_operation_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '当前占用订单的资金操作号',
  `user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `product_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押产品编号快照',
  `product_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '质押产品名称快照',
  `product_type` TINYINT NOT NULL DEFAULT '1' COMMENT '产品类型快照：1活期 2定期',
  `coin_name` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '质押币种名称快照',
  `coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '质押币种符号快照',
  `reward_coin_name` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '收益币种名称快照',
  `reward_coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '收益币种符号快照',
  `stake_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '质押数量',
  `apr` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '年化收益率快照',
  `lock_days` INT NOT NULL DEFAULT '0' COMMENT '锁仓天数快照',
  `interest_mode` TINYINT NOT NULL DEFAULT '1' COMMENT '计息方式快照：1按天计息 2到期一次性计息',
  `reward_mode` TINYINT NOT NULL DEFAULT '1' COMMENT '发息方式快照：1每日发放 2到期发放',
  `allow_early_redeem` TINYINT NOT NULL DEFAULT '2' COMMENT '是否允许提前赎回快照：1是 2否',
  `early_redeem_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '提前赎回手续费率快照',
  `interest_days` INT NOT NULL DEFAULT '0' COMMENT '已计息天数',
  `start_times` BIGINT NOT NULL DEFAULT '0' COMMENT '起息时间戳（毫秒）',
  `end_times` BIGINT NOT NULL DEFAULT '0' COMMENT '到期时间戳（毫秒），活期可为0',
  `last_reward_times` BIGINT NOT NULL DEFAULT '0' COMMENT '最后一次收益发放时间戳（毫秒）',
  `next_reward_times` BIGINT NOT NULL DEFAULT '0' COMMENT '下一次收益发放时间戳（毫秒）',
  `total_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计收益',
  `pending_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '待发放收益',
  `redeem_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '赎回本金数量',
  `redeem_fee` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '赎回手续费',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '订单状态：1质押中 2已到期 3已赎回 4提前赎回 5已取消 6锁仓处理中',
  `redeem_type` TINYINT NOT NULL DEFAULT '0' COMMENT '赎回类型：1未赎回 2到期赎回 3提前赎回 4手动赎回',
  `redeem_apply_times` BIGINT NOT NULL DEFAULT '0' COMMENT '申请赎回时间戳（毫秒）',
  `redeem_times` BIGINT NOT NULL DEFAULT '0' COMMENT '实际赎回时间戳（毫秒）',
  `source` TINYINT NOT NULL DEFAULT '1' COMMENT '来源：1后台 2H5 3APP 4API',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  `version` BIGINT NOT NULL DEFAULT '1' COMMENT '并发控制版本',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_order_no` (`tenant_id`, `order_no`),
  UNIQUE KEY `uk_tenant_user_request` (`tenant_id`, `user_id`, `request_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `user_id`),
  KEY `idx_tenant_product_id` (`tenant_id`, `product_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_tenant_uid_status` (`tenant_id`, `user_id`, `status`),
  KEY `idx_tenant_start_times` (`tenant_id`, `start_times`),
  KEY `idx_tenant_end_times` (`tenant_id`, `end_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押订单表';

-- ----------------------------
-- Table structure for t_stake_reward_log
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_reward_log`;
CREATE TABLE `t_stake_reward_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `operation_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '收益资金操作号',
  `user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `product_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '质押产品名称快照',
  `coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '质押币种符号快照',
  `reward_coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '收益币种符号快照',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '本次收益数量',
  `before_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '发放前累计收益',
  `after_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '发放后累计收益',
  `reward_type` TINYINT NOT NULL DEFAULT '1' COMMENT '收益类型：1日收益 2到期收益 3补发收益 4手动发放',
  `reward_status` TINYINT NOT NULL DEFAULT '1' COMMENT '发放状态：1失败 2成功 3处理中',
  `reward_times` BIGINT NOT NULL DEFAULT '0' COMMENT '收益发放时间戳（毫秒）',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_operation_no` (`tenant_id`, `operation_no`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `user_id`),
  KEY `idx_tenant_reward_times` (`tenant_id`, `reward_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押收益记录表';

-- ----------------------------
-- Table structure for t_stake_redeem_log
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_redeem_log`;
CREATE TABLE `t_stake_redeem_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `redeem_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '赎回单号',
  `redeem_type` TINYINT NOT NULL DEFAULT '1' COMMENT '赎回类型：1未赎回 2到期赎回 3提前赎回 4手动赎回',
  `stake_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '原始质押数量',
  `redeem_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '实际赎回本金数量',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '本次一并发放收益数量',
  `fee_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '手续费率',
  `fee_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '手续费数量',
  `redeem_status` TINYINT NOT NULL DEFAULT '1' COMMENT '赎回状态：1失败 2成功 3处理中',
  `redeem_times` BIGINT NOT NULL DEFAULT '0' COMMENT '赎回时间戳（毫秒）',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_redeem_no` (`tenant_id`, `redeem_no`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `user_id`),
  KEY `idx_tenant_redeem_times` (`tenant_id`, `redeem_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押赎回记录表';

-- ----------------------------
-- Table structure for t_stake_operation
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_operation`;
CREATE TABLE `t_stake_operation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '用户ID',
  `order_id` BIGINT NOT NULL DEFAULT '0' COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `operation_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '资金操作号',
  `request_no` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '调用方幂等号',
  `operation_type` TINYINT NOT NULL DEFAULT '0' COMMENT '操作类型：1申购 2日收益 3到期收益 4到期赎回 5提前赎回 6人工收益 7人工赎回 8申购回滚',
  `principal_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '本金金额',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '收益金额',
  `fee_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '手续费金额',
  `principal_status` TINYINT NOT NULL DEFAULT '0' COMMENT '本金步骤：0无需 1待处理 2成功',
  `reward_status` TINYINT NOT NULL DEFAULT '0' COMMENT '收益步骤：0无需 1待处理 2成功',
  `fee_status` TINYINT NOT NULL DEFAULT '0' COMMENT '手续费步骤：0无需 1待处理 2成功',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '操作状态：1待处理 2处理中 3成功 4可重试失败 5需人工处理',
  `period_end` BIGINT NOT NULL DEFAULT '0' COMMENT '收益周期结束时间（毫秒）',
  `retry_count` INT NOT NULL DEFAULT '0' COMMENT '重试次数',
  `next_retry_at` BIGINT NOT NULL DEFAULT '0' COMMENT '下次重试时间（毫秒）',
  `last_error` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '最近错误',
  `operator_user_id` BIGINT NOT NULL DEFAULT '0' COMMENT '操作人ID，用户操作为用户ID',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `version` BIGINT NOT NULL DEFAULT '1' COMMENT '并发控制版本',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳（毫秒）',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳（毫秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_operation_no` (`tenant_id`, `operation_no`),
  UNIQUE KEY `uk_tenant_user_type_request` (`tenant_id`, `user_id`, `operation_type`, `request_no`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  KEY `idx_status_retry` (`status`, `next_retry_at`, `id`),
  KEY `idx_tenant_period` (`tenant_id`, `operation_type`, `period_end`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押可恢复资金操作单';

-- ----------------------------
-- Table structure for t_stake_reconciliation
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_reconciliation`;
CREATE TABLE `t_stake_reconciliation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT '0' COMMENT '租户ID',
  `reconciliation_date` INT NOT NULL DEFAULT '0' COMMENT '对账日期，UTC YYYYMMDD',
  `coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '对账币种',
  `active_principal` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '质押中订单本金',
  `product_staked` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '产品汇总已质押本金',
  `position_staked` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '用户持仓汇总本金',
  `asset_locked` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT 'Asset剩余锁仓本金',
  `reward_log_amount` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '成功收益日志累计金额',
  `reward_platform_amount` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '奖励平台账户累计支出',
  `fee_log_amount` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '成功赎回手续费累计金额',
  `fee_platform_amount` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '手续费平台账户累计收入',
  `product_diff` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '产品汇总与订单本金差额',
  `position_diff` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '用户持仓与订单本金差额',
  `lock_diff` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT 'Asset锁仓与订单本金差额',
  `reward_diff` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '奖励平台支出与收益日志差额',
  `fee_diff` DECIMAL(36,18) NOT NULL DEFAULT '0.000000000000000000' COMMENT '手续费平台收入与赎回日志差额',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '状态：1一致 2存在差额 3执行失败',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '差额或执行错误说明',
  `create_times` BIGINT NOT NULL DEFAULT '0' COMMENT '创建时间戳（毫秒）',
  `update_times` BIGINT NOT NULL DEFAULT '0' COMMENT '更新时间戳（毫秒）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_date_coin` (`tenant_id`, `reconciliation_date`, `coin_symbol`),
  KEY `idx_date_status` (`reconciliation_date`, `status`, `id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押账实对账快照';

SET FOREIGN_KEY_CHECKS = 1;
