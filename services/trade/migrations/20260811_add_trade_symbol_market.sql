ALTER TABLE `t_trade_symbol`
  ADD COLUMN `market` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '行情市场代码，例如BA、SH、HK、US' AFTER `category_type`;

-- These categories have one canonical iTick region in the current market
-- configuration. Stock and futures rows remain empty because their market
-- cannot be inferred safely from symbol alone.
UPDATE `t_trade_symbol`
SET `market` = CASE `category_type`
  WHEN 1 THEN 'GB'
  WHEN 2 THEN 'BA'
  WHEN 5 THEN 'GB'
  WHEN 6 THEN 'US'
  ELSE `market`
END
WHERE `market` = ''
  AND `category_type` IN (1, 2, 5, 6);

ALTER TABLE `t_trade_symbol`
  DROP INDEX `uk_tenant_symbol_product`,
  ADD UNIQUE KEY `uk_tenant_market_symbol_product` (
    `tenant_id`,
    `category_type`,
    `market`,
    `symbol`,
    `product_type`,
    `contract_type`,
    `contract_value_type`
  );
