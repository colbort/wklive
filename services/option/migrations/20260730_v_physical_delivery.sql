-- 实物交割所需合约、订单、担保批次与结算审计字段。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_contract' AND column_name='underlying_coin'),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN underlying_coin VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''实物交割标的币种'' AFTER underlying_symbol,
     ADD COLUMN physical_delivery_policy TINYINT NOT NULL DEFAULT 0 COMMENT ''实物交割策略：0不适用 1严格全额交收'' AFTER liquidation_deficit_policy'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_order' AND column_name='margin_coin'),
  'SELECT 1',
  'ALTER TABLE t_option_order ADD COLUMN margin_coin VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''订单冻结资产币种'' AFTER margin_amount'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_margin_lot' AND column_name='collateral_coin'),
  'SELECT 1',
  'ALTER TABLE t_option_margin_lot ADD COLUMN collateral_coin VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''实际冻结的担保资产币种'' AFTER freeze_biz_no'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_settlement_detail' AND column_name='delivery_coin'),
  'SELECT 1',
  'ALTER TABLE t_option_settlement_detail
     ADD COLUMN delivery_coin VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''实物交割标的币种'' AFTER instruction_no,
     ADD COLUMN delivery_quantity DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''实物交割标的数量'' AFTER delivery_coin,
     ADD COLUMN payment_coin VARCHAR(16) NOT NULL DEFAULT '''' COMMENT ''行权款币种'' AFTER delivery_quantity,
     ADD COLUMN payment_amount DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''行权款金额'' AFTER payment_coin'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
