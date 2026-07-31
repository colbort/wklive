-- OPT-P2-005 合约系列治理后台权限。只追加，不覆盖已有角色授权。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 743, 600, 1, '期权合约系列', 2, 'GET',
       '/option/contract-series', 'option:contract-series:list',
       'option/contract-series', 'Grid', 743
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=743 OR perms='option:contract-series:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 744, 743, 1, '创建期权合约系列', 3, 'POST',
       '/option/contract-series', 'option:contract-series:create', '', '', 744
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=744 OR perms='option:contract-series:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 745, 743, 1, '复核并生成期权合约系列', 3, 'POST',
       '/option/contract-series/review', 'option:contract-series:review', '', '', 745
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=745 OR perms='option:contract-series:review'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 746, 743, 1, '查询期权系列生成谱系', 3, 'GET',
       '/option/contract-series/details', 'option:contract-series:detail:list', '', '', 746
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=746 OR perms='option:contract-series:detail:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 747, 743, 1, '复核期权系列上市', 3, 'POST',
       '/option/contract-series/launch-review', 'option:contract-series:launch-review', '', '', 747
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id=747 OR perms='option:contract-series:launch-review'
);
