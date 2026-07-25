-- 修复历史衍生品交易对缺失的合约价值类型。
--
-- 规则：
--   线性合约：保证金和结算资产均为计价资产（例如 BTCUSDT 使用 USDT）。
--   反向合约：保证金和结算资产均为基础资产（例如 BTCUSD 使用 BTC）。
--
-- 无法由资产关系唯一判断的记录不会被修改，需人工确认后再处理。

UPDATE `t_trade_symbol`
SET
  `contract_value_type` = CASE
    WHEN `margin_asset` = `quote_asset`
      AND `settle_asset` = `quote_asset` THEN 1
    WHEN `margin_asset` = `base_asset`
      AND `settle_asset` = `base_asset` THEN 2
    ELSE `contract_value_type`
  END,
  `update_times` = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE `product_type` = 2
  AND `contract_type` IN (1, 2)
  AND `contract_value_type` = 0;

-- 此查询应返回空集；若仍有数据，必须人工确定线性/反向，不能按 symbol 名称猜测。
SELECT
  `id`,
  `tenant_id`,
  `symbol`,
  `display_symbol`,
  `contract_type`,
  `base_asset`,
  `quote_asset`,
  `margin_asset`,
  `settle_asset`,
  `contract_value_type`
FROM `t_trade_symbol`
WHERE `product_type` = 2
  AND `contract_type` IN (1, 2)
  AND `contract_value_type` NOT IN (1, 2)
ORDER BY `tenant_id`, `id`;
