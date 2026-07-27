-- 支付系统金额统一使用自然货币单位，例如 1000 USDT 存储为 1000。
-- 历史 BIGINT 金额使用“分”，因此本迁移会且只能执行一次除以 100。

ALTER TABLE `t_tenant_pay_channel`
  MODIFY `single_min_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单笔最小金额，自然货币单位',
  MODIFY `single_max_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单笔最大金额，0表示不限制，自然货币单位',
  MODIFY `daily_max_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单日最大金额，0表示不限制，自然货币单位',
  MODIFY `fee_fixed_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '固定手续费，自然货币单位';
UPDATE `t_tenant_pay_channel`
SET `single_min_amount` = `single_min_amount` / 100,
    `single_max_amount` = `single_max_amount` / 100,
    `daily_max_amount` = `daily_max_amount` / 100,
    `fee_fixed_amount` = `fee_fixed_amount` / 100;

ALTER TABLE `t_tenant_pay_channel_rule`
  MODIFY `single_amount_min` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单笔充值最小金额，自然货币单位',
  MODIFY `single_amount_max` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '单笔充值最大金额，0表示不限制，自然货币单位',
  MODIFY `user_total_recharge_min` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '用户累计充值最小金额，自然货币单位',
  MODIFY `user_total_recharge_max` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '用户累计充值最大金额，0表示不限制，自然货币单位';
UPDATE `t_tenant_pay_channel_rule`
SET `single_amount_min` = `single_amount_min` / 100,
    `single_amount_max` = `single_amount_max` / 100,
    `user_total_recharge_min` = `user_total_recharge_min` / 100,
    `user_total_recharge_max` = `user_total_recharge_max` / 100;

ALTER TABLE `t_user_recharge_stat`
  MODIFY `success_total_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '成功累计充值金额，自然货币单位',
  MODIFY `today_success_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '今日成功充值金额，自然货币单位';
UPDATE `t_user_recharge_stat`
SET `success_total_amount` = `success_total_amount` / 100,
    `today_success_amount` = `today_success_amount` / 100;

ALTER TABLE `t_recharge_order`
  MODIFY `order_amount` DECIMAL(36,18) NOT NULL COMMENT '订单金额，自然货币单位',
  MODIFY `pay_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '实际支付金额，自然货币单位',
  MODIFY `fee_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '手续费金额，自然货币单位';
UPDATE `t_recharge_order`
SET `order_amount` = `order_amount` / 100,
    `pay_amount` = `pay_amount` / 100,
    `fee_amount` = `fee_amount` / 100;

ALTER TABLE `t_withdraw_order`
  MODIFY `amount` DECIMAL(36,18) NOT NULL COMMENT '订单金额，自然货币单位',
  MODIFY `fee_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '手续费金额，自然货币单位',
  MODIFY `actual_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '实际到账金额，自然货币单位';
UPDATE `t_withdraw_order`
SET `amount` = `amount` / 100,
    `fee_amount` = `fee_amount` / 100,
    `actual_amount` = `actual_amount` / 100;
