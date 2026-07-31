-- 拆分标的、标记价格和 Greeks 的新鲜度。旧 snapshot_time 无法证明
-- 标记价/Greeks 的来源，因此只回填标的时间；卖方交易需等待新的标记价。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_market'
      AND column_name='underlying_snapshot_time'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_market
     ADD COLUMN underlying_snapshot_time BIGINT NOT NULL DEFAULT 0 COMMENT ''标的价格快照时间'' AFTER snapshot_time,
     ADD COLUMN mark_snapshot_time BIGINT NOT NULL DEFAULT 0 COMMENT ''期权标记价格快照时间'' AFTER underlying_snapshot_time,
     ADD COLUMN greeks_snapshot_time BIGINT NOT NULL DEFAULT 0 COMMENT ''IV与Greeks快照时间'' AFTER mark_snapshot_time,
     ADD KEY idx_tenant_underlying_time (tenant_id, underlying_snapshot_time),
     ADD KEY idx_tenant_mark_time (tenant_id, mark_snapshot_time)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

UPDATE t_option_market
SET underlying_snapshot_time = snapshot_time
WHERE underlying_snapshot_time = 0
  AND underlying_price > 0
  AND snapshot_time > 0;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_market_snapshot'
      AND column_name='source_type'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_market_snapshot
     ADD COLUMN source_type TINYINT NOT NULL DEFAULT 0 COMMENT ''来源：0未知 1权威标的行情 2管理行情 3结算审计'' AFTER snapshot_time,
     ADD COLUMN source_snapshot_id VARCHAR(128) NOT NULL DEFAULT '''' COMMENT ''来源快照唯一标识'' AFTER source_type,
     ADD KEY idx_contract_source_time (tenant_id, contract_id, source_type, snapshot_time)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
