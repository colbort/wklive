-- P2-001：版本化交易日历、周会话、节假日/特别开市和临时休市审计。
-- 迁移为历史租户建立显式 24x7 日历，只保持原行为，不代表生产审批。

SET @option_calendar_col_exists = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 't_option_contract'
    AND COLUMN_NAME = 'trading_calendar_code'
);
SET @option_calendar_col_sql = IF(
  @option_calendar_col_exists = 0,
  'ALTER TABLE t_option_contract ADD COLUMN trading_calendar_code VARCHAR(64) NOT NULL DEFAULT ''CONTINUOUS_24_7'' COMMENT ''不可变交易日历代码'' AFTER physical_delivery_cure_seconds',
  'SELECT 1'
);
PREPARE option_calendar_col_stmt FROM @option_calendar_col_sql;
EXECUTE option_calendar_col_stmt;
DEALLOCATE PREPARE option_calendar_col_stmt;

CREATE TABLE IF NOT EXISTS `t_option_trading_calendar` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `calendar_code` VARCHAR(64) NOT NULL DEFAULT '',
  `version` BIGINT NOT NULL DEFAULT 1,
  `status` TINYINT NOT NULL DEFAULT 1,
  `timezone` VARCHAR(64) NOT NULL DEFAULT 'UTC',
  `effective_from` BIGINT NOT NULL DEFAULT 0,
  `effective_until` BIGINT NOT NULL DEFAULT 0,
  `supersedes_id` BIGINT NOT NULL DEFAULT 0,
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_calendar_version` (`tenant_id`,`calendar_code`,`version`),
  KEY `idx_trading_calendar_effective` (`tenant_id`,`calendar_code`,`status`,`effective_from`,`effective_until`,`id`),
  CONSTRAINT `chk_option_trading_calendar` CHECK (
    `tenant_id` > 0 AND `calendar_code` <> '' AND `version` > 0
    AND `status` IN (1,2,3,4) AND `timezone` <> ''
    AND (`effective_until` = 0 OR `effective_until` > `effective_from`)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可覆盖的期权交易日历版本';

CREATE TABLE IF NOT EXISTS `t_option_trading_calendar_session` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `calendar_id` BIGINT NOT NULL DEFAULT 0,
  `weekday` TINYINT NOT NULL DEFAULT 0,
  `open_second` INT NOT NULL DEFAULT 0,
  `close_second` INT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_calendar_session` (`tenant_id`,`calendar_id`,`weekday`,`open_second`,`close_second`),
  KEY `idx_trading_calendar_session` (`tenant_id`,`calendar_id`,`weekday`,`id`),
  CONSTRAINT `chk_option_trading_calendar_session` CHECK (
    `tenant_id` > 0 AND `calendar_id` > 0 AND `weekday` BETWEEN 0 AND 6
    AND `open_second` BETWEEN 0 AND 86399
    AND `close_second` > `open_second` AND `close_second` <= 172800
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易日历周会话';

CREATE TABLE IF NOT EXISTS `t_option_trading_calendar_exception` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `calendar_id` BIGINT NOT NULL DEFAULT 0,
  `exception_type` TINYINT NOT NULL DEFAULT 0,
  `start_time` BIGINT NOT NULL DEFAULT 0,
  `end_time` BIGINT NOT NULL DEFAULT 0,
  `reason` VARCHAR(500) NOT NULL DEFAULT '',
  `announcement_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_trading_calendar_exception` (`tenant_id`,`calendar_id`,`start_time`,`end_time`,`exception_type`,`id`),
  CONSTRAINT `chk_option_trading_calendar_exception` CHECK (
    `tenant_id` > 0 AND `calendar_id` > 0 AND `exception_type` IN (1,2)
    AND `start_time` > 0 AND `end_time` > `start_time` AND `reason` <> ''
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权交易日历节假日与特别开市窗口';

CREATE TABLE IF NOT EXISTS `t_option_trading_halt` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `halt_no` VARCHAR(96) NOT NULL DEFAULT '',
  `active_key` VARCHAR(96) NOT NULL DEFAULT '',
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `source` TINYINT NOT NULL DEFAULT 1,
  `status` TINYINT NOT NULL DEFAULT 1,
  `reason` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `started_at` BIGINT NOT NULL DEFAULT 0,
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `cancel_total` BIGINT NOT NULL DEFAULT 0,
  `cancel_success` BIGINT NOT NULL DEFAULT 0,
  `cancel_failed` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `lifted_at` BIGINT NOT NULL DEFAULT 0,
  `lifted_by` BIGINT NOT NULL DEFAULT 0,
  `lift_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_trading_halt_no` (`tenant_id`,`halt_no`),
  UNIQUE KEY `uk_trading_halt_active` (`tenant_id`,`active_key`),
  KEY `idx_trading_halt_contract` (`tenant_id`,`contract_id`,`status`,`id`),
  CONSTRAINT `chk_option_trading_halt` CHECK (
    `tenant_id` > 0 AND `halt_no` <> '' AND `active_key` <> '' AND `contract_id` > 0
    AND `source` IN (1,2,3) AND `status` IN (1,2) AND `reason` <> ''
    AND `started_at` > 0 AND `cancel_total` >= 0 AND `cancel_success` >= 0 AND `cancel_failed` >= 0
    AND (
      (`status` = 1 AND `active_key` = CONCAT('CONTRACT:',`contract_id`) AND `lifted_at` = 0 AND `lifted_by` = 0)
      OR
      (`status` = 2 AND `active_key` = CONCAT('HALT:',`halt_no`) AND `lifted_at` > 0 AND `lifted_by` >= 0 AND `lift_reason` <> '')
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权临时休市与恢复审计';

-- option.sql 或早期开发库可能已含同名表；重建 CHECK 以保证系统到期可用 lifted_by=0。
UPDATE t_option_trading_halt
SET cancel_total=cancel_success+cancel_failed
WHERE cancel_total < cancel_success+cancel_failed;
SET @option_halt_check_exists = (
  SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA=DATABASE()
    AND TABLE_NAME='t_option_trading_halt'
    AND CONSTRAINT_NAME='chk_option_trading_halt'
    AND CONSTRAINT_TYPE='CHECK'
);
SET @option_halt_drop_check_sql = IF(
  @option_halt_check_exists>0,
  'ALTER TABLE t_option_trading_halt DROP CHECK chk_option_trading_halt',
  'SELECT 1'
);
PREPARE option_halt_drop_check_stmt FROM @option_halt_drop_check_sql;
EXECUTE option_halt_drop_check_stmt;
DEALLOCATE PREPARE option_halt_drop_check_stmt;
ALTER TABLE t_option_trading_halt
  ADD CONSTRAINT chk_option_trading_halt CHECK (
    `tenant_id` > 0 AND `halt_no` <> '' AND `active_key` <> '' AND `contract_id` > 0
    AND `source` IN (1,2,3) AND `status` IN (1,2) AND `reason` <> ''
    AND `started_at` > 0 AND `cancel_total` >= 0 AND `cancel_success` >= 0 AND `cancel_failed` >= 0
    AND `cancel_success` + `cancel_failed` <= `cancel_total`
    AND (
      (`status` = 1 AND `active_key` = CONCAT('CONTRACT:',`contract_id`) AND `lifted_at` = 0 AND `lifted_by` = 0)
      OR
      (`status` = 2 AND `active_key` = CONCAT('HALT:',`halt_no`) AND `lifted_at` > 0 AND `lifted_by` >= 0 AND `lift_reason` <> '')
    )
  );

INSERT INTO t_option_trading_calendar
(tenant_id,calendar_code,version,status,timezone,effective_from,effective_until,
 supersedes_id,change_reason,evidence_ref,created_by,reviewed_by,review_reason,
 reviewed_at,create_times,update_times)
SELECT tenants.tenant_id,'CONTINUOUS_24_7',1,2,'UTC',0,0,0,
       'MIGRATION_PRESERVE_EXISTING_24X7','20260731_zk_option_trading_calendar',
       0,0,'SYSTEM_BOOTSTRAP',UNIX_TIMESTAMP(),UNIX_TIMESTAMP(),UNIX_TIMESTAMP()
FROM (SELECT DISTINCT tenant_id FROM t_option_contract WHERE tenant_id > 0) tenants
WHERE NOT EXISTS (
  SELECT 1 FROM t_option_trading_calendar calendar
  WHERE calendar.tenant_id=tenants.tenant_id
    AND calendar.calendar_code='CONTINUOUS_24_7'
    AND calendar.version=1
);

INSERT INTO t_option_trading_calendar_session
(tenant_id,calendar_id,weekday,open_second,close_second,create_times)
SELECT calendar.tenant_id,calendar.id,weekdays.weekday,0,86400,UNIX_TIMESTAMP()
FROM t_option_trading_calendar calendar
JOIN (
  SELECT 0 weekday UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3
  UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6
) weekdays
WHERE calendar.calendar_code='CONTINUOUS_24_7' AND calendar.version=1
  AND NOT EXISTS (
    SELECT 1 FROM t_option_trading_calendar_session session
    WHERE session.tenant_id=calendar.tenant_id
      AND session.calendar_id=calendar.id
      AND session.weekday=weekdays.weekday
      AND session.open_second=0 AND session.close_second=86400
  );

UPDATE t_option_contract
SET trading_calendar_code='CONTINUOUS_24_7'
WHERE trading_calendar_code='';

DROP TRIGGER IF EXISTS trg_option_trading_calendar_update;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_update
BEFORE UPDATE ON t_option_trading_calendar
FOR EACH ROW
BEGIN
  IF NEW.tenant_id <> OLD.tenant_id
    OR NEW.calendar_code <> OLD.calendar_code
    OR NEW.version <> OLD.version
    OR NEW.timezone <> OLD.timezone
    OR NEW.effective_from <> OLD.effective_from
    OR (
      NEW.effective_until <> OLD.effective_until
      AND NOT (
        OLD.status=2 AND NEW.status=4 AND OLD.effective_until=0
        AND NEW.effective_until>OLD.effective_from
      )
    )
    OR (
      NEW.supersedes_id <> OLD.supersedes_id
      AND NOT (OLD.status=1 AND NEW.status=2 AND OLD.supersedes_id=0 AND NEW.supersedes_id>0)
    )
    OR NEW.change_reason <> OLD.change_reason
    OR NEW.evidence_ref <> OLD.evidence_ref
    OR NEW.created_by <> OLD.created_by
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trading calendar economic fields are immutable';
  END IF;
  IF OLD.status = 1 AND NEW.status NOT IN (2,3) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid trading calendar review transition';
  ELSEIF OLD.status = 2 AND (
    NEW.status <> 4 OR NEW.effective_until <= NEW.effective_from
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'approved trading calendar may only be superseded';
  ELSEIF OLD.status IN (3,4) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'terminal trading calendar version is immutable';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_session_insert;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_session_insert
BEFORE INSERT ON t_option_trading_calendar_session
FOR EACH ROW
BEGIN
  DECLARE draft_count BIGINT DEFAULT 0;
  SELECT COUNT(*) INTO draft_count
  FROM t_option_trading_calendar
  WHERE id=NEW.calendar_id AND tenant_id=NEW.tenant_id AND status=1;
  IF draft_count <> 1 THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trading calendar sessions may only be added to a draft';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_no_delete
BEFORE DELETE ON t_option_trading_calendar
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading calendar history cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_exception_insert;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_exception_insert
BEFORE INSERT ON t_option_trading_calendar_exception
FOR EACH ROW
BEGIN
  DECLARE draft_count BIGINT DEFAULT 0;
  SELECT COUNT(*) INTO draft_count
  FROM t_option_trading_calendar
  WHERE id=NEW.calendar_id AND tenant_id=NEW.tenant_id AND status=1;
  IF draft_count <> 1 THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trading calendar exceptions may only be added to a draft';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_session_no_update;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_session_no_update
BEFORE UPDATE ON t_option_trading_calendar_session
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading calendar sessions are append-only';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_session_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_session_no_delete
BEFORE DELETE ON t_option_trading_calendar_session
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading calendar sessions cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_exception_no_update;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_exception_no_update
BEFORE UPDATE ON t_option_trading_calendar_exception
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading calendar exceptions are append-only';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_calendar_exception_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trading_calendar_exception_no_delete
BEFORE DELETE ON t_option_trading_calendar_exception
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading calendar exceptions cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_halt_update;
DELIMITER $$
CREATE TRIGGER trg_option_trading_halt_update
BEFORE UPDATE ON t_option_trading_halt
FOR EACH ROW
BEGIN
  IF NEW.tenant_id <> OLD.tenant_id
    OR NEW.halt_no <> OLD.halt_no
    OR NEW.contract_id <> OLD.contract_id
    OR NEW.source <> OLD.source
    OR NEW.reason <> OLD.reason
    OR NEW.evidence_ref <> OLD.evidence_ref
    OR NEW.started_at <> OLD.started_at
    OR NEW.created_by <> OLD.created_by
    OR NEW.create_times <> OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trading halt identity is immutable';
  END IF;
  IF OLD.status <> 1 OR NEW.status NOT IN (1,2) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid trading halt transition';
  END IF;
  IF NEW.cancel_total < OLD.cancel_total
    OR NEW.cancel_success < OLD.cancel_success
    OR NEW.cancel_failed < OLD.cancel_failed
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trading halt cancellation counters are monotonic';
  END IF;
  IF OLD.status=1 AND NEW.status=1 AND (
    NEW.active_key <> OLD.active_key
    OR NEW.lifted_at <> OLD.lifted_at
    OR NEW.lifted_by <> OLD.lifted_by
    OR NEW.lift_reason <> OLD.lift_reason
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'active trading halt lift fields are immutable';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trading_halt_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trading_halt_no_delete
BEFORE DELETE ON t_option_trading_halt
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trading halt history cannot be deleted';
END$$
DELIMITER ;

-- 合并既有交易控制门禁：交易态必须有显式日历，且上市后不能换日历代码。
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
      OR NEW.trading_calendar_code = ''
    THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires controls and trading calendar';
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
      OR NEW.trading_calendar_code = ''
    THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'TRADING contract requires controls and trading calendar';
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
