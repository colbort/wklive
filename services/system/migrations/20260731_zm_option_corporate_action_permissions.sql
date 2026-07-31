-- P2-004 公司行动后台权限与执行任务。只追加，不覆盖已有角色授权。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 739, 600, 1, '期权公司行动', 2, 'GET',
       '/option/corporate-actions', 'option:corporate-action:list',
       'option/corporate-action', 'Switch', 739
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=739 OR perms='option:corporate-action:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 740, 739, 1, '登记期权公司行动', 3, 'POST',
       '/option/corporate-actions', 'option:corporate-action:create', '', '', 740
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=740 OR perms='option:corporate-action:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 741, 739, 1, '复核期权公司行动', 3, 'POST',
       '/option/corporate-actions/review', 'option:corporate-action:review', '', '', 741
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=741 OR perms='option:corporate-action:review'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 742, 739, 1, '查询公司行动持仓迁移', 3, 'GET',
       '/option/corporate-actions/positions', 'option:corporate-action:position:list', '', '', 742
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=742 OR perms='option:corporate-action:position:list'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '期权公司行动迁移', 'OPTION', 'option.ProcessCorporateActions', '*/1 * * * * *', 1,
  '每秒按100个持仓批次执行已复核且到达生效时间的公司行动；失败三次转人工',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target='option.ProcessCorporateActions'
);
