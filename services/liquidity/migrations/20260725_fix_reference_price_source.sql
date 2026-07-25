-- 修复早期界面错误地把 authority 名称 price-engine 写入行情产品来源的问题。
-- authority 由 liquidity.rpc 的 PriceEngineAuthority 配置控制；
-- reference_price_source 必须是 category:market:symbol。

UPDATE `t_liquidity_symbol_config`
SET
  `reference_price_source` = CONCAT(
    'crypto:BA:',
    CASE
      WHEN TRIM(`external_symbol`) <> '' THEN TRIM(`external_symbol`)
      ELSE TRIM(`symbol`)
    END
  ),
  `update_times` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  `version` = `version` + 1
WHERE TRIM(`reference_price_source`) IN ('', 'price-engine');

-- 应返回空集。存在结果时需按实际行情市场人工配置，例如 crypto:BA:BTCUSDT。
SELECT
  `id`,
  `symbol_id`,
  `symbol`,
  `external_symbol`,
  `reference_price_source`
FROM `t_liquidity_symbol_config`
WHERE `reference_price_source` NOT REGEXP '^[^:]+:[^:]+:[^:]+$';
