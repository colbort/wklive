-- OPT-P2-002/P2-003：拆分最后交易、行权截止、到期与交割时点。
-- 可重复执行；存量数据以已审批的行权截止时间作为最后交易时间。

DROP TRIGGER IF EXISTS `trg_option_contract_series_expiry_no_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_last_trade_default_insert`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_last_trade_default_insert`;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract'
      AND column_name='last_trade_time'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN last_trade_time BIGINT NOT NULL DEFAULT 0 COMMENT ''最后可交易时间'' AFTER list_time'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract_series_expiry'
      AND column_name='last_trade_time'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract_series_expiry
     ADD COLUMN last_trade_time BIGINT NOT NULL DEFAULT 0 COMMENT ''最后可交易时间'' AFTER list_time'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `t_option_contract`
SET `last_trade_time` = CASE
  WHEN `exercise_cutoff_time` > `list_time` THEN `exercise_cutoff_time`
  ELSE `expire_time`
END
WHERE `last_trade_time` = 0;

UPDATE `t_option_contract_series_expiry`
SET `last_trade_time` = CASE
  WHEN `exercise_cutoff_time` > `list_time` THEN `exercise_cutoff_time`
  ELSE `expire_time`
END
WHERE `last_trade_time` = 0;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='t_option_contract'
      AND index_name='idx_tenant_last_trade_time'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD KEY idx_tenant_last_trade_time (tenant_id, last_trade_time)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='t_option_contract'
      AND index_name='idx_option_contract_last_trade_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD KEY idx_option_contract_last_trade_monitor (status, last_trade_time, tenant_id, id)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=DATABASE() AND table_name='t_option_contract'
      AND constraint_name='chk_option_contract_lifecycle_times'
  ),
  'ALTER TABLE t_option_contract DROP CHECK chk_option_contract_lifecycle_times',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `t_option_contract`
  ADD CONSTRAINT `chk_option_contract_lifecycle_times` CHECK (
    `list_time` > 0 AND `last_trade_time` > `list_time`
    AND `exercise_cutoff_time` >= `last_trade_time`
    AND `expire_time` >= `exercise_cutoff_time` AND `deliver_time` >= `expire_time`
  );

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=DATABASE() AND table_name='t_option_contract_series_expiry'
      AND constraint_name='chk_option_contract_series_expiry'
  ),
  'ALTER TABLE t_option_contract_series_expiry DROP CHECK chk_option_contract_series_expiry',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `t_option_contract_series_expiry`
  ADD CONSTRAINT `chk_option_contract_series_expiry` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `sequence_no` > 0 AND `cycle_code` <> ''
    AND `list_time` > 0 AND `last_trade_time` > `list_time`
    AND `exercise_cutoff_time` >= `last_trade_time`
    AND `expire_time` >= `exercise_cutoff_time` AND `deliver_time` >= `expire_time`
  );

-- 发布过渡期兼容旧的内部写入者；新 RPC/管理端始终显式传入该字段。
-- 触发器在 CHECK 约束前执行，因此旧写入不会生成 0 值。
DELIMITER $$
CREATE TRIGGER `trg_option_contract_last_trade_default_insert`
BEFORE INSERT ON `t_option_contract`
FOR EACH ROW
BEGIN
  IF NEW.`last_trade_time` = 0 THEN
    SET NEW.`last_trade_time` = NEW.`exercise_cutoff_time`;
  END IF;
END$$

CREATE TRIGGER `trg_option_contract_series_last_trade_default_insert`
BEFORE INSERT ON `t_option_contract_series_expiry`
FOR EACH ROW
BEGIN
  IF NEW.`last_trade_time` = 0 THEN
    SET NEW.`last_trade_time` = NEW.`exercise_cutoff_time`;
  END IF;
END$$
DELIMITER ;

DELIMITER $$
CREATE TRIGGER `trg_option_contract_series_expiry_no_update`
BEFORE UPDATE ON `t_option_contract_series_expiry`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series expiry is immutable';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS `trg_option_contract_series_contract_guard_update`;
DELIMITER $$
CREATE TRIGGER `trg_option_contract_series_contract_guard_update`
BEFORE UPDATE ON `t_option_contract`
FOR EACH ROW
BEGIN
  IF EXISTS (
    SELECT 1 FROM `t_option_contract_series_detail`
    WHERE `tenant_id` = OLD.`tenant_id` AND `contract_id` = OLD.`id` LIMIT 1
  ) AND NOT (
    OLD.`tenant_id` <=> NEW.`tenant_id`
    AND OLD.`contract_code` <=> NEW.`contract_code`
    AND OLD.`underlying_symbol` <=> NEW.`underlying_symbol`
    AND OLD.`underlying_coin` <=> NEW.`underlying_coin`
    AND OLD.`settle_coin` <=> NEW.`settle_coin`
    AND OLD.`quote_coin` <=> NEW.`quote_coin`
    AND OLD.`option_type` <=> NEW.`option_type`
    AND OLD.`exercise_style` <=> NEW.`exercise_style`
    AND OLD.`settlement_type` <=> NEW.`settlement_type`
    AND OLD.`strike_price` <=> NEW.`strike_price`
    AND OLD.`contract_unit` <=> NEW.`contract_unit`
    AND OLD.`min_order_qty` <=> NEW.`min_order_qty`
    AND OLD.`max_order_qty` <=> NEW.`max_order_qty`
    AND OLD.`price_tick` <=> NEW.`price_tick`
    AND OLD.`qty_step` <=> NEW.`qty_step`
    AND OLD.`multiplier` <=> NEW.`multiplier`
    AND OLD.`list_time` <=> NEW.`list_time`
    AND OLD.`last_trade_time` <=> NEW.`last_trade_time`
    AND OLD.`expire_time` <=> NEW.`expire_time`
    AND OLD.`deliver_time` <=> NEW.`deliver_time`
    AND OLD.`exercise_cutoff_time` <=> NEW.`exercise_cutoff_time`
    AND OLD.`auto_exercise_threshold` <=> NEW.`auto_exercise_threshold`
    AND OLD.`settlement_price_source` <=> NEW.`settlement_price_source`
    AND OLD.`settlement_price_method` <=> NEW.`settlement_price_method`
    AND OLD.`settlement_window_seconds` <=> NEW.`settlement_window_seconds`
    AND OLD.`settlement_min_samples` <=> NEW.`settlement_min_samples`
    AND OLD.`is_auto_exercise` <=> NEW.`is_auto_exercise`
    AND OLD.`maker_fee_rate` <=> NEW.`maker_fee_rate`
    AND OLD.`taker_fee_rate` <=> NEW.`taker_fee_rate`
    AND OLD.`exercise_fee_rate` <=> NEW.`exercise_fee_rate`
    AND OLD.`fee_user_id` <=> NEW.`fee_user_id`
    AND OLD.`fee_account_id` <=> NEW.`fee_account_id`
    AND OLD.`seller_margin_mode` <=> NEW.`seller_margin_mode`
    AND OLD.`initial_margin_rate` <=> NEW.`initial_margin_rate`
    AND OLD.`maintenance_margin_rate` <=> NEW.`maintenance_margin_rate`
    AND OLD.`min_margin_rate` <=> NEW.`min_margin_rate`
    AND OLD.`liquidation_fee_rate` <=> NEW.`liquidation_fee_rate`
    AND OLD.`insurance_user_id` <=> NEW.`insurance_user_id`
    AND OLD.`insurance_account_id` <=> NEW.`insurance_account_id`
    AND OLD.`liquidation_deficit_policy` <=> NEW.`liquidation_deficit_policy`
    AND OLD.`physical_delivery_policy` <=> NEW.`physical_delivery_policy`
    AND OLD.`physical_delivery_cure_seconds` <=> NEW.`physical_delivery_cure_seconds`
    AND OLD.`trading_calendar_code` <=> NEW.`trading_calendar_code`
    AND OLD.`is_deleted` <=> NEW.`is_deleted`
    AND OLD.`create_times` <=> NEW.`create_times`
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'contract series generated contract economics are immutable';
  END IF;
END$$
DELIMITER ;
