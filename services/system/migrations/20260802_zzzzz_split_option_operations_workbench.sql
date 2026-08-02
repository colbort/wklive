-- 将原期权运营工作台拆分为四个职责单一的页面。

SET NAMES utf8mb4;

UPDATE sys_menu
SET parent_id = 600,
    name = '期权运营中心',
    menu_type = 1,
    method = '',
    path = '',
    perms = '',
    component = '',
    icon = 'Monitor',
    sort = 729,
    visible = 1,
    enabled = 1
WHERE id = 729;

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 761, 729, 1, '运营概览', 2, 'GET',
       '/option/operations/overview', 'option:operations:view',
       'option/operations-overview', 'DataBoard', 10
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 761);

UPDATE sys_menu
SET parent_id = 729,
    name = '运营概览',
    menu_type = 2,
    method = 'GET',
    path = '/option/operations/overview',
    perms = 'option:operations:view',
    component = 'option/operations-overview',
    icon = 'DataBoard',
    sort = 10,
    visible = 1,
    enabled = 1
WHERE id = 761;

UPDATE sys_menu
SET parent_id = 729,
    name = '组合父单',
    menu_type = 2,
    method = 'GET',
    path = '/option/combo-orders',
    perms = 'option:operations:combo-list',
    component = 'option/combo-orders',
    icon = 'List',
    sort = 20,
    visible = 1,
    enabled = 1
WHERE id = 748;

UPDATE sys_menu
SET parent_id = 729,
    name = '资产指令',
    menu_type = 2,
    method = 'GET',
    path = '/option/operations/asset-instructions',
    perms = 'option:operations:asset-list',
    component = 'option/asset-instructions',
    icon = 'Tickets',
    sort = 30,
    visible = 1,
    enabled = 1
WHERE id = 730;

UPDATE sys_menu
SET parent_id = 729,
    name = '对账差异',
    menu_type = 2,
    method = 'GET',
    path = '/option/operations/reconciliation-issues',
    perms = 'option:operations:reconciliation-list',
    component = 'option/reconciliation-issues',
    icon = 'Warning',
    sort = 40,
    visible = 1,
    enabled = 1
WHERE id = 731;

UPDATE sys_menu SET parent_id = 730 WHERE id = 732;
UPDATE sys_menu SET parent_id = 748 WHERE id IN (749, 750);

-- 原工作台页面权限等价迁移到新的概览页面。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT tenant_id, role_id, 761
FROM sys_role_menu
WHERE menu_id = 729;

-- 若角色以前只分配了动作权限，补齐对应页面和父目录权限。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 748
FROM sys_role_menu
WHERE menu_id IN (749, 750);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 730
FROM sys_role_menu
WHERE menu_id = 732;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 729
FROM sys_role_menu
WHERE menu_id IN (730, 731, 748, 749, 750, 761, 732);
