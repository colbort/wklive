-- OPT-P2-007：独立复杂策略簿父单、不可变腿和影子子单关联。
-- 可重复执行；不修改已有普通订单的业务语义。

DROP PROCEDURE IF EXISTS `sp_option_combo_add_column`;
DELIMITER $$
CREATE PROCEDURE `sp_option_combo_add_column`(
  IN p_table VARCHAR(64),
  IN p_column VARCHAR(64),
  IN p_definition VARCHAR(500)
)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name=p_table AND column_name=p_column
  ) THEN
    SET @ddl = CONCAT('ALTER TABLE `',p_table,'` ADD COLUMN `',p_column,'` ',p_definition);
    PREPARE stmt FROM @ddl;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$
DELIMITER ;

CALL `sp_option_combo_add_column`(
  't_option_order','combo_order_id',
  'BIGINT NOT NULL DEFAULT 0 COMMENT ''组合父单ID；0为普通单'' AFTER `mmp_group`'
);
CALL `sp_option_combo_add_column`(
  't_option_order','combo_leg_no',
  'BIGINT NOT NULL DEFAULT 0 COMMENT ''组合腿序号；普通单为0'' AFTER `combo_order_id`'
);
CALL `sp_option_combo_add_column`(
  't_option_trade','combo_match_no',
  'VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''组合成交组号；普通成交为空'' AFTER `trade_no`'
);
CALL `sp_option_combo_add_column`(
  't_option_trade','combo_leg_no',
  'BIGINT NOT NULL DEFAULT 0 COMMENT ''组合成交腿序号；普通成交为0'' AFTER `combo_match_no`'
);
DROP PROCEDURE `sp_option_combo_add_column`;

SET @combo_child_index_exists = (
  SELECT COUNT(1) FROM information_schema.statistics
  WHERE table_schema=DATABASE() AND table_name='t_option_order'
    AND index_name='idx_option_combo_child'
);
SET @combo_child_index_sql = IF(
  @combo_child_index_exists=0,
  'ALTER TABLE `t_option_order` ADD INDEX `idx_option_combo_child` (`tenant_id`,`combo_order_id`,`combo_leg_no`,`id`)',
  'SELECT 1'
);
PREPARE combo_child_index_stmt FROM @combo_child_index_sql;
EXECUTE combo_child_index_stmt;
DEALLOCATE PREPARE combo_child_index_stmt;

SET @combo_trade_index_exists = (
  SELECT COUNT(1) FROM information_schema.statistics
  WHERE table_schema=DATABASE() AND table_name='t_option_trade'
    AND index_name='idx_option_combo_match_trade'
);
SET @combo_trade_index_sql = IF(
  @combo_trade_index_exists=0,
  'ALTER TABLE `t_option_trade` ADD INDEX `idx_option_combo_match_trade` (`tenant_id`,`combo_match_no`,`combo_leg_no`,`id`)',
  'SELECT 1'
);
PREPARE combo_trade_index_stmt FROM @combo_trade_index_sql;
EXECUTE combo_trade_index_stmt;
DEALLOCATE PREPARE combo_trade_index_stmt;

SET @combo_child_check_exists = (
  SELECT COUNT(1) FROM information_schema.table_constraints
  WHERE constraint_schema=DATABASE() AND table_name='t_option_order'
    AND constraint_name='chk_option_order_combo_link' AND constraint_type='CHECK'
);
SET @combo_child_check_sql = IF(
  @combo_child_check_exists=0,
  'ALTER TABLE `t_option_order` ADD CONSTRAINT `chk_option_order_combo_link` CHECK ((`combo_order_id`=0 AND `combo_leg_no`=0) OR (`combo_order_id`>0 AND `combo_leg_no` BETWEEN 1 AND 4))',
  'SELECT 1'
);
PREPARE combo_child_check_stmt FROM @combo_child_check_sql;
EXECUTE combo_child_check_stmt;
DEALLOCATE PREPARE combo_child_check_stmt;

CREATE TABLE IF NOT EXISTS `t_option_combo_order` (
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
  KEY `idx_option_combo_user` (`tenant_id`,`user_id`,`account_id`,`id`),
  CONSTRAINT `chk_option_combo_order` CHECK (
    `tenant_id`>0 AND `combo_no`<>'' AND `user_id`>0 AND `account_id`>0
    AND `client_combo_id`<>'' AND CHAR_LENGTH(`strategy_key`)=64
    AND CHAR_LENGTH(`inverse_strategy_key`)=64 AND `underlying_symbol`<>''
    AND `expire_time`>0 AND `settle_coin`<>'' AND `quote_coin`<>''
    AND `order_type` IN (1,2) AND `qty`>0
    AND `filled_qty`>=0 AND `unfilled_qty`>=0 AND `filled_qty`+`unfilled_qty`=`qty`
    AND `status` IN (1,2,3,4,5,6,7,8) AND CHAR_LENGTH(`payload_hash`)=64
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权组合父单';

CREATE TABLE IF NOT EXISTS `t_option_combo_order_leg` (
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
  KEY `idx_option_combo_leg_contract` (`tenant_id`,`contract_id`,`combo_order_id`),
  CONSTRAINT `chk_option_combo_order_leg` CHECK (
    `tenant_id`>0 AND `combo_order_id`>0 AND `leg_no` BETWEEN 1 AND 4
    AND `contract_id`>0 AND `side` IN (1,2) AND `position_effect`=1
    AND `ratio` BETWEEN 1 AND 8 AND `price`>0 AND `qty`>0
    AND `filled_qty`>=0 AND `unfilled_qty`>=0 AND `filled_qty`+`unfilled_qty`=`qty`
    AND `child_order_id`>0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权组合不可变腿及影子子单';

DROP TRIGGER IF EXISTS `trg_option_combo_order_guard_update`;
DROP TRIGGER IF EXISTS `trg_option_combo_order_no_delete`;
DROP TRIGGER IF EXISTS `trg_option_combo_leg_guard_update`;
DROP TRIGGER IF EXISTS `trg_option_combo_leg_no_delete`;

DELIMITER $$
CREATE TRIGGER `trg_option_combo_order_guard_update`
BEFORE UPDATE ON `t_option_combo_order`
FOR EACH ROW
BEGIN
  IF NOT (
    OLD.tenant_id<=>NEW.tenant_id AND OLD.combo_no<=>NEW.combo_no
    AND OLD.user_id<=>NEW.user_id AND OLD.account_id<=>NEW.account_id
    AND OLD.client_combo_id<=>NEW.client_combo_id
    AND OLD.strategy_key<=>NEW.strategy_key
    AND OLD.inverse_strategy_key<=>NEW.inverse_strategy_key
    AND OLD.underlying_symbol<=>NEW.underlying_symbol
    AND OLD.expire_time<=>NEW.expire_time
    AND OLD.settle_coin<=>NEW.settle_coin AND OLD.quote_coin<=>NEW.quote_coin
    AND OLD.order_type<=>NEW.order_type AND OLD.net_price<=>NEW.net_price
    AND OLD.qty<=>NEW.qty AND OLD.payload_hash<=>NEW.payload_hash
    AND OLD.create_times<=>NEW.create_times
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo order inputs are immutable';
  END IF;
  IF NEW.filled_qty<OLD.filled_qty OR NEW.unfilled_qty>OLD.unfilled_qty THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo execution progress cannot reverse';
  END IF;
  IF NOT (
    (OLD.status=1 AND NEW.status IN (1,2,5,7,8))
    OR (OLD.status=2 AND NEW.status IN (2,3,4,5,7,8))
    OR (OLD.status=3 AND NEW.status IN (3,4,5,8))
    OR (OLD.status=5 AND NEW.status IN (5,6,8))
    OR (OLD.status IN (4,6,7,8) AND NEW.status=OLD.status)
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='illegal combo order status transition';
  END IF;
END$$

CREATE TRIGGER `trg_option_combo_order_no_delete`
BEFORE DELETE ON `t_option_combo_order`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo order history cannot be deleted';
END$$

CREATE TRIGGER `trg_option_combo_leg_guard_update`
BEFORE UPDATE ON `t_option_combo_order_leg`
FOR EACH ROW
BEGIN
  IF NOT (
    OLD.tenant_id<=>NEW.tenant_id AND OLD.combo_order_id<=>NEW.combo_order_id
    AND OLD.leg_no<=>NEW.leg_no AND OLD.contract_id<=>NEW.contract_id
    AND OLD.side<=>NEW.side AND OLD.position_effect<=>NEW.position_effect
    AND OLD.ratio<=>NEW.ratio AND OLD.price<=>NEW.price
    AND OLD.qty<=>NEW.qty AND OLD.child_order_id<=>NEW.child_order_id
    AND OLD.create_times<=>NEW.create_times
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo leg inputs are immutable';
  END IF;
  IF NEW.filled_qty<OLD.filled_qty OR NEW.unfilled_qty>OLD.unfilled_qty THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo leg execution progress cannot reverse';
  END IF;
END$$

CREATE TRIGGER `trg_option_combo_leg_no_delete`
BEFORE DELETE ON `t_option_combo_order_leg`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='combo leg history cannot be deleted';
END$$
DELIMITER ;
