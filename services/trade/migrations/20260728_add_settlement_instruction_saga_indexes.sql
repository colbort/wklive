-- 覆盖 Funding/Delivery/Liquidation 指令扫描和批次级步骤屏障。
ALTER TABLE `t_trade_settlement_instruction`
  ADD INDEX `idx_settlement_biz_pending`
    (`tenant_id`, `biz_type`, `status`, `next_retry_at`, `id`),
  ADD INDEX `idx_settlement_batch_step`
    (`tenant_id`, `biz_type`, `batch_no`, `step_no`, `status`, `id`);
