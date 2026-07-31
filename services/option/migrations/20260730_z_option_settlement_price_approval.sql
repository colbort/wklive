-- 结算价改为追加版本、双人复核。已确认版本不覆盖；更正生成新版本。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_settlement_price'
      AND column_name='supersedes_id'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_settlement_price
     ADD COLUMN supersedes_id BIGINT NOT NULL DEFAULT 0
       COMMENT ''被本版本替代的结算价ID'' AFTER status,
     ADD COLUMN change_reason VARCHAR(500) NOT NULL DEFAULT ''''
       COMMENT ''计算、拒绝或人工更正原因'' AFTER supersedes_id,
     ADD COLUMN created_by BIGINT NOT NULL DEFAULT 0
       COMMENT ''创建人，0为系统计算'' AFTER change_reason'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=@schema_name
      AND table_name='t_option_settlement_price'
      AND index_name='uk_settlement_price_contract'
  ),
  'ALTER TABLE t_option_settlement_price DROP INDEX uk_settlement_price_contract',
  'SELECT 1'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=@schema_name
      AND table_name='t_option_settlement_price'
      AND index_name='uk_settlement_price_version'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_settlement_price
     ADD UNIQUE KEY uk_settlement_price_version (tenant_id, contract_id, version)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=@schema_name
      AND table_name='t_option_settlement_price'
      AND constraint_name='chk_option_settlement_price'
  ),
  'ALTER TABLE t_option_settlement_price DROP CHECK chk_option_settlement_price',
  'SELECT 1'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE t_option_settlement_price
  ADD CONSTRAINT chk_option_settlement_price CHECK (
    status IN (1,2,3,4)
    AND version > 0
    AND (status <> 2 OR (delivery_price > 0 AND sample_count > 0 AND confirmed_at > 0))
    AND (created_by = 0 OR confirmed_by = 0 OR created_by <> confirmed_by)
  );

DROP TRIGGER IF EXISTS trg_option_settlement_price_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_settlement_price_immutable
BEFORE UPDATE ON t_option_settlement_price
FOR EACH ROW
BEGIN
  IF NOT (
    OLD.tenant_id <=> NEW.tenant_id
    AND OLD.contract_id <=> NEW.contract_id
    AND OLD.price_source <=> NEW.price_source
    AND OLD.window_start <=> NEW.window_start
    AND OLD.window_end <=> NEW.window_end
    AND OLD.sample_count <=> NEW.sample_count
    AND OLD.calculation_method <=> NEW.calculation_method
    AND OLD.delivery_price <=> NEW.delivery_price
    AND OLD.source_snapshot_ids <=> NEW.source_snapshot_ids
    AND OLD.version <=> NEW.version
    AND OLD.supersedes_id <=> NEW.supersedes_id
    AND OLD.created_by <=> NEW.created_by
    AND OLD.create_times <=> NEW.create_times
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'settlement price economic fields are immutable; create a new version';
  END IF;
END$$
DELIMITER ;
