-- 期权实物交割与运营工作台权限；只追加，不覆盖既有授权。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 727, 680, 1, '查询实物交割单元', 3, 'GET',
       '/option/physical-delivery/units', 'option:physical-delivery:list', '', '', 727
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 727 OR perms = 'option:physical-delivery:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 728, 680, 1, '重试实物交割单元', 3, 'POST',
       '/option/physical-delivery/units/retry', 'option:physical-delivery:retry', '', '', 728
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 728 OR perms = 'option:physical-delivery:retry'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 729, 600, 1, '期权运营工作台', 2, 'GET',
       '/option/operations/overview', 'option:operations:view', 'option/operations', 'Monitor', 729
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 729 OR perms = 'option:operations:view'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 730, 729, 1, '查询资产指令', 3, 'GET',
       '/option/operations/asset-instructions', 'option:operations:asset-list', '', '', 730
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 730 OR perms = 'option:operations:asset-list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 731, 729, 1, '查询期权对账差异', 3, 'GET',
       '/option/operations/reconciliation-issues', 'option:operations:reconciliation-list', '', '', 731
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 731 OR perms = 'option:operations:reconciliation-list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 732, 729, 1, '重试异常资产指令', 3, 'POST',
       '/option/recovery/asset-instructions/retry', 'option:operations:asset-retry', '', '', 732
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 732 OR perms = 'option:operations:asset-retry'
);
