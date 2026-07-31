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
      AND table_name='t_option_mmp_config'
      AND index_name='idx_option_mmp_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_mmp_config
     ADD INDEX idx_option_mmp_monitor
       (status,triggered_at,update_times,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
