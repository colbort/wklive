-- 永续、交割合约的撮合、仓位、结算和事件恢复任务。
-- 使用 invoke_target 判断，允许迁移在已有环境安全重放。

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '交易撮合恢复', 'TRADE', 'trade.ProcessOrderMatching', '*/5 * * * * *', 1,
  '每5秒恢复尚未完成的订单撮合事件', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessOrderMatching'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '合约仓位与风险处理', 'TRADE', 'trade.ProcessPositions', '*/1 * * * * *', 1,
  '每秒恢复合约仓位投影并执行风险扫描', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessPositions'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '合约资金费与交割结算', 'TRADE', 'trade.ProcessContractSettlements', '*/1 * * * * *', 1,
  '每秒处理永续资金费、交割生命周期及未完成结算', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessContractSettlements'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '交易事件恢复', 'TRADE', 'trade.ProcessTradeEvents', '*/1 * * * * *', 1,
  '每秒恢复未投递或未完成的交易领域事件', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessTradeEvents'
);
