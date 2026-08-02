-- 将交易日历版本和临时休市记录拆分为独立页面。

SET NAMES utf8mb4;

UPDATE sys_menu
SET parent_id = 600,
    name = '期权交易时段管理',
    menu_type = 1,
    method = '',
    path = '',
    perms = '',
    component = '',
    icon = 'Calendar',
    sort = 733,
    visible = 1,
    enabled = 1
WHERE id = 733;

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 762, 733, 1, '交易日历', 2, 'GET',
       '/option/trading-calendars', 'option:trading-calendar:list',
       'option/trading-calendars', 'Calendar', 10
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 762);

UPDATE sys_menu
SET parent_id = 733,
    name = '交易日历',
    menu_type = 2,
    method = 'GET',
    path = '/option/trading-calendars',
    perms = 'option:trading-calendar:list',
    component = 'option/trading-calendars',
    icon = 'Calendar',
    sort = 10,
    visible = 1,
    enabled = 1
WHERE id = 762;

UPDATE sys_menu
SET parent_id = 733,
    name = '临时休市',
    menu_type = 2,
    method = 'GET',
    path = '/option/trading-halts',
    perms = 'option:trading-halt:list',
    component = 'option/trading-halts',
    icon = 'SwitchButton',
    sort = 20,
    visible = 1,
    enabled = 1
WHERE id = 736;

UPDATE sys_menu SET parent_id = 762 WHERE id IN (734, 735);
UPDATE sys_menu SET parent_id = 736 WHERE id IN (737, 738);

-- 原页面权限迁移到新的交易日历页面。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 762
FROM sys_role_menu
WHERE menu_id = 733;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 762
FROM sys_role_menu
WHERE menu_id IN (734, 735);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 736
FROM sys_role_menu
WHERE menu_id IN (737, 738);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 733
FROM sys_role_menu
WHERE menu_id IN (734, 735, 736, 737, 738, 762);
