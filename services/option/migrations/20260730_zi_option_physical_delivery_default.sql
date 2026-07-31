-- 实物交割配对单元、补资期限和独立步骤屏障。
-- 可重复执行；迁移后实物合约仍须完成业务规则批准与预生产验收才能开放。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
         WHERE table_schema=@schema_name AND table_name='t_option_contract'
           AND column_name='physical_delivery_cure_seconds'),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN physical_delivery_cure_seconds BIGINT NOT NULL DEFAULT 0
       COMMENT ''实物交割补资期限秒数，现金合约为0''
       AFTER physical_delivery_policy'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
         WHERE table_schema=@schema_name AND table_name='t_option_asset_instruction'
           AND column_name='delivery_unit_id'),
  'SELECT 1',
  'ALTER TABLE t_option_asset_instruction
     ADD COLUMN delivery_unit_id BIGINT NOT NULL DEFAULT 0
       COMMENT ''实物交割配对单元ID'' AFTER liquidation_id'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
         WHERE table_schema=@schema_name AND table_name='t_option_asset_instruction'
           AND column_name='execution_group'),
  'SELECT 1',
  'ALTER TABLE t_option_asset_instruction
     ADD COLUMN execution_group VARCHAR(96) NOT NULL DEFAULT ''''
       COMMENT ''步骤屏障执行域；空值回退biz_no'' AFTER delivery_unit_id'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.statistics
         WHERE table_schema=@schema_name AND table_name='t_option_asset_instruction'
           AND index_name='idx_instruction_delivery_unit'),
  'SELECT 1',
  'ALTER TABLE t_option_asset_instruction
     ADD KEY idx_instruction_delivery_unit
       (tenant_id,delivery_unit_id,step_no,status,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS `t_option_physical_delivery_unit` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `delivery_unit_no` VARCHAR(96) NOT NULL DEFAULT '',
  `batch_id` BIGINT NOT NULL DEFAULT 0,
  `batch_no` VARCHAR(96) NOT NULL DEFAULT '',
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `long_position_id` BIGINT NOT NULL DEFAULT 0,
  `long_user_id` BIGINT NOT NULL DEFAULT 0,
  `long_account_id` BIGINT NOT NULL DEFAULT 0,
  `short_position_id` BIGINT NOT NULL DEFAULT 0,
  `short_user_id` BIGINT NOT NULL DEFAULT 0,
  `short_account_id` BIGINT NOT NULL DEFAULT 0,
  `quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `delivery_coin` VARCHAR(16) NOT NULL DEFAULT '',
  `delivery_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `payment_coin` VARCHAR(16) NOT NULL DEFAULT '',
  `payment_amount` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `cure_deadline` BIGINT NOT NULL DEFAULT 0,
  `failed_instruction_id` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `manual_retry_count` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_physical_delivery_unit_no` (`tenant_id`,`delivery_unit_no`),
  KEY `idx_physical_delivery_batch` (`tenant_id`,`batch_id`,`id`),
  KEY `idx_physical_delivery_status` (`tenant_id`,`status`,`cure_deadline`,`id`),
  KEY `idx_physical_delivery_long` (`tenant_id`,`long_user_id`,`long_position_id`,`id`),
  KEY `idx_physical_delivery_short` (`tenant_id`,`short_user_id`,`short_position_id`,`id`),
  CONSTRAINT `chk_option_physical_delivery_unit` CHECK (
    `long_position_id` > 0 AND `short_position_id` > 0
    AND `quantity` > 0 AND `delivery_quantity` > 0 AND `payment_amount` > 0
    AND `delivery_coin` <> '' AND `payment_coin` <> ''
    AND `delivery_coin` <> `payment_coin`
    AND `status` IN (1,2,3,4,5,6) AND `cure_deadline` > 0
    AND `manual_retry_count` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='实物交割多空配对执行单元';

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
         WHERE table_schema=@schema_name AND table_name='t_option_physical_delivery_unit'
           AND column_name='manual_retry_count'),
  'SELECT 1',
  'ALTER TABLE t_option_physical_delivery_unit
     ADD COLUMN manual_retry_count BIGINT NOT NULL DEFAULT 0
       COMMENT ''逾期后人工重试代次'' AFTER completed_at'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

DROP TRIGGER IF EXISTS `trg_option_physical_unit_immutable`;
DELIMITER $$
CREATE TRIGGER `trg_option_physical_unit_immutable`
BEFORE UPDATE ON `t_option_physical_delivery_unit`
FOR EACH ROW
BEGIN
  IF NEW.tenant_id <> OLD.tenant_id
     OR NEW.delivery_unit_no <> OLD.delivery_unit_no
     OR NEW.batch_id <> OLD.batch_id
     OR NEW.batch_no <> OLD.batch_no
     OR NEW.contract_id <> OLD.contract_id
     OR NEW.long_position_id <> OLD.long_position_id
     OR NEW.long_user_id <> OLD.long_user_id
     OR NEW.long_account_id <> OLD.long_account_id
     OR NEW.short_position_id <> OLD.short_position_id
     OR NEW.short_user_id <> OLD.short_user_id
     OR NEW.short_account_id <> OLD.short_account_id
     OR NEW.quantity <> OLD.quantity
     OR NEW.delivery_coin <> OLD.delivery_coin
     OR NEW.delivery_quantity <> OLD.delivery_quantity
     OR NEW.payment_coin <> OLD.payment_coin
     OR NEW.payment_amount <> OLD.payment_amount
     OR NEW.cure_deadline <> OLD.cure_deadline
     OR NEW.create_times <> OLD.create_times THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'physical delivery economic fields are immutable';
  END IF;
  IF NOT (
    NEW.status = OLD.status
    OR (OLD.status = 1 AND NEW.status = 2)
    OR (OLD.status IN (1,2) AND NEW.status = 3)
    OR (OLD.status IN (2,3) AND NEW.status IN (4,6))
    OR (OLD.status IN (3,4,6) AND NEW.status = 2)
    OR (OLD.status IN (2,3,4,6) AND NEW.status = 5)
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid physical delivery status transition';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS `trg_option_physical_unit_no_delete`;
DELIMITER $$
CREATE TRIGGER `trg_option_physical_unit_no_delete`
BEFORE DELETE ON `t_option_physical_delivery_unit`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'physical delivery history cannot be deleted';
END$$
DELIMITER ;
