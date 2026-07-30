INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 710, 600, 1, '风险与强平', 2, 'GET', '/option/risk/accounts',
       'option:risk:list', 'option/risk', 'Warning', 710
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 710 OR perms = 'option:risk:list');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 711, 710, 1, '查询强平记录', 3, 'GET', '/option/risk/liquidations',
       'option:liquidation:list', '', '', 711
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 711 OR perms = 'option:liquidation:list');

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 712, 710, 1, '重试强平', 3, 'POST', '/option/risk/liquidations/retry',
       'option:liquidation:retry', '', '', 712
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 712 OR perms = 'option:liquidation:retry');
