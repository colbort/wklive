-- Insurance-fund flows store positive business magnitudes. Direction is
-- derived exclusively from flow_type: 1/3 are inflows and 2/4 are outflows.
-- Existing signed rows are intentionally not rewritten; readers use ABS so
-- historical evidence remains immutable during rollout.

ALTER TABLE `t_option_insurance_fund_flow`
  MODIFY COLUMN `amount` DECIMAL(32,16) NOT NULL DEFAULT 0
  COMMENT '业务绝对金额，方向由flow_type确定';

DROP TRIGGER IF EXISTS `trg_option_insurance_fund_flow_validate_insert`;
DROP TRIGGER IF EXISTS `trg_option_insurance_fund_flow_no_update`;
DROP TRIGGER IF EXISTS `trg_option_insurance_fund_flow_no_delete`;

DELIMITER $$
CREATE TRIGGER `trg_option_insurance_fund_flow_validate_insert`
BEFORE INSERT ON `t_option_insurance_fund_flow`
FOR EACH ROW
BEGIN
  IF NEW.`tenant_id` <= 0 OR NEW.`flow_no` = '' OR NEW.`flow_type` NOT IN (1,2,3,4)
    OR NEW.`coin` = '' OR NEW.`amount` <= 0 OR NEW.`asset_flow_no` = '' OR NEW.`create_times` <= 0
    OR (NEW.`flow_type` IN (1,2) AND (NEW.`contract_id` <= 0 OR NEW.`liquidation_id` <= 0)) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='invalid option insurance fund flow evidence';
  END IF;
END$$
DELIMITER ;

CREATE TRIGGER `trg_option_insurance_fund_flow_no_update`
BEFORE UPDATE ON `t_option_insurance_fund_flow`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option insurance fund flow is immutable';

CREATE TRIGGER `trg_option_insurance_fund_flow_no_delete`
BEFORE DELETE ON `t_option_insurance_fund_flow`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option insurance fund flow cannot be deleted';
