-- 异常成交治理：只允许追加式、资金守恒、双人复核的现金更正。
-- 原成交、订单、持仓和 Asset 流水不得删除或原地篡改。

CREATE TABLE IF NOT EXISTS `t_option_trade_correction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `case_no` VARCHAR(96) NOT NULL DEFAULT '',
  `trade_id` BIGINT NOT NULL DEFAULT 0,
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `action` TINYINT NOT NULL DEFAULT 1,
  `status` TINYINT NOT NULL DEFAULT 1,
  `reason` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `requested_by` BIGINT NOT NULL DEFAULT 0,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_trade_correction_case` (`tenant_id`, `case_no`),
  KEY `idx_option_trade_correction_trade` (`tenant_id`, `trade_id`, `id`),
  KEY `idx_option_trade_correction_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_trade_correction` CHECK (
    `tenant_id` > 0 AND `trade_id` > 0 AND `contract_id` > 0
    AND `action` = 1 AND `status` IN (1,2,3,4,5)
    AND `reason` <> '' AND `evidence_ref` <> '' AND `requested_by` > 0
    AND (`status` = 1 OR (`reviewed_by` > 0 AND `reviewed_by` <> `requested_by` AND `reviewed_at` > 0))
    AND (`status` <> 4 OR `completed_at` > 0)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异常成交现金更正案件';

CREATE TABLE IF NOT EXISTS `t_option_trade_correction_leg` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `correction_id` BIGINT NOT NULL DEFAULT 0,
  `leg_no` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `account_id` BIGINT NOT NULL DEFAULT 0,
  `coin` VARCHAR(16) NOT NULL DEFAULT '',
  `direction` TINYINT NOT NULL DEFAULT 0,
  `amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `instruction_no` VARCHAR(96) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_trade_correction_leg` (`tenant_id`, `correction_id`, `leg_no`),
  UNIQUE KEY `uk_option_trade_correction_instruction` (`tenant_id`, `instruction_no`),
  KEY `idx_option_trade_correction_leg_user` (`tenant_id`, `user_id`, `account_id`, `id`),
  CONSTRAINT `chk_option_trade_correction_leg` CHECK (
    `tenant_id` > 0 AND `correction_id` > 0 AND `leg_no` > 0
    AND `user_id` > 0 AND `coin` <> '' AND `direction` IN (1,2)
    AND `amount` > 0 AND `instruction_no` <> ''
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异常成交不可变现金更正分录';

DROP TRIGGER IF EXISTS trg_option_trade_correction_guard;
DELIMITER $$
CREATE TRIGGER trg_option_trade_correction_guard
BEFORE UPDATE ON t_option_trade_correction
FOR EACH ROW
BEGIN
  IF NEW.tenant_id <> OLD.tenant_id
    OR NEW.case_no <> OLD.case_no
    OR NEW.trade_id <> OLD.trade_id
    OR NEW.contract_id <> OLD.contract_id
    OR NEW.action <> OLD.action
    OR NEW.reason <> OLD.reason
    OR NEW.evidence_ref <> OLD.evidence_ref
    OR NEW.requested_by <> OLD.requested_by
    OR NEW.create_times <> OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trade correction economic fields are immutable';
  END IF;

  IF NOT (
    NEW.status = OLD.status
    OR (OLD.status = 1 AND NEW.status IN (2,3))
    OR (OLD.status = 3 AND NEW.status IN (4,5))
    OR (OLD.status = 5 AND NEW.status = 4)
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid trade correction status transition';
  END IF;

  IF NEW.status <> 1 AND (
    NEW.reviewed_by <= 0
    OR NEW.reviewed_by = NEW.requested_by
    OR NEW.reviewed_at <= 0
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'trade correction requires independent review';
  END IF;

  IF NEW.status = 4 AND NEW.completed_at <= 0 THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'completed trade correction requires completion time';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trade_correction_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trade_correction_no_delete
BEFORE DELETE ON t_option_trade_correction
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trade correction cases cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trade_correction_leg_no_update;
DELIMITER $$
CREATE TRIGGER trg_option_trade_correction_leg_no_update
BEFORE UPDATE ON t_option_trade_correction_leg
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trade correction legs are immutable';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_trade_correction_leg_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_trade_correction_leg_no_delete
BEFORE DELETE ON t_option_trade_correction_leg
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'trade correction legs cannot be deleted';
END$$
DELIMITER ;
