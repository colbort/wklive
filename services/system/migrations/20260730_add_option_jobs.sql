INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权资产指令处理', 'OPTION', 'option.ProcessAssetInstructions', '*/1 * * * * *', 1,
  '每秒执行和恢复期权资产指令', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessAssetInstructions'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权成交持仓事件处理', 'OPTION', 'option.ProcessTradeEvents', '*/1 * * * * *', 1,
  '每秒按合约撮合序号更新期权持仓', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessTradeEvents'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权卖方风险账户扫描', 'OPTION', 'option.ProcessRiskAccounts', '*/1 * * * * *', 1,
  '每秒聚合卖方权益、维持保证金与风险率',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessRiskAccounts'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权卖方强平处理', 'OPTION', 'option.ProcessLiquidations', '*/1 * * * * *', 1,
  '每秒执行卖方强平和保险账户接管',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessLiquidations'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权主动行权清算', 'OPTION', 'option.ProcessExercises', '*/1 * * * * *', 1,
  '每秒处理美式主动行权空头指派和资金清算',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessExercises'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权合约生命周期', 'OPTION', 'option.ProcessContractLifecycle', '*/1 * * * * *', 1,
  '每秒处理期权上市、到期、行权和结算', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.ProcessContractLifecycle'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权行情快照清理', 'OPTION', 'option.CleanMarketSnapshots', '0 10 3 * * *', 1,
  '每日清理超过保留期的期权行情快照', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'option.CleanMarketSnapshots'
);
