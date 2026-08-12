-- Repair stock products copied into the wrong iTick region by the historical
-- product-list sync. Product rows are retained for auditability and disabled;
-- tenant selections are moved to the canonical region before that happens.

DROP TEMPORARY TABLE IF EXISTS `tmp_itick_stock_product_repair`;

CREATE TEMPORARY TABLE `tmp_itick_stock_product_repair` (
  `wrong_product_id` BIGINT NOT NULL,
  `canonical_product_id` BIGINT NOT NULL,
  PRIMARY KEY (`wrong_product_id`),
  KEY `idx_canonical_product_id` (`canonical_product_id`)
) ENGINE=InnoDB;

INSERT INTO `tmp_itick_stock_product_repair` (`wrong_product_id`, `canonical_product_id`)
SELECT wrong_product.id, canonical_product.id
FROM `t_itick_product` AS wrong_product
JOIN `t_itick_product` AS canonical_product
  ON canonical_product.category_type = wrong_product.category_type
 AND canonical_product.symbol = wrong_product.symbol
 AND canonical_product.exchange = wrong_product.exchange
 AND canonical_product.market = CASE
   WHEN wrong_product.exchange = 'HKEX' THEN 'HK'
   WHEN wrong_product.exchange = 'SZSE' THEN 'SZ'
   WHEN wrong_product.exchange = 'SSE' THEN 'SH'
   WHEN wrong_product.exchange IN ('AMEX', 'CBOE', 'NASDAQ', 'NYSE', 'OTC') THEN 'US'
   WHEN wrong_product.exchange = 'SGX' THEN 'SG'
   WHEN wrong_product.exchange IN ('FSE', 'NAG', 'SAPSE', 'TSE') THEN 'JP'
   WHEN wrong_product.exchange IN ('TPEX', 'TWSE') THEN 'TW'
   WHEN wrong_product.exchange IN ('BSE', 'NSE') THEN 'IN'
   WHEN wrong_product.exchange = 'SET' THEN 'TH'
   WHEN wrong_product.exchange IN ('FWB', 'XETR') THEN 'DE'
   WHEN wrong_product.exchange IN ('BIVA', 'BMV') THEN 'MX'
   WHEN wrong_product.exchange = 'MYX' THEN 'MY'
   WHEN wrong_product.exchange = 'BIST' THEN 'TR'
   WHEN wrong_product.exchange = 'BME' THEN 'ES'
   WHEN wrong_product.exchange = 'EURONEXT' THEN 'NL'
   WHEN wrong_product.exchange IN ('LSE', 'LSIN') THEN 'GB'
   WHEN wrong_product.exchange = 'IDX' THEN 'ID'
   WHEN wrong_product.exchange IN ('HNX', 'HOSE', 'UPCOM') THEN 'VN'
   ELSE ''
 END
WHERE wrong_product.category_code = 'stock'
  AND wrong_product.market <> canonical_product.market;

INSERT INTO `t_itick_tenant_product` (
  `tenant_id`, `product_id`, `enabled`, `app_visible`, `display_name`, `sort`,
  `remark`, `create_times`, `update_times`
)
SELECT
  tenant_product.tenant_id,
  repair.canonical_product_id,
  tenant_product.enabled,
  tenant_product.app_visible,
  tenant_product.display_name,
  tenant_product.sort,
  tenant_product.remark,
  tenant_product.create_times,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
FROM `t_itick_tenant_product` AS tenant_product
JOIN `tmp_itick_stock_product_repair` AS repair
  ON repair.wrong_product_id = tenant_product.product_id
ON DUPLICATE KEY UPDATE
  `enabled` = LEAST(`t_itick_tenant_product`.`enabled`, VALUES(`enabled`)),
  `app_visible` = LEAST(`t_itick_tenant_product`.`app_visible`, VALUES(`app_visible`)),
  `display_name` = CASE
    WHEN `t_itick_tenant_product`.`display_name` = '' THEN VALUES(`display_name`)
    ELSE `t_itick_tenant_product`.`display_name`
  END,
  `sort` = LEAST(`t_itick_tenant_product`.`sort`, VALUES(`sort`)),
  `remark` = CASE
    WHEN `t_itick_tenant_product`.`remark` = '' THEN VALUES(`remark`)
    ELSE `t_itick_tenant_product`.`remark`
  END,
  `update_times` = VALUES(`update_times`);

DELETE tenant_product
FROM `t_itick_tenant_product` AS tenant_product
JOIN `tmp_itick_stock_product_repair` AS repair
  ON repair.wrong_product_id = tenant_product.product_id;

UPDATE `t_itick_product`
SET
  `enabled` = 2,
  `app_visible` = 2,
  `update_times` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE `category_code` = 'stock'
  AND `remark` LIKE '同步自 iTick%'
  AND NOT (
    (`market` = 'HK' AND `exchange` = 'HKEX') OR
    (`market` = 'SZ' AND `exchange` = 'SZSE') OR
    (`market` = 'SH' AND `exchange` = 'SSE') OR
    (`market` = 'US' AND `exchange` IN ('AMEX', 'CBOE', 'NASDAQ', 'NYSE', 'OTC')) OR
    (`market` = 'SG' AND `exchange` = 'SGX') OR
    (`market` = 'JP' AND `exchange` IN ('FSE', 'NAG', 'SAPSE', 'TSE')) OR
    (`market` = 'TW' AND `exchange` IN ('TPEX', 'TWSE')) OR
    (`market` = 'IN' AND `exchange` IN ('BSE', 'NSE')) OR
    (`market` = 'TH' AND `exchange` = 'SET') OR
    (`market` = 'DE' AND `exchange` IN ('FWB', 'XETR')) OR
    (`market` = 'MX' AND `exchange` IN ('BIVA', 'BMV')) OR
    (`market` = 'MY' AND `exchange` = 'MYX') OR
    (`market` = 'TR' AND `exchange` = 'BIST') OR
    (`market` = 'ES' AND `exchange` = 'BME') OR
    (`market` = 'NL' AND `exchange` = 'EURONEXT') OR
    (`market` = 'GB' AND `exchange` IN ('LSE', 'LSIN')) OR
    (`market` = 'ID' AND `exchange` = 'IDX') OR
    (`market` = 'VN' AND `exchange` IN ('HNX', 'HOSE', 'UPCOM'))
  );

DROP TEMPORARY TABLE `tmp_itick_stock_product_repair`;
