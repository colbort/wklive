-- 为历史流动性配置补齐一个默认启用的一级报价档位。
-- 新配置由 CreateSymbolConfig 自动创建默认档位。

INSERT INTO `t_liquidity_strategy_level` (
  `config_id`,
  `level_no`,
  `bid_spread_bps`,
  `ask_spread_bps`,
  `bid_qty`,
  `ask_qty`,
  `enabled`,
  `version`,
  `create_times`,
  `update_times`
)
SELECT
  config.`id`,
  1,
  config.`base_spread_bps`,
  config.`base_spread_bps`,
  config.`min_quote_qty`,
  config.`min_quote_qty`,
  1,
  1,
  config.`create_times`,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
FROM `t_liquidity_symbol_config` AS config
WHERE config.`min_quote_qty` > 0
  AND NOT EXISTS (
    SELECT 1
    FROM `t_liquidity_strategy_level` AS level
    WHERE level.`config_id` = config.`id`
  );

-- 应返回空集；仍有结果表示主配置的 min_quote_qty 非法，需要人工修正后补档位。
SELECT
  config.`id`,
  config.`symbol_id`,
  config.`symbol`,
  config.`base_spread_bps`,
  config.`min_quote_qty`
FROM `t_liquidity_symbol_config` AS config
WHERE NOT EXISTS (
  SELECT 1
  FROM `t_liquidity_strategy_level` AS level
  WHERE level.`config_id` = config.`id`
);
