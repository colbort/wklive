-- 保险接管库存受控主动退出：不可变申请、四眼复核、批准后幂等提交。

CREATE TABLE IF NOT EXISTS `t_option_insurance_inventory_exit` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `request_no` VARCHAR(96) NOT NULL DEFAULT '',
  `active_key` VARCHAR(128) NOT NULL DEFAULT '',
  `position_id` BIGINT NOT NULL DEFAULT 0,
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `insurance_user_id` BIGINT NOT NULL DEFAULT 0,
  `insurance_account_id` BIGINT NOT NULL DEFAULT 0,
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `limit_price` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `reason` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `requested_by` BIGINT NOT NULL DEFAULT 0,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0,
  `order_id` BIGINT NOT NULL DEFAULT 0,
  `submitted_by` BIGINT NOT NULL DEFAULT 0,
  `submitted_at` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_insurance_exit_request` (`tenant_id`, `request_no`),
  UNIQUE KEY `uk_option_insurance_exit_active` (`tenant_id`, `active_key`),
  KEY `idx_option_insurance_exit_order` (`tenant_id`, `order_id`),
  KEY `idx_option_insurance_exit_position` (`tenant_id`, `position_id`, `id`),
  KEY `idx_option_insurance_exit_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_insurance_inventory_exit` CHECK (
    `tenant_id` > 0 AND `request_no` <> '' AND `active_key` <> ''
    AND `position_id` > 0 AND `contract_id` > 0
    AND `insurance_user_id` > 0 AND `insurance_account_id` > 0
    AND `quantity` > 0 AND `limit_price` > 0 AND `status` IN (1,2,3,4)
    AND `reason` <> '' AND `evidence_ref` <> '' AND `requested_by` > 0
    AND `active_key` = IF(`status` IN (1,2),
      CONCAT('POSITION:', `position_id`), CONCAT('REQUEST:', `request_no`))
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0
        AND `order_id` = 0 AND `submitted_by` = 0 AND `submitted_at` = 0)
      OR (`status` IN (2,3) AND `reviewed_by` > 0 AND `reviewed_by` <> `requested_by`
        AND `review_reason` <> '' AND `reviewed_at` > 0
        AND `order_id` = 0 AND `submitted_by` = 0 AND `submitted_at` = 0)
      OR (`status` = 4 AND `reviewed_by` > 0 AND `reviewed_by` <> `requested_by`
        AND `review_reason` <> '' AND `reviewed_at` > 0
        AND `order_id` > 0 AND `submitted_by` > 0 AND `submitted_at` > 0)
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='保险接管库存受控主动退出申请';

DROP TRIGGER IF EXISTS `trg_option_insurance_exit_insert_guard`;
DELIMITER $$
CREATE TRIGGER `trg_option_insurance_exit_insert_guard`
BEFORE INSERT ON `t_option_insurance_inventory_exit`
FOR EACH ROW
BEGIN
  IF NEW.status <> 1 OR NEW.active_key <> CONCAT('POSITION:', NEW.position_id)
    OR NEW.reviewed_by <> 0 OR NEW.review_reason <> ''
    OR NEW.reviewed_at <> 0 OR NEW.order_id <> 0 OR NEW.submitted_by <> 0
    OR NEW.submitted_at <> 0 OR NEW.last_error_msg <> ''
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'insurance inventory exit must start as an unreviewed request';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS `trg_option_insurance_exit_guard`;
DELIMITER $$
CREATE TRIGGER `trg_option_insurance_exit_guard`
BEFORE UPDATE ON `t_option_insurance_inventory_exit`
FOR EACH ROW
BEGIN
  DECLARE duplicate_order BIGINT DEFAULT 0;
  IF NEW.tenant_id <> OLD.tenant_id
    OR NEW.request_no <> OLD.request_no
    OR NEW.position_id <> OLD.position_id
    OR NEW.contract_id <> OLD.contract_id
    OR NEW.insurance_user_id <> OLD.insurance_user_id
    OR NEW.insurance_account_id <> OLD.insurance_account_id
    OR NEW.quantity <> OLD.quantity
    OR NEW.limit_price <> OLD.limit_price
    OR NEW.reason <> OLD.reason
    OR NEW.evidence_ref <> OLD.evidence_ref
    OR NEW.requested_by <> OLD.requested_by
    OR NEW.create_times <> OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'insurance inventory exit economic fields are immutable';
  END IF;

  IF NOT (
    NEW.status = OLD.status
    OR (OLD.status = 1 AND NEW.status IN (2,3))
    OR (OLD.status = 2 AND NEW.status = 4)
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid insurance inventory exit status transition';
  END IF;

  IF NEW.active_key <> IF(NEW.status IN (1,2),
    CONCAT('POSITION:', NEW.position_id), CONCAT('REQUEST:', NEW.request_no))
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid insurance inventory exit active key';
  END IF;

  IF OLD.status = 1 AND NEW.status = 1 AND (
    NEW.reviewed_by <> OLD.reviewed_by OR NEW.review_reason <> OLD.review_reason
    OR NEW.reviewed_at <> OLD.reviewed_at OR NEW.order_id <> OLD.order_id
    OR NEW.submitted_by <> OLD.submitted_by OR NEW.submitted_at <> OLD.submitted_at
    OR NEW.last_error_msg <> OLD.last_error_msg
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'pending insurance inventory exit cannot be mutated';
  END IF;

  IF OLD.status = 1 AND NEW.status IN (2,3) AND (
    NEW.reviewed_by <= 0 OR NEW.reviewed_by = NEW.requested_by
    OR NEW.review_reason = '' OR NEW.reviewed_at <= 0
    OR NEW.order_id <> 0 OR NEW.submitted_by <> 0 OR NEW.submitted_at <> 0
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'insurance inventory exit requires independent review';
  END IF;

  IF OLD.status IN (2,3,4) AND (
    NEW.reviewed_by <> OLD.reviewed_by OR NEW.review_reason <> OLD.review_reason
    OR NEW.reviewed_at <> OLD.reviewed_at
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'insurance inventory exit review is immutable';
  END IF;

  IF OLD.status = 2 AND NEW.status = 2 AND (
    NEW.order_id <> OLD.order_id OR NEW.submitted_by <> OLD.submitted_by
    OR NEW.submitted_at <> OLD.submitted_at
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'approved insurance inventory exit has no order yet';
  END IF;

  IF OLD.status = 2 AND NEW.status = 4 THEN
    IF NEW.order_id <= 0 OR NEW.submitted_by <= 0 OR NEW.submitted_at <= 0 THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'submitted insurance inventory exit requires order evidence';
    END IF;
    SELECT COUNT(*) INTO duplicate_order
    FROM t_option_insurance_inventory_exit other
    WHERE other.tenant_id = NEW.tenant_id AND other.order_id = NEW.order_id
      AND other.id <> OLD.id;
    IF duplicate_order <> 0 THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'insurance inventory exit order is already linked';
    END IF;
  END IF;

  IF OLD.status IN (3,4) AND NEW.last_error_msg <> OLD.last_error_msg THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'terminal insurance inventory exit is immutable';
  END IF;

  IF OLD.status IN (3,4) AND (
    NEW.order_id <> OLD.order_id OR NEW.submitted_by <> OLD.submitted_by
    OR NEW.submitted_at <> OLD.submitted_at
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'terminal insurance inventory exit order evidence is immutable';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS `trg_option_insurance_exit_no_delete`;
DELIMITER $$
CREATE TRIGGER `trg_option_insurance_exit_no_delete`
BEFORE DELETE ON `t_option_insurance_inventory_exit`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'insurance inventory exit history cannot be deleted';
END$$
DELIMITER ;
