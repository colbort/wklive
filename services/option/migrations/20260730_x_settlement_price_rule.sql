-- 为每个合约固化最终结算价来源、算法和到期取价窗口。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_contract'
      AND column_name='settlement_price_source'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN settlement_price_source VARCHAR(32) NOT NULL DEFAULT ''authoritative-market'' COMMENT ''最终结算价来源'' AFTER deliver_time,
     ADD COLUMN settlement_price_method VARCHAR(16) NOT NULL DEFAULT ''MEDIAN'' COMMENT ''最终结算价算法'' AFTER settlement_price_source,
     ADD COLUMN settlement_window_seconds INT NOT NULL DEFAULT 60 COMMENT ''到期前取价窗口秒数'' AFTER settlement_price_method,
     ADD COLUMN settlement_min_samples INT NOT NULL DEFAULT 3 COMMENT ''最终结算价最小样本数'' AFTER settlement_window_seconds'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
