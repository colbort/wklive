-- 期权交易日历和临时休市治理。只追加菜单与接口权限，不覆盖既有角色授权。

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 733, 600, 1, '期权交易日历与休市', 2, 'GET',
       '/option/trading-calendars', 'option:trading-calendar:list',
       'option/trading-calendar', 'Calendar', 733
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 733 OR perms = 'option:trading-calendar:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 734, 733, 1, '创建期权交易日历版本', 3, 'POST',
       '/option/trading-calendars', 'option:trading-calendar:create', '', '', 734
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 734 OR perms = 'option:trading-calendar:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 735, 733, 1, '复核期权交易日历版本', 3, 'POST',
       '/option/trading-calendars/review', 'option:trading-calendar:review', '', '', 735
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 735 OR perms = 'option:trading-calendar:review'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 736, 733, 1, '查询期权临时休市', 3, 'GET',
       '/option/trading-halts', 'option:trading-halt:list', '', '', 736
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 736 OR perms = 'option:trading-halt:list'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 737, 733, 1, '暂停期权合约交易', 3, 'POST',
       '/option/trading-halts', 'option:trading-halt:create', '', '', 737
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 737 OR perms = 'option:trading-halt:create'
);

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 738, 733, 1, '恢复期权合约交易', 3, 'POST',
       '/option/trading-halts/resume', 'option:trading-halt:resume', '', '', 738
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menu WHERE id = 738 OR perms = 'option:trading-halt:resume'
);
