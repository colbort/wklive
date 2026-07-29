-- dbinit:baseline-safe
-- 旧版 Asset 的部分冻结扣减曾把 remain_amount>0 的记录提前置为 status=4，
-- 后续现货手续费只能从可用余额扣除；Trade 预留账本仍把手续费记为冻结消费，
-- 导致终态订单遗留真实 frozen_amount，释放指令还可能被误标成功。
--
-- 仅恢复同时满足以下事实的历史现货订单：
-- 1. 订单已经终态；
-- 2. Asset 冻结为旧版异常形态 status=4 且 remain_amount>0；
-- 3. 对应手续费指令成功，且 Asset 流水确认手续费实际从可用余额扣除；
-- 4. Trade 预留与 Asset 冻结身份、总额一致。
--
-- 本迁移不直接修改用户余额或伪造 Asset 流水，只把 Trade 账本恢复为 Asset
-- 已确认事实并生成待释放指令；资金变化仍由 Asset RPC 的原子事务完成。

UPDATE `t_trade_asset_reservation` AS r
JOIN `t_trade_order` AS o
  ON o.tenant_id = r.tenant_id
 AND o.id = r.order_id
JOIN `t_asset_freeze` AS f
  ON f.tenant_id = r.tenant_id
 AND f.user_id = o.user_id
 AND f.biz_type = 'trade'
 AND f.biz_no = r.reservation_no
 AND f.amount = r.reserved_amount
SET r.consumed_amount = f.used_amount,
    r.released_amount = f.unfreeze_amount,
    r.status = 4,
    r.retry_count = 0,
    r.next_retry_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    r.last_error_msg = 'recover legacy terminal freeze remainder',
    r.version = r.version + 1,
    r.update_times = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
WHERE o.product_type = 1
  AND o.status IN (3,4,5,6)
  AND f.status = 4
  AND f.remain_amount > 0
  AND EXISTS (
    SELECT 1
    FROM `t_trade_settlement_instruction` AS fee_instruction
    JOIN `t_asset_flow` AS fee_flow
      ON fee_flow.tenant_id = fee_instruction.tenant_id
     AND fee_flow.biz_no = fee_instruction.instruction_no
     AND fee_flow.op_type = 2
    WHERE fee_instruction.tenant_id = r.tenant_id
      AND fee_instruction.order_id = r.order_id
      AND fee_instruction.action = 4
      AND fee_instruction.status = 3
  );

INSERT INTO `t_trade_settlement_instruction` (
  `tenant_id`, `instruction_no`, `biz_type`, `biz_id`, `batch_no`,
  `fill_id`, `order_id`, `position_id`, `reservation_no`, `user_id`,
  `action`, `asset`, `amount`, `step_no`, `status`, `retry_count`,
  `next_retry_at`, `last_error_msg`, `asset_flow_no`, `reconciled_at`,
  `create_times`, `update_times`
)
SELECT
  r.tenant_id,
  CONCAT(r.reservation_no, '-RELEASE'),
  'order',
  r.reservation_no,
  '',
  0,
  r.order_id,
  0,
  r.reservation_no,
  o.user_id,
  2,
  r.asset,
  f.remain_amount,
  1,
  1,
  0,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  'recover legacy terminal freeze remainder',
  '',
  0,
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
  CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)
FROM `t_trade_asset_reservation` AS r
JOIN `t_trade_order` AS o
  ON o.tenant_id = r.tenant_id
 AND o.id = r.order_id
JOIN `t_asset_freeze` AS f
  ON f.tenant_id = r.tenant_id
 AND f.user_id = o.user_id
 AND f.biz_type = 'trade'
 AND f.biz_no = r.reservation_no
 AND f.amount = r.reserved_amount
WHERE o.product_type = 1
  AND o.status IN (3,4,5,6)
  AND f.status = 4
  AND f.remain_amount > 0
  AND EXISTS (
    SELECT 1
    FROM `t_trade_settlement_instruction` AS fee_instruction
    JOIN `t_asset_flow` AS fee_flow
      ON fee_flow.tenant_id = fee_instruction.tenant_id
     AND fee_flow.biz_no = fee_instruction.instruction_no
     AND fee_flow.op_type = 2
    WHERE fee_instruction.tenant_id = r.tenant_id
      AND fee_instruction.order_id = r.order_id
      AND fee_instruction.action = 4
      AND fee_instruction.status = 3
  )
ON DUPLICATE KEY UPDATE
  `amount` = VALUES(`amount`),
  `status` = 1,
  `retry_count` = 0,
  `next_retry_at` = VALUES(`next_retry_at`),
  `last_error_msg` = VALUES(`last_error_msg`),
  `asset_flow_no` = '',
  `reconciled_at` = 0,
  `update_times` = VALUES(`update_times`);
