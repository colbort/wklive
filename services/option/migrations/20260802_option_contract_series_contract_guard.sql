-- OPT-P2-005：系列生成合约经济字段与存续状态的数据库最终门禁。
-- 可重复执行；生命周期仍可更新 status/update_times，审计后的交易控制参数也可调整。

DROP TRIGGER IF EXISTS `trg_option_contract_series_contract_guard_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_contract_no_delete`;

DELIMITER $$
CREATE TRIGGER `trg_option_contract_series_contract_guard_update`
BEFORE UPDATE ON `t_option_contract`
FOR EACH ROW
BEGIN
  IF EXISTS (
    SELECT 1
    FROM `t_option_contract_series_detail`
    WHERE `tenant_id` = OLD.`tenant_id` AND `contract_id` = OLD.`id`
    LIMIT 1
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

CREATE TRIGGER `trg_option_contract_series_contract_no_delete`
BEFORE DELETE ON `t_option_contract`
FOR EACH ROW
BEGIN
  IF EXISTS (
    SELECT 1
    FROM `t_option_contract_series_detail`
    WHERE `tenant_id` = OLD.`tenant_id` AND `contract_id` = OLD.`id`
    LIMIT 1
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'contract series generated contract cannot be deleted';
  END IF;
END$$
DELIMITER ;
