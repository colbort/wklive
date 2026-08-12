-- 股票交易对必须来自租户产品目录，避免手工输入 market/symbol/settle_asset
-- 与 market-rpc 实际行情源不一致。
ALTER TABLE `t_trade_symbol`
  ADD COLUMN `tenant_product_id` BIGINT NOT NULL DEFAULT 0
    COMMENT '租户产品ID；股票交易对必须关联 t_itick_tenant_product.id'
    AFTER `tenant_id`,
  ADD KEY `idx_tenant_product_id` (`tenant_product_id`);

-- 历史手工股票交易对没有可靠的产品来源。先禁用，避免继续按错误市场交易；
-- 匹配到有效租户产品的记录会在下面重新绑定并启用。
UPDATE `t_trade_symbol`
SET `status` = 2,
    `update_times` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE `category_type` = 3
  AND `tenant_product_id` = 0;

-- 将当前启用且 APP 可见的租户股票产品创建为现货交易对。
INSERT INTO `t_trade_symbol` (
  `tenant_id`, `tenant_product_id`, `category_type`, `market`, `symbol`,
  `display_symbol`, `product_type`, `base_asset`, `quote_asset`, `settle_asset`,
  `margin_asset`, `contract_type`, `contract_value_type`, `status`,
  `price_scale`, `qty_scale`, `min_price`, `max_price`, `price_tick`,
  `min_qty`, `max_qty`, `qty_step`, `min_notional`, `max_notional`,
  `listing_time`, `trading_start_time`, `trading_end_time`, `sort`, `remark`,
  `create_times`, `update_times`
)
SELECT
  tp.`tenant_id`, tp.`id`, 3, UPPER(TRIM(p.`market`)), UPPER(TRIM(p.`symbol`)),
  UPPER(TRIM(p.`symbol`)), 1,
  COALESCE(NULLIF(UPPER(TRIM(p.`base_coin`)), ''), UPPER(TRIM(p.`symbol`))),
  COALESCE(NULLIF(UPPER(TRIM(p.`quote_coin`)), ''), CASE UPPER(TRIM(p.`market`))
    WHEN 'HK' THEN 'HKD' WHEN 'SZ' THEN 'CNY' WHEN 'SH' THEN 'CNY'
    WHEN 'US' THEN 'USD' WHEN 'SG' THEN 'SGD' WHEN 'JP' THEN 'JPY'
    WHEN 'TW' THEN 'TWD' WHEN 'IN' THEN 'INR' WHEN 'TH' THEN 'THB'
    WHEN 'DE' THEN 'EUR' WHEN 'MX' THEN 'MXN' WHEN 'MY' THEN 'MYR'
    WHEN 'TR' THEN 'TRY' WHEN 'ES' THEN 'EUR' WHEN 'NL' THEN 'EUR'
    WHEN 'GB' THEN 'GBP' WHEN 'ID' THEN 'IDR' WHEN 'VN' THEN 'VND'
    ELSE '' END),
  COALESCE(NULLIF(UPPER(TRIM(p.`quote_coin`)), ''), CASE UPPER(TRIM(p.`market`))
    WHEN 'HK' THEN 'HKD' WHEN 'SZ' THEN 'CNY' WHEN 'SH' THEN 'CNY'
    WHEN 'US' THEN 'USD' WHEN 'SG' THEN 'SGD' WHEN 'JP' THEN 'JPY'
    WHEN 'TW' THEN 'TWD' WHEN 'IN' THEN 'INR' WHEN 'TH' THEN 'THB'
    WHEN 'DE' THEN 'EUR' WHEN 'MX' THEN 'MXN' WHEN 'MY' THEN 'MYR'
    WHEN 'TR' THEN 'TRY' WHEN 'ES' THEN 'EUR' WHEN 'NL' THEN 'EUR'
    WHEN 'GB' THEN 'GBP' WHEN 'ID' THEN 'IDR' WHEN 'VN' THEN 'VND'
    ELSE '' END),
  '', 0, 0, 1,
  2, 0, 0.01, 1000000, 0.01,
  1, 100000, 1, 10, 1000000,
  0, 0, 0, tp.`sort`, CONCAT('由租户产品 #', tp.`id`, ' 创建'),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
FROM `t_itick_tenant_product` tp
JOIN `t_itick_product` p ON p.`id` = tp.`product_id`
WHERE p.`category_type` = 3
  AND tp.`enabled` = 1 AND tp.`app_visible` = 1
  AND p.`enabled` = 1 AND p.`app_visible` = 1
ON DUPLICATE KEY UPDATE
  `tenant_product_id` = VALUES(`tenant_product_id`),
  `base_asset` = VALUES(`base_asset`),
  `quote_asset` = VALUES(`quote_asset`),
  `settle_asset` = VALUES(`settle_asset`),
  `status` = 1,
  `update_times` = VALUES(`update_times`);

-- 股票属于现货产品；为新绑定的交易对补齐现货买卖配置。
INSERT INTO `t_trade_symbol_spot` (
  `tenant_id`, `symbol_id`, `maker_fee_rate`, `taker_fee_rate`,
  `buy_enabled`, `sell_enabled`, `create_times`, `update_times`
)
SELECT
  s.`tenant_id`, s.`id`, 0.01, 0.01, 1, 1,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
FROM `t_trade_symbol` s
LEFT JOIN `t_trade_symbol_spot` ss
  ON ss.`tenant_id` = s.`tenant_id` AND ss.`symbol_id` = s.`id`
WHERE s.`category_type` = 3
  AND s.`tenant_product_id` > 0
  AND ss.`id` IS NULL;
