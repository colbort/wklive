-- Add an explicit, product-approved Greeks freshness threshold. Existing rows
-- remain 0 (unconfigured); no business value is guessed during migration.
SET @schema_name = DATABASE();
SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_contract'
      AND column_name='greeks_max_age_seconds'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN greeks_max_age_seconds BIGINT NOT NULL DEFAULT 0
       COMMENT ''IV与Greeks最大允许陈旧秒数，0表示未审批/未配置''
       AFTER circuit_breaker_ratio'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Rebuild the latest combined contract gate from the calendar migration and
-- add Greeks configuration to the same atomic database admission rule.
DROP TRIGGER IF EXISTS trg_option_contract_trading_controls;
DELIMITER $$
CREATE TRIGGER trg_option_contract_trading_controls
BEFORE UPDATE ON t_option_contract
FOR EACH ROW
BEGIN
  DECLARE active_calendar_count BIGINT DEFAULT 0;
  IF OLD.status <> 1 AND NEW.trading_calendar_code <> OLD.trading_calendar_code THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'listed contract trading calendar code is immutable';
  END IF;
  IF NEW.max_user_long_qty < 0
    OR NEW.max_user_short_qty < 0
    OR NEW.max_open_interest < 0
    OR NEW.order_price_band_ratio < 0
    OR NEW.order_price_band_ratio > 1
    OR NEW.circuit_breaker_ratio < 0
    OR NEW.circuit_breaker_ratio > 1
    OR NEW.greeks_max_age_seconds < 0
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid contract trading controls';
  END IF;
  IF NEW.status = 2 THEN
    IF NEW.max_user_long_qty <= 0
      OR NEW.max_user_short_qty <= 0
      OR NEW.max_open_interest <= 0
      OR NEW.order_price_band_ratio <= 0
      OR NEW.circuit_breaker_ratio <= 0
      OR NEW.greeks_max_age_seconds <= 0
      OR NEW.trading_calendar_code = ''
    THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires controls, Greeks threshold and trading calendar';
    END IF;
    SELECT COUNT(*) INTO active_calendar_count
    FROM t_option_trading_calendar
    WHERE tenant_id=NEW.tenant_id
      AND calendar_code=NEW.trading_calendar_code
      AND status IN (2,4)
      AND effective_from <= UNIX_TIMESTAMP()
      AND (effective_until=0 OR effective_until > UNIX_TIMESTAMP());
    IF active_calendar_count <> 1 THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires exactly one active approved calendar';
    END IF;
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_contract_trading_controls_insert;
DELIMITER $$
CREATE TRIGGER trg_option_contract_trading_controls_insert
BEFORE INSERT ON t_option_contract
FOR EACH ROW
BEGIN
  DECLARE active_calendar_count BIGINT DEFAULT 0;
  IF NEW.max_user_long_qty < 0
    OR NEW.max_user_short_qty < 0
    OR NEW.max_open_interest < 0
    OR NEW.order_price_band_ratio < 0
    OR NEW.order_price_band_ratio > 1
    OR NEW.circuit_breaker_ratio < 0
    OR NEW.circuit_breaker_ratio > 1
    OR NEW.greeks_max_age_seconds < 0
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid contract trading controls';
  END IF;
  IF NEW.status = 2 THEN
    IF NEW.max_user_long_qty <= 0
      OR NEW.max_user_short_qty <= 0
      OR NEW.max_open_interest <= 0
      OR NEW.order_price_band_ratio <= 0
      OR NEW.circuit_breaker_ratio <= 0
      OR NEW.greeks_max_age_seconds <= 0
      OR NEW.trading_calendar_code = ''
    THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires controls, Greeks threshold and trading calendar';
    END IF;
    SELECT COUNT(*) INTO active_calendar_count
    FROM t_option_trading_calendar
    WHERE tenant_id=NEW.tenant_id
      AND calendar_code=NEW.trading_calendar_code
      AND status IN (2,4)
      AND effective_from <= UNIX_TIMESTAMP()
      AND (effective_until=0 OR effective_until > UNIX_TIMESTAMP());
    IF active_calendar_count <> 1 THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires exactly one active approved calendar';
    END IF;
  END IF;
END$$
DELIMITER ;
