CREATE UNIQUE INDEX `uk_internal_trade_user`
ON `t_liquidity_provider` (
  (CASE WHEN `provider_type` = 1 THEN `trade_user_id` ELSE NULL END)
);
