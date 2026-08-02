-- 将原先堆叠在一个页面中的期权风险工作台拆分为独立菜单。

SET NAMES utf8mb4;

INSERT INTO sys_menu
(id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort)
SELECT 760, 600, 1, '期权风控中心', 1, '', '', '', '', 'Warning', 710
WHERE NOT EXISTS (SELECT 1 FROM sys_menu WHERE id = 760);

UPDATE sys_menu
SET parent_id = 600,
    name = '期权风控中心',
    menu_type = 1,
    method = '',
    path = '',
    perms = '',
    component = '',
    icon = 'Warning',
    sort = 710,
    visible = 1,
    enabled = 1
WHERE id = 760;

UPDATE sys_menu
SET parent_id = 760,
    name = '风险账户',
    menu_type = 2,
    component = 'option/risk',
    icon = 'Warning',
    sort = 710,
    visible = 1,
    enabled = 1
WHERE id = 710;

UPDATE sys_menu
SET parent_id = 760,
    name = '强平记录',
    menu_type = 2,
    component = 'option/liquidations',
    icon = 'Histogram',
    sort = 720,
    visible = 1,
    enabled = 1
WHERE id = 711;

UPDATE sys_menu
SET parent_id = 760,
    name = '交易控制',
    menu_type = 2,
    component = 'option/trading-control',
    icon = 'Switch',
    sort = 730,
    visible = 1,
    enabled = 1
WHERE id = 720;

UPDATE sys_menu
SET parent_id = 760,
    name = '做市商保护（MMP）',
    menu_type = 2,
    component = 'option/mmp',
    icon = 'Aim',
    sort = 740,
    visible = 1,
    enabled = 1
WHERE id = 723;

UPDATE sys_menu
SET parent_id = 760,
    name = '异常成交更正',
    menu_type = 2,
    component = 'option/trade-correction',
    icon = 'Document',
    sort = 750,
    visible = 1,
    enabled = 1
WHERE id = 722;

UPDATE sys_menu
SET parent_id = 760,
    name = '组合保证金配置',
    menu_type = 2,
    component = 'option/portfolio-risk',
    icon = 'Grid',
    sort = 760,
    visible = 1,
    enabled = 1
WHERE id = 726;

UPDATE sys_menu
SET parent_id = 760,
    name = '保险库存退出',
    menu_type = 2,
    component = 'option/insurance-inventory-exit',
    icon = 'Wallet',
    sort = 770,
    visible = 1,
    enabled = 1
WHERE id = 758;

UPDATE sys_menu SET parent_id = 720 WHERE id IN (715, 721);
UPDATE sys_menu SET parent_id = 723 WHERE id IN (718, 719);
UPDATE sys_menu SET parent_id = 711 WHERE id = 712;
UPDATE sys_menu SET parent_id = 722 WHERE id IN (716, 717);
UPDATE sys_menu SET parent_id = 726 WHERE id IN (724, 725);
UPDATE sys_menu SET parent_id = 758 WHERE id IN (755, 756, 757);

-- 保留现有授权边界：拥有任一拆分页面权限的角色自动获得父目录。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT DISTINCT tenant_id, role_id, 760
FROM sys_role_menu
WHERE menu_id IN (710, 711, 720, 722, 723, 726, 758);
