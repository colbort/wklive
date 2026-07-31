-- 交易完整性控制：原子持仓/OI 限额、价格带、熔断参数、用户 kill switch
-- 和拒绝/处置审计。0 参数代表尚未配置；TRADING 合约必须显式设置正值。

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract' AND column_name='max_user_long_qty'),
  'SELECT 1',
  'ALTER TABLE t_option_contract ADD COLUMN max_user_long_qty DECIMAL(32,16) NOT NULL DEFAULT 0
    COMMENT ''单用户跨账户多头持仓及开仓委托上限，0表示未配置'' AFTER auto_exercise_threshold'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract' AND column_name='max_user_short_qty'),
  'SELECT 1',
  'ALTER TABLE t_option_contract ADD COLUMN max_user_short_qty DECIMAL(32,16) NOT NULL DEFAULT 0
    COMMENT ''单用户跨账户空头持仓及开仓委托上限，0表示未配置'' AFTER max_user_long_qty'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract' AND column_name='max_open_interest'),
  'SELECT 1',
  'ALTER TABLE t_option_contract ADD COLUMN max_open_interest DECIMAL(32,16) NOT NULL DEFAULT 0
    COMMENT ''合约单边持仓及开仓委托上限，0表示未配置'' AFTER max_user_short_qty'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract' AND column_name='order_price_band_ratio'),
  'SELECT 1',
  'ALTER TABLE t_option_contract ADD COLUMN order_price_band_ratio DECIMAL(20,10) NOT NULL DEFAULT 0
    COMMENT ''相对新鲜标记价的下单价格带比例，0表示未配置'' AFTER max_open_interest'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract' AND column_name='circuit_breaker_ratio'),
  'SELECT 1',
  'ALTER TABLE t_option_contract ADD COLUMN circuit_breaker_ratio DECIMAL(20,10) NOT NULL DEFAULT 0
    COMMENT ''标记价相对前值跳变熔断比例，0表示未配置'' AFTER order_price_band_ratio'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `t_option_user_trading_control` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `kill_switch` TINYINT NOT NULL DEFAULT 2,
  `reason` VARCHAR(255) NOT NULL DEFAULT '',
  `activated_at` BIGINT NOT NULL DEFAULT 0,
  `released_at` BIGINT NOT NULL DEFAULT 0,
  `activated_by` BIGINT NOT NULL DEFAULT 0,
  `released_by` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_user_trading_control` (`tenant_id`, `user_id`),
  CONSTRAINT `chk_option_user_trading_control` CHECK (
    `tenant_id` > 0 AND `user_id` > 0 AND `kill_switch` IN (1,2)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权用户交易控制';

CREATE TABLE IF NOT EXISTS `t_option_trading_control_event` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `order_id` BIGINT NOT NULL DEFAULT 0,
  `event_type` VARCHAR(32) NOT NULL DEFAULT '',
  `reason` VARCHAR(64) NOT NULL DEFAULT '',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '',
  `operator_id` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_option_control_event_contract` (`tenant_id`, `contract_id`, `id`),
  KEY `idx_option_control_event_user` (`tenant_id`, `user_id`, `id`),
  KEY `idx_option_control_event_reason` (`tenant_id`, `event_type`, `reason`, `id`),
  CONSTRAINT `chk_option_trading_control_event` CHECK (
    `tenant_id` > 0 AND `event_type` <> '' AND `reason` <> ''
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易控制审计事件';

-- 数据库最终门禁：历史 TRADING 行不会因安装迁移而失败，但任何新增或更新为
-- TRADING 的记录都必须显式配置全部控制参数。
DROP TRIGGER IF EXISTS trg_option_contract_trading_controls;
DELIMITER $$
CREATE TRIGGER trg_option_contract_trading_controls
BEFORE UPDATE ON t_option_contract
FOR EACH ROW
BEGIN
  IF NEW.max_user_long_qty < 0
    OR NEW.max_user_short_qty < 0
    OR NEW.max_open_interest < 0
    OR NEW.order_price_band_ratio < 0
    OR NEW.order_price_band_ratio > 1
    OR NEW.circuit_breaker_ratio < 0
    OR NEW.circuit_breaker_ratio > 1
    OR (
      NEW.status = 2 AND (
        NEW.max_user_long_qty <= 0
        OR NEW.max_user_short_qty <= 0
        OR NEW.max_open_interest <= 0
        OR NEW.order_price_band_ratio <= 0
        OR NEW.circuit_breaker_ratio <= 0
      )
    )
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'TRADING contract requires positive limits, price band and circuit breaker';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_contract_trading_controls_insert;
DELIMITER $$
CREATE TRIGGER trg_option_contract_trading_controls_insert
BEFORE INSERT ON t_option_contract
FOR EACH ROW
BEGIN
  IF NEW.max_user_long_qty < 0
    OR NEW.max_user_short_qty < 0
    OR NEW.max_open_interest < 0
    OR NEW.order_price_band_ratio < 0
    OR NEW.order_price_band_ratio > 1
    OR NEW.circuit_breaker_ratio < 0
    OR NEW.circuit_breaker_ratio > 1
    OR (
      NEW.status = 2 AND (
        NEW.max_user_long_qty <= 0
        OR NEW.max_user_short_qty <= 0
        OR NEW.max_open_interest <= 0
        OR NEW.order_price_band_ratio <= 0
        OR NEW.circuit_breaker_ratio <= 0
      )
    )
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'TRADING contract requires positive limits, price band and circuit breaker';
  END IF;
END$$
DELIMITER ;

-- 控制事件是追加式审计记录，禁止原地修改或删除。
DROP TRIGGER IF EXISTS trg_option_trading_control_event_no_update;
DELIMITER $$
CREATE TRIGGER trg_option_trading_control_event_no_update
BEFORE UPDATE ON t_option_trading_control_event
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading control audit events are immutable';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_control_event_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trading_control_event_no_delete
BEFORE DELETE ON t_option_trading_control_event
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading control audit events cannot be deleted';
END$$
DELIMITER ;
