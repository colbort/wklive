SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_stake_product
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_product`;
CREATE TABLE `t_stake_product` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '租户ID',
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
  `allow_early_redeem` TINYINT NOT NULL DEFAULT '0' COMMENT '是否允许提前赎回：0否 1是',
  `early_redeem_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '提前赎回手续费率，例如5.0000表示5%',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '状态：0禁用 1启用 2下架',
  `sort` INT NOT NULL DEFAULT '0' COMMENT '排序值，越大越靠前',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_product_no` (`tenant_id`, `product_no`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_tenant_coin_symbol` (`tenant_id`, `coin_symbol`),
  KEY `idx_tenant_sort` (`tenant_id`, `sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押产品表';

-- ----------------------------
-- Table structure for t_stake_order
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_order`;
CREATE TABLE `t_stake_order` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `uid` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '质押产品ID',
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
  `allow_early_redeem` TINYINT NOT NULL DEFAULT '0' COMMENT '是否允许提前赎回快照：0否 1是',
  `early_redeem_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '提前赎回手续费率快照',
  `interest_days` INT NOT NULL DEFAULT '0' COMMENT '已计息天数',
  `start_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '起息时间戳',
  `end_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '到期时间戳，活期可为0',
  `last_reward_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '最后一次收益发放时间戳',
  `next_reward_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '下一次收益发放时间戳',
  `total_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '累计收益',
  `pending_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '待发放收益',
  `redeem_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '赎回本金数量',
  `redeem_fee` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '赎回手续费',
  `status` TINYINT NOT NULL DEFAULT '1' COMMENT '订单状态：1质押中 2已到期 3已赎回 4提前赎回 5已取消',
  `redeem_type` TINYINT NOT NULL DEFAULT '0' COMMENT '赎回类型：0未赎回 1到期赎回 2提前赎回',
  `redeem_apply_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '申请赎回时间戳',
  `redeem_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '实际赎回时间戳',
  `source` TINYINT NOT NULL DEFAULT '1' COMMENT '来源：1后台 2H5 3APP 4API',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `uid`),
  KEY `idx_tenant_product_id` (`tenant_id`, `product_id`),
  KEY `idx_tenant_status` (`tenant_id`, `status`),
  KEY `idx_tenant_uid_status` (`tenant_id`, `uid`, `status`),
  KEY `idx_tenant_start_times` (`tenant_id`, `start_times`),
  KEY `idx_tenant_end_times` (`tenant_id`, `end_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押订单表';

-- ----------------------------
-- Table structure for t_stake_reward_log
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_reward_log`;
CREATE TABLE `t_stake_reward_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `uid` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `product_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '质押产品名称快照',
  `coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '质押币种符号快照',
  `reward_coin_symbol` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '收益币种符号快照',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '本次收益数量',
  `before_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '发放前累计收益',
  `after_reward` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '发放后累计收益',
  `reward_type` TINYINT NOT NULL DEFAULT '1' COMMENT '收益类型：1日收益 2到期收益 3补发收益 4手动发放',
  `reward_status` TINYINT NOT NULL DEFAULT '1' COMMENT '发放状态：0失败 1成功',
  `reward_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '收益发放时间戳',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `uid`),
  KEY `idx_tenant_reward_times` (`tenant_id`, `reward_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押收益记录表';

-- ----------------------------
-- Table structure for t_stake_redeem_log
-- ----------------------------
DROP TABLE IF EXISTS `t_stake_redeem_log`;
CREATE TABLE `t_stake_redeem_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '租户ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '质押订单ID',
  `order_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '质押订单号',
  `uid` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `product_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '质押产品ID',
  `redeem_no` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '赎回单号',
  `redeem_type` TINYINT NOT NULL DEFAULT '1' COMMENT '赎回类型：1到期赎回 2提前赎回 3手动赎回',
  `stake_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '原始质押数量',
  `redeem_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '实际赎回本金数量',
  `reward_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '本次一并发放收益数量',
  `fee_rate` DECIMAL(10,4) NOT NULL DEFAULT '0.0000' COMMENT '手续费率',
  `fee_amount` DECIMAL(30,8) NOT NULL DEFAULT '0.00000000' COMMENT '手续费数量',
  `redeem_status` TINYINT NOT NULL DEFAULT '1' COMMENT '赎回状态：0失败 1成功 2处理中',
  `redeem_times` INT UNSIGNED NOT NULL DEFAULT '0' COMMENT '赎回时间戳',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `create_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `update_user_id` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新人ID',
  `create_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间戳',
  `update_times` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '更新时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_redeem_no` (`tenant_id`, `redeem_no`),
  KEY `idx_tenant_order_id` (`tenant_id`, `order_id`),
  KEY `idx_tenant_order_no` (`tenant_id`, `order_no`),
  KEY `idx_tenant_uid` (`tenant_id`, `uid`),
  KEY `idx_tenant_redeem_times` (`tenant_id`, `redeem_times`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质押赎回记录表';

SET FOREIGN_KEY_CHECKS = 1;