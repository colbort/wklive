-- Option 租户级运营指标的全局状态/时间窗口索引。
-- 采样器按固定组数跨租户聚合；这些索引避免每15秒扫描完整审计和状态表。

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_contract'
      AND index_name='idx_option_contract_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD INDEX idx_option_contract_monitor (status,update_times,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_contract'
      AND index_name='idx_option_contract_lifecycle_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD INDEX idx_option_contract_lifecycle_monitor
       (status,expire_time,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_contract'
      AND index_name='idx_option_public_chain_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD INDEX idx_option_public_chain_monitor
       (status,is_deleted,tenant_id,underlying_symbol,expire_time,
        strike_price,option_type,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_user_trading_control'
      AND index_name='idx_option_user_control_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_user_trading_control
     ADD INDEX idx_option_user_control_monitor
       (kill_switch,activated_at,tenant_id,user_id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_risk_account'
      AND index_name='idx_option_risk_account_portfolio_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
     ADD INDEX idx_option_risk_account_portfolio_monitor
       (portfolio_risk_method,last_calc_time,tenant_id,settle_coin,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_portfolio_risk_config'
      AND index_name='idx_option_portfolio_config_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_portfolio_risk_config
     ADD INDEX idx_option_portfolio_config_monitor
       (status,effective_from,effective_until,tenant_id,settle_coin,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_trading_control_event'
      AND index_name='idx_option_control_event_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_trading_control_event
     ADD INDEX idx_option_control_event_monitor
       (event_type,reason,create_times,tenant_id,user_id,contract_id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_position'
      AND index_name='idx_option_position_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_position
     ADD INDEX idx_option_position_monitor
       (status,tenant_id,contract_id,user_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_settlement_price'
      AND index_name='idx_option_settlement_price_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_settlement_price
     ADD INDEX idx_option_settlement_price_monitor
       (status,tenant_id,contract_id,confirmed_at,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_trade_correction'
      AND index_name='idx_option_trade_correction_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_trade_correction
     ADD INDEX idx_option_trade_correction_monitor
       (status,update_times,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_corporate_action'
      AND index_name='idx_corporate_action_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_corporate_action
     ADD INDEX idx_corporate_action_monitor
       (status,effective_time,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_corporate_action_contract'
      AND index_name='idx_corporate_action_contract_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_corporate_action_contract
     ADD INDEX idx_corporate_action_contract_monitor
       (status,tenant_id,action_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_contract_series'
      AND index_name='idx_option_contract_series_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract_series
     ADD INDEX idx_option_contract_series_monitor
       (status,create_times,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_mmp_config'
      AND index_name='idx_option_mmp_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_mmp_config
     ADD INDEX idx_option_mmp_monitor
       (status,triggered_at,update_times,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
