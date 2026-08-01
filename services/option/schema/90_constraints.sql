-- goctl model mysql ddl does not support foreign keys or triggers.
-- Apply this file after all numbered table schema files.

ALTER TABLE `t_option_trading_calendar`
  ADD CONSTRAINT `chk_option_trading_calendar` CHECK (
    `tenant_id` > 0 AND `calendar_code` <> '' AND `version` > 0
    AND `status` IN (1,2,3,4) AND `timezone` <> ''
    AND (`effective_until` = 0 OR `effective_until` > `effective_from`)
  );

ALTER TABLE `t_option_trading_calendar_session`
  ADD CONSTRAINT `chk_option_trading_calendar_session` CHECK (
    `tenant_id` > 0 AND `calendar_id` > 0 AND `weekday` BETWEEN 0 AND 6
    AND `open_second` BETWEEN 0 AND 86399
    AND `close_second` > `open_second` AND `close_second` <= 172800
  );

ALTER TABLE `t_option_trading_calendar_exception`
  ADD CONSTRAINT `chk_option_trading_calendar_exception` CHECK (
    `tenant_id` > 0 AND `calendar_id` > 0 AND `exception_type` IN (1,2)
    AND `start_time` > 0 AND `end_time` > `start_time` AND `reason` <> ''
  );

ALTER TABLE `t_option_trading_halt`
  ADD CONSTRAINT `chk_option_trading_halt` CHECK (
    `tenant_id` > 0 AND `halt_no` <> '' AND `active_key` <> '' AND `contract_id` > 0
    AND `source` IN (1,2,3,4) AND `status` IN (1,2) AND `reason` <> ''
    AND `started_at` > 0 AND `cancel_total` >= 0 AND `cancel_success` >= 0 AND `cancel_failed` >= 0
    AND `cancel_success` + `cancel_failed` <= `cancel_total`
    AND (
      (`status` = 1 AND `active_key` = CONCAT('CONTRACT:',`contract_id`) AND `lifted_at` = 0 AND `lifted_by` = 0)
      OR
      (`status` = 2 AND `active_key` = CONCAT('HALT:',`halt_no`) AND `lifted_at` > 0 AND `lifted_by` >= 0 AND `lift_reason` <> '')
    )
  );

ALTER TABLE `t_option_corporate_action`
  ADD CONSTRAINT `chk_option_corporate_action` CHECK (
    `tenant_id` > 0 AND `event_no` <> '' AND `external_event_ref` <> '' AND `version` > 0
    AND `underlying_symbol` <> '' AND `action_type` IN (1,2,3,4,5,6,7,8,9,10)
    AND `status` IN (1,2,3,4,5,6,7)
    AND `announcement_time` > 0 AND `effective_time` > 0
    AND `evidence_ref` <> '' AND `evidence_hash` <> '' AND `description` <> ''
    AND `created_by` > 0
  );

ALTER TABLE `t_option_corporate_action_contract`
  ADD CONSTRAINT `chk_option_corporate_action_contract` CHECK (
    `tenant_id` > 0 AND `action_id` > 0 AND `source_contract_id` > 0
    AND `execution_mode` IN (1,2)
    AND ((`execution_mode` = 1 AND `successor_contract_id` > 0) OR `execution_mode` = 2)
    AND `quantity_numerator` > 0 AND `quantity_denominator` > 0
    AND `status` IN (1,2,3,4,5,6)
    AND `position_total` >= 0 AND `position_completed` >= 0 AND `position_failed` >= 0
    AND `retry_count` >= 0
  );

ALTER TABLE `t_option_corporate_action_position`
  ADD CONSTRAINT `chk_option_corporate_action_position` CHECK (
    `tenant_id` > 0 AND `action_id` > 0 AND `action_contract_id` > 0
    AND `source_position_id` > 0 AND `user_id` > 0 AND `side` IN (1,2)
    AND `source_quantity` > 0 AND `successor_quantity` > 0
    AND `source_available_quantity` >= 0 AND `successor_available_quantity` >= 0
    AND `source_effective_multiplier` > 0 AND `successor_effective_multiplier` > 0
    AND `cost_basis_before` >= 0 AND `cost_basis_after` >= 0
    AND `status` IN (1,2,3) AND `retry_count` >= 0
  );

ALTER TABLE `t_option_corporate_action_margin_lot`
  ADD CONSTRAINT `chk_option_corporate_action_margin_lot` CHECK (
    `tenant_id` > 0 AND `action_position_id` > 0 AND `margin_lot_id` > 0
    AND `source_contract_id` > 0 AND `successor_contract_id` > 0
    AND `source_position_id` > 0 AND `successor_position_id` > 0
    AND `source_quantity` > 0 AND `successor_quantity` > 0
    AND `source_remaining_quantity` >= 0 AND `successor_remaining_quantity` >= 0
  );

ALTER TABLE `t_option_order`
  ADD CONSTRAINT `chk_option_order_combo_link` CHECK (
    (`combo_order_id`=0 AND `combo_leg_no`=0)
    OR (`combo_order_id`>0 AND `combo_leg_no` BETWEEN 1 AND 4)
  ),
  ADD CONSTRAINT `chk_option_order_portfolio_config_pair` CHECK (
    (`portfolio_risk_config_id`=0 AND `portfolio_risk_config_version`=0)
    OR (`portfolio_risk_config_id`>0 AND `portfolio_risk_config_version`>0)
  ),
  ADD CONSTRAINT `chk_option_order_margin_coin` CHECK (
    `margin_amount`=0 OR `margin_coin`<>''
  );

ALTER TABLE `t_option_client_order_key`
  ADD CONSTRAINT `chk_option_client_order_key` CHECK (`client_order_id` <> '' AND `order_id` > 0);

ALTER TABLE `t_option_combo_order`
  ADD CONSTRAINT `chk_option_combo_order` CHECK (
    `tenant_id`>0 AND `combo_no`<>'' AND `user_id`>0 AND `account_id`>0
    AND `client_combo_id`<>'' AND CHAR_LENGTH(`strategy_key`)=64
    AND CHAR_LENGTH(`inverse_strategy_key`)=64 AND `underlying_symbol`<>''
    AND `expire_time`>0 AND `settle_coin`<>'' AND `quote_coin`<>''
    AND `order_type` IN (1,2) AND `qty`>0
    AND `filled_qty`>=0 AND `unfilled_qty`>=0 AND `filled_qty`+`unfilled_qty`=`qty`
    AND `status` IN (1,2,3,4,5,6,7,8) AND CHAR_LENGTH(`payload_hash`)=64
  );

ALTER TABLE `t_option_combo_order_leg`
  ADD CONSTRAINT `chk_option_combo_order_leg` CHECK (
    `tenant_id`>0 AND `combo_order_id`>0 AND `leg_no` BETWEEN 1 AND 4
    AND `contract_id`>0 AND `side` IN (1,2) AND `position_effect`=1
    AND `ratio` BETWEEN 1 AND 8 AND `price`>0 AND `qty`>0
    AND `filled_qty`>=0 AND `unfilled_qty`>=0 AND `filled_qty`+`unfilled_qty`=`qty`
    AND `child_order_id`>0
  );

ALTER TABLE `t_option_match_sequence`
  ADD CONSTRAINT `chk_option_match_sequence` CHECK (`next_sequence` > 0);

ALTER TABLE `t_option_outbox`
  ADD CONSTRAINT `chk_option_outbox` CHECK (
    `event_type` IN (1) AND `match_sequence` > 0
    AND `status` IN (1,2,3,4,5) AND `retry_count` >= 0
  );

ALTER TABLE `t_option_inbox`
  ADD CONSTRAINT `chk_option_inbox` CHECK (
    `event_type` IN (1) AND `match_sequence` > 0 AND `status` IN (1,2,3)
  );

ALTER TABLE `t_option_margin_lot`
  ADD CONSTRAINT `chk_option_margin_lot` CHECK (
    `quantity` > 0 AND `remaining_quantity` >= 0 AND `remaining_quantity` <= `quantity`
    AND `initial_margin` > 0 AND `remaining_margin` >= 0 AND `pending_margin` >= 0
    AND `pending_margin` <= `remaining_margin`
    AND `status` IN (1,2,3,4,5,6) AND `collateral_coin`<>''
  );

ALTER TABLE `t_option_margin_lot_application`
  ADD CONSTRAINT `chk_margin_lot_application` CHECK (`action` IN (2,3) AND `amount` > 0);

ALTER TABLE `t_option_risk_account`
  ADD CONSTRAINT `chk_option_risk_account` CHECK (
    `position_margin` >= 0 AND `maintenance_margin` >= 0
    AND `portfolio_risk_method` IN (0,1)
    AND `portfolio_risk_config_id` >= 0 AND `portfolio_risk_config_version` >= 0
    AND `portfolio_scenario_loss` >= 0 AND `portfolio_short_floor` >= 0
    AND `portfolio_concentration_addon` >= 0 AND `portfolio_liquidity_addon` >= 0
    AND `risk_rate` >= 0 AND `status` IN (1,2,3,4,5)
  );

ALTER TABLE `t_option_portfolio_risk_config`
  ADD CONSTRAINT `chk_option_portfolio_risk_config` CHECK (
    `tenant_id` > 0 AND `settle_coin` <> '' AND `version` > 0
    AND `status` IN (1,2,3,4) AND `model_method` = 1
    AND `initial_shock_rate` > 0 AND `initial_shock_rate` <= 10
    AND `maintenance_shock_rate` > 0
    AND `maintenance_shock_rate` <= `initial_shock_rate`
    AND `scenario_shocks` <> ''
    AND `concentration_threshold` >= 0
    AND `concentration_addon_rate` >= 0 AND `concentration_addon_rate` <= 1
    AND `liquidity_addon_rate` >= 0 AND `liquidity_addon_rate` <= 1
    AND `effective_from` > 0
    AND (`effective_until` = 0 OR `effective_until` > `effective_from`)
    AND `change_reason` <> '' AND `created_by` > 0
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0)
      OR (`status` IN (2,3,4) AND `reviewed_by` > 0
        AND `reviewed_by` <> `created_by` AND `reviewed_at` > 0)
    )
    AND (`status` NOT IN (2,4) OR `evidence_ref` <> '')
  );

ALTER TABLE `t_option_liquidation`
  ADD CONSTRAINT `chk_option_liquidation` CHECK (
    `quantity` > 0 AND `deficit_amount` >= 0 AND `liquidation_fee` >= 0
    AND `status` IN (1,2,3,4,5,6) AND `retry_count` >= 0 AND `insurance_attempt` >= 0
    AND `backstop_amount` >= 0 AND `deficit_resolution` IN (1,2,3,4,5)
  );

ALTER TABLE `t_option_insurance_fund_flow`
  ADD CONSTRAINT `chk_option_insurance_fund_flow` CHECK (`flow_type` IN (1,2,3,4) AND `amount` <> 0);

ALTER TABLE `t_option_exercise`
  ADD CONSTRAINT `chk_option_client_exercise_key` CHECK (`exercise_type` <> 1 OR `client_exercise_id` <> '');

ALTER TABLE `t_option_exercise_assignment`
  ADD CONSTRAINT `chk_option_exercise_assignment` CHECK (
    `quantity` > 0 AND `payoff` > 0 AND `status` IN (1,2,3,4)
  );

ALTER TABLE `t_option_exercise_instruction`
  ADD CONSTRAINT `chk_option_exercise_instruction` CHECK (
    `instruction_type` IN (1,2,3)
    AND `version` > 0
    AND `status` IN (1,2)
    AND `client_instruction_id` <> ''
  );

ALTER TABLE `t_option_user_trading_control`
  ADD CONSTRAINT `chk_option_user_trading_control` CHECK (
    `tenant_id` > 0 AND `user_id` > 0 AND `kill_switch` IN (1,2)
  );

ALTER TABLE `t_option_trading_control_event`
  ADD CONSTRAINT `chk_option_trading_control_event` CHECK (
    `tenant_id` > 0 AND `event_type` <> '' AND `reason` <> ''
  );

ALTER TABLE `t_option_asset_instruction`
  ADD CONSTRAINT `chk_option_asset_instruction` CHECK (
    `action` IN (1,2,3,4,5)
    AND `coin` <> '' AND `amount` > 0
    AND `step_no` > 0
    AND `status` IN (1,2,3,4,5,6)
    AND `retry_count` >= 0
    AND `reconciliation_status` IN (1,2,3)
  );

ALTER TABLE `t_option_settlement_price`
  ADD CONSTRAINT `chk_option_settlement_price` CHECK (
    `status` IN (1,2,3,4)
    AND `version` > 0
    AND (`status` NOT IN (1,2) OR (
      `delivery_price` > 0 AND `sample_count` > 0
      AND `window_start` > 0 AND `window_end` >= `window_start`
      AND JSON_VALID(`source_snapshot_ids`)
      AND JSON_TYPE(IF(JSON_VALID(`source_snapshot_ids`),`source_snapshot_ids`,'[]'))='ARRAY'
      AND JSON_LENGTH(IF(JSON_VALID(`source_snapshot_ids`),`source_snapshot_ids`,'[]'))=`sample_count`
      AND ((`price_source`='authoritative-market' AND `calculation_method`='MEDIAN' AND `created_by`=0)
        OR (`price_source`='manual-correction' AND `calculation_method`='MANUAL' AND `created_by`>0))
    ))
    AND (`status` <> 2 OR (`confirmed_by` > 0 AND `confirmed_at` > 0))
    AND (`created_by` = 0 OR `confirmed_by` = 0 OR `created_by` <> `confirmed_by`)
  );

ALTER TABLE `t_option_settlement_batch`
  ADD CONSTRAINT `chk_option_settlement_batch` CHECK (
    `status` IN (1,2,3,4,5,6,7,8)
    AND `total_credit` >= 0
    AND `total_debit` >= 0
    AND `instruction_count` >= 0
    AND `success_count` >= 0
    AND `success_count` <= `instruction_count`
  );

ALTER TABLE `t_option_settlement_detail`
  ADD CONSTRAINT `chk_option_settlement_detail` CHECK (
    `side` IN (1,2)
    AND `quantity` >= 0
    AND `payoff` >= 0
    AND `direction` IN (1,2,3)
  );

ALTER TABLE `t_option_physical_delivery_unit`
  ADD CONSTRAINT `chk_option_physical_delivery_unit` CHECK (
    `long_position_id` > 0 AND `short_position_id` > 0
    AND `quantity` > 0 AND `delivery_quantity` > 0 AND `payment_amount` > 0
    AND `delivery_coin` <> '' AND `payment_coin` <> '' AND `delivery_coin` <> `payment_coin`
    AND `status` IN (1,2,3,4,5,6) AND `cure_deadline` > 0
    AND `manual_retry_count` >= 0
  );

ALTER TABLE `t_option_reconciliation_issue`
  ADD CONSTRAINT `chk_option_reconciliation_issue` CHECK (
    `check_type` IN (1,2,3)
    AND `status` IN (1,2,3)
    AND `occurrence_count` > 0
  );

ALTER TABLE `t_option_reconciliation_run`
  ADD CONSTRAINT `chk_option_reconciliation_run` CHECK (
    `business_date` REGEXP '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'
    AND `scope` IN (1,2)
    AND `attempt_no` > 0
    AND `status` IN (1,2,3)
    AND `snapshot_time` > 0
    AND `coin_count` >= 0
    AND `account_count` >= 0
    AND `mismatch_count` >= 0
    AND `mismatch_count` <= `account_count`
    AND `completed_at` >= `snapshot_time`
  );

ALTER TABLE `t_option_reconciliation_run_detail`
  ADD CONSTRAINT `chk_option_reconciliation_detail` CHECK (
    `business_date` REGEXP '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'
    AND `scope`=2
    AND `dimension_type` IN (1,2,3)
    AND `dimension_key`<>''
    AND `flow_count`>=0
    AND `mismatch_count`>=0
    AND `status` IN (1,2,3)
    AND `expected_closing`=`opening_amount`+`external_net`+`option_net`+`manual_net`
    AND `difference_amount`=`actual_closing`-`expected_closing`
    AND ((`status`=1 AND `difference_amount`=0 AND `mismatch_count`=0) OR `status`<>1)
  );

ALTER TABLE `t_option_trade_correction`
  ADD CONSTRAINT `chk_option_trade_correction` CHECK (
    `tenant_id` > 0 AND `trade_id` > 0 AND `contract_id` > 0
    AND `action` = 1 AND `status` IN (1,2,3,4,5)
    AND `reason` <> '' AND `evidence_ref` <> '' AND `requested_by` > 0
    AND (`status` = 1 OR (`reviewed_by` > 0 AND `reviewed_by` <> `requested_by` AND `reviewed_at` > 0))
    AND (`status` <> 4 OR `completed_at` > 0)
  );

ALTER TABLE `t_option_trade_correction_leg`
  ADD CONSTRAINT `chk_option_trade_correction_leg` CHECK (
    `tenant_id` > 0 AND `correction_id` > 0 AND `leg_no` > 0
    AND `user_id` > 0 AND `coin` <> '' AND `direction` IN (1,2)
    AND `amount` > 0 AND `instruction_no` <> ''
  );

ALTER TABLE `t_option_mmp_config`
  ADD CONSTRAINT `chk_option_mmp_config` CHECK (
    `tenant_id` > 0 AND `user_id` > 0 AND `contract_id` > 0 AND `group_code` <> ''
    AND `enabled` IN (1,2) AND `status` IN (1,2,3)
    AND `qty_threshold` >= 0 AND `trade_count_threshold` >= 0 AND `loss_threshold` >= 0
    AND `window_seconds` > 0 AND `cooldown_seconds` >= 0
    AND (`enabled` = 2 OR (`qty_threshold` > 0 OR `trade_count_threshold` > 0 OR `loss_threshold` > 0))
    AND `accumulated_qty` >= 0 AND `trade_count` >= 0 AND `accumulated_loss` >= 0
  );

ALTER TABLE `t_option_contract_series_expiry`
  ADD CONSTRAINT `chk_option_contract_series_expiry` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `sequence_no` > 0 AND `cycle_code` <> ''
    AND `list_time` > 0 AND `exercise_cutoff_time` > `list_time`
    AND `expire_time` >= `exercise_cutoff_time` AND `deliver_time` >= `expire_time`
  );

ALTER TABLE `t_option_contract_series_strike_band`
  ADD CONSTRAINT `chk_option_contract_series_band` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `sequence_no` > 0
    AND `lower_strike` > 0 AND `upper_strike` >= `lower_strike` AND `strike_step` > 0
  );

ALTER TABLE `t_option_contract_series_detail`
  ADD CONSTRAINT `chk_option_contract_series_detail` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `expiry_id` > 0
    AND `option_type` IN (1,2) AND `strike_price` > 0
    AND `contract_code` <> '' AND `contract_id` > 0
  );

ALTER TABLE `t_option_reconciliation_run_detail`
  ADD CONSTRAINT `fk_option_reconciliation_detail_run`
  FOREIGN KEY (`run_id`,`tenant_id`,`business_date`,`scope`)
  REFERENCES `t_option_reconciliation_run` (`id`,`tenant_id`,`business_date`,`scope`)
  ON DELETE RESTRICT ON UPDATE RESTRICT;

ALTER TABLE `t_option_contract_series`
  ADD CONSTRAINT `chk_option_contract_series` CHECK (
    `tenant_id` > 0 AND `request_key` <> '' AND `series_code` <> ''
    AND `version` > 0 AND `status` IN (1,2,3)
    AND `template_contract_id` = 0 AND JSON_VALID(`template_snapshot`)
    AND `underlying_symbol` <> '' AND `reference_price` > 0
    AND `reference_source` <> '' AND `reference_time` > 0
    AND `evidence_ref` <> '' AND `change_reason` <> ''
    AND CHAR_LENGTH(`payload_hash`) = 64
    AND `expected_contract_count` BETWEEN 2 AND 500
    AND `generated_contract_count` BETWEEN 0 AND `expected_contract_count`
    AND `created_by` > 0
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0 AND `generated_contract_count` = 0 AND `generated_at` = 0
          AND `launch_status` = 0 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
      OR (`status` = 2 AND `reviewed_by` > 0 AND `reviewed_by` <> `created_by`
          AND `review_reason` <> '' AND `reviewed_at` > 0
          AND `generated_contract_count` = `expected_contract_count` AND `generated_at` > 0
          AND (
            (`launch_status` = 1 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
            OR (`launch_status` IN (2,3) AND `launch_reviewed_by` > 0
                AND `launch_reviewed_by` <> `created_by`
                AND `launch_review_reason` <> '' AND `launch_reviewed_at` > 0)
          ))
      OR (`status` = 3 AND `reviewed_by` > 0 AND `reviewed_by` <> `created_by`
          AND `review_reason` <> '' AND `reviewed_at` > 0
          AND `generated_contract_count` = 0 AND `generated_at` = 0
          AND `launch_status` = 0 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
    )
  );

DROP TRIGGER IF EXISTS `trg_option_reconciliation_run_no_update`;
CREATE TRIGGER `trg_option_reconciliation_run_no_update`
BEFORE UPDATE ON `t_option_reconciliation_run`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation run is immutable';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_run_no_delete`;
CREATE TRIGGER `trg_option_reconciliation_run_no_delete`
BEFORE DELETE ON `t_option_reconciliation_run`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation run cannot be deleted';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_detail_no_update`;
CREATE TRIGGER `trg_option_reconciliation_detail_no_update`
BEFORE UPDATE ON `t_option_reconciliation_run_detail`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation detail is immutable';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_detail_no_delete`;
CREATE TRIGGER `trg_option_reconciliation_detail_no_delete`
BEFORE DELETE ON `t_option_reconciliation_run_detail`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation detail cannot be deleted';
