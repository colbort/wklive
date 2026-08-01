-- Backfill deterministic collateral-coin evidence and reject future guesses.
-- Existing successful/processing Asset instructions are never rewritten.

UPDATE `t_option_order` order_item
JOIN `t_option_contract` contract
  ON contract.tenant_id=order_item.tenant_id AND contract.id=order_item.contract_id
SET order_item.margin_coin=CASE
  WHEN order_item.side=2 AND order_item.position_effect=1
       AND contract.seller_margin_mode=4 AND contract.settlement_type=2
       AND contract.option_type=1
    THEN contract.underlying_coin
  ELSE contract.settle_coin
END
WHERE order_item.margin_coin=''
  AND (CASE
    WHEN order_item.side=2 AND order_item.position_effect=1
         AND contract.seller_margin_mode=4 AND contract.settlement_type=2
         AND contract.option_type=1
      THEN contract.underlying_coin
    ELSE contract.settle_coin
  END)<>'';

UPDATE `t_option_margin_lot` lot
JOIN `t_option_contract` contract
  ON contract.tenant_id=lot.tenant_id AND contract.id=lot.contract_id
SET lot.collateral_coin=CASE
  WHEN contract.seller_margin_mode=4 AND contract.settlement_type=2
       AND contract.option_type=1
    THEN contract.underlying_coin
  ELSE contract.settle_coin
END
WHERE lot.collateral_coin=''
  AND (CASE
    WHEN contract.seller_margin_mode=4 AND contract.settlement_type=2
         AND contract.option_type=1
      THEN contract.underlying_coin
    ELSE contract.settle_coin
  END)<>'';

-- Empty-coin instructions cannot have succeeded at Asset. Only pending, failed,
-- or manual-review rows are repaired; processing/success history is immutable.
UPDATE `t_option_asset_instruction` instruction
JOIN `t_option_order` order_item
  ON order_item.tenant_id=instruction.tenant_id AND order_item.id=instruction.order_id
SET instruction.coin=order_item.margin_coin
WHERE instruction.coin='' AND order_item.margin_coin<>''
  AND instruction.status IN (1,4,5);

UPDATE `t_option_asset_instruction` instruction
JOIN `t_option_margin_lot` lot
  ON lot.tenant_id=instruction.tenant_id AND lot.id=instruction.margin_lot_id
SET instruction.coin=lot.collateral_coin
WHERE instruction.coin='' AND lot.collateral_coin<>''
  AND instruction.status IN (1,4,5);

DROP TRIGGER IF EXISTS `trg_option_order_margin_coin_insert`;
DROP TRIGGER IF EXISTS `trg_option_order_margin_coin_update`;
DROP TRIGGER IF EXISTS `trg_option_margin_lot_coin_insert`;
DROP TRIGGER IF EXISTS `trg_option_margin_lot_coin_update`;
DROP TRIGGER IF EXISTS `trg_option_asset_instruction_coin_insert`;
DROP TRIGGER IF EXISTS `trg_option_asset_instruction_coin_update`;

DELIMITER $$

CREATE TRIGGER `trg_option_order_margin_coin_insert`
BEFORE INSERT ON `t_option_order`
FOR EACH ROW
BEGIN
  DECLARE expected_coin VARCHAR(16) DEFAULT '';
  SET NEW.margin_coin=TRIM(NEW.margin_coin);
  IF NEW.margin_amount>0 THEN
    SELECT CASE
      WHEN NEW.side=2 AND NEW.position_effect=1
           AND contract.seller_margin_mode=4 AND contract.settlement_type=2
           AND contract.option_type=1
        THEN contract.underlying_coin
      ELSE contract.settle_coin
    END INTO expected_coin
    FROM `t_option_contract` contract
    WHERE contract.tenant_id=NEW.tenant_id AND contract.id=NEW.contract_id
    LIMIT 1;
    IF OCTET_LENGTH(expected_coin)=0 OR BINARY NEW.margin_coin<>BINARY expected_coin THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='order margin coin does not match frozen collateral';
    END IF;
  END IF;
END$$

CREATE TRIGGER `trg_option_order_margin_coin_update`
BEFORE UPDATE ON `t_option_order`
FOR EACH ROW
BEGIN
  DECLARE expected_coin VARCHAR(16) DEFAULT '';
  SET NEW.margin_coin=TRIM(NEW.margin_coin);
  IF OLD.margin_coin<>'' AND NEW.margin_coin<>OLD.margin_coin THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='order margin coin evidence is immutable';
  END IF;
  IF (OLD.margin_coin='' AND NEW.margin_coin<>'')
     OR (OLD.margin_amount<=0 AND NEW.margin_amount>0)
     OR NEW.contract_id<>OLD.contract_id
     OR NEW.side<>OLD.side
     OR NEW.position_effect<>OLD.position_effect THEN
    IF NEW.margin_amount>0 THEN
      SELECT CASE
        WHEN NEW.side=2 AND NEW.position_effect=1
             AND contract.seller_margin_mode=4 AND contract.settlement_type=2
             AND contract.option_type=1
          THEN contract.underlying_coin
        ELSE contract.settle_coin
      END INTO expected_coin
      FROM `t_option_contract` contract
      WHERE contract.tenant_id=NEW.tenant_id AND contract.id=NEW.contract_id
      LIMIT 1;
      IF OCTET_LENGTH(expected_coin)=0 OR BINARY NEW.margin_coin<>BINARY expected_coin THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='order margin coin repair does not match frozen collateral';
      END IF;
    END IF;
  END IF;
END$$

CREATE TRIGGER `trg_option_margin_lot_coin_insert`
BEFORE INSERT ON `t_option_margin_lot`
FOR EACH ROW
BEGIN
  DECLARE expected_coin VARCHAR(16) DEFAULT '';
  SET NEW.collateral_coin=TRIM(NEW.collateral_coin);
  SELECT CASE
    WHEN contract.seller_margin_mode=4 AND contract.settlement_type=2
         AND contract.option_type=1
      THEN contract.underlying_coin
    ELSE contract.settle_coin
  END INTO expected_coin
  FROM `t_option_contract` contract
  WHERE contract.tenant_id=NEW.tenant_id AND contract.id=NEW.contract_id
  LIMIT 1;
  IF OCTET_LENGTH(expected_coin)=0 OR BINARY NEW.collateral_coin<>BINARY expected_coin THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='margin lot collateral coin does not match contract';
  END IF;
END$$

CREATE TRIGGER `trg_option_margin_lot_coin_update`
BEFORE UPDATE ON `t_option_margin_lot`
FOR EACH ROW
BEGIN
  DECLARE expected_coin VARCHAR(16) DEFAULT '';
  SET NEW.collateral_coin=TRIM(NEW.collateral_coin);
  IF OLD.collateral_coin<>'' AND NEW.collateral_coin<>OLD.collateral_coin THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='margin lot collateral coin evidence is immutable';
  END IF;
  IF (OLD.collateral_coin='' AND NEW.collateral_coin<>'')
     OR NEW.contract_id<>OLD.contract_id THEN
    SELECT CASE
      WHEN contract.seller_margin_mode=4 AND contract.settlement_type=2
           AND contract.option_type=1
        THEN contract.underlying_coin
      ELSE contract.settle_coin
    END INTO expected_coin
    FROM `t_option_contract` contract
    WHERE contract.tenant_id=NEW.tenant_id AND contract.id=NEW.contract_id
    LIMIT 1;
    IF OCTET_LENGTH(expected_coin)=0 OR BINARY NEW.collateral_coin<>BINARY expected_coin THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='margin lot collateral coin repair does not match contract';
    END IF;
  END IF;
END$$

CREATE TRIGGER `trg_option_asset_instruction_coin_insert`
BEFORE INSERT ON `t_option_asset_instruction`
FOR EACH ROW
BEGIN
  SET NEW.coin=TRIM(NEW.coin);
  IF TRIM(NEW.coin)='' THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='asset instruction coin is required';
  END IF;
END$$

CREATE TRIGGER `trg_option_asset_instruction_coin_update`
BEFORE UPDATE ON `t_option_asset_instruction`
FOR EACH ROW
BEGIN
  SET NEW.coin=TRIM(NEW.coin);
  IF OLD.coin<>'' AND NEW.coin<>OLD.coin THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='asset instruction coin evidence is immutable';
  END IF;
END$$

DELIMITER ;
