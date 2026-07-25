-- 当前 goctl SQL 解析器不支持函数表达式索引。内部交易用户唯一绑定
-- 由 liquidity Provider model/创建逻辑保证，这里添加兼容的查询索引。

CREATE INDEX `idx_type_trade_user`
ON `t_liquidity_provider` (`provider_type`, `trade_user_id`);

-- 执行迁移前应返回空集；若存在结果，需先人工合并重复的内部提供方。
SELECT
  `trade_user_id`,
  COUNT(*) AS `provider_count`
FROM `t_liquidity_provider`
WHERE `provider_type` = 1
GROUP BY `trade_user_id`
HAVING COUNT(*) > 1;
